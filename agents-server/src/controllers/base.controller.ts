import { ProjectContext, AgentResult, TaskStatus } from '../models/project.model';

// Agent控制器基础接口
export interface AgentController {
  // 基础接口
  execute(context: ProjectContext): Promise<AgentResult>;
  validate(context: ProjectContext): Promise<boolean>;
  rollback(context: ProjectContext): Promise<void>;
  
  // 状态管理
  getStatus(taskId: string): Promise<TaskStatus>;
  updateProgress(taskId: string, progress: number): Promise<void>;
}

// PM Agent 专用接口
export interface PMAgentController extends AgentController {
  generatePRD(projectInfo: any, stageInput: any): Promise<string>;
  clarifyRequirements(questions: string[]): Promise<string>;
}

// UX Expert Agent 专用接口
export interface UXAgentController extends AgentController {
  generateUXSpec(prd: string, figmaUrl?: string): Promise<string>;
  createDesignSystem(uxSpec: string): Promise<string>;
}

// Architect Agent 专用接口
export interface ArchitectAgentController extends AgentController {
  designArchitecture(prd: string, uxSpec: string): Promise<string>;
  selectTechStack(requirements: string): Promise<string>;
}

// PO Agent 专用接口
export interface POAgentController extends AgentController {
  createEpics(prd: string, architecture: string): Promise<string>;
  createStories(epics: string): Promise<string>;
}

// Dev Agent 专用接口
export interface DevAgentController extends AgentController {
  implementStory(story: string, context: ProjectContext): Promise<string>;
  runTests(projectPath: string): Promise<string>;
  fixBugs(bugReport: string): Promise<string>;
}
