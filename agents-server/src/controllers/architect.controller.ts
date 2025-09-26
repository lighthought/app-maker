import { ArchitectAgentController } from './base.controller';
import { ProjectContext, AgentResult, DevStage } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class ArchitectController { // implements ArchitectAgentController 
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
      const message = '@bmad/architect.mdc 请根据 PRD 生成架构，输出到 docs/architecture.md';
      const content = await this.commandService.executeClaudeCommand(context.projectPath, message);
      const filePath = `${context.projectPath}/docs/architecture.md`;
      await this.fileService.writeFile(filePath, content);
      await this.gitService.commitAndPush(context.projectPath, 'docs: add/update ux-spec.md by UX agent');

      return {
        success: true,
        artifacts: [{
          id: `architecture_${context.projectId}`,
          type: 'architecture' as any,
          name: 'architecture.md',
          path: filePath,
          content,
          format: 'markdown' as any,
          createdAt: new Date(),
          updatedAt: new Date()
        }],
        nextStage: DevStage.ARCH_DESIGNING as any,
        metadata: { agentType: 'architect', stage: DevStage.ARCH_DESIGNING }
      };
    } catch (error) {
      logger.error('Architect Agent failed', error);
      return { success: false, artifacts: [], error: (error as Error).message, metadata: { agentType: 'architect' } };
    }
  }

  async validate(): Promise<boolean> { return true; }
  async rollback(): Promise<void> { /* noop */ }
  async getStatus(): Promise<any> { return 'running'; }
  async updateProgress(): Promise<void> { /* noop */ }
}

