package models

import (
	"autocodeweb-backend/internal/constants"
	"time"
)

type Agent struct {
	Name        string `json:"name" gorm:"size:100;not null"`
	Role        string `json:"role" gorm:"size:20;not null"`
	ChineseRole string `json:"chinese_role" gorm:"size:100;not null"`
}

type ProjectArtifact struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // 在 constants/agents.go 中定义
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Content   string    `json:"content"`
	Format    string    `json:"format"` // 在 constants/agents.go 中定义
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AgentResult struct {
	Success         bool              `json:"success"`
	MarkdownContent string            `json:"markdown_content"`
	Artifacts       []ProjectArtifact `json:"artifacts"`
	NextStage       string            `json:"next_stage"` // devstage
	Dependencies    []string          `json:"dependencies"`
	Error           string            `json:"error"`
	Metadata        map[string]any    `json:"metadata"`
}

// 根据结果组装 MarkdownContent
func (a *AgentResult) GetMarkdownContent() string {
	if a.MarkdownContent == "" {
		if a.Error != "" {
			a.MarkdownContent = a.Error
			return a.MarkdownContent
		}

		if a.Artifacts != nil {
			for _, artifact := range a.Artifacts {
				switch artifact.Type {
				case constants.ArtifactTypePRD:
					a.MarkdownContent += artifact.Content
				case constants.ArtifactTypeUXSpec:
					a.MarkdownContent += artifact.Content
				case constants.ArtifactTypeArchitecture:
					a.MarkdownContent += artifact.Content
					// TODO 根据 agents-server 返回的 artifact 类型，追加到 MarkdownContent 中去
				}
			}
		}
	}
	return a.MarkdownContent
}
