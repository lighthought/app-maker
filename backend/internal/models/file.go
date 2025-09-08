package models

// FileItem 文件项结构
type FileItem struct {
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Type       string     `json:"type"` // file, folder
	Size       int64      `json:"size"`
	ModifiedAt string     `json:"modifiedAt"`
	Children   []FileItem `json:"children,omitempty"`
}

// FileContent 文件内容结构
type FileContent struct {
	Path       string `json:"path"`
	Content    string `json:"content"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modifiedAt"`
}
