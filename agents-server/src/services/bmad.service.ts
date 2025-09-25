import { CommandExecutionService } from './command-execution.service';
import logger from '../utils/logger.util';

export class BmadService {
  private commandService: CommandExecutionService;

  constructor(commandService: CommandExecutionService) {
    this.commandService = commandService;
  }

  // 安装 BMAD 工具
  async install(projectPath: string): Promise<boolean> {
    try {
      const result = await this.commandService.executeCommand(
        `cd "${projectPath}" && npx bmad-method install -f -i claude -d .`,
        { timeout: 5 * 60 * 1000 }
      );
      return !!result.success;
    } catch (error) {
      logger.error('BMAD install failed', error);
      return false;
    }
  }

  // 根据 BMAD 角色执行提示（仅示意：不包含 QA）
  async runPrompt(projectPath: string, mdcCommand: string): Promise<string> {
    const result = await this.commandService.executeClaudeCommand(projectPath, mdcCommand);
    return result;
  }
}

