package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
	"github.com/lighthought/app-maker/shared-models/utils"

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
	case common.TaskTypeProjectDownload:
		return h.HandleProjectDownloadTask(ctx, task)
	case common.TaskTypeProjectBackup:
		return h.HandleProjectBackupTask(ctx, task)
	default:
		return fmt.Errorf("unexpected task type %s", task.Type())
	}
}

// HandleProjectBackupTask 处理项目备份任务
func (s *ProjectTaskHandler) HandleProjectBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resultWriter := t.ResultWriter()
	logger.Info("handle project backup task", logger.String("taskID", resultWriter.TaskID()))

	resultPath, projectPath, err := s.zipProjectPath(t)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to zip project file: "+err.Error())
		return fmt.Errorf("failed to zip project file: %s, projectID: %s", err.Error(), resultWriter.TaskID())
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "project file zipped to cache")

	// 删除项目目录
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "deleting project directory")
	if err := os.RemoveAll(projectPath); err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to delete project directory: "+err.Error())
		return fmt.Errorf("failed to delete project directory: %s, projectPath: %s", err.Error(), projectPath)
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, resultPath)
	return nil
}

// HandleProjectDownloadTask 处理项目下载任务
func (s *ProjectTaskHandler) HandleProjectDownloadTask(ctx context.Context, t *asynq.Task) error {
	resultWriter := t.ResultWriter()
	logger.Info("handle project download task", logger.String("taskID", resultWriter.TaskID()))

	resultPath, _, err := s.zipProjectPath(t)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to zip project file: "+err.Error())
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, resultPath)
	return nil
}

func (s *ProjectTaskHandler) zipProjectPath(t *asynq.Task) (string, string, error) {
	// 1. 解析任务负载
	var payload tasks.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return "", "", fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	projectID := payload.ProjectID
	projectGuid := payload.ProjectGuid
	projectPath := payload.ProjectPath
	resultWriter := t.ResultWriter()

	// 创建缓存目录
	cacheDir := utils.GetCachePath()
	// 生成缓存文件名
	cacheFileName := fmt.Sprintf("%s_%s", projectGuid, time.Now().Format("20060102_150405"))

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "zipping project file...")
	// 使用 utils 压缩到缓存
	resultPath, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "failed to zip project file: "+err.Error())
		return "", projectPath, fmt.Errorf("failed to zip project file: %s, projectID: %s, projectGuid: %s", err.Error(), projectID, projectGuid)
	}
	return resultPath, projectPath, nil
}
