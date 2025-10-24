import { httpService } from '@/utils/http'
import { defineStore } from 'pinia'

// 文件树节点类型
export interface FileTreeNode {
  name: string
  path: string
  type: 'file' | 'folder'
  size: number
  children?: FileTreeNode[]
  expanded?: boolean
  loaded?: boolean
  content?: string
}

export const useFilesStore = defineStore('projectFiles', () => {

// 下载文件
const downloadFile = async (filePath: string) => {
    try {
      // 使用 httpService 的 download 方法
      const blob = await httpService.download(`/files/download`, { filePath: filePath })
      
      // 验证blob数据
      if (!blob || blob.size === 0) {
        throw new Error('下载的文件为空')
      }

      const filename = filePath.split('/').pop()
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${filename}`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      console.log('文件下载成功:', filePath, filename)
    } catch (error) {
      console.error('下载文件失败:', error)
      throw error
    }
  }

  // 获取文件内容
  const getFileContent = async (projectGuid: string, filePath: string, encoding: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: {
          path: string
          content: string
          size: number
        }
      }>(`/files/filecontent/${projectGuid}`, {
        params: { filePath, encoding }
      })

      if (response.code === 0 && response.data) {
        return response.data
      } else {
        console.error('获取文件内容失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取文件内容失败:', error)
      return null
    }
  }

  // 获取项目文件列表
  const getProjectFiles = async (projectGuid: string, path?: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: Array<{
          name: string
          path: string
          type: 'file' | 'folder'
          size: number
        }>
      }>(`/files/files/${projectGuid}`, {
        params: path ? { path } : {}
      })

      if (response.code === 0 && response.data) {
        // 排序，把文件夹放前面
        return response.data.sort((a, b) => {
          return a.type === 'folder' ? -1 : 1
        })
      } else {
        console.error('获取项目文件失败:', response.message)
        return null
      }
    } catch (error) {
      console.error('获取项目文件失败:', error)
      return null
    }
  }

  // 组装文件树结构
  const buildFileTree = (files: Array<{
    name: string
    path: string
    type: 'file' | 'folder'
    size: number
  }>): FileTreeNode[] => {
    const tree: FileTreeNode[] = []
    const pathMap = new Map<string, FileTreeNode>()

    // 首先处理所有文件，创建节点
    files.forEach(file => {
      const node: FileTreeNode = {
        name: file.name,
        path: file.path,
        type: file.type,
        size: file.size,
        children: file.type === 'folder' ? [] : undefined,
        expanded: false,
        loaded: false
      }
      pathMap.set(file.path, node)
    })

    // 然后建立父子关系
    files.forEach(file => {
      const node = pathMap.get(file.path)!
      const pathParts = file.path.split('/')
      
      if (pathParts.length > 1) {
        // 有父目录
        const parentPath = pathParts.slice(0, -1).join('/')
        const parentNode = pathMap.get(parentPath)
        
        if (parentNode) {
          // 父节点已存在，直接添加子节点
          parentNode.children!.push(node)
        } else {
          // 父节点不存在，需要创建
          const parentName = pathParts[pathParts.length - 2]
          const newParentNode: FileTreeNode = {
            name: parentName,
            path: parentPath,
            type: 'folder',
            size: 0,
            children: [node],
            expanded: false,
            loaded: false
          }
          pathMap.set(parentPath, newParentNode)
          
          // 递归处理父节点的父节点
          if (pathParts.length > 2) {
            const grandParentPath = pathParts.slice(0, -2).join('/')
            const grandParentNode = pathMap.get(grandParentPath)
            if (grandParentNode) {
              grandParentNode.children!.push(newParentNode)
            } else {
              // 继续递归创建祖父节点
              const grandParentName = pathParts[pathParts.length - 3]
              const newGrandParentNode: FileTreeNode = {
                name: grandParentName,
                path: grandParentPath,
                type: 'folder',
                size: 0,
                children: [newParentNode],
                expanded: false,
                loaded: false
              }
              pathMap.set(grandParentPath, newGrandParentNode)
            }
          }
        }
      } else {
        // 根目录文件
        tree.push(node)
      }
    })

    // 收集所有根节点
    const rootNodes: FileTreeNode[] = []
    pathMap.forEach(node => {
      const pathParts = node.path.split('/')
      if (pathParts.length === 1) {
        rootNodes.push(node)
      }
    })

    // 按文件夹优先排序
    return rootNodes.sort((a, b) => {
      if (a.type === 'folder' && b.type !== 'folder') return -1
      if (a.type !== 'folder' && b.type === 'folder') return 1
      return a.name.localeCompare(b.name)
    })
  }

  // 获取项目文件树
  const getProjectFileTree = async (projectGuid: string): Promise<FileTreeNode[]> => {
    try {
      const files = await getProjectFiles(projectGuid)
      if (files) {
        return buildFileTree(files)
      }
      return []
    } catch (error) {
      console.error('获取项目文件树失败:', error)
      return []
    }
  }

  // 展开文件夹并加载子文件
  const expandFolder = async (projectGuid: string, folderPath: string, fileTree: FileTreeNode[]): Promise<void> => {
    try {
      const files = await getProjectFiles(projectGuid, folderPath)
      // 找到对应的文件夹节点
      const folderNode = findNodeByPath(fileTree, folderPath)
      if (folderNode && folderNode.type === 'folder') {
        if (files && files.length > 0) {
          folderNode.children = files.map(file => ({
            name: file.name,
            path: file.path,
            type: file.type,
            size: file.size,
            children: file.type === 'folder' ? [] : undefined,
            expanded: false,
            loaded: false
          }))
        } else {
          // 空文件夹，设置空数组
          folderNode.children = []
        }
        folderNode.loaded = true
      }
    } catch (error) {
      console.error('展开文件夹失败:', error)
    }
  }

  // 根据路径查找节点
  const findNodeByPath = (nodes: FileTreeNode[], path: string): FileTreeNode | null => {
    for (const node of nodes) {
      if (node.path === path) {
        return node
      }
      if (node.children) {
        const found = findNodeByPath(node.children, path)
        if (found) return found
      }
    }
    return null
  }

  return {
    downloadFile,
    getFileContent,
    getProjectFiles,
    getProjectFileTree,
    expandFolder,
    findNodeByPath,
    buildFileTree
  }
})