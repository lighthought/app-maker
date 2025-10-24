package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lighthought/app-maker/shared-models/logger"
)

// CompressDirectory 压缩指定目录到zip文件
// sourceDir: 要压缩的源目录路径
// outputPath: 输出的zip文件路径
// workingDir: 工作目录（可选，默认为sourceDir）
func CompressDirectory(ctx context.Context, sourceDir, outputPath string, workingDir ...string) error {
	logger.Info("开始压缩目录", logger.String("sourceDir", sourceDir), logger.String("outputPath", outputPath))

	// 检查源目录是否存在
	if !IsDirectoryExists(sourceDir) {
		logger.Error("源目录不存在", logger.String("sourceDir", sourceDir))
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("创建输出目录失败", logger.String("outputDir", outputDir), logger.ErrorField(err))
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 设置工作目录
	workDir := sourceDir
	if len(workingDir) > 0 && workingDir[0] != "" {
		workDir = workingDir[0]
	}

	// 使用系统zip命令压缩
	cmd := exec.CommandContext(ctx, "zip", "-r", outputPath, ".")
	cmd.Dir = workDir

	if err := cmd.Run(); err != nil {
		logger.Error("执行zip命令失败",
			logger.String("sourceDir", sourceDir),
			logger.String("outputPath", outputPath),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to execute zip command: %w", err)
	}

	logger.Info("目录压缩完成", logger.String("sourceDir", sourceDir), logger.String("outputPath", outputPath))
	return nil
}

// CompressDirectoryToDir 压缩指定目录到缓存目录
// sourceDir: 要压缩的源目录路径
// cacheDir: 缓存目录路径
// fileName: 缓存文件名（不包含扩展名）
// workingDir: 工作目录（可选，默认为sourceDir）
func CompressDirectoryToDir(ctx context.Context, sourceDir, cacheDir, fileName string, workingDir ...string) (string, error) {
	logger.Info("开始压缩目录到缓存", logger.String("sourceDir", sourceDir), logger.String("cacheDir", cacheDir), logger.String("fileName", fileName))

	// 检查源目录是否存在
	if !IsDirectoryExists(sourceDir) {
		logger.Error("源目录不存在", logger.String("sourceDir", sourceDir))
		return "", fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		logger.Error("创建缓存目录失败", logger.String("cacheDir", cacheDir), logger.ErrorField(err))
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 生成缓存文件路径
	cacheFilePath := filepath.Join(cacheDir, fileName+".zip")

	// 设置工作目录
	workDir := sourceDir
	if len(workingDir) > 0 && workingDir[0] != "" {
		workDir = workingDir[0]
	}

	// 使用系统zip命令压缩到缓存
	cmd := exec.CommandContext(ctx, "zip", "-r", cacheFilePath, ".", "-x@exclude_list.txt")
	cmd.Dir = workDir

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute zip command: %w", err)
	}
	logger.Info("目录压缩到缓存完成", logger.String("sourceDir", sourceDir), logger.String("cacheFilePath", cacheFilePath))

	// 从 cacheFilePath 中去掉 baseDir 前缀
	baseDir := GetEnvOrDefault("APP_DATA_HOME", "/app/data")
	resultPath := strings.Replace(cacheFilePath, baseDir, "", 1)
	return resultPath, nil
}
