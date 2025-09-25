import { UXAgentController } from './base.controller';
import { ProjectContext, AgentResult } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class UXController implements UXAgentController {
  private commandService: CommandExecutionService;
  private fileService: FileSystemService;
  private notificationService: NotificationService;
  private gitService: GitService;

  constructor(
    commandService: CommandExecutionService,
    fileService: FileSystemService,
    notificationService: NotificationService,
    gitService: GitService
  ) {
    this.commandService = commandService;
    this.fileService = fileService;
    this.notificationService = notificationService;
    this.gitService = gitService;
  }

  async execute(context: ProjectContext): Promise<AgentResult> {
    try {
      const message = '@bmad/ux.mdc 请根据 PRD 生成 UX 规范，输出到 docs/ux-spec.md';
      const content = await this.commandService.executeClaudeCommand(context.projectPath, message);
      const filePath = `${context.projectPath}/docs/ux-spec.md`;
      await this.fileService.writeFile(filePath, content);
      await this.gitService.commitAndPush(context.projectPath, 'docs: add/update ux-spec.md by UX agent');

      return {
        success: true,
        artifacts: [{
          id: `ux_${context.projectId}`,
          type: 'ux_spec' as any,
          name: 'ux-spec.md',
          path: filePath,
          content,
          format: 'markdown' as any,
          createdAt: new Date(),
          updatedAt: new Date()
        }],
        nextStage: 'arch_designing' as any,
        metadata: { agentType: 'ux', stage: 'ux_defining' }
      };
    } catch (error) {
      logger.error('UX Agent failed', error);
      return { success: false, artifacts: [], error: (error as Error).message, metadata: { agentType: 'ux' } };
    }
  }

  async validate(): Promise<boolean> { return true; }
  async rollback(): Promise<void> { /* noop */ }
  async getStatus(): Promise<any> { return 'running'; }
  async updateProgress(): Promise<void> { /* noop */ }
}

