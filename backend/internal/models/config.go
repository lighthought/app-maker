package models

// PreviewFilesConfig 预览项目文件配置
type PreviewFilesConfig struct {
	Folders []string `json:"folders"`
	Files   []string `json:"files"`
}
