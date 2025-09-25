import { createLogger, format, transports } from 'winston';
import * as fs from 'fs';
import * as path from 'path';

const logFilePath = process.env.AGENTS_SERVER_LOG || path.resolve(process.cwd(), 'logs', 'agents-server.log');
try {
  const dir = path.dirname(logFilePath);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
} catch {}

const logger = createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: format.combine(
    format.timestamp(),
    format.errors({ stack: true }),
    format.json()
  ),
  transports: [
    new transports.Console({
      format: format.combine(
        format.colorize(),
        format.simple()
      )
    }),
    new transports.File({ 
      filename: logFilePath,
      maxsize: 5242880, // 5MB
      maxFiles: 5
    })
  ]
});

export { logger };
export default logger;
