package models

import "time"

// PreviewFilesConfig 预览项目文件配置
type PreviewFilesConfig struct {
	Folders []string `json:"folders"`
	Files   []string `json:"files"`
}

// JenkinsConfig Jenkins 配置
type JenkinsConfig struct {
	BaseURL     string
	Username    string
	APIToken    string
	JobName     string
	RemoteToken string
	Timeout     time.Duration
}
