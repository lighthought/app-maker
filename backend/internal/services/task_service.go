package services

import (
	"context"
	"fmt"
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
)

// TaskService 任务服务接口
type TaskService interface {
	// 任务查询
	GetProjectTasks(ctx context.Context, projectID, userID string, limit, offset int) ([]*models.Task, error)
	GetTaskDetails(ctx context.Context, taskID, userID string) (*models.Task, error)
	GetTaskLogs(ctx context.Context, taskID, userID string, limit, offset int) ([]*models.TaskLog, error)
	
	// 任务控制
	CancelTask(ctx context.Context, taskID, userID string) error
}

// taskService 任务服务实现
type taskService struct {
	taskRepo       repositories.TaskRepository
	projectRepo    repositories.ProjectRepository
}

// NewTaskService 创建任务服务实例
func NewTaskService(taskRepo repositories.TaskRepository, projectRepo repositories.ProjectRepository) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
	}
}

// GetProjectTasks 获取项目任务列表
func (s *taskService) GetProjectTasks(ctx context.Context, projectID, userID string, limit, offset int) ([]*models.Task, error) {
	// 验证项目所有权
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	
	if project == nil {
		return nil, fmt.Errorf("项目不存在")
	}
	
	if project.UserID != userID {
		return nil, fmt.Errorf("无权限访问此项目")
	}
	
	// 获取任务列表
	tasks, err := s.taskRepo.List(ctx, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}
	
	return tasks, nil
}

// GetTaskDetails 获取任务详情
func (s *taskService) GetTaskDetails(ctx context.Context, taskID, userID string) (*models.Task, error) {
	// 获取任务详情
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务详情失败: %w", err)
	}
	
	if task == nil {
		return nil, nil
	}
	
	// 验证项目所有权
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	
	if project.UserID != userID {
		return nil, fmt.Errorf("无权限访问此任务")
	}
	
	return task, nil
}

// GetTaskLogs 获取任务日志
func (s *taskService) GetTaskLogs(ctx context.Context, taskID, userID string, limit, offset int) ([]*models.TaskLog, error) {
	// 验证任务访问权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}
	
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}
	
	// 验证项目所有权
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	
	if project.UserID != userID {
		return nil, fmt.Errorf("无权限访问此任务")
	}
	
	// 获取任务日志
	logs, err := s.taskRepo.GetLogs(ctx, taskID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("获取任务日志失败: %w", err)
	}
	
	return logs, nil
}

// CancelTask 取消任务
func (s *taskService) CancelTask(ctx context.Context, taskID, userID string) error {
	// 验证任务访问权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务信息失败: %w", err)
	}
	
	if task == nil {
		return fmt.Errorf("任务不存在")
	}
	
	// 验证项目所有权
	project, err := s.projectRepo.GetByID(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("获取项目信息失败: %w", err)
	}
	
	if project.UserID != userID {
		return fmt.Errorf("无权限操作此任务")
	}
	
	// 检查任务状态
	if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
		return fmt.Errorf("任务已结束，无法取消")
	}
	
	// 更新任务状态为已取消
	err = s.taskRepo.UpdateStatus(ctx, taskID, "cancelled")
	if err != nil {
		return fmt.Errorf("取消任务失败: %w", err)
	}
	
	return nil
}
