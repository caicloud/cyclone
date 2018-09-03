package resource

import (
	"io/ioutil"
	"sync"
	"time"
)

type MultiTenantConfig struct {
	multiTenantEnabled bool
	updateTime         time.Time
	lock               *sync.Mutex
}

var multiTenantConfig = MultiTenantConfig{
	multiTenantEnabled: true,
	lock:               &sync.Mutex{},
}

func IsMultiTenantEnabled() bool {
	multiTenantConfig.lock.Lock()
	if multiTenantConfig.updateTime.Add(time.Minute).Before(time.Now()) {
		multiTenantConfig.updateTime = time.Now()
		multiTenantConfig.multiTenantEnabled = getMultiTenantConfig()
	}
	multiTenantConfig.lock.Unlock()
	return multiTenantConfig.multiTenantEnabled
}

func getMultiTenantConfig() bool {
	data, err := ioutil.ReadFile("/platform/etc/multi_tenant")
	if err != nil {
		return true
	}
	return string(data) == "enabled"
}
