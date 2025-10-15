export default {
  // 通用词语
  common: {
    confirm: '确认',
    cancel: '取消',
    save: '保存',
    edit: '编辑',
    delete: '删除',
    create: '创建',
    search: '搜索',
    loading: '加载中...',
    submit: '提交',
    reset: '重置',
    close: '关闭',
    open: '打开',
    upload: '上传',
    download: '下载',
    preview: '预览',
    retry: '重试',
    success: '成功',
    error: '错误',
    warning: '警告',
    info: '提示',
    completed: '已完成',
    inProgress: '进行中',
    pending: '待处理',
    failed: '失败',
    draft: '草稿',
    user: '用户',
    status: '状态',
    websocketDebug: 'WebSocket 调试',
    unknownError: '未知错误',
    and: '和',
    back: '返回',
    websocketError: 'WebSocket 连接错误',
    reconnect: '重连',
    reconnecting: '正在重连... ({attempts}/{max})',
    aiThinking: 'AI Agent 正在思考中...',
    inputRequirements: '输入您的需求或问题...',
    selectAgent: '选择要对话的 Agent',
    agentLocked: 'Agent 已锁定，正在等待回复',
    manualRefresh: '手动刷新',
    refresh: '刷新',
    expand: '展开',
    collapse: '折叠',
    copy: '复制',
    copySuccess: '复制成功',
    copyFailed: '复制失败',
    copyRetry: '复制失败，请重试'
  },

  // 导航和菜单
  nav: {
    home: '首页',
    dashboard: '控制台',
    createProject: '创建项目',
    profile: '个人中心',
    logout: '退出登录'
  },

  // 首页相关
  hero: {
    title: '多Agent自动实现APP和网站项目',
    subtitle: '用自然语言描述需求，AI Agent 自动生成完整项目',
    inputPlaceholder: '描述你的项目需求，例如：创建一个电商网站...',
    createButton: '开始创建',
    recentProjects: '最近项目'
  },

  // 使用流程
  process: {
    title: '使用流程',
    describe: '描述需求',
    describeDescription: '用自然语言描述你的项目需求',
    agentAnalysis: 'Agent分析',
    agentAnalysisDescription: '多Agent协作分析需求并制定方案',
    generateCode: '生成代码',
    generateCodeDescription: '自动生成高质量的代码和文档',
    testDeploy: '测试部署',
    testDeployDescription: '自动测试并部署到目标环境'
  },

  // 功能特性
  features: {
    title: '功能特性',
    smartCodeGeneration: '智能代码生成',
    smartCodeGenerationDescription: '基于自然语言描述，自动生成高质量的代码',
    multiAgentCollaboration: '多Agent协作',
    multiAgentCollaborationDescription: '产品经理、架构师、UX专家、开发工程师等多角色协作',
    fastDeployment: '快速部署',
    fastDeploymentDescription: '支持CI/CD容器部署，或下载后本地部署',
    secureReliable: '安全可靠',
    secureReliableDescription: '企业级安全保障，数据加密传输',
    codeRepository: '代码托管',
    codeRepositoryDescription: '支持GitLab代码托管，自动提交推送文档和代码',
    efficientDevelopment: '高效开发',
    efficientDevelopmentDescription: '提升开发效率，减少重复工作'
  },

  // 页脚
  footer: {
    description: '让编程变得更简单，让创意更快实现',
    contact: '联系我们',
    follow: '关注我们',
    rights: '保留所有权利',
    email: '邮箱',
    account: '账号',
    xiaohongshu: '小红书',
    bilibili: 'B站',
    douyin: '抖音'
  },


  // 控制台
  dashboard: {
    title: '控制台',
    welcomeBack: '欢迎回来，{name}',
    todayStats: '今天是 {date}，您有 {count} 个项目',
    totalProjects: '总项目数',
    newThisMonth: '本月新增',
    myProjects: '我的项目',
    statusFilter: '状态筛选',
    allStatus: '全部状态',
    systemStatus: '系统状态',
    currentProject: '当前项目',
    backendService: '后端服务',
    database: '数据库连接',
    agentOnline: 'AI Agent 在线',
    normal: '正常',
    abnormal: '异常',
    checking: '检查中',
    doubleClickEdit: '双击编辑',
    createdAt: '创建时间',
    progress: '进度',
    actionEdit: '编辑',
    actionPreview: '预览',
    actionDownload: '下载',
    actionDelete: '删除',
    deleteConfirm: '确定要删除此项目吗？',
    createFirstProject: '创建第一个项目',
    noProjects: '暂无项目',
    noProjectsDesc: '您还没有创建任何项目，开始您的第一个项目吧！'
  },

  // 项目相关
  project: {
    createProject: '创建项目',
    editProject: '编辑项目',
    searchProjects: '搜索项目...',
    projectName: '项目名称',
    projectDescription: '项目描述',
    projectType: '项目类型',
    projectTypeWeb: '网站项目',
    projectTypeApp: '移动应用',
    createButton: '创建项目',
    saveChanges: '保存修改',
    deleteProject: '删除项目',
    uploadFiles: '上传文件',
    downloadProject: '下载项目',
    projectCreated: '项目创建成功',
    projectUpdated: '项目更新成功',
    projectDeleted: '项目删除成功',
    projectCreateError: '项目创建失败',
    projectUpdateError: '项目更新失败',
    projectDeleteError: '项目删除失败',
    nameRequired: '请输入项目名称',
    descriptionRequired: '请输入项目描述',
    creating: '正在创建项目，请稍候...',
    loadProjectFailed: '加载项目失败',
    projectInfoUpdated: '项目信息已更新',
    envSetupCompleted: '项目环境配置完成，刷新文件树',
    loadStagesFailed: '加载开发阶段失败',
    projectWithId: '项目 {id}',
    devProgress: '开发进度',
    projectFiles: '项目文件',
    loadingFiles: '加载文件列表中...',
    noFileData: '暂无文件数据',
    previewUnavailable: '预览暂不可用',
    previewDevelopingNote: '项目正在开发中，预览功能将在部署完成后可用',
    previewLoadFailed: '预览加载失败',
    projectDataLoaded: '项目数据已加载，开始加载文件',
    projectDataExists: '组件挂载时项目数据已存在，开始加载文件',
    projectDataNotLoaded: '组件挂载时项目数据尚未加载，等待项目数据...',
    projectSettings: '项目设置',
    devConfiguration: '开发配置',
    devConfigNote: '留空表示使用用户默认设置。这些配置仅在项目环境初始化时生效。'
  },

  // 按钮和交互文本
  buttons: {
    experience: '立即体验',
    enterConsole: '进入控制台',
    createProject: '创建项目',
    startCreating: '开始创建',
    viewAll: '查看全部'
  },

  // 关于我们
  about: {
    title: '关于我们'
  },

  // Header 相关
  header: {
    notifications: '消息通知',
    noMessages: '暂无消息',
    userSettings: '用户设置',
    logout: '退出登录'
  },

  // 用户设置
  userSettings: {
    nickname: '昵称姓名',
    nicknamePlaceholder: '请输入昵称姓名',
    nicknameRequired: '请输入昵称姓名',
    nicknameLength: '昵称长度在 3 到 20 个字符',
    email: '邮箱地址',
    emailPlaceholder: '请输入邮箱地址',
    emailRequired: '请输入邮箱地址',
    emailFormat: '请输入正确的邮箱格式',
    phoneBinding: '手机号绑定',
    phoneBound: '已绑定手机号',
    saveSuccess: '保存成功',
    saveFailed: '保存失败',
    networkError: '保存失败，请检查网络连接',
    developmentSettings: '开发设置',
    cliTool: 'CLI 工具',
    cliToolPlaceholder: '选择 CLI 工具',
    modelProvider: '模型提供商',
    modelProviderPlaceholder: '选择模型提供商',
    aiModel: 'AI 模型',
    aiModelPlaceholder: '输入模型名称，如: glm-4.6',
    modelApiUrl: 'API 地址',
    modelApiUrlPlaceholder: '输入 API 地址',
    apiToken: 'API Token',
    apiTokenPlaceholder: '输入 API Token (如: sk-...)',
    devSettingsNote: '这些设置将作为新项目的默认配置。支持本地 Ollama、Zhipu AI、Claude 等多种模型。API Token 为可选项，某些云端 API 厂商需要。'
  },

  // 认证相关
  auth: {
    welcomeBack: '欢迎回来',
    createAccount: '创建账户',
    login: '登录',
    register: '注册',
    email: '邮箱',
    emailPlaceholder: '请输入邮箱',
    username: '用户名',
    usernamePlaceholder: '请输入用户名',
    password: '密码',
    passwordPlaceholder: '请输入密码',
    confirmPassword: '确认密码',
    confirmPasswordPlaceholder: '请再次输入密码',
    rememberMe: '记住我',
    forgotPassword: '忘记密码？',
    agreeTerms: '注册即表示您同意我们的',
    userAgreement: '《用户协议》',
    privacyPolicy: '《隐私政策》',
    welcomeToAppMaker: '欢迎使用 App-Maker！',
    termsDescription: '本协议是您与 App-Maker 平台之间的法律协议，请您仔细阅读。',
    privacyImportance: '我们非常重视您的隐私保护。',
    privacyDescription: '本政策说明了我们如何收集、使用和保护您的个人信息。',
    
    // 表单验证
    usernameRequired: '请输入用户名',
    usernameMinLength: '用户名至少需要3个字符',
    usernameMaxLength: '用户名不能超过20个字符',
    emailRequired: '请输入邮箱',
    emailFormatError: '请输入有效的邮箱地址',
    passwordRequired: '请输入密码',
    passwordMinLength: '密码至少需要6个字符',
    confirmPasswordRequired: '请确认密码',
    passwordMismatch: '两次输入的密码不一致',
    
    // 操作结果
    loginSuccess: '登录成功',
    loginFailed: '登录失败',
    registerSuccess: '注册成功',
    registerFailed: '注册失败',
    agreeTermsRequired: '请先同意用户协议和隐私政策',
    forgotPasswordFeature: '密码重置功能开发中...',
    socialLoginFeature: '{provider} 登录功能开发中...'
  },

  // 编辑器相关
  editor: {
    code: '代码',
    preview: '预览',
    openInNewWindow: '新窗口打开',
    refresh: '刷新',
    previewUnavailable: '预览暂不可用',
    previewDevelopingNote: '项目正在开发中，预览功能将在部署完成后可用',
    viewCode: '查看代码',
    selectFile: '选择文件查看代码',
    loadingFile: '正在加载文件内容...',
    selectFileToView: '选择一个文件查看代码内容',
    loadingFileFailed: '加载文件内容失败',
    loadError: '文件加载出错',
    rawContent: '原始内容（可能包含乱码）',
    encodingConversionFailed: '编码转换失败',
    fallbackLoadFailed: '备用加载也失败',
    previewMode: '预览模式',
    editMode: '编辑模式'
  },

  // AI Agent 角色
  agent: {
    devEngineer: '开发工程师',
    productManager: '产品经理',
    productOwner: '产品负责人',
    architect: '架构师',
    uxExpert: 'UX专家',
    analyst: '分析师',
    testEngineer: '测试工程师',
    opsEngineer: '运维工程师'
  },

  // 开发阶段
  stage: {
    initializing: '初始化',
    setupEnvironment: '环境配置',
    pendingAgents: 'agent初始化',
    checkRequirement: '需求检查',
    generatePrd: '生成PRD',
    defineUxStandard: 'UX标准',
    designArchitecture: '架构设计',
    defineDataModel: '数据模型',
    defineApi: 'API设计',
    planEpicAndStory: '任务规划',
    developStory: '功能开发',
    fixBug: '问题修复',
    runTest: '自动测试',
    deploy: '项目部署',
    done: '完成',
    failed: '失败'
  },

  // 预览相关
  preview: {
    shareLink: '分享预览链接',
    generatingLink: '正在生成分享链接...',
    linkGenerated: '分享链接已生成',
    generateFailed: '生成分享链接失败',
    generateError: '生成分享链接时发生错误',
    expiresAt: '过期时间',
    token: '令牌',
    shareNote: '此链接将在过期时间后失效，任何拥有此链接的人都可以预览项目。',
    openLink: '打开链接',
    desktop: '桌面视图',
    tablet: '平板视图',
    mobile: '手机视图',
    deviceView: '设备视图',
    sharePreview: '分享预览',
    copyUrl: '复制链接',
    urlCopied: '链接已复制到剪贴板',
    copyFailed: '复制失败，请手动复制',
    deploy: '一键部署',
    deploying: '正在部署项目...',
    deploySuccess: '部署成功！',
    deployFailed: '部署失败'
  }
}