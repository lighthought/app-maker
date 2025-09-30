package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"
	"shared-models/utils"
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
	logger.Info("处理项目备份任务", logger.String("taskID", resultWriter.TaskID()))

	resultPath, projectPath, err := s.zipProjectPath(t)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "打包项目文件失败: "+err.Error())
		return fmt.Errorf("打包项目文件失败: %w, projectID: %s", err, resultWriter.TaskID())
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 60, "项目已打包到缓存")

	// 删除项目目录
	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 80, "正在删除项目目录")
	if err := os.RemoveAll(projectPath); err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "删除项目目录失败: "+err.Error())
		return fmt.Errorf("删除项目目录失败: %w, projectPath: %s", err, projectPath)
	}
	tasks.UpdateResult(resultWriter, common.CommonStatusDone, 100, resultPath)
	return nil
}

// HandleProjectDownloadTask 处理项目下载任务
func (s *ProjectTaskHandler) HandleProjectDownloadTask(ctx context.Context, t *asynq.Task) error {
	resultWriter := t.ResultWriter()
	logger.Info("处理项目下载任务", logger.String("taskID", resultWriter.TaskID()))

	resultPath, _, err := s.zipProjectPath(t)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "打包项目文件失败: "+err.Error())
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

	tasks.UpdateResult(resultWriter, common.CommonStatusInProgress, 30, "正在打包项目文件...")
	// 使用 utils 压缩到缓存
	resultPath, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
	if err != nil {
		tasks.UpdateResult(resultWriter, common.CommonStatusFailed, 0, "打包项目文件失败: "+err.Error())
		return "", projectPath, fmt.Errorf("打包项目文件失败: %w, projectID: %s, projectGuid: %s", err, projectID, projectGuid)
	}
	return resultPath, projectPath, nil
}
