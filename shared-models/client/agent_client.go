package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lighthought/app-maker/shared-models/agent"
	"github.com/lighthought/app-maker/shared-models/common"
	"github.com/lighthought/app-maker/shared-models/logger"
	"github.com/lighthought/app-maker/shared-models/tasks"
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

// CheckVersion 检查后台版本
func (c *AgentClient) CheckVersion(ctx context.Context) (*agent.AgentHealthResp, error) {
	resp, err := c.httpClient.Get(ctx, "/api/v1/version")
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

// CheckVersion 检查后台版本
func (c *AgentClient) CheckHealth(ctx context.Context) error {
	resp, err := c.httpClient.Get(ctx, "/api/v1/health")
	if err != nil {
		return err
	}

	if resp.Code != common.SUCCESS_CODE {
		return fmt.Errorf("agent 健康检查失败: %s", resp.Message)
	}

	return nil
}

// 等待任务完成或失败
// 注意：此方法使用独立的 background context，不受 HTTP 请求超时限制
func (c *AgentClient) WaitForTaskCompletion(ctx context.Context, taskID string) (*tasks.TaskResult, error) {
	// 使用 background context 替代传入的 ctx，避免 HTTP 请求超时导致长时间运行的任务被中断
	// 原始的 ctx 仅用于检查是否被主动取消
	bgCtx := context.Background()

	iRetryTimes := 0
	iMaxRetryTimes := 600 // 最多等待约 83 分钟 (600 * 5 秒)

	for iRetryTimes < iMaxRetryTimes {
		// 检查原始 context 是否被取消（允许主动取消任务）
		select {
		case <-ctx.Done():
			logger.Info("task waiting cancelled", logger.String("taskID", taskID))
			return nil, fmt.Errorf("task waiting cancelled: %w", ctx.Err())
		default:
			// 继续执行
		}

		// 使用 background context 进行 HTTP 请求，避免超时
		resp, err := c.httpClient.Get(bgCtx, "/api/v1/tasks/"+taskID)
		if err != nil {
			logger.Info("failed to get task status",
				logger.String("taskID", taskID),
				logger.String("error", err.Error()),
				logger.Int("retryTimes", iRetryTimes))
			return nil, err
		}
		if resp.Code != common.SUCCESS_CODE {
			return nil, fmt.Errorf("agent task execution failed: %s", resp.Message)
		}

		result := &tasks.TaskResult{}
		if err := parseResponseData(resp, result); err != nil {
			logger.Info("failed to parse task status",
				logger.String("taskID", taskID),
				logger.String("error", err.Error()))
			return nil, err
		}

		if result.Status == common.CommonStatusDone {
			logger.Info("task completed",
				logger.String("taskID", taskID),
				logger.Int("totalRetries", iRetryTimes))
			return result, nil
		}
		if result.Status == common.CommonStatusFailed {
			logger.Info("task failed",
				logger.String("taskID", taskID),
				logger.String("message", result.Message))
			return result, fmt.Errorf("agent task execution failed: %s", result.Message)
		}

		// 等待 5 秒后重试
		time.Sleep(5 * time.Second)
		iRetryTimes++

		// 每 10 次重试记录一次日志
		if iRetryTimes%10 == 0 {
			logger.Info("task still executing",
				logger.String("taskID", taskID),
				logger.String("status", result.Status),
				logger.String("message", result.Message),
				logger.Int("retryTimes", iRetryTimes))
		}
	}

	return nil, fmt.Errorf("task timeout: waited %d seconds", iMaxRetryTimes*5)
}

func (c *AgentClient) innerPost(ctx context.Context, url string, req interface{}) (string, error) {
	resp, err := c.httpClient.Post(ctx, url, req)
	if err != nil {
		return "", err
	}

	if resp.Code != common.SUCCESS_CODE {
		return "", fmt.Errorf("agent execution failed: %s", resp.Message)
	}

	taskID := resp.Data.(string)
	return taskID, nil
}

// SetupProjectEnvironment 项目环境准备
func (c *AgentClient) SetupProjectEnvironment(ctx context.Context, req *agent.SetupProjEnvReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/project/setup", req)
}

// AnalyseProjectBrief 分析项目简介
func (c *AgentClient) AnalyseProjectBrief(ctx context.Context, req *agent.GetProjBriefReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/analyse/project-brief", req)
}

// GetPRD 获取 PRD
func (c *AgentClient) GetPRD(ctx context.Context, req *agent.GetPRDReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/pm/prd", req)
}

// GetUXStandard 获取 UX 标准
func (c *AgentClient) GetUXStandard(ctx context.Context, req *agent.GetUXStandardReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/ux-expert/ux-standard", req)
}

// GetArchitecture 获取架构设计
func (c *AgentClient) GetArchitecture(ctx context.Context, req *agent.GetArchitectureReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/architect/architect", req)
}

// GetDatabaseDesign 获取数据库设计
func (c *AgentClient) GetDatabaseDesign(ctx context.Context, req *agent.GetDatabaseDesignReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/architect/database", req)
}

// GetAPIDefinition 获取 API 定义
func (c *AgentClient) GetAPIDefinition(ctx context.Context, req *agent.GetAPIDefinitionReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/architect/apidefinition", req)
}

// GetEpicsAndStories 获取史诗和故事
func (c *AgentClient) GetEpicsAndStories(ctx context.Context, req *agent.GetEpicsAndStoriesReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/po/epicsandstories", req)
}

// ImplementStory 实现用户故事
func (c *AgentClient) ImplementStory(ctx context.Context, req *agent.ImplementStoryReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/dev/implstory", req)
}

// FixBug 修复 Bug
func (c *AgentClient) FixBug(ctx context.Context, req *agent.FixBugReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/dev/fixbug", req)
}

// RunTest 运行测试
func (c *AgentClient) RunTest(ctx context.Context, req *agent.RunTestReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/dev/runtest", req)
}

// Deploy 部署项目
func (c *AgentClient) Deploy(ctx context.Context, req *agent.DeployReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/dev/deploy", req)
}

// ChatWithAgent 与 Agent 对话
func (c *AgentClient) ChatWithAgent(ctx context.Context, req *agent.ChatReq) (string, error) {
	return c.innerPost(ctx, "/api/v1/agent/chat", req)
}
