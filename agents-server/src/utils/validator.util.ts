import { Request, Response, NextFunction } from 'express';
import { DevStage } from '../models/project.model';

// 验证工具类
export class Validator {
  static isString(value: any): value is string {
    return typeof value === 'string' && value.trim().length > 0;
  }

  static isNumber(value: any): value is number {
    return typeof value === 'number' && !isNaN(value);
  }

  static isArray(value: any): value is any[] {
    return Array.isArray(value);
  }

  static isObject(value: any): value is object {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
  }

  static isValidProjectId(projectId: string): boolean {
    return this.isString(projectId) && projectId.length >= 10;
  }

  static isValidUserId(userId: string): boolean {
    return this.isString(userId) && userId.length >= 10;
  }

  static isValidAgentType(agentType: string): boolean {
    const validTypes = ['pm', 'ux', 'architect', 'po', 'dev'];
    return this.isString(agentType) && validTypes.includes(agentType);
  }

  static isValidStage(stage: DevStage): boolean {
    const validStages = [
      DevStage.PRD_GENERATING,
      DevStage.UX_DEFINING, 
      DevStage.ARCH_DESIGNING,
      DevStage.DATA_MODELING,
      DevStage.API_DEFINING,
      DevStage.EPIC_PLANNING,
      DevStage.STORY_DEVELOPING,
      DevStage.BUG_FIXING,
      DevStage.TESTING,
      DevStage.PACKAGING
    ];
    return this.isString(stage) && validStages.includes(stage);
  }
}

// 错误处理中间件
export const errorHandler = (
  error: Error,
  req: Request,
  res: Response,
  next: NextFunction
) => {
  console.error('Error:', error);
  
  res.status(500).json({
    success: false,
    error: 'Internal Server Error',
    message: process.env.NODE_ENV === 'development' ? error.message : 'Something went wrong'
  });
};

// 请求验证中间件
export const validateRequest = (requiredFields: string[]) => {
  return (req: Request, res: Response, next: NextFunction) => {
    const missingFields = requiredFields.filter(field => !req.body[field]);
    
    if (missingFields.length > 0) {
      return res.status(400).json({
        success: false,
        error: 'Missing required fields',
        missingFields
      });
    }
    
    next();
  };
};
