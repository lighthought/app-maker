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
      const message = '@bmad/dev.mdc 请你始终记得项目的前后端框架及约束：\n1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；\n2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。\n3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。\n4. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程；\n\n请你基于 @docs/PRD.md 和架构师的设计 @docs/arch/ ，以及 UX 标准 @docs/ux/ 实现 @docs/epics/ 中的用户故事，优先实现 @docs/stories/ 下的故事。实现完，编译确认下验收的标准是否都达到了，达到了以后，更新用户故事文档，勾上对应的验收标准。然后再询问我，是否继续。不要每次生成多余的总结文档。';
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

