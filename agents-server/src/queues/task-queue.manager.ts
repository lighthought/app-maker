import Queue from 'bull';
import { AgentTask, AgentType, DevStage } from '../models/project.model';
import { PMController } from '../controllers/pm.controller';
import { UXController } from '../controllers/ux.controller';
import { ArchitectController } from '../controllers/architect.controller';
import { POController } from '../controllers/po.controller';
import { DevController } from '../controllers/dev.controller';
import { CommandExecutionService } from '../services/command-execution.service';
import { FileSystemService } from '../services/file-system.service';
import { NotificationService } from '../services/notification.service';
import { GitService } from '../services/git.service';
import logger from '../utils/logger.util';

export class TaskQueueManager {
  private queues: Map<AgentType, Queue.Queue<AgentTask>> = new Map();
  private controllers: Map<AgentType, any> = new Map();
  private notificationService: NotificationService | undefined;

  constructor(
    redisUrl: string,
    commandService: CommandExecutionService,
    fileService: FileSystemService,
    notificationService: NotificationService
  ) {
    this.notificationService = notificationService;
    this.initializeQueues(redisUrl);
    this.initializeControllers(commandService, fileService, notificationService);
    this.setupProcessors();
  }

  private initializeQueues(redisUrl: string): void {
    const agentTypes: AgentType[] = [AgentType.PM, AgentType.UX, AgentType.ARCHITECT, AgentType.PO, AgentType.DEV];
    
    agentTypes.forEach(agentType => {
      const queue = new Queue<AgentTask>(`agent-${agentType}`, redisUrl, {
        defaultJobOptions: {
          removeOnComplete: 10,
          removeOnFail: 5,
          attempts: 3,
          backoff: {
            type: 'exponential',
            delay: 2000
          }
        }
      });

      this.queues.set(agentType, queue);
      logger.info(`Queue initialized for agent: ${agentType}`);
    });
  }

  private initializeControllers(
    commandService: CommandExecutionService,
    fileService: FileSystemService,
    notificationService: NotificationService
  ): void {
    // 初始化PM控制器
    const gitService = new GitService(commandService, fileService);
    const pmController = new PMController(commandService, fileService, notificationService, gitService);
    this.controllers.set(AgentType.PM, pmController);
    const uxController = new UXController(commandService, fileService, notificationService, gitService);
    this.controllers.set(AgentType.UX, uxController);
    const archController = new ArchitectController(commandService, fileService, notificationService, gitService);
    this.controllers.set(AgentType.ARCHITECT, archController);
    const poController = new POController(commandService, fileService, notificationService, gitService);
    this.controllers.set(AgentType.PO, poController);
    const devController = new DevController(commandService, fileService, notificationService, gitService);
    this.controllers.set(AgentType.DEV, devController);
  }

  private setupProcessors(): void {
    this.queues.forEach((queue, agentType) => {
      queue.process(async (job) => {
        try {
          logger.info(`Processing ${agentType} task: ${job.id}`);
          
          const controller = this.controllers.get(agentType);
          if (!controller) {
            throw new Error(`No controller found for agent type: ${agentType}`);
          }

          // 初始进度
          await job.progress(5);
          if (this.notificationService) {
            await this.notificationService.broadcastProgress(job.data.projectId, {
              taskId: job.id as string,
              projectId: job.data.projectId,
              progress: 5,
              message: `Task started: ${agentType}`,
              stage: job.data.stage as unknown as string
            });
          }

          const result = await controller.execute(job.data.context);
          
          // 更新任务进度
          await job.progress(100);
          if (this.notificationService) {
            await this.notificationService.broadcastProgress(job.data.projectId, {
              taskId: job.id as string,
              projectId: job.data.projectId,
              progress: 100,
              message: `Task completed: ${agentType}`,
              stage: job.data.stage as unknown as string
            });
            await this.notificationService.broadcastTaskCompleted(job.data.projectId, {
              taskId: job.id as string,
              projectId: job.data.projectId,
              result,
              artifacts: (result?.artifacts || []).map((a: any) => a.path)
            });
          }
          
          logger.info(`${agentType} task done: ${job.id}`);
          return result;
        } catch (error) {
          logger.error(`${agentType} task failed: ${job.id}`, error);
          if (this.notificationService) {
            try {
              await this.notificationService.broadcastTaskFailed((job as any).data?.projectId || 'unknown', {
                taskId: job.id as string,
                projectId: (job as any).data?.projectId || 'unknown',
                error: (error as Error).message,
                retryable: true
              });
            } catch {}
          }
          throw error;
        }
      });

      // 监听队列事件
      queue.on('done', (job, result) => {
        logger.info(`Task done: ${job.id}`);
      });

      queue.on('failed', (job, err) => {
        logger.error(`Task failed: ${job.id}`, err);
      });

      queue.on('progress', (job, progress) => {
        logger.info(`Task progress: ${job.id} - ${progress}%`);
      });
    });
  }

  async addTask(task: AgentTask): Promise<void> {
    const queue = this.queues.get(task.agentType);
    if (!queue) {
      throw new Error(`No queue found for agent type: ${task.agentType}`);
    }

    await queue.add(task, {
      jobId: task.id,
      delay: 0
    });

    logger.info(`Task added to queue: ${task.id} (${task.agentType})`);
  }

  async getQueueStats(agentType: AgentType): Promise<any> {
    const queue = this.queues.get(agentType);
    if (!queue) {
      throw new Error(`No queue found for agent type: ${agentType}`);
    }

    return {
      waiting: await queue.getWaiting(),
      active: await queue.getActive(),
      completed: await queue.getCompleted(),
      failed: await queue.getFailed()
    };
  }

  async pauseQueue(agentType: AgentType): Promise<void> {
    const queue = this.queues.get(agentType);
    if (queue) {
      await queue.pause();
      logger.info(`Queue paused: ${agentType}`);
    }
  }

  async resumeQueue(agentType: AgentType): Promise<void> {
    const queue = this.queues.get(agentType);
    if (queue) {
      await queue.resume();
      logger.info(`Queue resumed: ${agentType}`);
    }
  }

  async closeQueues(): Promise<void> {
    const closePromises = Array.from(this.queues.values()).map(queue => queue.close());
    await Promise.all(closePromises);
    logger.info('All queues closed');
  }

  getController(agentType: AgentType): any | undefined {
    return this.controllers.get(agentType);
  }
}
