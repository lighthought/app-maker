import { FileSystemService } from './file-system.service';
import { logger } from '../utils/logger.util';

// 文档服务 - 读取现有项目文档并按结构输出
export class DocumentService {
  private fileService: FileSystemService;

  constructor(fileService: FileSystemService) {
    this.fileService = fileService;
  }

  // 读取项目中的现有文档
  async readProjectDocument(projectPath: string, docType: string): Promise<string> {
    try {
      const docPath = this.getDocumentPath(projectPath, docType);
      
      if (await this.fileService.fileExists(docPath)) {
        return await this.fileService.readFile(docPath);
      }
      
      logger.warn(`Document not found: ${docPath}`);
      return '';
    } catch (error) {
      logger.error(`Failed to read project document: ${docType}`, error);
      throw error;
    }
  }

  // 根据文档类型生成输出
  async generateDocumentOutput(projectPath: string, docType: string, stageInput: any): Promise<string> {
    try {
      // 读取现有文档作为参考
      const existingDoc = await this.readProjectDocument(projectPath, docType);
      
      // 根据文档类型和输入生成输出
      switch (docType) {
        case 'prd':
          return await this.generatePRDOutput(projectPath, stageInput, existingDoc);
        case 'ux-spec':
          return await this.generateUXOutput(projectPath, stageInput, existingDoc);
        case 'architecture':
          return await this.generateArchitectureOutput(projectPath, stageInput, existingDoc);
        case 'story':
          return await this.generateStoryOutput(projectPath, stageInput, existingDoc);
        default:
          throw new Error(`Unsupported document type: ${docType}`);
      }
    } catch (error) {
      logger.error(`Failed to generate document output: ${docType}`, error);
      throw error;
    }
  }

  // 获取文档路径
  private getDocumentPath(projectPath: string, docType: string): string {
    const docPaths: Record<string, string> = {
      'prd': `${projectPath}/docs/PRD.md`,
      'ux-spec': `${projectPath}/docs/UX_Design_Specifications.md`,
      'architecture': `${projectPath}/docs/Backend_Architecture.md`,
      'story': `${projectPath}/docs/stories/epic1-project-creation-stories.md`
    };
    
    return docPaths[docType] || `${projectPath}/docs/${docType}.md`;
  }

  // 生成PRD输出
  private async generatePRDOutput(projectPath: string, stageInput: any, existingDoc: string): Promise<string> {
    // 基于现有PRD结构和stageInput生成新的PRD内容
    const projectInfo = await this.readProjectInfo(projectPath);
    
    return `# 产品需求文档 (PRD)

## 项目概述
**项目名称**: ${projectInfo.name}
**版本**: 1.0.0
**创建时间**: ${new Date().toISOString()}
**负责人**: PM Agent

## 需求描述
${stageInput.requirements || '项目需求描述'}

## 功能需求
### 核心功能
- [ ] 功能1
- [ ] 功能2
- [ ] 功能3

## 技术约束
### 技术栈
${projectInfo.techStack}

## 验收标准
1. 所有核心功能正常工作
2. 性能指标达标
3. 用户界面友好

## 项目计划
### 阶段1: 需求分析 (1周)
### 阶段2: 开发实现 (4周)
### 阶段3: 测试部署 (1周)
`;
  }

  // 生成UX输出
  private async generateUXOutput(projectPath: string, stageInput: any, existingDoc: string): Promise<string> {
    return `# UX设计规范

## 设计原则
1. **简洁性**: 界面简洁明了，避免冗余元素
2. **一致性**: 保持设计元素的一致性
3. **可用性**: 确保用户能够轻松完成任务

## 视觉设计
### 色彩规范
- **主色调**: #EDF3F8 (淡蓝灰)
- **辅助色**: #254D72 (深蓝)
- **强调色**: #FFA940 (橙黄)

## 组件设计
### 按钮组件
- **主要按钮**: 背景色 #254D72, 文字白色
- **次要按钮**: 边框色 #254D72, 文字 #254D72

## 交互设计
### 状态反馈
- **悬停**: 轻微阴影变化
- **点击**: 轻微缩放效果
- **加载**: 旋转动画
`;
  }

  // 生成架构输出
  private async generateArchitectureOutput(projectPath: string, stageInput: any, existingDoc: string): Promise<string> {
    const projectInfo = await this.readProjectInfo(projectPath);
    
    return `# 系统架构设计

## 架构概述
### 系统架构图
\`\`\`
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用       │    │   后端API       │    │   数据库        │
│   ${projectInfo.frontend || 'Vue.js'}        │◄──►│   ${projectInfo.backend || 'Node.js'}       │◄──►│   ${projectInfo.database || 'PostgreSQL'}    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
\`\`\`

## 技术栈
### 前端技术栈
- **框架**: ${projectInfo.frontend || 'Vue.js 3.x'}
- **语言**: TypeScript
- **构建工具**: Vite

### 后端技术栈
- **运行时**: Node.js 18+
- **框架**: Express.js
- **语言**: TypeScript

## 系统架构
### 分层架构
1. **表现层**: 前端应用
2. **API层**: REST API
3. **业务层**: 业务逻辑
4. **数据层**: 数据库
`;
  }

  // 生成Story输出
  private async generateStoryOutput(projectPath: string, stageInput: any, existingDoc: string): Promise<string> {
    return `# Epic和Story规划

## Epic 1: 核心功能开发

### Epic描述
实现项目的核心功能模块

### Stories

#### Story 1.1: 基础功能实现
**优先级**: P0
**估算工时**: 3天
**负责人**: 开发工程师

**用户故事**
作为用户，我希望能够使用核心功能，以便完成主要任务。

**验收标准**
- [ ] 功能正常工作
- [ ] 界面友好
- [ ] 性能达标

## 项目计划
### 第一阶段 (2周)
- Epic 1: 核心功能开发

### 第二阶段 (2周)
- 测试和优化
`;
  }

  // 读取项目信息
  private async readProjectInfo(projectPath: string): Promise<any> {
    try {
      const packageJsonPath = `${projectPath}/package.json`;
      if (await this.fileService.fileExists(packageJsonPath)) {
        const packageContent = await this.fileService.readFile(packageJsonPath);
        const packageJson = JSON.parse(packageContent);
        
        return {
          name: packageJson.name || 'Project',
          description: packageJson.description || 'A web application project',
          frontend: this.extractFrontendTech(packageJson),
          backend: this.extractBackendTech(packageJson),
          database: 'PostgreSQL',
          techStack: this.extractTechStack(packageJson)
        };
      }
      
      return {
        name: 'Project',
        description: 'A web application project',
        frontend: 'Vue.js',
        backend: 'Node.js',
        database: 'PostgreSQL',
        techStack: 'Vue.js + Node.js'
      };
    } catch (error) {
      logger.error('Failed to read project info', error);
      return {
        name: 'Project',
        description: 'A web application project',
        frontend: 'Vue.js',
        backend: 'Node.js',
        database: 'PostgreSQL',
        techStack: 'Vue.js + Node.js'
      };
    }
  }

  private extractFrontendTech(packageJson: any): string {
    const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };
    if (dependencies.vue) return 'Vue.js';
    if (dependencies.react) return 'React';
    if (dependencies.angular) return 'Angular';
    return 'Vue.js';
  }

  private extractBackendTech(packageJson: any): string {
    const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };
    if (dependencies.express) return 'Express.js';
    if (dependencies['@nestjs/core']) return 'NestJS';
    return 'Node.js';
  }

  private extractTechStack(packageJson: any): string {
    const dependencies = { ...packageJson.dependencies, ...packageJson.devDependencies };
    const techStack: string[] = [];
    
    if (dependencies.vue) techStack.push('Vue.js');
    if (dependencies.react) techStack.push('React');
    if (dependencies.express) techStack.push('Express.js');
    if (dependencies.typescript) techStack.push('TypeScript');
    
    return techStack.join(' + ') || 'Vue.js + Node.js';
  }
}
