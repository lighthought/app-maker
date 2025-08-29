package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Monitor Redis 监控信息
type Monitor struct {
	client *redis.Client
}

// NewMonitor 创建新的监控实例
func NewMonitor(client *redis.Client) *Monitor {
	return &Monitor{client: client}
}

// HealthCheck 健康检查
func (m *Monitor) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.client.Ping(ctx).Err()
}

// GetInfo 获取 Redis 信息
func (m *Monitor) GetInfo() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := m.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取Redis信息失败: %w", err)
	}

	// 解析 INFO 命令的输出
	result := make(map[string]string)
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
	}

	return result, nil
}

// GetMemoryUsage 获取内存使用情况
func (m *Monitor) GetMemoryUsage() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取内存信息
	memoryInfo, err := m.client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("获取内存信息失败: %w", err)
	}

	// 获取数据库大小
	dbSize, err := m.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取数据库大小失败: %w", err)
	}

	// 获取连接数
	clientList, err := m.client.ClientList(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("获取客户端列表失败: %w", err)
	}

	// 计算连接数
	clientCount := len(strings.Split(clientList, "\n"))

	result := map[string]interface{}{
		"db_size":      dbSize,
		"client_count": clientCount,
		"memory_info":  memoryInfo,
	}

	return result, nil
}

// GetKeyspaceStats 获取键空间统计
func (m *Monitor) GetKeyspaceStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取键空间信息
	keyspaceInfo, err := m.client.Info(ctx, "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("获取键空间信息失败: %w", err)
	}

	// 获取所有数据库的键数量
	databases := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	dbStats := make(map[string]int64)

	for _, db := range databases {
		// 切换到指定数据库
		if err := m.client.Do(ctx, "SELECT", db).Err(); err != nil {
			continue
		}

		// 获取键数量
		if count, err := m.client.DBSize(ctx).Result(); err == nil {
			if count > 0 {
				dbStats[fmt.Sprintf("db%d", db)] = count
			}
		}
	}

	// 切换回默认数据库
	m.client.Do(ctx, "SELECT", 0)

	result := map[string]interface{}{
		"keyspace_info":  keyspaceInfo,
		"database_stats": dbStats,
	}

	return result, nil
}

// GetPerformanceMetrics 获取性能指标
func (m *Monitor) GetPerformanceMetrics() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取命令统计
	commandStats, err := m.client.Info(ctx, "commandstats").Result()
	if err != nil {
		return nil, fmt.Errorf("获取命令统计失败: %w", err)
	}

	// 获取延迟统计
	latencyStats, err := m.client.Info(ctx, "latencystats").Result()
	if err != nil {
		return nil, fmt.Errorf("获取延迟统计失败: %w", err)
	}

	result := map[string]interface{}{
		"command_stats": commandStats,
		"latency_stats": latencyStats,
	}

	return result, nil
}

// GetFullStats 获取完整统计信息
func (m *Monitor) GetFullStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 基本信息
	if info, err := m.GetInfo(); err == nil {
		stats["basic_info"] = info
	}

	// 内存使用
	if memory, err := m.GetMemoryUsage(); err == nil {
		stats["memory_usage"] = memory
	}

	// 键空间统计
	if keyspace, err := m.GetKeyspaceStats(); err == nil {
		stats["keyspace_stats"] = keyspace
	}

	// 性能指标
	if performance, err := m.GetPerformanceMetrics(); err == nil {
		stats["performance_metrics"] = performance
	}

	// 健康状态
	stats["health_status"] = "healthy"
	if err := m.HealthCheck(); err != nil {
		stats["health_status"] = "unhealthy"
		stats["health_error"] = err.Error()
	}

	return stats, nil
}
