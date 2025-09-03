# 数据目录结构说明

## 目录结构

```
/app/data/
├── template.zip              # 项目模板文件
├── tmp/                      # 临时文件目录
│   └── compress_*.zip        # 压缩时的临时文件
├── projects/                 # 项目文件目录
│   ├── {user_id}/           # 用户项目目录
│   │   └── {project_id}/    # 具体项目目录
│   └── cache/               # 项目缓存目录
│       └── {project_id}_{name}_{timestamp}.zip  # 删除前的备份文件
└── logs/                    # 日志文件目录
```

## 目录权限

所有目录都使用 `appuser:appgroup` (UID: 1001, GID: 1001) 用户和组权限。

## 数据持久化

在开发环境中，`/app/data` 目录通过 Docker 卷映射到宿主机，确保容器重启后数据不丢失：

```yaml
volumes:
  - app_data:/app/data
```

## 目录用途

### `/app/data/tmp/`
- **用途**: 临时文件存储
- **内容**: 压缩过程中的临时 zip 文件
- **生命周期**: 压缩完成后自动删除

### `/app/data/projects/`
- **用途**: 项目文件存储
- **结构**: `{user_id}/{project_id}/`
- **内容**: 解压后的项目文件，包含前端和后端代码

### `/app/data/projects/cache/`
- **用途**: 项目删除前的备份
- **命名**: `{project_id}_{name}_{timestamp}.zip`
- **生命周期**: 手动清理或定期清理策略

### `/app/data/template.zip`
- **用途**: 项目模板文件
- **来源**: 构建时从 `backend/data/template.zip` 复制
- **内容**: 包含项目的基础结构和配置文件

## 开发环境注意事项

1. **数据持久化**: 使用 Docker 卷确保数据不丢失
2. **权限管理**: 容器内使用非 root 用户运行
3. **目录创建**: Dockerfile 中自动创建所有必要目录
4. **清理策略**: 临时文件自动清理，缓存文件需要手动管理

## 生产环境建议

1. **备份策略**: 定期备份 `/app/data/projects/cache/` 目录
2. **存储优化**: 考虑使用对象存储服务替代本地文件存储
3. **监控**: 监控磁盘使用情况，避免空间不足
4. **清理**: 实现自动清理策略，删除过期的缓存文件
