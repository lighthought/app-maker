import { UXAgentController } from './base.controller';
import { ProjectContext, AgentResult, DevStage } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class UXController { // implements UXAgentController 
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
      const message = '@bmad/ux-expert.mdc 帮我基于这个 @docs/PRD.md 和参考页面设计(如果需求有提及的话)，输出前端的 UX Spec 到 docs/ux/ux-spec.md。关键web页面的文生网站提示词到 docs/ux/page-prompt.md。';
      const content = await this.commandService.executeClaudeCommand(context.projectPath, message);
      const filePath = `${context.projectPath}/docs/ux/ux-spec.md`;
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
        nextStage: DevStage.ARCH_DESIGNING as any,
        metadata: { agentType: 'ux', stage: DevStage.UX_DEFINING }
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

