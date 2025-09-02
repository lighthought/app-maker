package services

import (
	"autocodeweb-backend/internal/models"
	"autocodeweb-backend/internal/utils"

	"math/rand"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ProjectNameGenerator 项目名生成器接口
type ProjectNameGenerator interface {
	GenerateProjectConfig(requirements string, projectConfig *models.Project) bool
}

// projectNameGenerator 项目名生成器实现
type projectNameGenerator struct {
	// 预定义的项目类型关键词
	projectTypes map[string][]string
	// 预定义的功能关键词
	features map[string][]string
	// 预定义的技术关键词
	technologies map[string][]string
	// 预定义的后缀
	suffixes []string
}

// NewProjectNameGenerator 创建项目名生成器实例
func NewProjectNameGenerator() ProjectNameGenerator {
	return &projectNameGenerator{
		projectTypes: map[string][]string{
			"app":      {"app", "application", "mobile", "web"},
			"platform": {"platform", "system", "service", "hub"},
			"tool":     {"tool", "utility", "helper", "assistant"},
			"game":     {"game", "play", "entertainment"},
			"social":   {"social", "community", "network"},
			"business": {"business", "enterprise", "management"},
		},
		features: map[string][]string{
			"ai":       {"ai", "smart", "intelligent", "auto"},
			"design":   {"design", "style", "fashion", "beauty"},
			"shopping": {"shop", "mall", "store", "market"},
			"social":   {"social", "share", "connect", "chat"},
			"media":    {"media", "photo", "video", "image"},
			"3d":       {"3d", "vr", "ar", "visual"},
			"mobile":   {"mobile", "phone", "app", "wechat"},
			"web":      {"web", "online", "cloud", "web"},
			"data":     {"data", "analytics", "insight", "report"},
			"security": {"secure", "safe", "protect", "guard"},
		},
		technologies: map[string][]string{
			"react":    {"react", "vue", "angular", "frontend"},
			"node":     {"node", "express", "backend", "api"},
			"python":   {"python", "django", "flask", "ml"},
			"java":     {"java", "spring", "android", "kotlin"},
			"database": {"db", "sql", "mongo", "redis"},
			"cloud":    {"cloud", "aws", "azure", "gcp"},
		},
		suffixes: []string{
			"pro", "plus", "max", "premium", "elite", "studio", "hub", "center", "suite", "kit",
			"lab", "works", "tools", "app", "web", "mobile", "cloud", "ai", "smart", "go",
		},
	}
}

// GenerateProjectName 根据需求生成项目名
func (g *projectNameGenerator) GenerateProjectConfig(requirements string, projectConfig *models.Project) bool {
	// 1. 提取关键词
	keywords := g.extractKeywords(requirements)

	// 2. 选择项目类型
	projectType := g.selectProjectType(keywords)

	// 3. 选择主要功能
	feature := g.selectFeature(keywords)

	// 4. 选择后缀
	suffix := g.selectSuffix()

	// 5. 组合项目名
	projectName := g.combineName(projectType, feature, suffix)
	projectDescription := requirements

	projectConfig.Name = projectName
	projectConfig.Description = projectDescription
	projectConfig.ApiBaseUrl = "/api/v1"
	projectConfig.FrontendPort = 3000
	projectConfig.BackendPort = 8080
	passwordUtils := utils.NewPasswordUtils()
	projectConfig.AppSecretKey = passwordUtils.GenerateRandomPassword("app")
	projectConfig.RedisPassword = passwordUtils.GenerateRandomPassword("redis")
	projectConfig.JwtSecretKey = passwordUtils.GenerateRandomPassword("jwt")
	projectConfig.DatabasePassword = passwordUtils.GenerateRandomPassword("database")
	projectConfig.Subnetwork = "172.20.0.0/16"

	return true
}

// extractKeywords 从需求中提取关键词
func (g *projectNameGenerator) extractKeywords(requirements string) []string {
	// 转换为小写
	text := strings.ToLower(requirements)

	// 定义关键词模式
	patterns := []string{
		`\b(app|application|mobile|web|platform|system|service|tool|utility|game|social|business)\b`,
		`\b(ai|smart|intelligent|auto|design|style|fashion|beauty|shop|mall|store|market)\b`,
		`\b(social|share|connect|chat|media|photo|video|image|3d|vr|ar|visual)\b`,
		`\b(mobile|phone|wechat|web|online|cloud|data|analytics|secure|safe)\b`,
		`\b(react|vue|angular|node|express|python|java|spring|database|sql|mongo)\b`,
	}

	var keywords []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)

		for _, match := range matches {
			if !seen[match] {
				keywords = append(keywords, match)
				seen[match] = true
			}
		}
	}

	return keywords
}

// selectProjectType 选择项目类型
func (g *projectNameGenerator) selectProjectType(keywords []string) string {
	// 根据关键词选择项目类型
	for _, keyword := range keywords {
		for _, types := range g.projectTypes {
			for _, projectType := range types {
				if keyword == projectType {
					// 随机选择一个该类型的项目名
					return types[rand.Intn(len(types))]
				}
			}
		}
	}

	// 默认返回 app
	return "app"
}

// selectFeature 选择主要功能
func (g *projectNameGenerator) selectFeature(keywords []string) string {
	// 根据关键词选择功能
	for _, keyword := range keywords {
		for _, features := range g.features {
			for _, feature := range features {
				if keyword == feature {
					// 随机选择一个该类型的功能
					return features[rand.Intn(len(features))]
				}
			}
		}
	}

	// 默认返回 smart
	return "smart"
}

// selectSuffix 选择后缀
func (g *projectNameGenerator) selectSuffix() string {
	return g.suffixes[rand.Intn(len(g.suffixes))]
}

// combineName 组合项目名
func (g *projectNameGenerator) combineName(projectType, feature, suffix string) string {
	// 确保单词首字母大写
	caser := cases.Title(language.English)
	projectType = caser.String(projectType)
	feature = caser.String(feature)
	suffix = caser.String(suffix)

	// 组合方式：Feature + ProjectType + Suffix
	return feature + projectType + suffix
}
