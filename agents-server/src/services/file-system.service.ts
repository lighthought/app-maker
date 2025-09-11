import * as fs from 'fs-extra';
import * as path from 'path';
import { FileOperationResult, DocumentFormat } from '../models/artifact.model';
import logger from '../utils/logger.util';

export class FileSystemService {
  private projectDataPath: string;

  constructor(projectDataPath: string) {
    this.projectDataPath = projectDataPath;
  }

  async readFile(filePath: string): Promise<string> {
    try {
      const fullPath = path.resolve(filePath);
      const content = await fs.readFile(fullPath, 'utf-8');
      logger.info(`File read successfully: ${fullPath}`);
      return content;
    } catch (error) {
      logger.error(`Failed to read file: ${filePath}`, error);
      throw error;
    }
  }

  async writeFile(filePath: string, content: string): Promise<FileOperationResult> {
    try {
      const fullPath = path.resolve(filePath);
      await fs.ensureDir(path.dirname(fullPath));
      await fs.writeFile(fullPath, content, 'utf-8');
      logger.info(`File written successfully: ${fullPath}`);
      return { success: true, path: fullPath };
    } catch (error) {
      logger.error(`Failed to write file: ${filePath}`, error);
      return { success: false, path: filePath, error: (error as Error).message };
    }
  }

  async createDirectory(dirPath: string): Promise<FileOperationResult> {
    try {
      const fullPath = path.resolve(dirPath);
      await fs.ensureDir(fullPath);
      logger.info(`Directory created successfully: ${fullPath}`);
      return { success: true, path: fullPath };
    } catch (error) {
      logger.error(`Failed to create directory: ${dirPath}`, error);
      return { success: false, path: dirPath, error: (error as Error).message };
    }
  }

  async copyTemplate(templatePath: string, targetPath: string): Promise<FileOperationResult> {
    try {
      const sourcePath = path.resolve(templatePath);
      const destPath = path.resolve(targetPath);
      
      await fs.ensureDir(path.dirname(destPath));
      await fs.copy(sourcePath, destPath);
      
      logger.info(`Template copied successfully: ${sourcePath} -> ${destPath}`);
      return { success: true, path: destPath };
    } catch (error) {
      logger.error(`Failed to copy template: ${templatePath} -> ${targetPath}`, error);
      return { success: false, path: targetPath, error: (error as Error).message };
    }
  }

  async replacePlaceholders(filePath: string, variables: Record<string, string>): Promise<FileOperationResult> {
    try {
      const content = await this.readFile(filePath);
      let newContent = content;

      // 替换占位符 ${VARIABLE_NAME}
      for (const [key, value] of Object.entries(variables)) {
        const placeholder = `\${${key.toUpperCase()}}`;
        newContent = newContent.replace(new RegExp(placeholder, 'g'), value);
      }

      await this.writeFile(filePath, newContent);
      logger.info(`Placeholders replaced successfully: ${filePath}`);
      return { success: true, path: filePath };
    } catch (error) {
      logger.error(`Failed to replace placeholders: ${filePath}`, error);
      return { success: false, path: filePath, error: (error as Error).message };
    }
  }

  async fileExists(filePath: string): Promise<boolean> {
    try {
      return await fs.pathExists(filePath);
    } catch (error) {
      logger.error(`Failed to check file existence: ${filePath}`, error);
      return false;
    }
  }

  async getProjectPath(projectId: string): Promise<string> {
    // 这个方法现在主要用于验证，实际路径由调用方提供
    return path.join(this.projectDataPath, projectId);
  }

  async ensureProjectDirectory(projectId: string): Promise<string> {
    // 项目目录应该已经存在，这里只做验证
    const projectPath = await this.getProjectPath(projectId);
    if (!await this.fileExists(projectPath)) {
      throw new Error(`Project directory does not exist: ${projectPath}`);
    }
    return projectPath;
  }
}
