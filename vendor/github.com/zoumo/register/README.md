# Register

register is a package for golang to build your own register



##  Example

```
package main

var (
	clouds = NewRegister(nil)
)

type Cloud interface{
	// some interface func
}

type Config struct {
	// some config field
}

type CloudFactory func(Config) (Cloud, error)


func RegisterCloud(name string, factory CloudFactory) {
	clouds.RegisterCloud(cloud, factory)
}

func GetCloud(name string, config Config) (Cloud，error) {
	v, found := clouds.Get(name)
	if !found {
		return nil, nil
	}
	factory := v.(CloudFactory)
	return factory(config)
}

func main() {
	RegisterCloud("aws", AwsCloudFactory)
	RegisterCloud("gce", GceCloudFactory)
	RegisterCloud("azure", AzureCloudFactory)
}
```

