package services

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	
	"autocodeweb-backend/internal/models"
)

// ProjectDevService 项目开发环境服务
type ProjectDevService struct {
	baseProjectsDir string
}

// NewProjectDevService 创建项目开发环境服务
func NewProjectDevService(baseProjectsDir string) *ProjectDevService {
	return &ProjectDevService{
		baseProjectsDir: baseProjectsDir,
	}
}

// SetupProjectDevEnvironment 设置项目开发环境
func (s *ProjectDevService) SetupProjectDevEnvironment(project *models.Project) error {
	projectDir := project.ProjectPath
	
	// 检查项目目录是否存在
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("项目目录不存在: %s", projectDir)
	}
	
	// 创建日志文件
	logFile, err := os.OpenFile("/app/logs/task.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %v", err)
	}
	defer logFile.Close()
	
	// 运行开发环境设置脚本
	scriptPath := filepath.Join("/app/scripts", "project-dev-setup.sh")
	cmd := exec.Command("bash", scriptPath, projectDir, project.ID)
	cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
	cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
	
	// 记录开始时间
	startTime := time.Now()
	logFile.WriteString(fmt.Sprintf("[%s] 开始设置项目开发环境: %s\n", startTime.Format("2006-01-02 15:04:05"), project.ID))
	
	if err := cmd.Run(); err != nil {
		endTime := time.Now()
		errorMsg := fmt.Sprintf("[%s] 设置开发环境失败: %v\n", endTime.Format("2006-01-02 15:04:05"), err)
		logFile.WriteString(errorMsg)
		return fmt.Errorf("设置开发环境失败: %v", err)
	}
	
	// 记录完成时间
	endTime := time.Now()
	logFile.WriteString(fmt.Sprintf("[%s] 项目开发环境设置完成: %s (耗时: %v)\n", 
		endTime.Format("2006-01-02 15:04:05"), project.ID, endTime.Sub(startTime)))
	
	return nil
}

// InstallBmadMethod 安装 bmad-method
func (s *ProjectDevService) InstallBmadMethod(projectDir string) error {
	// 检查是否已安装
	if s.isBmadMethodInstalled(projectDir) {
		return nil
	}
	
	// 进入项目目录
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(projectDir); err != nil {
		return err
	}
	
	// 初始化 package.json
	cmd := exec.Command("npm", "init", "-y")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("初始化 package.json 失败: %v", err)
	}
	
	// 安装 bmad-method
	cmd = exec.Command("npm", "install", "bmad-method")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 bmad-method 失败: %v", err)
	}
	
	return nil
}

// InstallCursorCLI 安装 cursor-cli
func (s *ProjectDevService) InstallCursorCLI() error {
	// 检查是否已安装
	if s.isCursorCLIInstalled() {
		return nil
	}
	
	// 全局安装 cursor-cli
	cmd := exec.Command("npm", "install", "-g", "@cursor/cli")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 cursor-cli 失败: %v", err)
	}
	
	return nil
}

// StartCursorChat 启动 Cursor CLI 聊天
func (s *ProjectDevService) StartCursorChat(projectDir string) error {
	if !s.isCursorCLIInstalled() {
		return fmt.Errorf("cursor-cli 未安装")
	}
	
	// 启动 cursor chat
	cmd := exec.Command("cursor", "chat", "--project", projectDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

// ExecuteCommand 在项目目录中执行命令
func (s *ProjectDevService) ExecuteCommand(projectDir, command string, args ...string) error {
	// 检查项目目录是否存在
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("项目目录不存在: %s", projectDir)
	}
	
	// 进入项目目录
	originalDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer os.Chdir(originalDir)
	
	if err := os.Chdir(projectDir); err != nil {
		return err
	}
	
	// 执行命令
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}

// GetProjectDevStatus 获取项目开发环境状态
func (s *ProjectDevService) GetProjectDevStatus(projectDir string) map[string]interface{} {
	status := map[string]interface{}{
		"projectDir": projectDir,
		"bmadMethod": map[string]interface{}{
			"installed": s.isBmadMethodInstalled(projectDir),
		},
		"cursorCLI": map[string]interface{}{
			"installed": s.isCursorCLIInstalled(),
		},
		"node": map[string]interface{}{
			"installed": s.isNodeInstalled(),
		},
		"npm": map[string]interface{}{
			"installed": s.isNpmInstalled(),
		},
	}
	
	// 获取版本信息
	if s.isNodeInstalled() {
		if version, err := s.getNodeVersion(); err == nil {
			status["node"].(map[string]interface{})["version"] = version
		}
	}
	
	if s.isNpmInstalled() {
		if version, err := s.getNpmVersion(); err == nil {
			status["npm"].(map[string]interface{})["version"] = version
		}
	}
	
	return status
}

// 检查工具是否已安装
func (s *ProjectDevService) isBmadMethodInstalled(projectDir string) bool {
	packageJSON := filepath.Join(projectDir, "package.json")
	nodeModules := filepath.Join(projectDir, "node_modules", "bmad-method")
	
	_, err1 := os.Stat(packageJSON)
	_, err2 := os.Stat(nodeModules)
	
	return err1 == nil && err2 == nil
}

func (s *ProjectDevService) isCursorCLIInstalled() bool {
	_, err := exec.LookPath("cursor")
	return err == nil
}

func (s *ProjectDevService) isNodeInstalled() bool {
	_, err := exec.LookPath("node")
	return err == nil
}

func (s *ProjectDevService) isNpmInstalled() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

// 获取版本信息
func (s *ProjectDevService) getNodeVersion() (string, error) {
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (s *ProjectDevService) getNpmVersion() (string, error) {
	cmd := exec.Command("npm", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
