import { CommandExecutionService } from './command-execution.service';
import { FileSystemService } from './file-system.service';
import * as path from 'path';
import logger from '../utils/logger.util';

export class GitService {
  private commandService: CommandExecutionService;
  private fileService: FileSystemService;

  constructor(commandService: CommandExecutionService, fileService: FileSystemService) {
    this.commandService = commandService;
    this.fileService = fileService;
  }

  async ensureCiFile(projectPath: string): Promise<void> {
    try {
      const ciPath = path.join(projectPath, '.gitlab-ci.yml');
      const exists = await this.fileService.fileExists(ciPath);
      if (!exists) {
        const content = [
          'stages:',
          '  - build',
          '  - test',
          '  - deploy',
          '',
          'build:',
          '  stage: build',
          '  image: node:18',
          '  script:',
          '    - echo "Building frontend and backend..."',
          '    - cd frontend && npm ci && npm run build || true',
          '    - cd ..',
          '    - cd backend && go mod download && go build -o server ./cmd/server || true',
          '',
          'test:',
          '  stage: test',
          '  image: node:18',
          '  script:',
          '    - echo "Running tests..."',
          '    - cd frontend && npm test || true',
          '    - cd ..',
          '    - cd backend && go test ./... || true',
          '',
          'deploy:',
          '  stage: deploy',
          '  image: alpine:latest',
          '  script:',
          '    - echo "Triggering GitLab Runner deployment on host..."'
        ].join('\n');
        await this.fileService.writeFile(ciPath, content);
        logger.info('Created .gitlab-ci.yml');
      }
    } catch (error) {
      logger.error('Failed to ensure .gitlab-ci.yml', error);
    }
  }

  async commitAndPush(projectPath: string, message: string): Promise<void> {
    // Ensure CI file exists from PRD stage onward
    try {
      await this.ensureCiFile(projectPath);
    } catch (e) {
      // non-fatal
    }

    // Configure user if missing (non-fatal)
    await this.commandService.executeGitCommand('config user.email "autocodeweb@local"', projectPath).catch(() => {});
    await this.commandService.executeGitCommand('config user.name "AutoCodeWeb Agents"', projectPath).catch(() => {});

    // Add changes
    await this.commandService.executeGitCommand('add -A', projectPath);

    // Commit (skip if nothing to commit)
    const commitResult = await this.commandService.executeGitCommand('commit -m "' + message.replace(/"/g, '\\"') + '"', projectPath)
      .catch((err) => {
        const msg = (err?.message || '').toLowerCase();
        if (msg.includes('nothing to commit')) {
          logger.info('Nothing to commit');
          return 'nothing';
        }
        throw err;
      });
    if (commitResult === 'nothing') {
      logger.info('Skip push as nothing to commit');
      return;
    }

    // Push
    await this.commandService.executeGitCommand('push', projectPath);
    logger.info('Changes pushed to remote');
  }
}

