import { PMAgentController } from './base.controller';
import { ProjectContext, AgentResult, TaskStatus } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class PMController implements PMAgentController {
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
      logger.info(`PM Agent executing for project: ${context.projectId}`);
      
      const projectPath = context.projectPath;
      // 基于项目信息生成PRD文档
      const prdContent = await this.generatePRD(projectPath, context.stageInput);
      
      // 保存PRD文档到项目目录（按官方约定路径）
      const prdFilePath = `${context.projectPath}/docs/prd.md`;
      await this.fileService.writeFile(prdFilePath, prdContent);
      
      // 更新进度
      await this.updateProgress(context.projectId, 100);
      
      // 提交并推送到 GitLab
      await this.gitService.commitAndPush(context.projectPath, 'docs: add/update prd.md by PM agent');

      return {
        success: true,
        artifacts: [{
          id: `prd_${context.projectId}`,
          type: 'prd' as any,
          name: 'prd.md',
          path: prdFilePath,
          content: prdContent,
          format: 'markdown' as any,
          createdAt: new Date(),
          updatedAt: new Date()
        }],
        nextStage: 'ux_defining' as any,
        metadata: {
          agentType: 'pm',
          stage: 'prd_generating',
          projectPath: context.projectPath
        }
      };
    } catch (error) {
      logger.error(`PM Agent execution failed for project: ${context.projectId}`, error);
      return {
        success: false,
        artifacts: [],
        error: (error as Error).message,
        metadata: {
          agentType: 'pm',
          stage: 'prd_generating'
        }
      };
    }
  }

  async generatePRD(projectInfo: any, stageInput: any): Promise<string> {
    try {
      const message = `@bmad/pm.mdc 请根据以下需求生成PRD 到 docs/PRD.md：${stageInput.requirements} \r\n\r\n技术选型我后续再和架构师深入讨论，主题颜色我后续再和 ux 专家讨论。`;

      // 使用文档服务生成PRD
      const prdContent = await this.commandService.executeClaudeCommand(
        projectInfo.projectPath || '',
        message
      );

      return prdContent;
    } catch (error) {
      logger.error('Failed to generate PRD', error);
      throw error;
    }
  }


  async clarifyRequirements(questions: string[]): Promise<string> {
    // TODO: 实现需求澄清逻辑
    return 'Requirements clarified';
  }

  async validate(context: ProjectContext): Promise<boolean> {
    return context.projectPath && context.stageInput && context.stageInput.requirements;
  }

  async rollback(context: ProjectContext): Promise<void> {
    // TODO: 实现回滚逻辑
    logger.info(`PM Agent rollback for project: ${context.projectId}`);
  }

  async getStatus(taskId: string): Promise<TaskStatus> {
    // TODO: 实现状态查询逻辑
    return 'done' as TaskStatus;
  }

  async updateProgress(taskId: string, progress: number): Promise<void> {
    try {
      await this.notificationService.broadcastProgress(taskId, {
        taskId,
        projectId: taskId,
        progress,
        message: `PM Agent progress: ${progress}%`,
        stage: 'prd_generating'
      });
    } catch (error) {
      logger.error('Failed to update progress', error);
    }
  }
}
