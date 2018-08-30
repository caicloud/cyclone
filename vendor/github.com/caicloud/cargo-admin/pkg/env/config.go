package env

// Mongo Safe struct in mgo.v2
type Safe struct {
	W        int    `json:"w"`        // Min # of servers to ack before success
	WMode    string `json:"wmode"`    // Write mode for MongoDB 2.0+ (e.g. "majority")
	WTimeout int    `json:"wtimeout"` // Milliseconds to wait for W before timing out
	FSync    bool   `json:"fsync"`    // Sync via the journal if present, or via data files sync otherwise
	J        bool   `json:"j"`        // Sync via the journal if present
}

// Mongo config
type MgoConfig struct {
	Addrs string `json:"addrs"`
	DB    string `json:"db"`
	Mode  string `json:"mode"`
	Safe  *Safe  `json:"safe"`
}

type ConflictStrategy string

const (
	IgnoreStrategy ConflictStrategy = "ignore"
	ForceStrategy  ConflictStrategy = "force"
)

type DefaultPublicProject struct {
	Name     string           `json:"name"`
	IfExists ConflictStrategy `json:"if_exists"`
	Harbor   string           `json:"harbor"`
}

type RegistrySpec struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminConfig struct {
	Address         string                  `json:"address"`
	SystemTenant    string                  `json:"system_tenant"`
	Mongo           *MgoConfig              `json:"mongo"`
	Projects        []*DefaultPublicProject `json:"default_public_projects"`
	DefaultRegistry *RegistrySpec           `json:"default_registry"`
}

type TokenConfig struct {
	Address        string     `json:"address"`
	SystemTenant   string     `json:"system_tenant"`
	PrivateKeyFile string     `json:"private_key"`
	CauthAddress   string     `json:"cauth_addr"`
	TokenExpire    int        `json:"token_expiration"`
	Mongo          *MgoConfig `json:"mongo"`
}
