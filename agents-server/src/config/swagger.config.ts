import swaggerJsdoc from 'swagger-jsdoc';
import { SwaggerDefinition } from 'swagger-jsdoc';

const swaggerDefinition: SwaggerDefinition = {
  openapi: '3.0.0',
  info: {
    title: 'Agents Server API',
    version: '1.0.0',
    description: 'BMad-Method Multi-Agent Collaboration Development Service API',
    contact: {
      name: 'BMad Team',
      email: 'support@bmad.com'
    }
  },
  servers: [
    {
      url: 'http://localhost:3001',
      description: 'Development server'
    }
  ],
  components: {
    schemas: {
      TaskRequest: {
        type: 'object',
        required: ['projectId', 'userId', 'agentType', 'stage', 'context'],
        properties: {
          projectId: {
            type: 'string',
            description: '项目ID',
            example: 'project-001'
          },
          userId: {
            type: 'string',
            description: '用户ID',
            example: 'user-001'
          },
          agentType: {
            type: 'string',
            enum: ['pm', 'ux', 'architect', 'po', 'dev'],
            description: 'Agent类型',
            example: 'pm'
          },
          stage: {
            type: 'string',
            description: '开发阶段',
            example: 'prd_generating'
          },
          context: {
            type: 'object',
            required: ['projectPath'],
            properties: {
              projectPath: {
                type: 'string',
                description: '项目路径',
                example: 'F:/app-maker/app_data/projects/project-001'
              },
              projectName: {
                type: 'string',
                description: '项目名称',
                example: 'My Project'
              },
              artifacts: {
                type: 'array',
                description: '项目工件',
                items: {
                  type: 'object'
                }
              },
              stageInput: {
                type: 'object',
                description: '阶段输入数据'
              },
              previousStageOutput: {
                type: 'object',
                description: '前一阶段输出'
              }
            }
          },
          parameters: {
            type: 'object',
            description: '额外参数',
            additionalProperties: true
          }
        }
      },
      TaskResponse: {
        type: 'object',
        properties: {
          taskId: {
            type: 'string',
            description: '任务ID',
            example: 'task_1234567890_abc123'
          },
          status: {
            type: 'string',
            enum: ['pending', 'running', 'done', 'failed', 'cancelled'],
            description: '任务状态',
            example: 'pending'
          },
          message: {
            type: 'string',
            description: '响应消息',
            example: 'Agent task added to queue successfully'
          }
        }
      },
      ApiResponse: {
        type: 'object',
        properties: {
          success: {
            type: 'boolean',
            description: '请求是否成功',
            example: true
          },
          data: {
            type: 'object',
            description: '响应数据'
          },
          error: {
            type: 'string',
            description: '错误信息'
          },
          message: {
            type: 'string',
            description: '响应消息'
          }
        }
      },
      HealthResponse: {
        type: 'object',
        properties: {
          success: {
            type: 'boolean',
            example: true
          },
          status: {
            type: 'string',
            example: 'healthy'
          },
          timestamp: {
            type: 'string',
            format: 'date-time',
            example: '2024-01-01T00:00:00.000Z'
          },
          version: {
            type: 'string',
            example: '1.0.0'
          },
          service: {
            type: 'string',
            example: 'agents-server'
          }
        }
      },
      ProjectSetupRequest: {
        type: 'object',
        required: ['projectId', 'projectPath'],
        properties: {
          projectId: {
            type: 'string',
            description: '项目ID',
            example: 'project-001'
          },
          projectPath: {
            type: 'string',
            description: '项目完整路径',
            example: 'F:/app-maker/app_data/projects/project-001'
          },
          userId: {
            type: 'string',
            description: '用户ID',
            example: 'user-001'
          },
          options: {
            type: 'object',
            description: '安装选项',
            properties: {
              installBmad: {
                type: 'boolean',
                description: '是否安装bmad-method',
                default: true
              },
              installFrontend: {
                type: 'boolean',
                description: '是否安装前端依赖',
                default: true
              },
              installBackend: {
                type: 'boolean',
                description: '是否安装后端依赖',
                default: true
              }
            }
          }
        }
      },
      ProjectSetupResponse: {
        type: 'object',
        properties: {
          projectId: {
            type: 'string',
            description: '项目ID'
          },
          status: {
            type: 'string',
            enum: ['success', 'partial', 'failed'],
            description: '整体状态'
          },
          results: {
            type: 'object',
            properties: {
              bmad: {
                type: 'object',
                properties: {
                  installed: {
                    type: 'boolean',
                    description: '是否安装成功'
                  },
                  message: {
                    type: 'string',
                    description: '安装结果消息'
                  }
                }
              },
              frontend: {
                type: 'object',
                properties: {
                  installed: {
                    type: 'boolean',
                    description: '是否安装成功'
                  },
                  message: {
                    type: 'string',
                    description: '安装结果消息'
                  }
                }
              },
              backend: {
                type: 'object',
                properties: {
                  installed: {
                    type: 'boolean',
                    description: '是否安装成功'
                  },
                  message: {
                    type: 'string',
                    description: '安装结果消息'
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  tags: [
    {
      name: 'Agents',
      description: 'Agent任务管理相关接口'
    },
    {
      name: 'Projects',
      description: '项目管理相关接口'
    },
    {
      name: 'Health',
      description: '健康检查接口'
    }
  ]
};

const options = {
  definition: swaggerDefinition,
  apis: ['./src/routes/*.ts', './src/controllers/*.ts'], // 扫描路径
};

export const swaggerSpec = swaggerJsdoc(options);
