import { DevAgentController } from './base.controller';
import { ProjectContext, AgentResult, DevStage } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class DevController { // implements DevAgentController 
  private commandService: CommandExecutionService;
  //private fileService: FileSystemService;
  //private notificationService: NotificationService;
  private gitService: GitService;

  constructor(
    commandService: CommandExecutionService,
    //fileService: FileSystemService,
    //notificationService: NotificationService,
    gitService: GitService
  ) {
    this.commandService = commandService;
    //this.fileService = fileService;
    //this.notificationService = notificationService;
    this.gitService = gitService;
  }

  async execute(context: ProjectContext): Promise<AgentResult> {
    try {
      // 开发阶段：简单示意，实际应按 stories 分片循环处理
      const message = '@bmad/dev.mdc 根据 docs/stories/ 的内容实现功能，并更新对应代码';
      await this.commandService.executeClaudeCommand(context.projectPath, message);

      // 运行测试（可选）
      try { await this.commandService.executeNPMCommand('test -w', context.projectPath); } catch {}

      // Commit and push
      await this.gitService.commitAndPush(context.projectPath, 'feat: implement stories by Dev agent');

      return {
        success: true,
        artifacts: [],
        nextStage: DevStage.TESTING as any,
        metadata: { agentType: 'dev', stage: DevStage.TESTING }
      };
    } catch (error) {
      logger.error('Dev Agent failed', error);
      return { success: false, artifacts: [], error: (error as Error).message, metadata: { agentType: 'dev' } };
    }
  }

  async validate(): Promise<boolean> { return true; }
  async rollback(): Promise<void> { /* noop */ }
  async getStatus(): Promise<any> { return 'running'; }
  async updateProgress(): Promise<void> { /* noop */ }
}

