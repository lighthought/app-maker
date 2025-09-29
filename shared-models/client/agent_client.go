package client

import (
	"context"
	"fmt"
	"time"

	"shared-models/agent"
	"shared-models/common"
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

// AnalyseProjectBrief 分析项目简介
func (c *AgentClient) AnalyseProjectBrief(ctx context.Context, req *agent.GetProjBriefReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/analyse/project-brief", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// GetPRD 获取 PRD
func (c *AgentClient) GetPRD(ctx context.Context, req *agent.GetPRDReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/pm/prd", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// GetUXStandard 获取 UX 标准
func (c *AgentClient) GetUXStandard(ctx context.Context, req *agent.GetUXStandardReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/ux-expert/ux-standard", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// GetArchitecture 获取架构设计
func (c *AgentClient) GetArchitecture(ctx context.Context, req *agent.GetArchitectureReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/architect/architect", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// GetDatabaseDesign 获取数据库设计
func (c *AgentClient) GetDatabaseDesign(ctx context.Context, req *agent.GetDatabaseDesignReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/architect/database", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// GetAPIDefinition 获取 API 定义
func (c *AgentClient) GetAPIDefinition(ctx context.Context, req *agent.GetAPIDefinitionReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/architect/apidefinition", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// GetEpicsAndStories 获取史诗和故事
func (c *AgentClient) GetEpicsAndStories(ctx context.Context, req *agent.GetEpicsAndStoriesReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/po/epicsandstories", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// ImplementStory 实现用户故事
func (c *AgentClient) ImplementStory(ctx context.Context, req *agent.ImplementStoryReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/implstory", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// FixBug 修复 Bug
func (c *AgentClient) FixBug(ctx context.Context, req *agent.FixBugReq) (*common.AgentResult, error) {
	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/fixbug", req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// RunTest 运行测试
func (c *AgentClient) RunTest(ctx context.Context, req *agent.RunTestReq) (*common.AgentResult, error) {
	// 转换为 FixBugReq 格式（临时方案）
	fixBugReq := &agent.FixBugReq{
		ProjectGuid:    req.ProjectGuid,
		BugDescription: "执行项目测试",
	}

	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/runtest", fixBugReq)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}

// Deploy 部署项目
func (c *AgentClient) Deploy(ctx context.Context, req *agent.DeployReq) (*common.AgentResult, error) {
	// 转换为 FixBugReq 格式（临时方案）
	fixBugReq := &agent.FixBugReq{
		ProjectGuid:    req.ProjectGuid,
		BugDescription: "打包部署项目",
	}

	resp, err := c.httpClient.Post(ctx, "/api/v1/agent/dev/deploy", fixBugReq)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("Agent 执行失败: %s", resp.Error)
	}

	return &common.AgentResult{
		Success: true,
		Output:  fmt.Sprintf("%v", resp.Data),
	}, nil
}
