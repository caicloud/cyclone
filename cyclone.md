## Cyclone 对接 github webhook

### 前提

Cyclone pipeline 的触发需要支持 PR.ref、PR.baseref、PR.number、PR.Head.SHA、PR.login (better，在显示日志的时候看到 PR 作者的名字...)

#### 流程:

> Todo(zjj2wry): 添加流程图

1. liubo 创建一个 pr 到 cyclone，github 发送 create pr 事件到 hook 服务
2. hook 服务接受到 create pr 的事件，根据 pr 的信息创建 pipeline records，由于 liubo 没有 cyclone 的权限，records 状态为 pending
3. zhangjun comments `ok to test`，pipeline records 状态改为 triggered，什么时候运行由 cyclone 来管
4. controller list workspace 下 pipeline 的 records，controller 主要的任务是让 triggered 的 pipeline 进入 running 状态，更新 pipeline records( failed 或者 success 或者 pending ) 以及日志的链接更新到 pr 的 status 中
5. 当执行失败, liubo 修改 PR 后 comments `retest`，hook 接收到 comments 的事件，更新 pipeline records 的 status 为 triggered
6. records 执行成功，controller 更新 pr 的 status
7. zhangjun comments `lgtm`，hook 接收到 comments 事件，添加 `lgtm` label, 最后由 caicloud-bot 来负责merge pr 和 cherry-pick 至 release 分支

### 实现
#### 接口如下：
​```golang

type cycloneClient interface {
	// 构建 Pipeline, 创建 pipeline records
	Build(*Pipeline) error
	// list 工作区中的 pipeline records
	ListBuilds(pipeline []string,ops *Option)(map[string]PipelineRecords, error)
	// 退出当前构建
	Abort(pipeline string, records *CycloneRecords) error
}

type githubClient interface {
	// 更新 github pr 的 status
	CreateStatus(org, repo, ref string, s github.Status) error
	ListIssueComments(org, repo string, number int) ([]github.IssueComment, error)
	CreateComment(org, repo string, number int, comment string) error
	DeleteComment(org, repo string, ID int) error
	EditComment(org, repo string, ID int, comment string) error
	// 获取 PR 的状态变化
	GetPullRequestChanges(org, repo string, number int) ([]github.PullRequestChange, error)
}```

hook:

cyclone-controller:

### 目标:

### 非目标

