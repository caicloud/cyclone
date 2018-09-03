package types

type Dockerfile struct {
	Group   string `yaml:"group" json:"group"`
	Image   string `yaml:"image" json:"image"`
	Content string `yaml:"content" json:"content"`
}

type DockerfileGroup struct {
	Group       string        `yaml:"group" json:"group"`
	Dockerfiles []*Dockerfile `yaml:"dockerfiles" json:"dockerfiles"`
}
