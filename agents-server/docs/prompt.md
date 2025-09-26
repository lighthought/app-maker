# agents-server 各阶段提示词

## 0. analyse 检查用户需求
```
@bmad/analyst.mdc 我的需求是这样，请你为我生成项目简介，再执行市场研究。输出对应的文档到 docs/analyse/ 目录下。
```


## 1. PM 根据需求生产 PRD
```
@bmad/pm.mdc 我希望你根据我的需求帮我输出 PRD 文档到  docs/PRD.md。
简化部署和运维、商业模式、成功指标、风险评估中的市场和运营风险。
技术选型我后续再和架构师深入讨论，主题颜色我后续再和 ux 专家讨论。
我的需求是：{requirement}
```

## 2. UX-expert 输出 UI/UX标准
```
@bmad/ux-expert.mdc 帮我基于这个 @docs/PRD.md 和参考页面设计(如果需求有提及的话)，输出前端的 UX Spec 到 docs/ux/ux-spec.md。关键web页面的文生网站提示词到 docs/ux/page-prompt.md。
```

## 3. arichitect 输出架构设计
```
@bmad/architect.mdc 请你基于最新的 @docs/PRD.md 和 UX 专家的设计文档 @docs/ux-spec.md，帮我把整体架构设计 Architect.md, 前端架构设计frontend_arch.md, 后端架构设计backend_arch.md输出到 docs/arch/ 目录下。
当前的项目代码是由模板生成，技术架构是：
1. 前端：vue.js+ vite ；
2. 后端服务和 API： GO + Gin 框架实现 API、数据库用 PostgreSql、缓存用 Redis。
3. 部署相关的脚本已经有了，用的 docker，前端用一个 nginx ，配置 /api 重定向到 /backend:port ，这样就能在前端项目中访问后端 API 了。引用关系是：前端依赖后端，后端依赖 Redis 和 PostgreSql。
```


## 4. po 划分Epic和story
```
@bmad/po.mdc 我希望你基于 @docs/PRD.md 和 @docs/arch/ 目录下的架构设计创建 Epics（史诗）和 Stories（用户故事）。生成分片的 Epics，输出到 docs/epics/ 下多个文件。再根据 Epics 生成分片的 Stories，输出到 docs/stories/ 下多个文件。不要考虑安全、合规。
```

## 5. arichitect 定义数据模型
```
@bmad/architect.mdc 请你基于最新的 @docs/PRD.md，和 @docs/stories/ 目录下的用户故事。输出数据模型设计(可以用 sql 脚本代替)，输出到 docs/arch/ 目录下
```

## 6. arichitect 定义API接口
```
@bmad/architect.mdc 请你基于最新的 @docs/PRD.md，和 @docs/stories/ 目录下的用户故事和 docs/arch/目录下的数据模型，生成 API 接口定义，输出到 docs/arch/ 下多个文件（按控制器分类）。
```

## 7. dev 开发 story 功能
```
@bmad/dev.mdc 请你始终记得项目的前后端框架及约束：
1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；
2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。
3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。
4. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程；

请你基于 @docs/PRD.md 和架构师的设计 @docs/arch/ ，以及 UX 标准 @docs/ux/ 实现 @{epic_file_name} 中的用户故事 @{story_file_name}。  
实现完，编译确认下验收的标准是否都达到了，达到了以后，更新用户故事文档，勾上对应的验收标准。  
然后再询问我，是否继续。不要每次生成多余的总结文档，你可以总结做了什么事，但是不要新增不必要的说明文件。
```


## 8. dev 修复开发问题
```
@bmad/dev.mdc 请你始终记得项目的前后端框架及约束：
1. 后端 Handler -> service -> repository 分层，引用和依赖关系都在 container 依赖注入容器中维护；
2. 后端的服务和repository 一般都有接口，供上一层调用。接口的定义和实现放在同一个文件中，不用为了定义服务接口或 repository 接口而单独新建文件。
3. 后端部分每个文件夹的具体作用可以参考 @backend/ReadMe.md。前端部分参考 @frontend/ReadMe.md。
4. 每次修改之前，先理解当前项目中已有的公共组件、框架约束，不要新增不必要的框架和技术流程；

我当前遇到了 {bug_description}，请你帮我修复下
```

## 8. dev 执行自动测试
```
@bmad/dev.mdc 请你使用项目现有的测试脚本，完成项目的自动测试过程。包括前端的 lint 和后端的测试过程。
如果有 make test 命令，直接执行即可
```

## 9. dev 打包项目
```
@bmad/dev.mdc 请你使用项目现有的打包脚本，完成项目的打包过程。
如果有类似 make build-dev 或 make build-prod 命令，直接执行即可
```