import { Router, Request, Response } from 'express';
import { TaskRequest, TaskResponse } from '../models/task.model';
import { AgentTask, AgentType, DevStage, TaskStatus } from '../models/project.model';
import { TaskQueueManager } from '../queues/task-queue.manager';
import { validateRequest } from '../utils/validator.util';
import logger from '../utils/logger.util';

export class AgentRoutes {
  private router: Router;
  private taskQueueManager: TaskQueueManager;

  constructor(taskQueueManager: TaskQueueManager) {
    this.router = Router();
    this.taskQueueManager = taskQueueManager;
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

    // 获取任务状态
    this.router.get('/tasks/:taskId', this.getTaskStatus.bind(this));

    // 取消任务
    this.router.delete('/tasks/:taskId', this.cancelTask.bind(this));

    // 获取项目所有任务
    this.router.get('/projects/:projectId/tasks', this.getProjectTasks.bind(this));

    // 获取队列状态
    this.router.get('/queues/:agentType/stats', this.getQueueStats.bind(this));

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

  private async healthCheck(req: Request, res: Response): Promise<void> {
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
