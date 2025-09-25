import { Router, Request, Response } from 'express';
import { TaskRequest, TaskResponse } from '../models/task.model';
import { AgentTask, AgentType, DevStage, TaskStatus } from '../models/project.model';
import { TaskQueueManager } from '../queues/task-queue.manager';
import { CommandExecutionService } from '../services/command-execution.service';
import { validateRequest } from '../utils/validator.util';
import logger from '../utils/logger.util';

export class AgentRoutes {
  private router: Router;
  private taskQueueManager: TaskQueueManager;
  private commandService: CommandExecutionService;

  constructor(taskQueueManager: TaskQueueManager, commandService: CommandExecutionService) {
    this.router = Router();
    this.taskQueueManager = taskQueueManager;
    this.commandService = commandService;
    this.setupRoutes();
  }

  private setupRoutes(): void {
    // 启动Agent任务
    /**
     * @swagger
     * /api/v1/agents/execute:
     *   post:
     *     summary: 执行Agent任务
     *     description: 创建一个新的Agent任务并添加到队列中执行
     *     tags: [Agents]
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             $ref: '#/components/schemas/TaskRequest'
     *     responses:
     *       200:
     *         description: 任务创建成功
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/ApiResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       $ref: '#/components/schemas/TaskResponse'
     *       400:
     *         description: 请求参数错误
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ApiResponse'
     *       500:
     *         description: 服务器内部错误
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ApiResponse'
     */
    this.router.post('/execute', 
      validateRequest(['projectId', 'userId', 'agentType', 'stage', 'context']),
      this.executeTask.bind(this)
    );

    // 同步执行一次性任务（不入队，直接执行并返回结果）
    this.router.post('/execute-sync', 
      validateRequest(['projectId', 'userId', 'agentType', 'stage', 'context']),
      this.executeTaskSync.bind(this)
    );

    // 获取任务状态
    this.router.get('/tasks/:taskId', this.getTaskStatus.bind(this));

    // 取消任务
    this.router.delete('/tasks/:taskId', this.cancelTask.bind(this));

    // 获取项目所有任务
    this.router.get('/projects/:projectId/tasks', this.getProjectTasks.bind(this));

    // 获取队列状态
    this.router.get('/queues/:agentType/stats', this.getQueueStats.bind(this));

    // 项目环境准备
    /**
     * @swagger
     * /api/v1/agents/projects/setup:
     *   post:
     *     summary: 准备项目开发环境
     *     description: 为项目安装和配置bmad-method、前端依赖、后端依赖等开发环境
     *     tags: [Projects]
     *     requestBody:
     *       required: true
     *       content:
     *         application/json:
     *           schema:
     *             type: object
     *             required:
     *               - projectId
     *               - projectPath
     *             properties:
     *               projectId:
     *                 type: string
     *                 description: 项目ID
     *                 example: "project-001"
     *               projectPath:
     *                 type: string
     *                 description: 项目完整路径
     *                 example: "F:/app-maker/app_data/projects/project-001"
     *               userId:
     *                 type: string
     *                 description: 用户ID
     *                 example: "user-001"
     *               options:
     *                 type: object
     *                 description: 安装选项
     *                 properties:
     *                   installBmad:
     *                     type: boolean
     *                     description: 是否安装bmad-method
     *                     default: true
     *                   installFrontend:
     *                     type: boolean
     *                     description: 是否安装前端依赖
     *                     default: true
     *                   installBackend:
     *                     type: boolean
     *                     description: 是否安装后端依赖
     *                     default: true
     *     responses:
     *       200:
     *         description: 环境准备成功
     *         content:
     *           application/json:
     *             schema:
     *               allOf:
     *                 - $ref: '#/components/schemas/ApiResponse'
     *                 - type: object
     *                   properties:
     *                     data:
     *                       type: object
     *                       properties:
     *                         projectId:
     *                           type: string
     *                         status:
     *                           type: string
     *                           enum: [success, partial, failed]
     *                         results:
     *                           type: object
     *                           properties:
     *                             bmad:
     *                               type: object
     *                               properties:
     *                                 installed:
     *                                   type: boolean
     *                                 message:
     *                                   type: string
     *                             frontend:
     *                               type: object
     *                               properties:
     *                                 installed:
     *                                   type: boolean
     *                                 message:
     *                                   type: string
     *                             backend:
     *                               type: object
     *                               properties:
     *                                 installed:
     *                                   type: boolean
     *                                 message:
     *                                   type: string
     *       400:
     *         description: 请求参数错误
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ApiResponse'
     *       500:
     *         description: 服务器内部错误
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/ApiResponse'
     */
    this.router.post('/projects/setup', 
      validateRequest(['projectId', 'projectPath']),
      this.setupProjectEnvironment.bind(this)
    );

    /**
     * @swagger
     * /api/v1/agents/health:
     *   get:
     *     summary: 健康检查
     *     description: 检查Agents Server服务状态
     *     tags: [Health]
     *     responses:
     *       200:
     *         description: 服务正常
     *         content:
     *           application/json:
     *             schema:
     *               $ref: '#/components/schemas/HealthResponse'
     */
    this.router.get('/health', this.healthCheck.bind(this));
  }

  private async executeTask(req: Request, res: Response): Promise<void> {
    try {
      const taskRequest: TaskRequest = req.body;
      
      // 验证项目路径是否存在
      if (!taskRequest.context.projectPath) {
        res.status(400).json({
          success: false,
          error: 'Project path is required'
        });
        return;
      }

      // 创建任务
      const task: AgentTask = {
        id: this.generateTaskId(),
        projectId: taskRequest.projectId,
        userId: taskRequest.userId,
        agentType: taskRequest.agentType as AgentType,
        stage: taskRequest.stage as DevStage,
        status: 'pending' as TaskStatus,
        progress: 0,
        parameters: taskRequest.parameters,
        context: {
          projectId: taskRequest.projectId,
          userId: taskRequest.userId,
          projectPath: taskRequest.context.projectPath,
          projectName: taskRequest.context.projectName || `Project-${taskRequest.projectId}`,
          currentStage: taskRequest.stage as DevStage,
          artifacts: taskRequest.context.artifacts || [],
          stageInput: taskRequest.context.stageInput,
          previousStageOutput: taskRequest.context.previousStageOutput
        },
        createdAt: new Date()
      };

      // 添加到队列
      await this.taskQueueManager.addTask(task);

      const response: TaskResponse = {
        taskId: task.id,
        status: 'pending' as TaskStatus,
        message: 'Agent task added to queue successfully'
      };

      res.json({
        success: true,
        data: response
      });

      logger.info(`Agent task created: ${task.id} (${task.agentType}) for stage: ${task.stage}`);
    } catch (error) {
      logger.error('Failed to execute agent task', error);
      res.status(500).json({
        success: false,
        error: 'Failed to execute agent task',
        message: (error as Error).message
      });
    }
  }

  private async executeTaskSync(req: Request, res: Response): Promise<void> {
    try {
      const taskRequest: TaskRequest = req.body;
      if (!taskRequest.context.projectPath) {
        res.status(400).json({ success: false, error: 'Project path is required' });
        return;
      }

      const controller = this.taskQueueManager.getController(taskRequest.agentType as AgentType);
      if (!controller) {
        res.status(400).json({ success: false, error: `Unsupported agentType: ${taskRequest.agentType}` });
        return;
      }

      const result = await controller.execute({
        projectId: taskRequest.projectId,
        userId: taskRequest.userId,
        projectPath: taskRequest.context.projectPath,
        projectName: taskRequest.context.projectName || `Project-${taskRequest.projectId}`,
        currentStage: taskRequest.stage as DevStage,
        artifacts: taskRequest.context.artifacts || [],
        stageInput: taskRequest.context.stageInput,
        previousStageOutput: taskRequest.context.previousStageOutput
      });

      res.json({ success: true, data: result });
    } catch (error) {
      logger.error('Failed to execute agent task sync', error);
      res.status(500).json({ success: false, error: 'Failed to execute agent task sync', message: (error as Error).message });
    }
  }

  private async getTaskStatus(req: Request, res: Response): Promise<void> {
    try {
      const { taskId } = req.params;
      
      // TODO: 实现任务状态查询
      res.json({
        success: true,
        data: {
          taskId,
          status: 'running',
          progress: 50,
          message: 'Task is running'
        }
      });
    } catch (error) {
      logger.error('Failed to get task status', error);
      res.status(500).json({
        success: false,
        error: 'Failed to get task status'
      });
    }
  }

  private async cancelTask(req: Request, res: Response): Promise<void> {
    try {
      const { taskId } = req.params;
      
      // TODO: 实现任务取消
      res.json({
        success: true,
        message: 'Task cancelled successfully'
      });
    } catch (error) {
      logger.error('Failed to cancel task', error);
      res.status(500).json({
        success: false,
        error: 'Failed to cancel task'
      });
    }
  }

  private async getProjectTasks(req: Request, res: Response): Promise<void> {
    try {
      const { projectId } = req.params;
      
      // TODO: 实现项目任务查询
      res.json({
        success: true,
        data: []
      });
    } catch (error) {
      logger.error('Failed to get project tasks', error);
      res.status(500).json({
        success: false,
        error: 'Failed to get project tasks'
      });
    }
  }

  private async getQueueStats(req: Request, res: Response): Promise<void> {
    try {
      const { agentType } = req.params;
      
      const stats = await this.taskQueueManager.getQueueStats(agentType as AgentType);
      
      res.json({
        success: true,
        data: stats
      });
    } catch (error) {
      logger.error('Failed to get queue stats', error);
      res.status(500).json({
        success: false,
        error: 'Failed to get queue stats'
      });
    }
  }

  private async setupProjectEnvironment(req: Request, res: Response): Promise<void> {
    try {
      const { projectId, projectPath, userId, options = {} } = req.body;
      
      // 默认选项
      const setupOptions = {
        installBmad: true,
        installFrontend: true,
        installBackend: true,
        ...options
      };

      logger.info(`Setting up project environment for: ${projectId} at ${projectPath}`);

      const results = {
        bmad: { installed: false, message: '' },
        frontend: { installed: false, message: '' },
        backend: { installed: false, message: '' }
      };

      let overallStatus = 'success';

      // 1. 安装 bmad-method
      if (setupOptions.installBmad) {
        try {
          logger.info('Installing bmad-method...');
          const bmadResult = await this.commandService.executeCommand(
            `cd "${projectPath}" && npx bmad-method install -f -i claude -d .`,
            { timeout: 300000 } // 5分钟超时
          );
          
          if (bmadResult.success) {
            results.bmad.installed = true;
            results.bmad.message = 'bmad-method installed successfully';
            logger.info('bmad-method installation completed');
          } else {
            results.bmad.message = `bmad-method installation failed: ${bmadResult.error}`;
            logger.error('bmad-method installation failed', bmadResult.error);
            overallStatus = 'partial';
          }
        } catch (error) {
          results.bmad.message = `bmad-method installation error: ${(error as Error).message}`;
          logger.error('bmad-method installation error', error);
          overallStatus = 'partial';
        }
      }

      // 2. 安装前端依赖
      if (setupOptions.installFrontend) {
        try {
          const frontendPath = `${projectPath}/frontend`;
          logger.info('Installing frontend dependencies...');
          
          const frontendResult = await this.commandService.executeCommand(
            `cd "${frontendPath}" && npm install`,
            { timeout: 300000 } // 5分钟超时
          );
          
          if (frontendResult.success) {
            results.frontend.installed = true;
            results.frontend.message = 'Frontend dependencies installed successfully';
            logger.info('Frontend dependencies installation completed');
          } else {
            results.frontend.message = `Frontend installation failed: ${frontendResult.error}`;
            logger.error('Frontend installation failed', frontendResult.error);
            overallStatus = 'partial';
          }
        } catch (error) {
          results.frontend.message = `Frontend installation error: ${(error as Error).message}`;
          logger.error('Frontend installation error', error);
          overallStatus = 'partial';
        }
      }

      // 3. 安装后端依赖
      if (setupOptions.installBackend) {
        try {
          const backendPath = `${projectPath}/backend`;
          logger.info('Installing backend dependencies...');
          
          // 安装Go依赖
          const goModResult = await this.commandService.executeCommand(
            `cd "${backendPath}" && go mod download`,
            { timeout: 300000 } // 5分钟超时
          );
          
          // 安装swagger工具
          const swaggerResult = await this.commandService.executeCommand(
            'go install github.com/swaggo/swag/cmd/swag@latest',
            { timeout: 120000 } // 2分钟超时
          );
          
          // 构建后端项目
          const buildResult = await this.commandService.executeCommand(
            `cd "${backendPath}" && go build -o server ./cmd/server`,
            { timeout: 300000 } // 5分钟超时
          );
          
          if (goModResult.success && buildResult.success) {
            results.backend.installed = true;
            results.backend.message = 'Backend dependencies and build completed successfully';
            logger.info('Backend setup completed');
          } else {
            results.backend.message = `Backend setup failed: Go mod: ${goModResult.success ? 'OK' : goModResult.error}, Build: ${buildResult.success ? 'OK' : buildResult.error}`;
            logger.error('Backend setup failed', { goModResult, buildResult });
            overallStatus = 'partial';
          }
        } catch (error) {
          results.backend.message = `Backend setup error: ${(error as Error).message}`;
          logger.error('Backend setup error', error);
          overallStatus = 'partial';
        }
      }

      res.json({
        success: true,
        data: {
          projectId,
          status: overallStatus,
          results
        }
      });

      logger.info(`Project environment setup completed for ${projectId} with status: ${overallStatus}`);
    } catch (error) {
      logger.error('Failed to setup project environment', error);
      res.status(500).json({
        success: false,
        error: 'Failed to setup project environment',
        message: (error as Error).message
      });
    }
  }

  private async healthCheck(req: Request, res: Response): Promise<void> {
    logger.info('Health check');
    // 检查 claude 命令执行是否正常，检查 claude --version 版本号
    const result = await this.commandService.executeCommand('claude --version');
    if (result.success) {
      logger.info('Claude command executed successfully');
    } else {
      logger.error('Claude command executed failed');
    }
    res.json({
      success: true,
      status: 'healthy',
      timestamp: new Date().toISOString(),
      version: '1.0.0',
      service: 'agents-server'
    });
  }

  private generateTaskId(): string {
    return `task_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  getRouter(): Router {
    return this.router;
  }
}
