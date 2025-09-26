import express from 'express';
import cors from 'cors';
import compression from 'compression';
import morgan from 'morgan';
import { createServer } from 'http';
import { Server as SocketIOServer } from 'socket.io';
import dotenv from 'dotenv';

import { defaultConfig } from './config/app.config';
import { RedisConnection } from './config/redis.config';
import { CommandExecutionService } from './services/command-execution.service';
import { FileSystemService } from './services/file-system.service';
import { NotificationService } from './services/notification.service';
import { TaskQueueManager } from './queues/task-queue.manager';
import { AgentRoutes } from './routes/agent.routes';
import logger from './utils/logger.util';
import swaggerUi from 'swagger-ui-express';
import { swaggerSpec } from './config/swagger.config';
import { errorHandler } from './utils/validator.util';

// 加载环境变量
dotenv.config();

export class AgentsServer {
  private app: express.Application;
  private server: any;
  private io: SocketIOServer;
  private redisConnection: RedisConnection;
  private taskQueueManager!: TaskQueueManager;
  private config = defaultConfig;
  private commandService: CommandExecutionService;

  constructor() {
    this.app = express();
    this.server = createServer(this.app);
    this.io = new SocketIOServer(this.server, {
      cors: this.config.app.cors
    });
    this.redisConnection = new RedisConnection(this.config.redis);
    this.commandService = new CommandExecutionService(this.config.tools);
    this.setupMiddleware();
    this.setupServices();
    this.setupRoutes();
    this.setupSocketIO();
  }

  private setupMiddleware(): void {
    // 基础中间件
    this.app.use(compression());
    this.app.use(morgan('combined'));
    this.app.use(cors(this.config.app.cors));
    this.app.use(express.json({ limit: '10mb' }));
    this.app.use(express.urlencoded({ extended: true }));

    // 错误处理
    this.app.use(errorHandler);
  }

  private setupServices(): void {
    // 初始化服务
    const fileService = new FileSystemService(this.config.projectDataPath);
    const notificationService = new NotificationService(this.config.backendApiUrl, this.io);

    // 初始化任务队列管理器
    this.taskQueueManager = new TaskQueueManager(
      this.config.redis.url,
      this.commandService,
      fileService,
      notificationService
    );
  }

  private setupRoutes(): void {
    // Swagger文档
    this.app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerSpec, {
      customCss: '.swagger-ui .topbar { display: none }',
      customSiteTitle: 'Agents Server API Documentation'
    }));

    // API路由
    const agentRoutes = new AgentRoutes(this.taskQueueManager, this.commandService);
    this.app.use('/api/v1/agents', agentRoutes.getRouter());

    // 根路径
    this.app.get('/', (req, res) => {
      res.json({
        name: 'Agents Server',
        version: '1.0.0',
        status: 'running',
        timestamp: new Date().toISOString()
      });
    });

    // 404处理
    this.app.use('*', (req, res) => {
      res.status(404).json({
        success: false,
        error: 'Not Found',
        message: `Route ${req.originalUrl} not found`
      });
    });
  }

  private setupSocketIO(): void {
    this.io.on('connection', (socket) => {
      logger.info(`Client connected: ${socket.id}`);

      // 加入项目房间
      socket.on('join-project', (projectId: string) => {
        socket.join(projectId);
        logger.info(`Client ${socket.id} joined project: ${projectId}`);
      });

      // 离开项目房间
      socket.on('leave-project', (projectId: string) => {
        socket.leave(projectId);
        logger.info(`Client ${socket.id} left project: ${projectId}`);
      });

      socket.on('disconnect', () => {
        logger.info(`Client disconnected: ${socket.id}`);
      });
    });
  }

  async start(): Promise<void> {
    try {
      // 连接Redis（可选）
      try {
        await this.redisConnection.connect();
        logger.info('Redis connected successfully');
      } catch (redisError) {
        logger.warn('Redis not available, proceeding without queues');
      }

      // 启动服务器
      this.server.listen(this.config.app.port, () => {
        logger.info(`Agents Server started on port ${this.config.app.port}`);
        logger.info(`Environment: ${this.config.app.nodeEnv}`);
        logger.info(`Project data path: ${this.config.projectDataPath}`);
        logger.info(`Backend API URL: ${this.config.backendApiUrl}`);
      });

      // 优雅关闭处理
      process.on('SIGTERM', this.gracefulShutdown.bind(this));
      process.on('SIGINT', this.gracefulShutdown.bind(this));
    } catch (error) {
      logger.error('Failed to start server', error);
      process.exit(1);
    }
  }

  private async gracefulShutdown(): Promise<void> {
    logger.info('Shutting down gracefully...');
    
    try {
      // 关闭任务队列
      await this.taskQueueManager.closeQueues();
      
      // 关闭Redis连接
      await this.redisConnection.disconnect();
      
      // 关闭服务器
      this.server.close(() => {
        logger.info('Server closed');
        process.exit(0);
      });
    } catch (error) {
      logger.error('Error during shutdown', error);
      process.exit(1);
    }
  }
}

// 启动服务器
if (require.main === module) {
  const server = new AgentsServer();
  server.start().catch((error) => {
    logger.error('Failed to start server', error);
    process.exit(1);
  });
}
