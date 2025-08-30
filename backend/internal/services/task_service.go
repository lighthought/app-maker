package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/repositories"
)

type TaskService interface {
	// 基础CRUD操作
	CreateTask(ctx context.Context, req *models.CreateTaskRequest, userID string) (*models.Task, error)
	GetTask(ctx context.Context, taskID, userID string) (*models.Task, error)
	UpdateTask(ctx context.Context, taskID string, req *models.UpdateTaskRequest, userID string) (*models.Task, error)
	DeleteTask(ctx context.Context, taskID, userID string) error
	ListTasks(ctx context.Context, req *models.TaskListRequest, userID string) ([]*models.Task, int64, error)

	// 状态管理
	UpdateTaskStatus(ctx context.Context, taskID string, req *models.TaskStatusUpdateRequest, userID string) error
	StartTask(ctx context.Context, taskID, userID string) error
	CompleteTask(ctx context.Context, taskID, userID string, result string) error
	FailTask(ctx context.Context, taskID, userID string, errorMessage string) error
	CancelTask(ctx context.Context, taskID, userID string) error

	// 依赖关系管理
	AddTaskDependency(ctx context.Context, taskID, dependencyID, userID string) error
	RemoveTaskDependency(ctx context.Context, taskID, dependencyID, userID string) error
	GetTaskDependencies(ctx context.Context, taskID, userID string) ([]*models.Task, error)
	GetTaskDependents(ctx context.Context, taskID, userID string) ([]*models.Task, error)

	// 重试和回滚
	RetryTask(ctx context.Context, taskID string, req *models.TaskRetryRequest, userID string) error
	RollbackTask(ctx context.Context, taskID string, req *models.TaskRollbackRequest, userID string) error
	GetFailedTasks(ctx context.Context, userID string) ([]*models.Task, error)

	// 任务日志
	CreateTaskLog(ctx context.Context, req *models.CreateTaskLogRequest, userID string) error
	GetTaskLogs(ctx context.Context, taskID, userID string) ([]*models.TaskLog, error)
	GetTaskLogsByLevel(ctx context.Context, taskID, level, userID string) ([]*models.TaskLog, error)

	// 查询和统计
	GetTasksByProject(ctx context.Context, projectID, userID string) ([]*models.Task, error)
	GetTasksByStatus(ctx context.Context, status string, userID string) ([]*models.Task, error)
	GetTasksByPriority(ctx context.Context, priority int, userID string) ([]*models.Task, error)
	GetOverdueTasks(ctx context.Context, userID string) ([]*models.Task, error)
	GetReadyTasks(ctx context.Context) ([]*models.Task, error)
}

type taskService struct {
	taskRepo    repositories.TaskRepository
	taskLogRepo repositories.TaskLogRepository
	projectRepo repositories.ProjectRepository
}

func NewTaskService(taskRepo repositories.TaskRepository, taskLogRepo repositories.TaskLogRepository, projectRepo repositories.ProjectRepository) TaskService {
	return &taskService{
		taskRepo:    taskRepo,
		taskLogRepo: taskLogRepo,
		projectRepo: projectRepo,
	}
}

// CreateTask 创建任务
func (s *taskService) CreateTask(ctx context.Context, req *models.CreateTaskRequest, userID string) (*models.Task, error) {
	// 验证项目是否存在且用户有权限
	project, err := s.projectRepo.GetByID(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	if project.UserID != userID {
		return nil, errors.New("access denied")
	}

	// 设置默认值
	if req.Priority == 0 {
		req.Priority = int(models.TaskPriorityNormal)
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}
	if req.RetryDelay == 0 {
		req.RetryDelay = 60
	}

	// 创建任务
	task := &models.Task{
		ProjectID:    req.ProjectID,
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		Status:       models.TaskStatusPending,
		Priority:     models.TaskPriority(req.Priority),
		Dependencies: req.Dependencies,
		MaxRetries:   req.MaxRetries,
		RetryCount:   0,
		RetryDelay:   req.RetryDelay,
		Deadline:     req.Deadline,
		Metadata:     req.Metadata,
		Tags:         req.Tags,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	// 创建初始日志
	log := &models.TaskLog{
		TaskID:    task.ID,
		Level:     "info",
		Message:   "Task created",
		Data:      fmt.Sprintf(`{"name":"%s","priority":%d}`, task.Name, task.Priority),
		CreatedAt: time.Now(),
	}
	s.taskLogRepo.Create(ctx, log)

	return task, nil
}

// GetTask 获取任务
func (s *taskService) GetTask(ctx context.Context, taskID, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	return task, nil
}

// UpdateTask 更新任务
func (s *taskService) UpdateTask(ctx context.Context, taskID string, req *models.UpdateTaskRequest, userID string) (*models.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	// 检查任务状态是否允许更新
	if task.Status == models.TaskStatusCompleted || task.Status == models.TaskStatusCancelled || task.Status == models.TaskStatusRolledBack {
		return nil, errors.New("cannot update completed/cancelled/rolled back task")
	}

	// 更新字段
	if req.Name != "" {
		task.Name = req.Name
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Priority > 0 {
		task.Priority = models.TaskPriority(req.Priority)
	}
	if req.Dependencies != nil {
		task.Dependencies = req.Dependencies
	}
	if req.MaxRetries > 0 {
		task.MaxRetries = req.MaxRetries
	}
	if req.RetryDelay > 0 {
		task.RetryDelay = req.RetryDelay
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}
	if req.Metadata != "" {
		task.Metadata = req.Metadata
	}
	if req.Tags != nil {
		task.Tags = req.Tags
	}

	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// 创建更新日志
	log := &models.TaskLog{
		TaskID:    task.ID,
		Level:     "info",
		Message:   "Task updated",
		Data:      fmt.Sprintf(`{"name":"%s","priority":%d}`, task.Name, task.Priority),
		CreatedAt: time.Now(),
	}
	s.taskLogRepo.Create(ctx, log)

	return task, nil
}

// DeleteTask 删除任务
func (s *taskService) DeleteTask(ctx context.Context, taskID, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// 检查任务状态
	if task.Status == models.TaskStatusRunning {
		return errors.New("cannot delete running task")
	}

	// 删除任务日志
	s.taskLogRepo.DeleteByTaskID(ctx, taskID)

	// 删除任务
	return s.taskRepo.Delete(ctx, taskID)
}

// ListTasks 获取任务列表
func (s *taskService) ListTasks(ctx context.Context, req *models.TaskListRequest, userID string) ([]*models.Task, int64, error) {
	// 如果未指定用户ID，使用当前用户ID
	if req.UserID == "" {
		req.UserID = userID
	}

	// 检查权限（只能查看自己的任务）
	if req.UserID != userID {
		return nil, 0, errors.New("access denied")
	}

	return s.taskRepo.List(ctx, req)
}

// UpdateTaskStatus 更新任务状态
func (s *taskService) UpdateTaskStatus(ctx context.Context, taskID string, req *models.TaskStatusUpdateRequest, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// 验证状态转换是否有效
	if !s.isValidStatusTransition(task.Status, models.TaskStatus(req.Status)) {
		return errors.New("invalid status transition")
	}

	// 更新状态
	if err := s.taskRepo.UpdateStatus(ctx, taskID, models.TaskStatus(req.Status), req.ErrorMessage, req.Result); err != nil {
		return err
	}

	// 创建状态变更日志
	log := &models.TaskLog{
		TaskID:    taskID,
		Level:     "info",
		Message:   fmt.Sprintf("Task status changed to %s", req.Status),
		Data:      fmt.Sprintf(`{"old_status":"%s","new_status":"%s","error_message":"%s"}`, task.Status, req.Status, req.ErrorMessage),
		CreatedAt: time.Now(),
	}
	s.taskLogRepo.Create(ctx, log)

	return nil
}

// StartTask 启动任务
func (s *taskService) StartTask(ctx context.Context, taskID, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// 检查任务状态
	if task.Status != models.TaskStatusPending {
		return errors.New("task is not in pending status")
	}

	// 检查依赖是否完成
	completed, err := s.taskRepo.CheckDependenciesCompleted(ctx, taskID)
	if err != nil {
		return err
	}
	if !completed {
		return errors.New("task dependencies not completed")
	}

	// 更新状态为运行中
	return s.UpdateTaskStatus(ctx, taskID, &models.TaskStatusUpdateRequest{
		Status: string(models.TaskStatusRunning),
	}, userID)
}

// CompleteTask 完成任务
func (s *taskService) CompleteTask(ctx context.Context, taskID, userID string, result string) error {
	return s.UpdateTaskStatus(ctx, taskID, &models.TaskStatusUpdateRequest{
		Status: string(models.TaskStatusCompleted),
		Result: result,
	}, userID)
}

// FailTask 任务失败
func (s *taskService) FailTask(ctx context.Context, taskID, userID string, errorMessage string) error {
	return s.UpdateTaskStatus(ctx, taskID, &models.TaskStatusUpdateRequest{
		Status:       string(models.TaskStatusFailed),
		ErrorMessage: errorMessage,
	}, userID)
}

// CancelTask 取消任务
func (s *taskService) CancelTask(ctx context.Context, taskID, userID string) error {
	return s.UpdateTaskStatus(ctx, taskID, &models.TaskStatusUpdateRequest{
		Status: string(models.TaskStatusCancelled),
	}, userID)
}

// AddTaskDependency 添加任务依赖
func (s *taskService) AddTaskDependency(ctx context.Context, taskID, dependencyID, userID string) error {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// 检查依赖任务权限
	depTask, err := s.taskRepo.GetByID(ctx, dependencyID)
	if err != nil {
		return err
	}
	if depTask == nil {
		return errors.New("dependency task not found")
	}
	if depTask.UserID != userID {
		return errors.New("access denied to dependency task")
	}

	// 检查循环依赖
	if taskID == dependencyID {
		return errors.New("cannot add self as dependency")
	}

	// 检查是否会导致循环依赖
	if s.wouldCreateCycle(ctx, taskID, dependencyID) {
		return errors.New("adding this dependency would create a cycle")
	}

	return s.taskRepo.AddDependency(ctx, taskID, dependencyID)
}

// RemoveTaskDependency 移除任务依赖
func (s *taskService) RemoveTaskDependency(ctx context.Context, taskID, dependencyID, userID string) error {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	if task.UserID != userID {
		return errors.New("access denied")
	}

	return s.taskRepo.RemoveDependency(ctx, taskID, dependencyID)
}

// GetTaskDependencies 获取任务依赖
func (s *taskService) GetTaskDependencies(ctx context.Context, taskID, userID string) ([]*models.Task, error) {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.taskRepo.GetDependencies(ctx, taskID)
}

// GetTaskDependents 获取依赖此任务的任务
func (s *taskService) GetTaskDependents(ctx context.Context, taskID, userID string) ([]*models.Task, error) {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.taskRepo.GetDependents(ctx, taskID)
}

// RetryTask 重试任务
func (s *taskService) RetryTask(ctx context.Context, taskID string, req *models.TaskRetryRequest, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// 检查任务状态
	if task.Status != models.TaskStatusFailed {
		return errors.New("task is not in failed status")
	}

	// 检查重试次数
	if task.RetryCount >= task.MaxRetries && !req.Force {
		return errors.New("max retry count exceeded")
	}

	// 如果不是强制重试，检查依赖
	if !req.Force {
		completed, err := s.taskRepo.CheckDependenciesCompleted(ctx, taskID)
		if err != nil {
			return err
		}
		if !completed {
			return errors.New("task dependencies not completed")
		}
	}

	// 增加重试次数
	if err := s.taskRepo.IncrementRetryCount(ctx, taskID); err != nil {
		return err
	}

	// 更新状态为重试中
	return s.UpdateTaskStatus(ctx, taskID, &models.TaskStatusUpdateRequest{
		Status: string(models.TaskStatusRetrying),
	}, userID)
}

// RollbackTask 回滚任务
func (s *taskService) RollbackTask(ctx context.Context, taskID string, req *models.TaskRollbackRequest, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}

	// 检查权限
	if task.UserID != userID {
		return errors.New("access denied")
	}

	// 检查任务状态
	if task.Status != models.TaskStatusCompleted {
		return errors.New("task is not in completed status")
	}

	// 更新状态为已回滚
	if err := s.UpdateTaskStatus(ctx, taskID, &models.TaskStatusUpdateRequest{
		Status: string(models.TaskStatusRolledBack),
		Result: req.Reason,
	}, userID); err != nil {
		return err
	}

	// 创建回滚日志
	log := &models.TaskLog{
		TaskID:    taskID,
		Level:     "warn",
		Message:   "Task rolled back",
		Data:      fmt.Sprintf(`{"reason":"%s"}`, req.Reason),
		CreatedAt: time.Now(),
	}
	s.taskLogRepo.Create(ctx, log)

	return nil
}

// GetFailedTasks 获取失败的任务
func (s *taskService) GetFailedTasks(ctx context.Context, userID string) ([]*models.Task, error) {
	tasks, err := s.taskRepo.GetFailedTasks(ctx)
	if err != nil {
		return nil, err
	}

	// 过滤用户权限
	var userTasks []*models.Task
	for _, task := range tasks {
		if task.UserID == userID {
			userTasks = append(userTasks, task)
		}
	}

	return userTasks, nil
}

// CreateTaskLog 创建任务日志
func (s *taskService) CreateTaskLog(ctx context.Context, req *models.CreateTaskLogRequest, userID string) error {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, req.TaskID)
	if err != nil {
		return err
	}
	if task == nil {
		return errors.New("task not found")
	}
	if task.UserID != userID {
		return errors.New("access denied")
	}

	log := &models.TaskLog{
		TaskID:    req.TaskID,
		Level:     req.Level,
		Message:   req.Message,
		Data:      req.Data,
		CreatedAt: time.Now(),
	}

	return s.taskLogRepo.Create(ctx, log)
}

// GetTaskLogs 获取任务日志
func (s *taskService) GetTaskLogs(ctx context.Context, taskID, userID string) ([]*models.TaskLog, error) {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.taskLogRepo.GetByTaskID(ctx, taskID)
}

// GetTaskLogsByLevel 根据级别获取任务日志
func (s *taskService) GetTaskLogsByLevel(ctx context.Context, taskID, level, userID string) ([]*models.TaskLog, error) {
	// 检查任务权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.taskLogRepo.GetByTaskIDAndLevel(ctx, taskID, level)
}

// GetTasksByProject 根据项目获取任务
func (s *taskService) GetTasksByProject(ctx context.Context, projectID, userID string) ([]*models.Task, error) {
	// 检查项目权限
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, errors.New("project not found")
	}
	if project.UserID != userID {
		return nil, errors.New("access denied")
	}

	return s.taskRepo.GetByProjectID(ctx, projectID)
}

// GetTasksByStatus 根据状态获取任务
func (s *taskService) GetTasksByStatus(ctx context.Context, status string, userID string) ([]*models.Task, error) {
	tasks, err := s.taskRepo.GetByStatus(ctx, models.TaskStatus(status))
	if err != nil {
		return nil, err
	}

	// 过滤用户权限
	var userTasks []*models.Task
	for _, task := range tasks {
		if task.UserID == userID {
			userTasks = append(userTasks, task)
		}
	}

	return userTasks, nil
}

// GetTasksByPriority 根据优先级获取任务
func (s *taskService) GetTasksByPriority(ctx context.Context, priority int, userID string) ([]*models.Task, error) {
	tasks, err := s.taskRepo.GetByPriority(ctx, models.TaskPriority(priority))
	if err != nil {
		return nil, err
	}

	// 过滤用户权限
	var userTasks []*models.Task
	for _, task := range tasks {
		if task.UserID == userID {
			userTasks = append(userTasks, task)
		}
	}

	return userTasks, nil
}

// GetOverdueTasks 获取超期任务
func (s *taskService) GetOverdueTasks(ctx context.Context, userID string) ([]*models.Task, error) {
	tasks, err := s.taskRepo.GetOverdueTasks(ctx)
	if err != nil {
		return nil, err
	}

	// 过滤用户权限
	var userTasks []*models.Task
	for _, task := range tasks {
		if task.UserID == userID {
			userTasks = append(userTasks, task)
		}
	}

	return userTasks, nil
}

// GetReadyTasks 获取可执行的任务
func (s *taskService) GetReadyTasks(ctx context.Context) ([]*models.Task, error) {
	return s.taskRepo.GetReadyTasks(ctx)
}

// isValidStatusTransition 检查状态转换是否有效
func (s *taskService) isValidStatusTransition(from, to models.TaskStatus) bool {
	validTransitions := map[models.TaskStatus][]models.TaskStatus{
		models.TaskStatusPending: {
			models.TaskStatusRunning,
			models.TaskStatusCancelled,
		},
		models.TaskStatusRunning: {
			models.TaskStatusCompleted,
			models.TaskStatusFailed,
			models.TaskStatusCancelled,
		},
		models.TaskStatusRetrying: {
			models.TaskStatusRunning,
			models.TaskStatusFailed,
			models.TaskStatusCancelled,
		},
		models.TaskStatusCompleted: {
			models.TaskStatusRolledBack,
		},
		models.TaskStatusFailed: {
			models.TaskStatusRetrying,
			models.TaskStatusCancelled,
		},
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowed {
		if allowedStatus == to {
			return true
		}
	}

	return false
}

// wouldCreateCycle 检查是否会导致循环依赖
func (s *taskService) wouldCreateCycle(ctx context.Context, taskID, dependencyID string) bool {
	// 简单的循环检测：检查依赖任务是否直接或间接依赖当前任务
	visited := make(map[string]bool)
	return s.hasPath(ctx, dependencyID, taskID, visited)
}

// hasPath 检查是否存在从start到target的路径
func (s *taskService) hasPath(ctx context.Context, start, target string, visited map[string]bool) bool {
	if start == target {
		return true
	}

	if visited[start] {
		return false
	}

	visited[start] = true

	dependents, err := s.taskRepo.GetDependents(ctx, start)
	if err != nil {
		return false
	}

	for _, dep := range dependents {
		if s.hasPath(ctx, dep.ID, target, visited) {
			return true
		}
	}

	return false
}
