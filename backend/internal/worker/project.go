package worker

import (
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

func NewProjectWorker() *ProjectTaskHandler {
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
	projectID := payload.ProjectID
	projectPath := payload.ProjectPath
	resultWriter := t.ResultWriter()
	logger.Info("处理项目下载任务", logger.String("taskID", resultWriter.TaskID()))

	if projectPath == "" {
		s.updateResult(resultWriter, models.TaskStatusFailed, 0, "项目路径为空")
		return fmt.Errorf("项目路径为空, projectID: %s", projectID)
	}

	var resultPath string

	if _, err := os.Stat(projectPath); err == nil {
		// 创建缓存目录
		cacheDir := utils.GetCachePath()

		// 生成缓存文件名
		cacheFileName := fmt.Sprintf("%s_%s", projectID, time.Now().Format("20060102_150405"))

		// 使用 utils 压缩到缓存
		s.updateResult(resultWriter, models.TaskStatusInProgress, 30, "正在打包项目文件...")
		_resultPath, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
		if err != nil {
			s.updateResult(resultWriter, models.TaskStatusFailed, 60, "打包项目文件失败: "+err.Error())
			return fmt.Errorf("打包项目文件失败: %w, projectID: %s", err, projectID)
		}
		resultPath = _resultPath
		s.updateResult(resultWriter, models.TaskStatusInProgress, 60, "项目已打包到缓存")
	}

	// 删除项目目录
	s.updateResult(resultWriter, models.TaskStatusInProgress, 80, "正在删除项目目录")
	if err := os.RemoveAll(projectPath); err != nil {
		s.updateResult(resultWriter, models.TaskStatusFailed, 90, "删除项目目录失败: "+err.Error())
		return fmt.Errorf("删除项目目录失败: %w, projectID: %s", err, projectID)
	}
	s.updateResult(resultWriter, models.TaskStatusDone, 100, resultPath)
	return nil
}

// HandleProjectDownloadTask 处理项目下载任务
func (s *ProjectTaskHandler) HandleProjectDownloadTask(ctx context.Context, t *asynq.Task) error {
	// 1. 解析任务负载
	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	projectID := payload.ProjectID
	projectPath := payload.ProjectPath

	resultWriter := t.ResultWriter()
	logger.Info("处理项目下载任务", logger.String("taskID", resultWriter.TaskID()))

	// 创建缓存目录
	cacheDir := utils.GetCachePath()
	// 生成缓存文件名
	cacheFileName := fmt.Sprintf("%s_%s", projectID, time.Now().Format("20060102_150405"))

	s.updateResult(resultWriter, models.TaskStatusInProgress, 30, "正在打包项目文件...")
	// 使用 utils 压缩到缓存
	resultPath, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
	if err != nil {
		s.updateResult(resultWriter, models.TaskStatusFailed, 0, "打包项目文件失败: "+err.Error())
		return fmt.Errorf("打包项目文件失败: %w, projectID: %s", err, projectID)
	}

	s.updateResult(resultWriter, models.TaskStatusDone, 100, resultPath)
	return nil
}

// updateResult 是一个帮助函数，用于将任务进度更新到Redis。
// 这里假设使用一个Redis Hash结构，key为`task:progress:<task_id>`。
func (h *ProjectTaskHandler) updateResult(resultWriter *asynq.ResultWriter, status string, progress int, message string) {
	if resultWriter == nil {
		logger.Error("resultWriter is nil, can't update result")
		return
	}

	data := models.TaskResult{
		TaskID:    resultWriter.TaskID(),
		Status:    status,
		Progress:  progress,
		Message:   message,
		UpdatedAt: utils.GetCurrentTime(),
	}
	resultWriter.Write(data.ToBytes())
}
