package harbor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/caicloud/cargo-admin/pkg/errors"
	"github.com/caicloud/cargo-admin/pkg/models"
	"github.com/caicloud/cargo-admin/pkg/utils/matcher"

	"github.com/caicloud/nirvana/log"
)

const (
	tickerDuration = time.Minute * 30
)

var hclosed chan struct{}

func InitHarbor(closing chan struct{}) (chan struct{}, error) {
	hclosed = make(chan struct{})

	rInfos, err := models.Registry.FindAll()
	if err != nil {
		log.Errorf("list registries from mongodb error: %v", err)
		return nil, err
	}
	ClientMgr.Refresh(rInfos)
	log.Infof("%s", rInfos)

	go backgroundHarbors(closing)

	return hclosed, nil
}

func LoginAndGetCookies(conf *Config) ([]*http.Cookie, error) {
	url := LoginUrl(conf.Host, conf.Username, conf.Password)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		msg := matcher.MaskPwd(err.Error())
		log.Error(msg)
		return nil, ErrorUnknownInternal.Error(msg)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorUnknownInternal.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Errorf("login harbor: %s error: %s", conf.Host, b)
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s", b))
	}

	// If status code is 200 and no cookies set, it's not a valid harbor. For example, 1.1.1.1
	if string(b) != "" || len(resp.Cookies()) == 0 {
		return nil, ErrorUnknownInternal.Error(fmt.Sprintf("%s is not a valid harbor", conf.Host))
	}

	return resp.Cookies(), nil
}

func backgroundHarbors(closing chan struct{}) {
	ticker := time.NewTicker(tickerDuration)

	for {
		select {
		case <-ticker.C:
			rInfos, err := models.Registry.FindAll()
			if err != nil {
				log.Errorf("list registries from mongodb error: %v", err)
				continue
			}
			ClientMgr.Refresh(rInfos)
		case <-closing:
			log.Info("capture closing signal, backgroundHarbors gorutine will exit")
			close(hclosed)
			return
		}
	}
}
