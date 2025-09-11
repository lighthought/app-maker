import axios from 'axios';
import { Server as SocketIOServer } from 'socket.io';
import { ProgressUpdate, TaskCompletedEvent, TaskFailedEvent, AgentStatusUpdate } from '../models/task.model';
import logger from '../utils/logger.util';

export class NotificationService {
  private backendApiUrl: string;
  private io: SocketIOServer;

  constructor(backendApiUrl: string, io: SocketIOServer) {
    this.backendApiUrl = backendApiUrl;
    this.io = io;
  }

  async notifyBackend(event: any): Promise<void> {
    try {
      const response = await axios.post(`${this.backendApiUrl}/api/v1/agents/notifications`, event, {
        timeout: 5000,
        headers: {
          'Content-Type': 'application/json'
        }
      });

      logger.info('Backend notification sent successfully');
    } catch (error) {
      logger.error('Failed to notify backend', error);
      // 不抛出错误，避免影响主流程
    }
  }

  async broadcastProgress(projectId: string, progress: ProgressUpdate): Promise<void> {
    try {
      // 通过WebSocket广播进度更新
      this.io.to(projectId).emit('task:progress', progress);
      
      // 通知后端API
      await this.notifyBackend({
        type: 'progress_update',
        projectId,
        data: progress
      });

      logger.info(`Progress broadcasted for project ${projectId}: ${progress.progress}%`);
    } catch (error) {
      logger.error(`Failed to broadcast progress for project ${projectId}`, error);
    }
  }

  async broadcastTaskCompleted(projectId: string, event: TaskCompletedEvent): Promise<void> {
    try {
      // 通过WebSocket广播任务完成
      this.io.to(projectId).emit('task:completed', event);
      
      // 通知后端API
      await this.notifyBackend({
        type: 'task_completed',
        projectId,
        data: event
      });

      logger.info(`Task completion broadcasted for project ${projectId}: ${event.taskId}`);
    } catch (error) {
      logger.error(`Failed to broadcast task completion for project ${projectId}`, error);
    }
  }

  async broadcastTaskFailed(projectId: string, event: TaskFailedEvent): Promise<void> {
    try {
      // 通过WebSocket广播任务失败
      this.io.to(projectId).emit('task:failed', event);
      
      // 通知后端API
      await this.notifyBackend({
        type: 'task_failed',
        projectId,
        data: event
      });

      logger.info(`Task failure broadcasted for project ${projectId}: ${event.taskId}`);
    } catch (error) {
      logger.error(`Failed to broadcast task failure for project ${projectId}`, error);
    }
  }

  async broadcastAgentStatus(agentType: string, status: AgentStatusUpdate): Promise<void> {
    try {
      // 广播Agent状态更新
      this.io.emit('agent:status', status);
      
      logger.info(`Agent status broadcasted: ${agentType} - ${status.status}`);
    } catch (error) {
      logger.error(`Failed to broadcast agent status for ${agentType}`, error);
    }
  }

  async sendErrorAlert(error: Error, context: any): Promise<void> {
    try {
      const errorEvent = {
        type: 'error_alert',
        timestamp: new Date().toISOString(),
        error: {
          name: error.name,
          message: error.message,
          stack: error.stack
        },
        context
      };

      // 通知后端API
      await this.notifyBackend(errorEvent);
      
      logger.error('Error alert sent', error);
    } catch (notifyError) {
      logger.error('Failed to send error alert', notifyError);
    }
  }
}
