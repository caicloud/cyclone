package resource

import (
	"context"
	"fmt"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/admin/types"
	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/harbor"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"github.com/davecgh/go-spew/spew"
)

const (
	TriggerKindManual   = "Manual"
	TriggerKindOnPush   = "OnPush"
	TriggerKindSchedule = "Scheduled"
	TriggerKindAll      = "AllTriggerType"

	DirectionForward  = "forward"
	DirectionBackword = "backward"

	ActionStart = "Start"
)

// =================================================================================================

// 关于时间的问题：

// 背景
// 1. list record 时，非常依赖 ReplicationInfo 的这三个时间：LastUpdateTime、LastTriggerTime 和 LastListRecordsTime
// 2. cargo 节点和控制集群的节点可能会存在时间不一致，这个现象比较普遍
// 3. harbor 返回的时间精度为秒

// 解决方法：
// 1. infra 负责 cargo 节点和控制集群各个节点的定时时间同步；
// 2. replication collection 中的所有时间，均为控制节点的时间，不使用 harbor 返回的时间；
// 3. 即便对各个节点进行了定时时间同步，也无法保证节点之间存在时间误差，因此，在 list replication record 时，starTime 向前推 500ms，endTime 向后推 500ms；
// 4. record collection 中的时间，使用 harbor repJobs 的时间（因为 record 和 harbor 和 repJobs 就是通过时间联系起来的）；
// 5. 其他和时间相关的操作（比如 docker login 时前 service token），暂时不做修改。

// =================================================================================================

type CreateReplicationReq struct {
	Alias             string
	Name              string
	Project           string
	ReplicateNow      bool
	ReplicateDeletion bool
	SourceRegistry    string
	TargetRegistry    string
	Trigger           *types.ReplicationTrigger
}

// 主要步骤：
// 1. 获取创建 replication 的一些必要信息
// 2. check target registry 是否存在 project，如果存在，则直接执行；如果不存在，则需要从 target registry 创建一个同名的 project；
// TODO：如果存在，该租户是否有权限操作 target registry 的 project（由于前端进行了一些限制，所以这个 bug 暂时不明显）
// 3. 向 source registry 发请求，创建该 replication policy；
// 4. 从 source registry 获取 刚刚创建的 replication policy;
// 5. 将改 replication 的信息写入 mongodb 数据库；
// 6. 按 API 返回执行格式的 replication 结构体。
func CreateReplication(ctx context.Context, tenant string, crReq *CreateReplicationReq) (*types.Replication, error) {
	cli, err := harbor.ClientMgr.GetClient(crReq.SourceRegistry)
	if err != nil {
		return nil, err
	}
	hcrpReq, sreginfo, treginfo, spinfo, err := getHarborCreateRepPolicyReq(tenant, crReq)
	if err != nil {
		log.Error("get harbor create replication policy request error: %v", err)
		log.Infof("user create replicaiotn request info: %v", spew.Sdump(crReq))
		return nil, err
	}

	// 注：这个 createTime 一定要在向 harbor 发请求前获取
	// list replication records 时，会依赖这三个时间 LastUpdateTime、LastTriggerTime 和 LastListRecordsTime
	// 如果这三个时间比实际时间晚，那么会影响到 harbor list repjobs 的结果，从而影响到 list replication records。
	createTime := time.Now()
	hrpid, err := createReplicationPolicy(ctx, tenant, crReq, cli, hcrpReq, spinfo)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// 从 registry 中获取刚刚创建的 replication policy，以此作为是否成功创建 replication policy 的依据
	hrp, err := cli.GetReplicationPolicy(hrpid)
	if err != nil {
		log.Errorf("get replication policy: %d from registry: %s error: %v", hrpid, crReq.SourceRegistry, err)
		return nil, err
	}
	repinfo := &models.ReplicationInfo{
		Name:                crReq.Name,
		Tenant:              tenant,
		ReplicationPolicyId: hrpid,
		Alias:               crReq.Alias,
		Project:             crReq.Project,
		TriggerKind:         crReq.Trigger.Kind,
		SourceRegistry:      crReq.SourceRegistry,
		TargetRegistry:      crReq.TargetRegistry,
		CreationTime:        createTime,
		LastUpdateTime:      createTime,
		LastListRecordsTime: createTime,
	}
	err = models.Replication.Save(repinfo)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, ErrorUnknownInternal.Error(err)
	}
	return getReplication(repinfo, hrp, sreginfo, treginfo)
}

// =================================================================================================

// 创建 replication：
// 1. 如果此 tenant 中不存在 target project，那么需要手动创建；如果存在，直接手动创建即可；
// 2. 判断此 target 中有无同名的 project，如果没有，则创建一个同名的 project；
// 3. 向 registry 发送创建 replication policy 的请求。
func createReplicationPolicy(ctx context.Context, tenant string, crReq *CreateReplicationReq, cli *harbor.Client, hcrpReq *harbor.HarborCreateRepPolicyReq, spinfo *models.ProjectInfo) (int64, error) {
	b, err := models.Project.IsExist(tenant, crReq.TargetRegistry, crReq.Project)
	if err != nil {
		return 0, err
	}
	if !b {
		if spinfo.IsPublic {
			_, err := CreatePublicProject(ctx, tenant, crReq.TargetRegistry, crReq.Project, "")
			if err != nil {
				log.Infof("create project: %s into target registry: %s, error: %v", crReq.Project, crReq.TargetRegistry, err)
				return 0, err
			}
		} else {
			_, err := CreateProject(ctx, tenant, crReq.TargetRegistry, crReq.Project, "", spinfo.IsPublic)
			if err != nil {
				log.Infof("create project: %s into target registry: %s, error: %v", crReq.Project, crReq.TargetRegistry, err)
				return 0, err
			}
		}
		log.Infof("create project: %s into target registry: %s successfully", crReq.Project, crReq.TargetRegistry)
	} else {
		log.Infof("target registry: %s has already created project: %s")
	}
	hrpid, err := cli.CreateReplicationPolicy(hcrpReq)
	if err != nil {
		log.Errorf("create replication policy to harbor error: %v", err)
		log.Errorf("HarborCreateRepPolicyReq: %s", spew.Sdump(hcrpReq))
		return 0, err
	}

	log.Infof("create replication policy: %s successfully", crReq.Name)
	return hrpid, nil
}

// =================================================================================================

// 本函数的关键是生成 harbor.HarborCreateRepPolicyReq，其余的返回值意义不大，只是为了减少重复调用
// 生成 harbor.HarborCreateRepPolicyReq 的步骤：
// 1. 从 registry collection 中获取 registry 相关信息;
// 2. 从 project  collection 中获取源 project 的 projectId;
// 3. 从 source registry 中获取 targetId;（此步骤也可不直接请求 source registry，直接从 registry collection 也能获取到）
// 4. 转换 trigger 和 filter（filter 在目前的 API 中并未体现，但为了向前兼容，加了此项）
func getHarborCreateRepPolicyReq(tenant string, crReq *CreateReplicationReq) (*harbor.HarborCreateRepPolicyReq, *models.RegistryInfo, *models.RegistryInfo, *models.ProjectInfo, error) {
	sreginfo, err := models.Registry.FindByName(crReq.SourceRegistry)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, nil, nil, nil, err
	}
	log.Infof("soure registry info, name: %v, domain: %v, username: %s", sreginfo.Name, sreginfo.Domain, sreginfo.Username)
	spinfo, err := models.Project.FindByName(tenant, sreginfo.Name, crReq.Project)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, nil, nil, nil, err
	}
	log.Infof("source registry project info: %v", spew.Sdump(spinfo))
	treginfo, err := models.Registry.FindByName(crReq.TargetRegistry)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, nil, nil, nil, err
	}
	log.Infof("target registry, name: %v, domain: %v, username: %s", treginfo.Name, treginfo.Domain, treginfo.Username)
	target, err := getHarborTarget(sreginfo.Name, treginfo)
	if err != nil {
		log.Errorf("get target error: %v from registry: %s", err, sreginfo.Name)
		return nil, nil, nil, nil, err
	}
	log.Infof("get target: %s from regitry: %s successfully", crReq.Project, treginfo.Name)
	log.Infof("target info: %s", spew.Sdump(target))

	trigger, err := getHarborTrigger(crReq.Trigger)
	if err != nil {
		log.Errorf("get trigger error: %v from registry: %s", err, sreginfo.Name)
		return nil, nil, nil, nil, err
	}
	log.Infof("trigger info: %s", spew.Sdump(trigger))

	filters, err := getHarborFilters()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return &harbor.HarborCreateRepPolicyReq{
		Name:                      crReq.Name,
		Description:               "",
		Filters:                   filters,
		ReplicateDeletion:         crReq.ReplicateDeletion,
		Trigger:                   trigger,
		Projects:                  []*harbor.HarborProject{&harbor.HarborProject{ProjectID: spinfo.ProjectId}},
		Targets:                   []*harbor.HarborRepTarget{target},
		ReplicateExistingImageNow: crReq.ReplicateNow,
	}, sreginfo, treginfo, spinfo, nil
}

// =================================================================================================

func getSourceHarborProject(tenant, registry, project string) (*harbor.HarborProject, error) {
	pInfo, err := models.Project.FindByName(tenant, registry, project)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, ErrorUnknownInternal.Error(err)
	}
	sCli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		log.Errorf("get registry: %s 's clinet error: %v", registry, err)
		return nil, err
	}
	sProject, err := sCli.GetProject(pInfo.ProjectId)
	if err != nil {
		log.Errorf("get project from harbor error: %v, registry: %s, projectId: %s, projectName: %s",
			err, registry, pInfo.ProjectId, pInfo.Name)
		return nil, err
	}
	return sProject, nil
}

func getHarborTarget(registry string, tRegInfo *models.RegistryInfo) (*harbor.HarborRepTarget, error) {
	sCli, err := harbor.ClientMgr.GetClient(registry)
	if err != nil {
		return nil, err
	}
	hTargets, err := sCli.ListTargets()
	if err != nil {
		log.Errorf("list targets error: %v, registry: %s", err, registry)
		return nil, err
	}

	var hTarget *harbor.HarborRepTarget
	for _, t := range hTargets {
		if t.URL == tRegInfo.Host {
			hTarget = t
		}
	}
	if hTarget == nil {
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("not found target in registry: %s", registry))
	}
	return hTarget, nil
}

func getHarborTrigger(trigger *types.ReplicationTrigger) (*harbor.HarborTrigger, error) {
	var ret *harbor.HarborTrigger
	switch {
	case trigger.Kind == TriggerKindManual:
		ret = &harbor.HarborTrigger{Kind: harbor.TriggerKindManual}
	case trigger.Kind == TriggerKindOnPush:
		ret = &harbor.HarborTrigger{Kind: harbor.TriggerKindImmediate}
	case trigger.Kind == TriggerKindSchedule:
		ret = &harbor.HarborTrigger{
			Kind: harbor.TriggerKindSchedule,
			ScheduleParam: &harbor.HarborScheduleParam{
				Type:    trigger.ScheduleParam.Type,
				Weekday: trigger.ScheduleParam.Weekday,
				Offtime: trigger.ScheduleParam.Offtime,
			},
		}
	default:
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("not found the trigger type: %s", trigger.Kind))
	}
	return ret, nil
}

// TODO(li-ang): implements on v2.7.1 or later
func getHarborFilters() ([]harbor.HarborFilter, error) {
	return make([]harbor.HarborFilter, 0), nil
}

// =================================================================================================

type ListReplicationsParams struct {
	Direction   string
	Registry    string
	Project     string
	TriggerType string
	Prefix      string
	Start       int
	Limit       int
}

func ListReplications(ctx context.Context, tenant string, param *ListReplicationsParams) (int, []*types.Replication, error) {
	var total int
	var repinfos []*models.ReplicationInfo
	var err error
	var triggerType string

	if param.TriggerType == TriggerKindAll {
		triggerType = ""
	} else {
		triggerType = param.TriggerType
	}
	mparam := &models.FindOnePageParams{
		Project:     param.Project,
		TriggerType: triggerType,
		Prefix:      param.Prefix,
		Start:       param.Start,
		Limit:       param.Limit,
	}
	switch {
	// 同步方向为：主动
	case param.Direction == DirectionForward:
		total, repinfos, err = models.Replication.FindOnePageBySourceRegistry(tenant, param.Registry, mparam)
		if err != nil {
			log.Errorf("mongo error: %v", err)
			return 0, nil, ErrorUnknownInternal.Error(err)
		}
	// 同步方向为：被动
	case param.Direction == DirectionBackword:
		total, repinfos, err = models.Replication.FindOnePageByTargetRegistry(tenant, param.Registry, mparam)
		if err != nil {
			log.Errorf("mongo error: %v", err)
			return 0, nil, ErrorUnknownInternal.Error(err)
		}
	default:
		return 0, nil, ErrorUnknownRequest.Error("direction only allowed \"forward\" or \"backward\"")
	}

	ret, err := getReplications(repinfos)
	if err != nil {
		log.Errorf("getRetReplications: %v", err)
		return 0, nil, err
	}

	return total, ret, nil
}

// getRetReplications 会请求多层次 registry，耗时可能会有些长，
// 但限制了只能集成三个镜像仓库，耗时可以接受
// getRetReplications 请求步骤
// 1. 从 registry collection 中去除所有的 registry info 信息；
// 2. 从所有的 registry 中 list 所有的 replication policy
// 3. 遍历传入的 repinfos ，遍历的同时，更新和 repinfo 相关的 record 和 record image （getRetReplication会做这个事情);
// 4. 组合 registry、repinfo 等信息，返回结果。
func getReplications(repinfos []*models.ReplicationInfo) ([]*types.Replication, error) {
	ret := make([]*types.Replication, 0, len(repinfos))
	reginfoMap, err := getAllRegistryInfoMap()
	if err != nil {
		return nil, err
	}
	hrpMap, err := getAllHarborReplicationPolicyMap(reginfoMap)
	if err != nil {
		log.Errorf("getAllHarborReplicationPolicyMap error: %v", err)
		return nil, err
	}
	for _, repinfo := range repinfos {
		// TODO(li-ang): 数据不一致的情况下改怎么办
		sreginfo, ok := reginfoMap[repinfo.SourceRegistry]
		if !ok {
			log.Errorf("replication: %s 's source registry: %s can not found in mongodb.", repinfo.Name, repinfo.SourceRegistry)
			continue
		}
		treginfo, ok := reginfoMap[repinfo.TargetRegistry]
		if !ok {
			log.Errorf("replication: %s 's target registry: %s can not found in mongodb.", repinfo.Name, repinfo.TargetRegistry)
			continue
		}
		hrpmap, ok := hrpMap[repinfo.SourceRegistry]
		if !ok {
			log.Errorf("replication: %s 's source registry: %s can not found", repinfo.Name, repinfo.SourceRegistry)
			continue
		}
		hrp, ok := hrpmap[repinfo.ReplicationPolicyId]
		if !ok {
			log.Errorf("replication: %s 's can not found in registry: %s.", repinfo.Name, repinfo.SourceRegistry)
			continue
		}
		replication, err := getReplication(repinfo, hrp, sreginfo, treginfo)
		if err != nil {
			log.Errorf("get ret replication error: %v, registry: %s, replication name: %s, replicaiotn policy id: %d",
				err, repinfo.SourceRegistry, repinfo.Name, repinfo.ReplicationPolicyId)
			continue
		}
		ret = append(ret, replication)
	}
	return ret, nil
}

func getAllRegistryInfoMap() (map[string]*models.RegistryInfo, error) {
	reginfos, err := models.Registry.FindAll()
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	ret := make(map[string]*models.RegistryInfo)
	for _, reginfo := range reginfos {
		ret[reginfo.Name] = reginfo
	}
	return ret, nil
}

func getAllHarborReplicationPolicyMap(reginfoMap map[string]*models.RegistryInfo) (map[string]map[int64]*harbor.HarborReplicationPolicy, error) {
	ret := make(map[string]map[int64]*harbor.HarborReplicationPolicy)
	for _, repinfo := range reginfoMap {
		cli, err := harbor.ClientMgr.GetClient(repinfo.Name)
		if err != nil {
			return nil, err
		}
		hrps, err := cli.ListReplicationPolicies()
		if err != nil {
			log.Errorf("list replicaiton policeis from registry: %s error: %v", repinfo.Name, err)
			return nil, ErrorUnknownInternal.Error(err)
		}
		hrpMap := make(map[int64]*harbor.HarborReplicationPolicy)
		for _, hrp := range hrps {
			hrpMap[hrp.ID] = hrp
		}
		ret[repinfo.Name] = hrpMap
	}
	return ret, nil
}

func GetReplication(ctx context.Context, tenant string, replication string) (*types.Replication, error) {
	repInfo, err := models.Replication.FindByName(tenant, replication)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}

	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return nil, err
	}
	hRep, err := cli.GetReplicationPolicy(repInfo.ReplicationPolicyId)
	if err != nil {
		log.Errorf("get replication policy: %d from registry: %s error: %v", repInfo.ReplicationPolicyId, repInfo.SourceRegistry, err)
		return nil, err
	}
	sRegInfo, err := models.Registry.FindByName(repInfo.SourceRegistry)
	if err != nil {
		return nil, err
	}
	tRegInfo, err := models.Registry.FindByName(repInfo.TargetRegistry)
	if err != nil {
		return nil, err
	}
	return getReplication(repInfo, hRep, sRegInfo, tRegInfo)
}

func getReplication(repInfo *models.ReplicationInfo, hRep *harbor.HarborReplicationPolicy,
	sRegInfo, tRegInfo *models.RegistryInfo) (*types.Replication, error) {
	isSyncing, err := replicationIsSyncing(repInfo, hRep, sRegInfo)
	if err != nil {
		log.Errorf("replicationIsSyncing error: %v", err)
		return nil, err
	}

	return &types.Replication{
		Metadata: &types.ReplicationMetadata{
			Name:           repInfo.Name,
			Alias:          repInfo.Alias,
			CreationTime:   repInfo.CreationTime,
			LastUpdateTime: repInfo.LastUpdateTime,
		},
		Spec: &types.ReplicationSpec{
			Project:           repInfo.Project,
			ReplicateNow:      hRep.ReplicateExistingImageNow,
			ReplicateDeletion: hRep.ReplicateDeletion,
			Source: &types.ReplicationSource{
				Name:   sRegInfo.Name,
				Alias:  sRegInfo.Alias,
				Domain: sRegInfo.Domain,
			},
			Target: &types.ReplicationTarget{
				Name:   tRegInfo.Name,
				Alias:  tRegInfo.Alias,
				Domain: tRegInfo.Domain,
			},
			Trigger: getReplicationTrigger(hRep.Trigger),
		},
		Status: &types.ReplicationStatus{
			IsSyncing:           isSyncing,
			LastReplicationTime: repInfo.LastTriggerTime,
		},
	}, nil
}

// To determine whether a replication is in syncing status.
// - If the replication is triggered within 3 seconds, the replication is regarded as is syncing. This
//   help judge the status when replication is triggered manually but before any jobs are scheduled in harbor.
// - First check latest unprocessed replication logs and check whether there are unfinished ones.
// - Check previous generated replication records, whether there still are some records in syncing status.
func replicationIsSyncing(repInfo *models.ReplicationInfo, hRep *harbor.HarborReplicationPolicy, sRegInfo *models.RegistryInfo) (bool, error) {
	if repInfo.LastTriggerTime.Add(time.Second * 3).After(time.Now()) {
		return true, nil
	}

	_, repJobs, err := makeLatestRecords(repInfo, hRep, sRegInfo)
	if err != nil {
		log.Errorf("make records from latest replication jobs error: %v", err)
		return false, err
	}
	if getRecordStatus(repJobs) == StatusSyncing {
		return true, nil
	}
	records, err := models.Record.FindByStatus(repInfo.Name, StatusSyncing)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return false, err
	}
	if len(records) != 0 {
		return true, nil
	}
	return false, nil
}

func getReplicationTrigger(ht *harbor.HarborTrigger) *types.ReplicationTrigger {
	var ret *types.ReplicationTrigger
	switch {
	case ht.Kind == harbor.TriggerKindManual:
		ret = &types.ReplicationTrigger{Kind: TriggerKindManual}
	case ht.Kind == harbor.TriggerKindImmediate:
		ret = &types.ReplicationTrigger{Kind: TriggerKindOnPush}
	case ht.Kind == harbor.TriggerKindSchedule:
		ret = &types.ReplicationTrigger{
			Kind: TriggerKindSchedule,
			ScheduleParam: &types.ScheduleParam{
				Type:    ht.ScheduleParam.Type,
				Weekday: ht.ScheduleParam.Weekday,
				Offtime: ht.ScheduleParam.Offtime,
			},
		}
	default:
		log.Errorf("not found the trigger type: %s", ht.Kind)
	}
	return ret
}

func UpdateReplication(ctx context.Context, tenant string, repinfo *models.ReplicationInfo, urReq *types.UpdateReplicationReq) (*types.Replication, error) {
	sreginfo, err := models.Registry.FindByName(repinfo.SourceRegistry)
	if err != nil {
		return nil, err
	}
	cli, err := harbor.ClientMgr.GetClient(urReq.Spec.Source.Name)
	if err != nil {
		return nil, err
	}
	oldhrp, err := cli.GetReplicationPolicy(repinfo.ReplicationPolicyId)
	if err != nil {
		return nil, err
	}

	treginfo, err := models.Registry.FindByName(urReq.Spec.Target.Name)
	if err != nil {
		log.Errorf("mongo error: %v", err)
		return nil, err
	}
	hurpReq, err := getHarborUpdateRepPolicyReq(ctx, tenant, repinfo.Name, sreginfo, treginfo, urReq)
	if err != nil {
		log.Error("get harbor update replication policy request error: %v", err)
		log.Infof("update replicaiotn request info: %v", spew.Sdump(urReq))
		return nil, err
	}
	isSyncing, err := replicationIsSyncing(repinfo, oldhrp, sreginfo)
	if err != nil {
		log.Errorf("replicationIsSyncing error: %v", err)
		return nil, err
	}
	if isSyncing {
		return nil, ErrorUnknownRequest.Error(fmt.Sprintf("replication: %s is sycning", repinfo.Name))
	}
	updateTime := time.Now()
	err = cli.UpdateReplicationPolicy(repinfo.ReplicationPolicyId, hurpReq)
	if err != nil {
		log.Errorf("update replication policy to harbor error: %v", err)
		log.Errorf("HarborUpdateRepPolicyReq: %s", spew.Sdump(hurpReq))
		return nil, err
	}
	hrp, err := cli.GetReplicationPolicy(repinfo.ReplicationPolicyId)
	if err != nil {
		log.Errorf("get replication policy: %d from registry: %s error: %v", repinfo.ReplicationPolicyId, urReq.Spec.Source.Name, err)
		return nil, err
	}

	err = models.Replication.UpdateReplication(tenant, repinfo.Name, urReq.Metadata.Alias, urReq.Spec.Target.Name, urReq.Spec.Trigger.Kind, updateTime)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	log.Infof("update replication: %s successfully", repinfo.Name)
	return getReplication(repinfo, hrp, sreginfo, treginfo)
}

func getHarborUpdateRepPolicyReq(ctx context.Context, tenant, replication string,
	sreginfo, treginfo *models.RegistryInfo, urReq *types.UpdateReplicationReq) (*harbor.HarborUpdateRepPolicyReq, error) {
	sproject, err := getSourceHarborProject(tenant, sreginfo.Name, urReq.Spec.Project)
	if err != nil {
		log.Errorf("get source project error: %v from registry: %s", err, sreginfo.Name)
		return nil, err
	}
	log.Infof("get source project: %s from regitry: %s successfully", urReq.Spec.Project, sreginfo.Name)
	log.Infof("source project info: %s", spew.Sdump(sproject))

	target, err := getHarborTarget(sreginfo.Name, treginfo)
	if err != nil {
		log.Errorf("get target error: %v from registry: %s", err, sreginfo.Name)
		return nil, err
	}
	log.Infof("get target: %s from regitry: %s successfully", urReq.Spec.Project, sreginfo.Name)
	log.Infof("target info: %s", spew.Sdump(target))

	trigger, err := getHarborTrigger(urReq.Spec.Trigger)
	if err != nil {
		log.Errorf("get trigger error: %v from registry: %s", err, sreginfo.Name)
		return nil, err
	}
	log.Infof("trigger info: %s", spew.Sdump(trigger))

	filters, err := getHarborFilters()
	if err != nil {
		return nil, err
	}
	return &harbor.HarborUpdateRepPolicyReq{
		Name:                      replication,
		Description:               "",
		Filters:                   filters,
		ReplicateDeletion:         urReq.Spec.ReplicateDeletion,
		Trigger:                   trigger,
		Projects:                  []*harbor.HarborProject{sproject},
		Targets:                   []*harbor.HarborRepTarget{target},
		ReplicateExistingImageNow: urReq.Spec.ReplicateNow,
	}, nil
}

// =================================================================================================

func deleteReplication(repinfo *models.ReplicationInfo) error {
	cli, err := harbor.ClientMgr.GetClient(repinfo.SourceRegistry)
	if err != nil {
		return err
	}

	err = cli.StopRepJobs(repinfo.ReplicationPolicyId)
	if err != nil {
		log.Errorf("stop replication job error: %v", err)
		return err
	}
	err = cli.DeleteReplicationPolicy(repinfo.ReplicationPolicyId)
	if err != nil {
		log.Errorf("delete replication policy: %s from registry: %s error: %v, replication: %s",
			repinfo.ReplicationPolicyId, repinfo.SourceRegistry, err, repinfo.Name)
		return err
	}

	err = models.Replication.Delete(repinfo.Tenant, repinfo.Name)
	if err != nil {
		log.Errorf("Replication.Delete: %v", err)
		return ErrorUnknownInternal.Error(err)
	}
	err = models.Record.DeleteAllByReplication(repinfo.Name)
	if err != nil {
		log.Errorf("Record.DeleteAllByReplication: %v", err)
		return ErrorUnknownInternal.Error(err)
	}
	err = models.RecordImage.DeleteAllByReplication(repinfo.Name)
	if err != nil {
		log.Errorf("RecordImage.DeleteAllByReplication: %v", err)
		return ErrorUnknownInternal.Error(err)
	}

	return nil
}

func DeleteReplication(ctx context.Context, tenant string, replication string) error {
	repinfo, err := models.Replication.FindByName(tenant, replication)
	if err != nil {
		return ErrorUnknownInternal.Error(err)
	}

	return deleteReplication(repinfo)
}

func TriggerReplication(ctx context.Context, tenant string, replication string, triggerRep *types.TriggerReplicationReq) error {
	repInfo, err := models.Replication.FindByName(tenant, replication)
	if err != nil {
		log.Errorf("Find replication %s error: %v", replication, err)
		return ErrorUnknownInternal.Error(err)
	}
	cli, err := harbor.ClientMgr.GetClient(repInfo.SourceRegistry)
	if err != nil {
		return err
	}

	if triggerRep.Action != ActionStart {
		return ErrorUnknownRequest.Error("only support \"Start\" action ")
	}

	// Trigger time should be obtained before sending request to Harbor. It will be used to list
	// replication jobs to make replication records.
	triggerTime := time.Now()
	err = cli.TriggerReplicationPolicy(repInfo.ReplicationPolicyId)
	if err != nil {
		log.Errorf("Trigger replication %s err: %v", repInfo.Name, err)
		return err
	}

	// Harbor doesn't have trigger time information, so store it in Cargo-Admin.
	err = models.Replication.UpdateLastTriggerTime(tenant, replication, triggerTime)
	if err != nil {
		log.Errorf("Update last trigger time error: %v", err)
		return err
	}

	return nil
}

func DeleteAllReplicationsBySourceRegistry(ctx context.Context, sregistry string) error {
	repinfos, err := models.Replication.FindAllBySourceRegistry(sregistry)
	if err != nil {
		log.Errorf("FindAllBySourceRegistry error: %v", err)
		return err
	}

	for _, repinfo := range repinfos {
		err = DeleteReplication(ctx, repinfo.Tenant, repinfo.Name)
		if err != nil {
			log.Errorf("delete replication policy: %d error: %v", err)
		}
	}
	return nil
}

func DeleteAllReplicationsByTargetRegistry(ctx context.Context, tregistry string) error {
	repinfos, err := models.Replication.FindAllByTargetRegistry(tregistry)
	if err != nil {
		log.Errorf("FindAllByTargetRegistry error: %v", err)
		return err
	}

	for _, repinfo := range repinfos {
		err = DeleteReplication(ctx, repinfo.Tenant, repinfo.Name)
		if err != nil {
			log.Errorf("delete replication policy: %d error: %v", err)
		}
	}
	return nil
}

func DeleteAllReplications(project string) error {
	replications, err := models.Replication.FindAllByProject(project)
	if err != nil {
		return err
	}

	total := len(replications)
	deleted := 0
	for _, replication := range replications {
		e := deleteReplication(replication)
		if e != nil {
			log.Errorf("Delete replication %s error: %v", replication.Name, err)
			continue
		}
		deleted++
	}

	if total > 0 {
		log.Infof("%d out of %d replications from project [%s] deleted", deleted, total, project)
	}
	if deleted < total {
		msg := fmt.Sprintf("delete all replications for project %s error, only %d out of %d deleted", project, deleted, total)
		log.Errorf(msg)
		return fmt.Errorf("%s", msg)
	}

	return nil
}
