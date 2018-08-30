package models

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/caicloud/nirvana/log"
)

var Project Projecter

type Projecter interface {
	EnsureIndexes()
	Save(pinfo *ProjectInfo) error
	GetTenantProjects(registry, tenant string) ([]*ProjectInfo, error)
	GetGroupedProjects(registry string) ([]*ProjectGroup, error)
	FindOnePage(tenant string, registry string, includePublic bool, start, limit int) (int, []*ProjectInfo, error)
	FindOnePageWithPrefix(tenant string, registry string, includePublic bool, prefix string, start, limit int) (int, []*ProjectInfo, error)
	FindOnePageOnlyPublic(registry string, start, limit int) (int, []*ProjectInfo, error)
	FindByName(tenant, registry, name string) (*ProjectInfo, error)
	FindByNameWithoutTenant(registry, name string) (*ProjectInfo, error)
	FindAllByRegistry(registry string) ([]*ProjectInfo, error)
	Delete(tenant, registry, name string) error
	DeleteWithoutTenant(registry, name string) error
	DeleteAllByRegistry(registry string) error
	Update(tenant, registry, name, desc string) error
	FindAllSortByName(tenant string, registry string, includePublic bool) ([]*ProjectInfo, error)
	IsExist(tenant, registry, name string) (bool, error)
}

type _Project struct {
	*mgo.Collection
}

var projectIndexes = []mgo.Index{
	{Key: []string{"name"}},
	{Key: []string{"registry"}},
	{Key: []string{"tenant"}},
}

func (p *_Project) EnsureIndexes() {
	EnsureIndexes(p.Collection, projectIndexes)
}

// If a project is protected, it can't be removed
type ProjectInfo struct {
	Name           string    `bson:"name"`
	Registry       string    `bson:"registry"`
	Tenant         string    `bson:"tenant"`
	ProjectId      int64     `bson:"project_id"`
	Description    string    `bson:"description"`
	IsPublic       bool      `bson:"is_public"`
	IsProtected    bool      `bson:"is_protected"`
	CreationTime   time.Time `bson:"creation_time"`
	LastUpdateTime time.Time `bson:"last_update_time"`
}

type ProjectGroup struct {
	Tenant  string  `bson:"_id"`
	PIDs    []int64 `bson:"projects"`
	Publics []bool  `bson:"publics"`
}

func (p *_Project) Save(pinfo *ProjectInfo) error {
	return p.Insert(pinfo)
}

func (p *_Project) GetGroupedProjects(registry string) ([]*ProjectGroup, error) {
	var pipeline []interface{}
	pipeline = append(pipeline, bson.M{"$match": bson.M{"registry": registry}})
	pipeline = append(pipeline, bson.M{"$group": bson.M{
		"_id":      "$tenant",
		"projects": bson.M{"$push": "$project_id"},
		"publics":  bson.M{"$push": "$is_public"},
	}})

	pGroups := make([]*ProjectGroup, 0)
	if err := p.Pipe(pipeline).All(&pGroups); err != nil {
		log.Errorf("mongo query all error: %v", err)
		return nil, err
	}
	return pGroups, nil
}

// Get all projects for a tenant, all projects that are public or belongs to the tenant would be returned
func (p *_Project) GetTenantProjects(registry, tenant string) ([]*ProjectInfo, error) {
	query := p.Find(bson.M{
		"$or": []bson.M{
			bson.M{
				"registry":  registry,
				"is_public": true,
			},
			bson.M{
				"registry": registry,
				"tenant":   tenant,
			},
		},
	})

	pInfos := make([]*ProjectInfo, 0)
	err := query.All(&pInfos)
	if err != nil {
		return nil, err
	}
	return pInfos, nil
}

func (p *_Project) FindOnePage(tenant string, registry string, includePublic bool, start, limit int) (int, []*ProjectInfo, error) {
	pinfos := make([]*ProjectInfo, 0)
	var query *mgo.Query

	if includePublic {
		query = p.Find(bson.M{
			"$or": []bson.M{
				bson.M{
					"registry":  registry,
					"is_public": true,
				},
				bson.M{
					"tenant":   tenant,
					"registry": registry,
				},
			},
		})
	} else {
		query = p.Find(bson.M{"tenant": tenant, "registry": registry, "is_public": false})
	}

	// If query.Count() is at after of either query.Skip() or query.Limit(), the return count will be impacted.
	// eg.total:100
	// query.Count() = 100
	// query.Limit(20) -> query.Count() = 20
	// query.Skip(90) -> query.Count() = 10
	// query.Skip(90) -> query.Limit(20) -> query.Count() = 10
	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(start).Limit(limit).Sort("is_public", "-creation_time").All(&pinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}

	return total, pinfos, err
}

func (p *_Project) FindOnePageWithPrefix(tenant string, registry string, includePublic bool, prefix string, start, limit int) (int, []*ProjectInfo, error) {
	pinfos := make([]*ProjectInfo, 0)
	var query *mgo.Query

	if includePublic {
		query = p.Find(bson.M{
			"$or": []bson.M{
				bson.M{
					"registry":  registry,
					"is_public": includePublic,
					"name":      bson.M{"$regex": "^" + prefix},
				},
				bson.M{
					"tenant":   tenant,
					"registry": registry,
					"name":     bson.M{"$regex": "^" + prefix},
				},
			},
		})
	} else {
		query = p.Find(bson.M{"tenant": tenant, "registry": registry, "name": bson.M{"$regex": "^" + prefix}})
	}

	// If query.Count() is at after of either query.Skip() or query.Limit(), the return count will be impacted.
	// eg.total:100
	// query.Count() = 100
	// query.Limit(20) -> query.Count() = 20
	// query.Skip(90) -> query.Count() = 10
	// query.Skip(90) -> query.Limit(20) -> query.Count() = 10
	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(start).Limit(limit).Sort("is_public", "-creation_time").All(&pinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}

	return total, pinfos, err
}

func (p *_Project) FindOnePageOnlyPublic(registry string, start, limit int) (int, []*ProjectInfo, error) {
	pinfos := make([]*ProjectInfo, 0)
	query := p.Find(bson.M{"registry": registry, "is_public": true})

	total, err := query.Count()
	if err != nil {
		log.Errorf("mongo query count error: %v", err)
		return 0, nil, err
	}
	log.Infof("mongo query count total: %d", total)

	err = query.Skip(start).Limit(limit).Sort("is_public", "-creation_time").All(&pinfos)
	if err != nil {
		log.Errorf("mongo query all error: %v", err)
		return 0, nil, err
	}

	return total, pinfos, err
}

func (p *_Project) FindByName(tenant, registry, name string) (*ProjectInfo, error) {
	pinfo := &ProjectInfo{}
	err := p.Find(bson.M{"tenant": tenant, "registry": registry, "name": name}).One(pinfo)
	return pinfo, err
}

func (p *_Project) FindByNameWithoutTenant(registry, name string) (*ProjectInfo, error) {
	pinfo := &ProjectInfo{}
	err := p.Find(bson.M{"registry": registry, "name": name}).One(pinfo)
	return pinfo, err
}

func (p *_Project) FindAllByRegistry(registry string) ([]*ProjectInfo, error) {
	pinfos := make([]*ProjectInfo, 0)
	err := p.Find(bson.M{"registry": registry}).All(&pinfos)
	return pinfos, err
}

func (p *_Project) Delete(tenant, registry, name string) error {
	return p.Remove(bson.M{"tenant": tenant, "registry": registry, "name": name})
}

func (p *_Project) DeleteWithoutTenant(registry, name string) error {
	return p.Remove(bson.M{"registry": registry, "name": name})
}

func (p *_Project) DeleteAllByRegistry(registry string) error {
	_, err := p.RemoveAll(bson.M{"registry": registry})
	return err
}

func (p *_Project) Update(tenant, registry, name, desc string) error {
	return p.Collection.Update(bson.M{"tenant": tenant, "registry": registry, "name": name},
		bson.M{
			"$set": bson.M{
				"description":      desc,
				"last_update_time": time.Now().Format(time.RFC3339),
			},
		},
	)
}

func (p *_Project) FindAllSortByName(tenant string, registry string, includePublic bool) ([]*ProjectInfo, error) {
	pinfos := make([]*ProjectInfo, 0)
	var query *mgo.Query

	if includePublic {
		query = p.Find(bson.M{"$or": []bson.M{bson.M{"registry": registry, "is_public": true}, bson.M{"tenant": tenant, "registry": registry}}})
	} else {
		query = p.Find(bson.M{"tenant": tenant, "registry": registry})
	}

	err := query.Sort("name").All(&pinfos)

	return pinfos, err
}

func (p *_Project) IsExist(tenant, registry, name string) (bool, error) {
	return IsExist(p.Collection, bson.M{"tenant": tenant, "registry": registry, "name": name})
}
