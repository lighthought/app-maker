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
      const message = '@bmad/architect.mdc 请你基于最新的 @docs/PRD.md 和 UX 专家的设计文档 @docs/ux/ux-spec.md，帮我把整体架构设计 Architect.md, 前端架构设计frontend_arch.md, 后端架构设计backend_arch.md输出到 docs/arch/ 目录下。\n当前的项目代码是由模板生成，技术架构是：\n1. 前端：vue.js+ vite ；\n2. 后端服务和 API： GO + Gin 框架实现 API、数据库用 PostgreSql、缓存用 Redis。\n3. 部署相关的脚本已经有了，用的 docker，前端用一个 nginx ，配置 /api 重定向到 /backend:port ，这样就能在前端项目中访问后端 API 了。引用关系是：前端依赖后端，后端依赖 Redis 和 PostgreSql。';
      const content = await this.commandService.executeClaudeCommand(context.projectPath, message);
      const filePath = `${context.projectPath}/docs/arch/Architect.md`;
      await this.fileService.writeFile(filePath, content);
      await this.gitService.commitAndPush(context.projectPath, 'docs: add/update architecture by Architect agent');

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
        nextStage: DevStage.DATA_MODELING as any,
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

