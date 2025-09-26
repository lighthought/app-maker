import { POAgentController } from './base.controller';
import { ProjectContext, AgentResult, DevStage } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class POController { // implements POAgentController 
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
      const epicsMsg = '@bmad/po-epics.mdc 请根据 PRD 和架构，生成分片的 Epics，输出到 docs/epics/ 下多个文件';
      const storiesMsg = '@bmad/po-stories.mdc 请根据 Epics，生成分片的 Stories，输出到 docs/stories/ 下多个文件';

      // 基于工具让 agent 输出内容；此处以执行命令示意
      const epicsContent = await this.commandService.executeClaudeCommand(context.projectPath, epicsMsg);
      const storiesContent = await this.commandService.executeClaudeCommand(context.projectPath, storiesMsg);

      // 简化：写入单文件，后续再按分片拆分
      const epicsFile = `${context.projectPath}/docs/epics/epics.md`;
      const storiesFile = `${context.projectPath}/docs/stories/stories.md`;
      await this.fileService.writeFile(epicsFile, epicsContent);
      await this.fileService.writeFile(storiesFile, storiesContent);

      await this.gitService.commitAndPush(context.projectPath, 'docs: add/update epics and stories by PO agent');

      return {
        success: true,
        artifacts: [
          { id: `epics_${context.projectId}`, type: 'epics' as any, name: 'epics.md', path: epicsFile, content: epicsContent, format: 'markdown' as any, createdAt: new Date(), updatedAt: new Date() },
          { id: `stories_${context.projectId}`, type: 'stories' as any, name: 'stories.md', path: storiesFile, content: storiesContent, format: 'markdown' as any, createdAt: new Date(), updatedAt: new Date() }
        ],
        nextStage: DevStage.STORY_DEVELOPING as any,
        metadata: { agentType: 'po', stage: DevStage.EPIC_PLANNING }
      };
    } catch (error) {
      logger.error('PO Agent failed', error);
      return { success: false, artifacts: [], error: (error as Error).message, metadata: { agentType: 'po' } };
    }
  }

  async validate(): Promise<boolean> { return true; }
  async rollback(): Promise<void> { /* noop */ }
  async getStatus(): Promise<any> { return 'running'; }
  async updateProgress(): Promise<void> { /* noop */ }
}

