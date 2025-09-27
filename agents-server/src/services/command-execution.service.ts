import { exec, ExecOptions, spawn, ChildProcessWithoutNullStreams } from 'child_process';
import { promisify } from 'util';
import { ExecResult } from '../models/artifact.model';
import { logger } from '../utils/logger.util';

const execAsync = promisify(exec);

export class CommandExecutionService {
  private toolPaths: {
    claudeCli: string;
    npm: string;
    git: string;
  };
  private sessions: Map<string, {
    shell: ChildProcessWithoutNullStreams;
    stdoutBuffer: string;
    stderrBuffer: string;
    queue: Promise<any>;
    killed: boolean;
  }> = new Map();

  constructor(toolPaths: { claudeCli: string; npm: string; git: string }) {
    this.toolPaths = toolPaths;
  }

  async executeCommand(command: string, options: ExecOptions = {}): Promise<ExecResult> {
    try {
      logger.info(`Executing command: ${command}`);
      
      const { stdout, stderr } = await execAsync(command, {
        timeout: 300000, // 5 minutes
        maxBuffer: 1024 * 1024 * 10, // 10MB
        ...options
      });

      return {
        success: true,
        stdout: stdout.toString(),
        stderr: stderr.toString(),
        exitCode: 0
      };
    } catch (error: any) {
      logger.error(`Command execution failed: ${command}`, error);
      
      return {
        success: false,
        stdout: error.stdout || '',
        stderr: error.stderr || '',
        exitCode: error.code || 1,
        error: error.message
      };
    }
  }

  private ensureProjectSession(projectPath: string): void {
    if (this.sessions.has(projectPath)) return;
    logger.info(`Creating persistent shell session for project: ${projectPath}`);
    const shell = spawn('bash', { cwd: projectPath, env: process.env, stdio: 'pipe' });
    const session = { shell, stdoutBuffer: '', stderrBuffer: '', queue: Promise.resolve(), killed: false };
    shell.stdout.on('data', (chunk: Buffer) => { session.stdoutBuffer += chunk.toString(); });
    shell.stderr.on('data', (chunk: Buffer) => { session.stderrBuffer += chunk.toString(); });
    shell.on('exit', (code, signal) => {
      session.killed = true;
      logger.warn(`Project session exited for ${projectPath} code=${code} signal=${signal}`);
      this.sessions.delete(projectPath);
    });
    this.sessions.set(projectPath, session);
  }

  private async runInProjectSession(projectPath: string, command: string, timeoutMs = 300000): Promise<ExecResult> {
    this.ensureProjectSession(projectPath);
    const session = this.sessions.get(projectPath)!;
    // serialize commands per project
    const run = async (): Promise<ExecResult> => {
      if (session.killed) {
        // recreate session
        this.sessions.delete(projectPath);
        this.ensureProjectSession(projectPath);
      }
      const token = `__CMD_DONE_${Date.now()}_${Math.random().toString(36).slice(2)}__`;
      const donePattern = new RegExp(`${token}:(-?\\d+)`);
      const startStdoutLen = session.stdoutBuffer.length;
      const startStderrLen = session.stderrBuffer.length;
      const cmd = `${command} ; echo ${token}:$?\n`;
      logger.info(`Session exec (${projectPath}): ${command}`);
      session.shell.stdin.write(cmd);

      const start = Date.now();
      return await new Promise<ExecResult>((resolve, reject) => {
        const check = () => {
          if (Date.now() - start > timeoutMs) {
            return reject(new Error(`Command timed out after ${timeoutMs}ms: ${command}`));
          }
          const out = session.stdoutBuffer.slice(startStdoutLen);
          const err = session.stderrBuffer.slice(startStderrLen);
          const m = out.match(donePattern);
          if (m) {
            const exitCode = parseInt(m[1], 10) || 0;
            // Trim output up to token
            const tokenIndex = out.indexOf(m[0]);
            const stdout = out.slice(0, tokenIndex).trimEnd();
            resolve({ success: exitCode === 0, stdout, stderr: err, exitCode });
            return;
          }
          setTimeout(check, 50);
        };
        check();
      });
    };
    // chain to session queue
    session.queue = session.queue.then(run, run);
    return session.queue;
  }

  async executeClaudeCommand(projectPath: string, message: string): Promise<string> {
    const command = `"${this.toolPaths.claudeCli}" "${message}"`;
    const result = await this.runInProjectSession(projectPath, command);

    if (!result.success) {
      throw new Error(`Cursor command failed: ${result.error}`);
    }

    return result.stdout;
  }

  async executeNPMCommand(command: string, projectPath: string): Promise<string> {
    const fullCommand = `"${this.toolPaths.npm}" ${command}`;
    const result = await this.runInProjectSession(projectPath, fullCommand);

    if (!result.success) {
      throw new Error(`NPM command failed: ${result.error}`);
    }

    return result.stdout;
  }

  async executeGitCommand(command: string, projectPath: string): Promise<string> {
    const fullCommand = `"${this.toolPaths.git}" ${command}`;
    const result = await this.runInProjectSession(projectPath, fullCommand);

    if (!result.success) {
      throw new Error(`Git command failed: ${result.error}`);
    }

    return result.stdout;
  }
}
