package test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"shared-models/agent"
	"shared-models/client"
)

const (
	// 测试项目 GUID
	testProjectGuid = "b0f152887703419e8a39d9718b024f7f"
	// Agent 服务 Base URL
	agentBaseURL = "http://localhost:8088"
	// 测试项目需求
	testRequirements = `
开发一个女性搭配应用+微信小程序：
1. 用户能够上传自己的自拍头部照片、全身正面和反面照片，以此来生成 3D 人物形象；
2. 可以通过淘宝、京东等购物网站分享链接到应用中，应用解析链接时，自动获取到衣服（上衣、裤子、裙装、连衣裙等）的图片，收集到我喜欢的衣服图片中；
3. 用户可为 3D 人物图像，自行从我喜欢的衣服图片中搭配衣服。 
4. 3D 人物，搭配衣服的过程中，支持各种 3D 操作，可以旋转、缩放，查看近身效果、远处查看效果、正面效果、背面效果。
5. 搭配好的效果，支持一键分享给微信好友。
`
)

// TestCompleteProjectDevelopment 测试完整的项目开发流程
func TestCompleteProjectDevelopment(t *testing.T) {
	// 创建 Agent 客户端
	agentClient := client.NewAgentClient(agentBaseURL, 10*time.Minute)

	ctx := context.Background()

	// 测试步骤按照 project_stage_service.go 中的顺序
	testSteps := []struct {
		name     string
		testFunc func(t *testing.T, client *client.AgentClient, ctx context.Context)
	}{
		{"1. 健康检查", testHealthCheck},
		{"2. 项目环境准备", testSetupProjectEnvironment},
		{"3. 分析项目概览", testAnalyseProjectBrief},
		{"4. 生成PRD文档", testGeneratePRD},
		{"5. 定义UX标准", testDefineUXStandards},
		{"6. 设计系统架构", testDesignArchitecture},
		{"7. 定义数据模型", testDefineDataModel},
		{"8. 定义API接口", testDefineAPIDefinition},
		{"9. 划分Epic和Story", testPlanEpicsAndStories},
		{"10. 开发Story功能", testDevelopStories},
		{"11. 修复开发问题", testFixBugs},
		{"12. 执行自动测试", testRunTests},
		{"13. 打包部署项目", testDeployProject},
	}

	for _, step := range testSteps {
		t.Run(step.name, func(t *testing.T) {
			log.Printf("=== 开始执行: %s ===", step.name)
			step.testFunc(t, agentClient, ctx)
			log.Printf("=== 完成执行: %s ===\n", step.name)

			// 每个步骤之间稍作停顿，避免请求过快
			time.Sleep(2 * time.Second)
		})
	}
}

// testHealthCheck 测试健康检查
func testHealthCheck(t *testing.T, client *client.AgentClient, ctx context.Context) {
	resp, err := client.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("健康检查失败: %v", err)
	}

	log.Printf("健康检查成功, 状态: %s, 版本号: %s", resp.Status, resp.Version)
}

// testSetupProjectEnvironment 测试项目环境准备
func testSetupProjectEnvironment(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.SetupProjEnvReq{
		ProjectGuid:     testProjectGuid,
		GitlabRepoUrl:   fmt.Sprintf("http://gitlab.app-maker.localhost/app-maker/%s.git", testProjectGuid),
		SetupBmadMethod: true,
		BmadCliType:     "claude",
	}

	result, err := client.SetupProjectEnvironment(ctx, req)
	if err != nil {
		t.Fatalf("项目环境准备失败: %v", err)
	}

	log.Printf("项目环境准备成功: %s", result.Message)
}

// testAnalyseProjectBrief 测试分析项目概览
func testAnalyseProjectBrief(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetProjBriefReq{
		Requirements: testRequirements,
		ProjectGuid:  testProjectGuid,
	}

	result, err := client.AnalyseProjectBrief(ctx, req)
	if err != nil {
		t.Fatalf("分析项目概览失败: %v", err)
	}

	log.Printf("分析项目概览成功: %s, message:\n%s", result.Status, result.Message)
}

// testGeneratePRD 测试生成PRD文档
func testGeneratePRD(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetPRDReq{
		ProjectGuid:  testProjectGuid,
		Requirements: testRequirements,
	}

	result, err := client.GetPRD(ctx, req)
	if err != nil {
		t.Fatalf("生成PRD文档失败: %v", err)
	}

	log.Printf("生成PRD文档成功: %s, message:\n%s", result.Status, result.Message)
}

// testDefineUXStandards 测试定义UX标准
func testDefineUXStandards(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetUXStandardReq{
		ProjectGuid:  testProjectGuid,
		Requirements: testRequirements,
		PrdPath:      "docs/PRD.md",
	}

	result, err := client.GetUXStandard(ctx, req)
	if err != nil {
		t.Fatalf("定义UX标准失败: %v", err)
	}

	log.Printf("定义UX标准成功: %s, message:\n%s", result.Status, result.Message)
}

// testDesignArchitecture 测试设计系统架构
func testDesignArchitecture(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetArchitectureReq{
		ProjectGuid:             testProjectGuid,
		PrdPath:                 "docs/PRD.md",
		UxSpecPath:              "docs/ux/ux-spec.md",
		TemplateArchDescription: "templates/architecture-template-v2.yaml",
	}

	result, err := client.GetArchitecture(ctx, req)
	if err != nil {
		t.Fatalf("设计系统架构失败: %v", err)
	}

	log.Printf("设计系统架构成功: %s, message:\n%s", result.Status, result.Message)
}

// testDefineDataModel 测试定义数据模型
func testDefineDataModel(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetDatabaseDesignReq{
		ProjectGuid:   testProjectGuid,
		PrdPath:       "docs/PRD.md",
		ArchFolder:    "docs/arch",
		StoriesFolder: "docs/stories",
	}

	result, err := client.GetDatabaseDesign(ctx, req)
	if err != nil {
		t.Fatalf("定义数据模型失败: %v", err)
	}

	log.Printf("定义数据模型成功: %s, message:\n%s", result.Status, result.Message)
}

// testDefineAPIDefinition 测试定义API接口
func testDefineAPIDefinition(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetAPIDefinitionReq{
		ProjectGuid:   testProjectGuid,
		PrdPath:       "docs/PRD.md",
		DbFolder:      "docs/db",
		StoriesFolder: "docs/stories",
	}

	result, err := client.GetAPIDefinition(ctx, req)
	if err != nil {
		t.Fatalf("定义API接口失败: %v", err)
	}

	log.Printf("定义API接口成功: %s, message:\n%s", result.Status, result.Message)
}

// testPlanEpicsAndStories 测试划分Epic和Story
func testPlanEpicsAndStories(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.GetEpicsAndStoriesReq{
		ProjectGuid: testProjectGuid,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch",
	}

	result, err := client.GetEpicsAndStories(ctx, req)
	if err != nil {
		t.Fatalf("划分Epic和Story失败: %v", err)
	}

	log.Printf("划分Epic和Story成功: %s, message:\n%s", result.Status, result.Message)
}

// testDevelopStories 测试开发Story功能
func testDevelopStories(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.ImplementStoryReq{
		ProjectGuid: testProjectGuid,
		PrdPath:     "docs/PRD.md",
		ArchFolder:  "docs/arch",
		DbFolder:    "docs/db",
		ApiFolder:   "docs/api",
		UxSpecPath:  "docs/ux/ux-spec.md",
		EpicFile:    "docs/epics.md",
		StoryFile:   "docs/stories/user-registration.md",
	}

	result, err := client.ImplementStory(ctx, req)
	if err != nil {
		t.Fatalf("开发Story功能失败: %v", err)
	}

	log.Printf("开发Story功能成功: %s, message:\n%s", result.Status, result.Message)
}

// testFixBugs 测试修复开发问题
func testFixBugs(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.FixBugReq{
		ProjectGuid:    testProjectGuid,
		BugDescription: "修复用户注册时邮箱验证的问题，确保邮箱格式正确性验证和重复邮箱检查",
	}

	result, err := client.FixBug(ctx, req)
	if err != nil {
		t.Fatalf("修复开发问题失败: %v", err)
	}

	log.Printf("修复开发问题成功: %s, message:\n%s", result.Status, result.Message)
}

// testRunTests 测试执行自动测试
func testRunTests(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.RunTestReq{
		ProjectGuid: testProjectGuid,
	}

	result, err := client.RunTest(ctx, req)
	if err != nil {
		t.Fatalf("执行自动测试失败: %v", err)
	}

	log.Printf("执行自动测试成功: %s, message:\n%s", result.Status, result.Message)
}

// testDeployProject 测试打包部署项目
func testDeployProject(t *testing.T, client *client.AgentClient, ctx context.Context) {
	req := &agent.DeployReq{
		ProjectGuid: testProjectGuid,
		Environment: "dev",
		DeployOptions: map[string]interface{}{
			"build_frontend": true,
			"build_backend":  true,
			"create_docker":  true,
		},
	}

	result, err := client.Deploy(ctx, req)
	if err != nil {
		t.Fatalf("打包部署项目失败: %v", err)
	}

	log.Printf("打包部署项目成功: %s, message:\n%s", result.Status, result.Message)
}

// TestSingleStep 测试单个步骤 - 用于单独调试某个阶段
func TestSingleStep(t *testing.T) {
	agentClient := client.NewAgentClient(agentBaseURL, 5*time.Minute)
	ctx := context.Background()

	// 修改这里来测试特定的步骤
	t.Run("测试健康检查", func(t *testing.T) {
		testHealthCheck(t, agentClient, ctx)
	})
}

// TestQuickFlow 测试快速流程 - 只测试关键步骤
func TestQuickFlow(t *testing.T) {
	agentClient := client.NewAgentClient(agentBaseURL, 5*time.Minute)
	ctx := context.Background()

	quickSteps := []struct {
		name     string
		testFunc func(t *testing.T, client *client.AgentClient, ctx context.Context)
	}{
		{"健康检查", testHealthCheck},
		{"项目环境准备", testSetupProjectEnvironment},
		{"分析项目概览", testAnalyseProjectBrief},
		{"生成PRD文档", testGeneratePRD},
	}

	for _, step := range quickSteps {
		t.Run(step.name, func(t *testing.T) {
			log.Printf("=== 快速测试: %s ===", step.name)
			step.testFunc(t, agentClient, ctx)
			time.Sleep(1 * time.Second)
		})
	}
}
