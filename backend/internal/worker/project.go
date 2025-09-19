package worker

import (
	"autocodeweb-backend/internal/constants"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

// ProjectTaskHandler 项目任务处理器
type ProjectTaskHandler struct {
}

func NewProjectTaskWorker() *ProjectTaskHandler {
	return &ProjectTaskHandler{}
}

// ProcessTask 处理项目任务
func (h *ProjectTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case models.TypeProjectDownload:
		return h.HandleProjectDownloadTask(ctx, task)
	case models.TypeProjectBackup:
		return h.HandleProjectBackupTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// HandleProjectBackupTask 处理项目备份任务
func (s *ProjectTaskHandler) HandleProjectBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("处理项目备份任务", logger.String("taskID", resultWriter.TaskID()))

	resultPath, projectPath, err := s.zipProjectPath(t)
	if err != nil {
		utils.UpdateResult(resultWriter, constants.CommandStatusFailed, 0, "打包项目文件失败: "+err.Error())
		return fmt.Errorf("打包项目文件失败: %w, projectID: %s", err, resultWriter.TaskID())
	}
	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 60, "项目已打包到缓存")

	// 删除项目目录
	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 80, "正在删除项目目录")
	if err := os.RemoveAll(projectPath); err != nil {
		utils.UpdateResult(resultWriter, constants.CommandStatusFailed, 0, "删除项目目录失败: "+err.Error())
		return fmt.Errorf("删除项目目录失败: %w, projectPath: %s", err, projectPath)
	}
	utils.UpdateResult(resultWriter, constants.CommandStatusDone, 100, resultPath)
	return nil
}

// HandleProjectDownloadTask 处理项目下载任务
func (s *ProjectTaskHandler) HandleProjectDownloadTask(ctx context.Context, t *asynq.Task) error {
	resultWriter := t.ResultWriter()
	logger.Info("处理项目下载任务", logger.String("taskID", resultWriter.TaskID()))

	resultPath, _, err := s.zipProjectPath(t)
	if err != nil {
		utils.UpdateResult(resultWriter, constants.CommandStatusFailed, 0, "打包项目文件失败: "+err.Error())
	}
	utils.UpdateResult(resultWriter, constants.CommandStatusDone, 100, resultPath)
	return nil
}

func (s *ProjectTaskHandler) zipProjectPath(t *asynq.Task) (string, string, error) {
	// 1. 解析任务负载
	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return "", "", fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	projectID := payload.ProjectID
	projectPath := payload.ProjectPath
	resultWriter := t.ResultWriter()

	// 创建缓存目录
	cacheDir := utils.GetCachePath()
	// 生成缓存文件名
	cacheFileName := fmt.Sprintf("%s_%s", projectID, time.Now().Format("20060102_150405"))

	utils.UpdateResult(resultWriter, constants.CommandStatusInProgress, 30, "正在打包项目文件...")
	// 使用 utils 压缩到缓存
	resultPath, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
	if err != nil {
		utils.UpdateResult(resultWriter, constants.CommandStatusFailed, 0, "打包项目文件失败: "+err.Error())
		return "", projectPath, fmt.Errorf("打包项目文件失败: %w, projectID: %s", err, projectID)
	}
	return resultPath, projectPath, nil
}
