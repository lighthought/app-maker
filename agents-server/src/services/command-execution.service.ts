import { exec, ExecOptions } from 'child_process';
import { promisify } from 'util';
import { ExecResult } from '../models/artifact.model';
import { logger } from '../utils/logger.util';

const execAsync = promisify(exec);

export class CommandExecutionService {
  private toolPaths: {
    cursorCli: string;
    npm: string;
    git: string;
  };

  constructor(toolPaths: { cursorCli: string; npm: string; git: string }) {
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

  async executeCursorCommand(projectPath: string, message: string): Promise<string> {
    const command = `"${this.toolPaths.cursorCli}" chat --project "${projectPath}" --message "${message}"`;
    
    const result = await this.executeCommand(command, {
      cwd: projectPath
    });

    if (!result.success) {
      throw new Error(`Cursor command failed: ${result.error}`);
    }

    return result.stdout;
  }

  async executeNPMCommand(command: string, projectPath: string): Promise<string> {
    const fullCommand = `"${this.toolPaths.npm}" ${command}`;
    
    const result = await this.executeCommand(fullCommand, {
      cwd: projectPath
    });

    if (!result.success) {
      throw new Error(`NPM command failed: ${result.error}`);
    }

    return result.stdout;
  }

  async executeGitCommand(command: string, projectPath: string): Promise<string> {
    const fullCommand = `"${this.toolPaths.git}" ${command}`;
    
    const result = await this.executeCommand(fullCommand, {
      cwd: projectPath
    });

    if (!result.success) {
      throw new Error(`Git command failed: ${result.error}`);
    }

    return result.stdout;
  }
}
