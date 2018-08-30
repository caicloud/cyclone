package cauth

import "github.com/caicloud/cargo-admin/pkg/utils/time"

const (
	RoleGroupAllProject   = "all-projects"
	RoleGroupAllWorkspace = "all-workspaces"
	RoleGroupOneProject   = "one-project"
	RoleTypeOwner         = "owner"
	RoleTypeUser          = "user"
	RoleTypeGuest         = "guest"
)

// Meta defines metadata for all resources
type Meta struct {
	LastModified time.Time         `json:"lastModified,omitempty"`
	CreateTime   time.Time         `json:"createTime"`
	DeleteTime   time.Time         `json:"deleteTime,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Finalizers   []string          `json:"finalizers,omitempty"`
}

// User defines user resource
type User struct {
	// Common metadata
	Meta `json:",inline"`

	// Username of the user
	//
	// required: true
	Username string `json:"username"`

	// Username of the user
	//
	// required: false
	Nickname string `json:"nickname,omitempty"`

	// Password of the user
	Password string `json:"password,omitempty"`

	// Email of the user
	Email string `json:"email"`

	// Whether if the user's email is verified or not
	EmailVerified bool `json:"emailVerified" request:"-"`

	// FIXME
	Invalid bool `json:"invalid"`

	// FIXME
	Remote string `json:"remote,omitempty"`

	// FIXME
	RemoteID string `json:"remoteId,omitempty"`

	Role MemberRoleType `json:"role,omitempty"`
}

// MemberRoleType defines role type of member, now
// sets of users have three kinds: system, tenant, team
// Now all kinds of user sets have two types of
// member: 'owner' or 'member'
type MemberRoleType string

const (
	// OwnerType defines member type which is owner
	OwnerType MemberRoleType = "owner"

	// MemberType defines member type which is member
	MemberType MemberRoleType = "member"
)

// Team is a set of user
// It belongs to a Tenant
type Team struct {
	Meta        `json:",inline"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tenant      string `json:"tenant"`

	Members []Member `json:"members"`
}

// Tenant is a bigger set of user
type Tenant struct {
	// Common metadata
	Meta `json:",inline"`

	// ID of the tenant
	ID string `json:"id"`

	// Name of the tenent
	//
	// required: true
	Name string `json:"name"`

	// User facing tenent description
	//
	// required: false
	Description string `json:"description"`

	// Members of the tenent
	Members []Member `json:"members"`
}

// ListMeta defines metadata for list
type ListMeta struct {
	Meta  `json:"-"`
	Total int `json:"total"`
}

// UserList is a list of User
type UserList struct {
	// Common list metadata
	ListMeta `json:",inline"`

	// List of users.
	Items []User `json:"items"`
}

// TenantList is a list of Tenant
type TenantList struct {
	// Common list metadata
	ListMeta `json:",inline"`

	// List of tenants
	Items []Tenant `json:"items"`
}

// TeamList is a list of Team
type TeamList struct {
	ListMeta `json:",inline"`
	Items    []Team `json:"items"`
}

// MemberList is a list of Member
type MemberList struct {
	ListMeta `json:",inline"`
	Items    []Member `json:"items"`
}

// RoleList is a list of Role
type RoleList struct {
	// Common list metadata
	ListMeta `json:",inline"`

	// List of roles.
	Items []Role `json:"items"`
}

// Member defines a user in Team or Tenant
type Member struct {
	Name              string         `json:"name"`
	Role              MemberRoleType `json:"role"`
	CreationTimestamp time.Time      `json:"creationTimestamp,omitempty"`
}

// Password defines passowrd for user
type Password struct {
	Username    string `json:"username"`
	OldPassword string `json:"oldPassword,omitempty"`
	NewPassword string `json:"newPassword,omitempty"`
}

// Paginator is a list option for paging
type Paginator struct {
	Start int `json:"start" url:"query"`
	Limit int `json:"limit" url:"query"`
}

// Sorter defines simple sorter
// e.g. ["name"] means sort by incresing order of name
// TODO(liubog2008): it is not supported now
type Sorter []string

// Selector defines simple selector
// e.g. ["name"] means select name field of object
// TODO(liubog2008): it is not supported now
type Selector []string

// UserListOptions defines options to LIST users
type UserListOptions struct {
	Paginator `json:",inline"`
	Sorter    `json:"sort" url:"query"`
	Selector  `json:"fields" url:"query"`

	Team        string `json:"team" url:"query"`
	Tenant      string `json:"x-tenant" url:"header"`
	NotInTenant string `json:"notInTenant" url:"query"`

	Remote   string `json:"remote" url:"query"`
	RemoteID string `json:"remoteId" url:"query"`

	// used for filter by tenant
	// only enabled when current tenant is system-tenant
	TenantFilter string `json:"tenant" url:"query"`
}

// TenantListOptions defines options to LIST tenants
type TenantListOptions struct {
	Paginator `json:",inline"`
	Sorter
	Selector

	User   string `json:"user" url:"query"`
	Status string `json:"status" url:"query"`
}

// TenantUpdateOptions defines options when update tenant
type TenantUpdateOptions struct {
	// UpdateMembers determine whether members will be updated
	// It is used for compatiblity
	UpdateMembers bool `json:"updateMembers" url:"query"`
}

// TeamListOptions defines options to LIST teams
type TeamListOptions struct {
	Paginator `json:",inline"`
	Sorter    `json:"sort" url:"query"`
	Selector  `json:"fields" url:"query"`

	Tenant string `json:"x-tenant" url:"header"`
	User   string `json:"user" url:"query"`
}

// MemberListOptions defines options to LIST members
type MemberListOptions struct {
	Paginator `json:",inline"`
	Sorter    `json:"sort" url:"query"`
	Selector  `json:"fields" url:"query"`

	Role      MemberRoleType `json:"role" url:"query"`
	NotInTeam string         `json:"notInTeam" url:"query"`
}

// RoleListOptions defines options to LIST roles
type RoleListOptions struct {
	Paginator `json:",inline"`
	Sorter    `json:"sort" url:"query"`
	Selector  `json:"fields" url:"query"`

	RoleType  string      `json:"roleType" url:"query"`
	RoleGroup string      `json:"roleGroup" url:"query"`
	SubType   SubjectType `json:"subType" url:"query"`
	SubID     string      `json:"subId" url:"query"`

	ResourceType string `json:"resourceType" url:"query"`
	ResourceName string `json:"resourceName" url:"query"`

	DataFilter []string `json:"dataFilter" url:"query"`
}

// Subject defines authorization subject
type Subject struct {
	SubType SubjectType `json:"subType"`
	SubID   string      `json:"subId"`
}

// SubjectType defines type of subject
type SubjectType string

const (
	// UserType defines subject which means all users in system
	UserType SubjectType = "users"

	// TeamType defines subject which means all members in team
	TeamType SubjectType = "teams"

	// TenantType defines subject which means all members in tenant
	TenantType SubjectType = "tenants"

	// TeamOwnerType defines subject which means team owners in team
	TeamOwnerType SubjectType = "teams.owner"

	// TenantOwnerType defines subject which means tenant owners in tenant
	TenantOwnerType SubjectType = "tenants.owner"
)

// Role defines authorization role which user can act as
type Role struct {
	// Common metadata
	Meta `json:",inline"`

	// Auto generated unique ID of the role.
	ID string `json:"id"`

	// Subject reference the identity that is bound to the role; e.g. tenant,
	// team, user etc.
	Subject Subject `json:"subject"`

	// Group and Type defines RoleType, which is the index of RoleTemplate. Each
	// Role must reference a RoleTemplate. Since we do not support custom role,
	// value of Group and Type is fixed and must correspond to pre-defined roles.
	// For example, group can be one of: all-clusters, one-cluster and Type can be
	// one of: owner, user, guest.
	Group string `json:"group"`
	Type  string `json:"type"`

	// Resource is the resource bound to the role.
	Resource Resource `json:"resource"`
}

// Resource defines resource for authorization
type Resource struct {
	ResourceID string `json:"resourceId"`
	// Deprecated
	ResourceName string `json:"resourceName"`
	// Deprecated
	ResourceType string `json:"resourceType"`

	Data map[string]Param `json:"data,omitempty"`
}

const (
	DataKeyTenant   = "tenant"
	DataKeyProject  = "projects"
	DataKeyRegistry = "registry"
)

// Param defines param which used by role template rule
type Param struct {
	Dynamic bool   `json:"dynamic"`
	Value   string `json:"value"`
}

// RoleTemplate defines a set of permission and named it
type RoleTemplate struct {
	Description string          `json:"description"`
	RoleType    RoleType        `json:"roleType"`
	Caicloud    *CaicloudSource `json:"caicloud,omitempty"`
	K8s         *K8sSource      `json:"k8s,omitempty"`
}

// K8sSource defines rule from k8s source
// NOTE(liubog2008): k8s source is not supported now
type K8sSource struct {
	Rules []Rule `json:"rules"`
}

// CaicloudSource defines rule from caicloud source
type CaicloudSource struct {
	Rules []Rule `json:"rules"`
}

// Rule defines caicloud permission rule of api request
type Rule struct {
	APIGroups []string `json:"apiGroups"`
	Paths     []string `json:"paths"`
	Verbs     []string `json:"verbs"`
	ParamType string   `json:"paramType"`
}

// ImpersonateType defines type can be impersonated
type ImpersonateType string

const (
	// ImpersonateTenantType defines tenant type
	ImpersonateTenantType = "tenant"
)

// ImpersonationRule defines impersonation rules
type ImpersonationRule struct {
	// Type defines impersonated type
	// Now only tenant is supported
	Type ImpersonateType `json:"type"`

	// Allowed defines allowed value to impersonate
	// '*' means all is allowed
	Allowed []string `json:"allowed"`
}

// RoleType is the index key of RoleTemplate. For list of pre-defined role
// templates, see "pkg/bootstrap".
type RoleType string

// Claim defines jwt claim
// TODO(liubog2008): move claim API out of core API
type Claim struct {
	// groups are used for kubernetes
	Groups  []string      `json:"groups"`
	Tenants []TenantClaim `json:"tenants"`
}

// TenantClaim defines claim which is from tenant
type TenantClaim struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Teams []TeamClaim `json:"teams"`
}

// TeamClaim defines claim which is from team
type TeamClaim struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
