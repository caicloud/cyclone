# 功能介绍

## 关联版本管理工具

Cyclone 可以关联多种常用版本管理工具（git、svn等），通过 OAuth 授权后，拉取代码，建立 webhook。当用户向代码库中提交 commit、pull request 和 release 版本时自动触发 CI/CD 工作流水线。

- 创建与版本管理工具关联的服务

<div align="center">
	<img src="./create_service.png" alt="create_service" width="500">
</div>

- 列出所有服务

<div align="center">
	<img src="./list_services.png" alt="list_services" width="500">
</div>

## 持续集成与安全扫描

全过程可视化的工作流水线："prebuild" 编译可执行文件，"build" 构建发布镜像， "integration" 集成测试，"publish" 发布镜像并安全扫描，"post build" 镜像发布后的关联操作，"deploy" 使用发布镜像部署应用。构建结果邮件通知。所有过程均以容器为载体，消除环境差异。

- 工作流水线日志

<div align="center">
	<img src="./logs.png" alt="logs" width="500">
</div>

- 安全扫描

<div align="center">
	<img src="./security.png" alt="security" width="500">
</div>

- 邮件通知

<div align="center">
	<img src="./pagging.png" alt="pagging" width="500">
</div>

## 资源管理

任务调度逻辑与构建任务分离。支持添加用户工作节点，多样化的构建配额方案。

- 用户资源设置

<div align="center">
	<img src="./quota.png" alt="quota" width="500">
</div>

- 单次构建资源限制

<div align="center">
	<img src="./create_version.png" alt="create_version" width="500">
</div>

## 联合发布与依赖管理

管理组件支持微服务多组件联合发布，使用图形化界面展示组件的依赖关系及联合发布的过程和状态，应用拓扑关系图形化。

<div align="center">
	<img src="./dependency.png" alt="dependency" width="500">
</div>

## 持续交付

基于发布策略和角色控制功能，提供灵活的持续部署方式。基于容器和镜像的版本控制，提供多种升级回滚策略。

- 多种部署方案

<div align="center">
	<img src="./deployment.png" alt="deployment" width="500">
</div>
