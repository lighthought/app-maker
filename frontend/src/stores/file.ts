import { httpService } from '@/utils/http'
import { defineStore } from 'pinia'

export const useFilesStore = defineStore('projectFiles', () => {

// 下载项目
const downloadProject = async (projectId: string) => {
    try {
      // 使用 httpService 的 download 方法
      const blob = await httpService.download(`/files/download/${projectId}`)
      
      // 验证blob数据
      if (!blob || blob.size === 0) {
        throw new Error('下载的文件为空')
      }
      
      // 创建下载链接
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `project-${projectId}.zip`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
      
      console.log('项目下载成功:', projectId)
    } catch (error) {
      console.error('下载项目失败:', error)
      throw error
    }
  }

  // 获取文件内容
  const getFileContent = async (projectId: string, filePath: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: {
          path: string
          content: string
          size: number
          modifiedAt: string
        }
      }>(`/files/filecontent/${projectId}`, {
        params: { filePath }
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
  const getProjectFiles = async (projectId: string, path?: string) => {
    try {
      const response = await httpService.get<{
        code: number
        message: string
        data: Array<{
          name: string
          path: string
          type: 'file' | 'folder'
          size: number
          modifiedAt: string
        }>
      }>(`/files/files/${projectId}`, {
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

  return {
    downloadProject,
    getFileContent,
    getProjectFiles,
  }
})