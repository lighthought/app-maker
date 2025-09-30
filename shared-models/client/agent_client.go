package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"shared-models/agent"
	"shared-models/common"
	"shared-models/logger"
	"shared-models/tasks"
)

// AgentClient Agent 服务客户端
type AgentClient struct {
	httpClient *HTTPClient
}

// NewAgentClient 创建 Agent 客户端
func NewAgentClient(baseURL string, timeout time.Duration) *AgentClient {
	return &AgentClient{
		httpClient: NewHTTPClient(baseURL, timeout),
	}
}

// SetHeader 设置请求头
func (c *AgentClient) SetHeader(key, value string) {
	c.httpClient.SetHeader(key, value)
}

// parseResponseData 安全地解析响应数据到目标结构体
func parseResponseData(resp *common.Response, target interface{}) error {
	// 将 Data 转换为 JSON 字节
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return fmt.Errorf("序列化响应数据失败: %w", err)
	}

	// 将 JSON 字节解析到目标结构体
	if err := json.Unmarshal(dataBytes, target); err != nil {
		return fmt.Errorf("解析响应数据失败: %w", err)
	}

	return nil
}

// HealthCheck 健康检查
func (c *AgentClient) HealthCheck(ctx context.Context) (*agent.AgentHealthResp, error) {
	resp, err := c.httpClient.Get(ctx, "/api/v1/health")
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 健康检查失败: %s", resp.Message)
	}

	result := &agent.AgentHealthResp{}
	if err := parseResponseData(resp, result); err != nil {
		return nil, err
	}

	return result, nil
}

// 等待任务完成或失败
func (c *AgentClient) waitForTaskCompletion(ctx context.Context, taskID string) (*tasks.TaskResult, error) {
	iRetryTimes := 0
	iMaxRetryTimes := 200
	for iRetryTimes < iMaxRetryTimes {
		resp, err := c.httpClient.Get(ctx, "/api/v1/tasks/"+taskID)
		if err != nil {
			logger.Info("获取任务状态失败", logger.String("taskID", taskID), logger.String("error", err.Error()))
			return nil, err
		}
		if resp.Code != common.SUCCESS_CODE {
			return nil, fmt.Errorf("agent 任务执行失败: %s", resp.Message)
		}

		result := &tasks.TaskResult{}
		if err := parseResponseData(resp, result); err != nil {
			logger.Info("解析任务状态失败", logger.String("taskID", taskID), logger.String("error", err.Error()))
			return nil, err
		}

		if result.Status == common.CommonStatusDone {
			logger.Info("任务完成", logger.String("taskID", taskID))
			return result, nil
		}
		if result.Status == common.CommonStatusFailed {
			logger.Info("任务失败", logger.String("taskID", taskID))
			return result, fmt.Errorf("agent 任务执行失败: %s", result.Message)
		}

		time.Sleep(5 * time.Second)
		iRetryTimes++
	}

	return nil, fmt.Errorf("任务超时")
}

// SetupProjectEnvironment 项目环境准备
func (c *AgentClient) SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/project/setup", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent setup project environment failed: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// AnalyseProjectBrief 分析项目简介
func (c *AgentClient) AnalyseProjectBrief(ctx context.Context, req *agent.GetProjBriefReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/analyse/project-brief", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// GetPRD 获取 PRD
func (c *AgentClient) GetPRD(ctx context.Context, req *agent.GetPRDReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/pm/prd", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// GetUXStandard 获取 UX 标准
func (c *AgentClient) GetUXStandard(ctx context.Context, req *agent.GetUXStandardReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/ux-expert/ux-standard", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// GetArchitecture 获取架构设计
func (c *AgentClient) GetArchitecture(ctx context.Context, req *agent.GetArchitectureReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/architect/architect", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// GetDatabaseDesign 获取数据库设计
func (c *AgentClient) GetDatabaseDesign(ctx context.Context, req *agent.GetDatabaseDesignReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/architect/database", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// GetAPIDefinition 获取 API 定义
func (c *AgentClient) GetAPIDefinition(ctx context.Context, req *agent.GetAPIDefinitionReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/architect/apidefinition", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// GetEpicsAndStories 获取史诗和故事
func (c *AgentClient) GetEpicsAndStories(ctx context.Context, req *agent.GetEpicsAndStoriesReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/po/epicsandstories", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// ImplementStory 实现用户故事
func (c *AgentClient) ImplementStory(ctx context.Context, req *agent.ImplementStoryReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/implstory", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// FixBug 修复 Bug
func (c *AgentClient) FixBug(ctx context.Context, req *agent.FixBugReq) (*tasks.TaskResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/fixbug", req)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// RunTest 运行测试
func (c *AgentClient) RunTest(ctx context.Context, req *agent.RunTestReq) (*tasks.TaskResult, error) {
	// 转换为 FixBugReq 格式（临时方案）
	fixBugReq := &agent.FixBugReq{
		ProjectGuid:    req.ProjectGuid,
		BugDescription: "执行项目测试",
	}

	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/runtest", fixBugReq)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}

// Deploy 部署项目
func (c *AgentClient) Deploy(ctx context.Context, req *agent.DeployReq) (*tasks.TaskResult, error) {
	// 转换为 FixBugReq 格式（临时方案）
	fixBugReq := &agent.FixBugReq{
		ProjectGuid:    req.ProjectGuid,
		BugDescription: "打包部署项目",
	}

	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/deploy", fixBugReq)
	if err != nil {
		return nil, err
	}

	if resp.Code != common.SUCCESS_CODE {
		return nil, fmt.Errorf("agent 执行失败: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return c.waitForTaskCompletion(ctx, taskID)
}
