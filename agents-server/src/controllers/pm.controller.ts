import { PMAgentController } from './base.controller';
import { ProjectContext, AgentResult, TaskStatus } from '../models/project.model';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { DocumentService } from '../services/document.service';
import { NotificationService } from '../services/notification.service';
import logger from '../utils/logger.util';

export class PMController implements PMAgentController {
  private commandService: CommandExecutionService;
  private fileService: FileSystemService;
  private documentService: DocumentService;
  private notificationService: NotificationService;

  constructor(
    commandService: CommandExecutionService,
    fileService: FileSystemService,
    documentService: DocumentService,
    notificationService: NotificationService
  ) {
    this.commandService = commandService;
    this.fileService = fileService;
    this.documentService = documentService;
    this.notificationService = notificationService;
  }

  async execute(context: ProjectContext): Promise<AgentResult> {
    try {
      logger.info(`PM Agent executing for project: ${context.projectId}`);
      
      // 从现有项目路径读取项目信息
      const projectInfo = await this.readProjectInfo(context.projectPath);
      
      // 基于项目信息生成PRD文档
      const prdContent = await this.generatePRD(projectInfo, context.stageInput);
      
      // 保存PRD文档到项目目录
      const prdFilePath = `${context.projectPath}/docs/PRD.md`;
      await this.fileService.writeFile(prdFilePath, prdContent);
      
      // 更新进度
      await this.updateProgress(context.projectId, 100);
      
      return {
        success: true,
        artifacts: [{
          id: `prd_${context.projectId}`,
          type: 'prd' as any,
          name: 'PRD.md',
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
      // 使用文档服务生成PRD
      const prdContent = await this.documentService.generateDocumentOutput(
        projectInfo.projectPath || '',
        'prd',
        stageInput
      );

      return prdContent;
    } catch (error) {
      logger.error('Failed to generate PRD', error);
      throw error;
    }
  }

  async readProjectInfo(projectPath: string): Promise<any> {
    try {
      // 读取项目package.json获取基本信息
      const packageJsonPath = `${projectPath}/package.json`;
      if (await this.fileService.fileExists(packageJsonPath)) {
        const packageContent = await this.fileService.readFile(packageJsonPath);
        const packageJson = JSON.parse(packageContent);
        
        return {
          name: packageJson.name,
          description: packageJson.description,
          type: packageJson.type || 'web-app',
          techStack: this.extractTechStack(packageJson)
        };
      }
      
      // 如果没有package.json，返回默认信息
      return {
        name: 'Project',
        description: 'A web application project',
        type: 'web-app',
        techStack: 'Vue.js + Node.js'
      };
    } catch (error) {
      logger.error('Failed to read project info', error);
      return {
        name: 'Project',
        description: 'A web application project',
        type: 'web-app',
        techStack: 'Vue.js + Node.js'
      };
    }
  }

  private extractTechStack(packageJson: any): string {
    const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };
    const techStack: string[] = [];
    
    if (dependencies.vue) techStack.push('Vue.js');
    if (dependencies.react) techStack.push('React');
    if (dependencies.angular) techStack.push('Angular');
    if (dependencies.express) techStack.push('Express.js');
    if (dependencies['@nestjs/core']) techStack.push('NestJS');
    if (dependencies.typescript) techStack.push('TypeScript');
    if (dependencies.node) techStack.push('Node.js');
    
    return techStack.join(' + ') || 'Vue.js + Node.js';
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
    return 'completed' as TaskStatus;
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
