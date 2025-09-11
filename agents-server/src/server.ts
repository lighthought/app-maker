import { AgentsServer } from './app';
import logger from './utils/logger.util';

// 启动Agents Server
async function startServer() {
  try {
    logger.info('Starting Agents Server...');
    
    const server = new AgentsServer();
    await server.start();
    
    logger.info('Agents Server started successfully');
  } catch (error) {
    logger.error('Failed to start Agents Server', error);
    process.exit(1);
  }
}

// 处理未捕获的异常
process.on('uncaughtException', (error) => {
  logger.error('Uncaught Exception:', error);
  process.exit(1);
});

process.on('unhandledRejection', (reason, promise) => {
  logger.error('Unhandled Rejection at:', promise, 'reason:', reason);
  process.exit(1);
});

// 启动服务器
startServer();
