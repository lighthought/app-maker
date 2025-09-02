/**
 * 时间工具类
 * 提供各种时间格式化方法
 */

/**
 * 格式化日期时间（带时分秒）
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期时间字符串，格式：2025-09-02 15:21:41
 */
export const formatDateTime = (date: string | Date | null | undefined): string => {
  if (!date) {
    return 'Invalid Date'
  }
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) {
    return 'Invalid Date'
  }
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  }).replace(/\//g, '-')
}

/**
 * 格式化日期（仅日期）
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期字符串，格式：2025年9月2日
 */
export const formatDate = (date: string | Date | null | undefined): string => {
  if (!date) {
    return 'Invalid Date'
  }
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) {
    return 'Invalid Date'
  }
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

/**
 * 格式化日期（简短格式）
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期字符串，格式：2025-09-02
 */
export const formatDateShort = (date: string | Date | null | undefined): string => {
  if (!date) {
    return 'Invalid Date'
  }
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) {
    return 'Invalid Date'
  }
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  }).replace(/\//g, '-')
}

/**
 * 获取相对时间描述
 * @param date 日期字符串或Date对象
 * @returns 相对时间描述，如：刚刚、5分钟前、1小时前等
 */
export const getRelativeTime = (date: string | Date | null | undefined): string => {
  if (!date) {
    return 'Invalid Date'
  }
  const d = typeof date === 'string' ? new Date(date) : date
  if (isNaN(d.getTime())) {
    return 'Invalid Date'
  }
  
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const diffMinutes = Math.floor(diff / (1000 * 60))
  const diffHours = Math.floor(diff / (1000 * 60 * 60))
  const diffDays = Math.floor(diff / (1000 * 60 * 60 * 24))
  
  if (diffMinutes < 1) {
    return '刚刚'
  } else if (diffMinutes < 60) {
    return `${diffMinutes}分钟前`
  } else if (diffHours < 24) {
    return `${diffHours}小时前`
  } else if (diffDays < 7) {
    return `${diffDays}天前`
  } else {
    return formatDateShort(d)
  }
}