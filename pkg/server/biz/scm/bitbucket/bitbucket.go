package bitbucket

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cyclone/pkg/server/apis/v1alpha1"
	"github.com/caicloud/cyclone/pkg/server/biz/scm"
	"github.com/caicloud/cyclone/pkg/server/biz/scm/bitbucket/server"
	"github.com/caicloud/cyclone/pkg/util/cerr"
)

type bitbucketType string

const (
	bitbucketCloudAddr = "https://bitbucket.org"

	bitbucketServer bitbucketType = "server"

	bitbucketCloud bitbucketType = "cloud"

	// EventTypeHeader represents the header key for event type of Bitbucket.
	EventTypeHeader = "X-Event-Key"

	// HookEventHeader represent the header key to decide whether it is a bitbucket hook event
	// value could be true or false
	HookEventHeader = "X-BitBucket-Hook-Event"
	// ServerAddressHeader represent the header key of bitbucket server address
	ServerAddressHeader = "X-Server-Address"
)

func init() {
	if err := scm.RegisterProvider(v1alpha1.Bitbucket, NewBitbucket); err != nil {
		log.Errorln(err)
	}
}

// NewBitbucket news BitBucket Server client.
func NewBitbucket(scmCfg *v1alpha1.SCMSource) (scm.Provider, error) {
	bitbucketType := getBitbucketType(scmCfg)

	switch bitbucketType {
	case bitbucketServer:
		client, err := newBitbucketServerClient(scmCfg)
		if err != nil {
			log.Errorf("fail to new BitBucket Server client as %v", err)
			return nil, err
		}

		return server.NewBitbucketServer(scmCfg, client), nil
	case bitbucketCloud:
		// TODO: support bitbucket cloud
		log.Errorln("bitbucet cloud is unsupported now")
		return nil, cerr.ErrorNotImplemented.Error("bitbucket cloud scm")
	default:
		log.Errorf("unsupported bitbucet product: %v", bitbucketType)
		return nil, cerr.ErrorNotImplemented.Error(fmt.Sprintf("unsupported bitbucket product: %v", bitbucketType))
	}
}

func newBitbucketServerClient(scmCfg *v1alpha1.SCMSource) (*server.V1Client, error) {
	if scmCfg.Password == "" && scmCfg.Token == "" {
		return nil, cerr.ErrorParamNotFound.Error("password or token")
	}
	var client *server.V1Client
	var config server.Config
	var err error
	config.BaseURL = scmCfg.Server
	baseClient := &http.Client{}
	switch scmCfg.AuthType {
	case v1alpha1.AuthTypeToken:
		config.AuthType = server.PersonalAccessToken
		config.Token = scmCfg.Token
		client, err = server.NewClient(baseClient, config)
	case v1alpha1.AuthTypePassword:
		config.Username = scmCfg.User
		if scmCfg.Password == "" {
			base, err := url.Parse(config.BaseURL)
			if err != nil {
				return nil, err
			}
			if !strings.HasSuffix(base.Path, "/") {
				base.Path += "/"
			}
			version, err := server.GetBitbucketVersion(baseClient, base)
			if err != nil {
				return nil, err
			}
			isSupportToken, err := server.IsHigherVersion(version, server.SupportAccessTokenVersion)
			if err != nil {
				return nil, err
			}
			if isSupportToken {
				config.AuthType = server.PersonalAccessToken
				config.Token = scmCfg.Token
				return server.NewClient(baseClient, config)
			}

			config.AuthType = server.BasicAuth
			config.Password = scmCfg.Token
			return server.NewClient(baseClient, config)
		}

		config.AuthType = server.BasicAuth
		config.Password = scmCfg.Password
		return server.NewClient(baseClient, config)
	default:
		return nil, cerr.ErrorNotImplemented.Error(fmt.Sprintf("unsupported auth type: %v", scmCfg.AuthType))
	}

	return client, err
}

func getBitbucketType(scmCfg *v1alpha1.SCMSource) bitbucketType {
	if strings.HasPrefix(scmCfg.Server, bitbucketCloudAddr) {
		return bitbucketCloud
	}
	return bitbucketServer
}

// ParseEvent parses data from Bitbucket events.
func ParseEvent(scmCfg *v1alpha1.SCMSource, request *http.Request) *scm.EventData {
	bitbucketType := getBitbucketType(scmCfg)

	switch bitbucketType {
	case bitbucketServer:
		return server.ParseEvent(request)
	case bitbucketCloud:
		// TODO: support bitbucket cloud
		err := fmt.Errorf("bitbucet cloud is unsupported now")
		log.Errorln(err)
		return nil
	default:
		err := fmt.Errorf("unsupported bitbucket product: %v", bitbucketType)
		log.Errorln(err)
		return nil
	}
}

// ParseHookEvent parses data from Bitbucket events.
func ParseHookEvent(request *http.Request) *scm.EventData {
	return server.ParseEvent(request)
}
