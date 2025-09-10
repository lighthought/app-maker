package worker

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"
	"autocodeweb-backend/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
)

// HandleProjectBackupTask 处理项目备份任务
func HandleProjectBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	projectID := payload.ProjectID
	projectPath := payload.ProjectPath

	if _, err := os.Stat(projectPath); err == nil {
		// 创建缓存目录
		cacheDir := utils.GetCachePath()

		// 生成缓存文件名
		cacheFileName := fmt.Sprintf("%s_%s", projectID, time.Now().Format("20060102_150405"))

		// 使用 utils 压缩到缓存
		_, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
		if err != nil {
			logger.Error("异步打包项目到缓存失败",
				logger.String("projectID", projectID),
				logger.ErrorField(err))
		} else {
			logger.Info("项目已异步打包到缓存",
				logger.String("projectID", projectID))
		}
	}
	return nil
}

func HandleProjectDownloadTask(ctx context.Context, t *asynq.Task) error {
	// 1. 解析任务负载
	var payload models.ProjectTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	projectID := payload.ProjectID
	projectPath := payload.ProjectPath

	// 创建缓存目录
	cacheDir := utils.GetCachePath()

	// 生成缓存文件名
	cacheFileName := fmt.Sprintf("%s_%s", projectID, time.Now().Format("20060102_150405"))

	// 使用 utils 压缩到缓存
	_, err := utils.CompressDirectoryToDir(context.Background(), projectPath, cacheDir, cacheFileName)
	if err != nil {
		return fmt.Errorf("打包项目文件失败: %w", err)
	}

	return nil
}

// updateProgress 是一个帮助函数，用于将任务进度更新到Redis。
// 这里假设使用一个Redis Hash结构，key为`task:progress:<task_id>`。
func updateProgress(taskID, status string, progress int, message string) {
	// 你需要一个Redis连接实例
	// rdb := redis.NewClient(...)
	// key := fmt.Sprintf("task:progress:%s", taskID)
	// data := map[string]interface{}{
	//     "status":   status,
	//     "progress": progress,
	//     "message":  message,
	//     "updated_at": time.Now().Unix(),
	// }
	// err := rdb.HSet(context.Background(), key, data).Err()
	// if err != nil {
	//     log.Printf("Failed to update progress for task %s: %v", taskID, err)
	// }
	log.Printf("[Progress] Task %s: %s (%d%%) - %s", taskID, status, progress, message)
}
