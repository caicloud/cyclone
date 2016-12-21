package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
)

const (
	stateInitial = iota
	stateName
	statePort
	stateTag
)

// Image represents Docker image
type Image struct {
	Registry string
	Name     string
	Tag      string
	FsLayers []FsLayer
	Token    string
	user     string
	password string
}

// FsLayer represents a layer in docker image
type FsLayer struct {
	BlobSum string
}

const dockerHub = "registry-1.docker.io"

var client = &http.Client{}
var tokenRe = regexp.MustCompile(`Bearer realm="(.*?)",service="(.*?)",scope="(.*?)"`)

// NewImage parses image name which could be the ful name registry:port/name:tag
// or in any other shorter forms and creates docker image entity without
// information about layers
func NewImage(qname, user, password string) (*Image, error) {
	registry := dockerHub
	tag := "latest"
	var name, port string
	state := stateInitial
	start := 0
	for i, c := range qname {
		if c == ':' || c == '/' || i == len(qname)-1 {
			if i == len(qname)-1 {
				// ignore a separator, include the last symbol
				i += 1
			}
			part := qname[start:i]
			start = i + 1
			switch state {
			case stateInitial:
				addrs, err := net.LookupHost(part)
				// not a hostname?
				if err != nil || len(addrs) == 0 {
					// it's an image name, if separator is /
					// next part is also part of the name
					// othrewise it's an offcial image
					if c == '/' {
						// we got just a part of name, till next time
						start = 0
						state = stateName
					} else {
						state = stateTag
						name = fmt.Sprintf("library/%s", part)
					}
				} else {
					// it's registry, let's check what's next =port of image name
					registry = part
					if c == ':' {
						state = statePort
					} else {
						state = stateName
					}
				}
			case stateTag:
				tag = part
			case statePort:
				state = stateName
				port = part
			case stateName:
				if c == ':' {
					state = stateTag
				}
				name = part
			}
		}
	}

	if port != "" {
		registry = fmt.Sprintf("%s:%s", registry, port)
	}

	registry = fmt.Sprintf("https://%s/v2", registry)
	return &Image{
		Registry: registry,
		Name:     name,
		Tag:      tag,
		user:     user,
		password: password,
	}, nil
}

// Pull retrieves information about layers from docker registry.
// It gets docker registry token if needed.
func (i *Image) Pull() error {
	resp, err := i.pullReq()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		i.Token, err = i.requestToken(resp)
		io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return err
		}
		// try again
		resp, err = i.pullReq()
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	if err = json.NewDecoder(resp.Body).Decode(i); err != nil {
		fmt.Println("Decode error")
		return err
	}
	return nil
}

func (i *Image) requestToken(resp *http.Response) (string, error) {
	authHeader := resp.Header.Get("Www-Authenticate")
	if authHeader == "" {
		return "", fmt.Errorf("Empty Www-Authenticate")
	}
	parts := tokenRe.FindStringSubmatch(authHeader)
	if parts == nil {
		return "", fmt.Errorf("Can't parse Www-Authenticate: %s", authHeader)
	}
	realm, service, scope := parts[1], parts[2], parts[3]
	url := fmt.Sprintf("%s?service=%s&scope=%s&account=%s", realm, service, scope, i.user)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Can't create a request")
		return "", err
	}
	req.SetBasicAuth(i.user, i.password)
	tResp, err := client.Do(req)
	if err != nil {
		io.Copy(ioutil.Discard, tResp.Body)
		return "", err
	}

	defer tResp.Body.Close()
	if tResp.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, tResp.Body)
		return "", fmt.Errorf("Token request returned %d", tResp.StatusCode)
	}
	var tokenEnv struct {
		Token string
	}

	if err = json.NewDecoder(tResp.Body).Decode(&tokenEnv); err != nil {
		fmt.Println("Token response decode error")
		return "", err
	}
	return fmt.Sprintf("Bearer %s", tokenEnv.Token), nil
}

func (i *Image) pullReq() (*http.Response, error) {
	url := fmt.Sprintf("%s/%s/manifests/%s", i.Registry, i.Name, i.Tag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Can't create a request")
		return nil, err
	}
	if i.Token == "" {
		req.SetBasicAuth(i.user, i.password)
	} else {
		req.Header.Set("Authorization", i.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Get error")
		return nil, err
	}
	return resp, nil
}

func dumpRequest(r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Printf("Can't dump HTTP request %s\n", err.Error())
	} else {
		fmt.Printf("request_dump: %s\n", string(dump[:]))
	}
}

func dumpResponse(r *http.Response) {
	dump, err := httputil.DumpResponse(r, true)
	if err != nil {
		fmt.Printf("Can't dump HTTP reqsponse %s\n", err.Error())
	} else {
		fmt.Printf("response_dump: %s\n", string(dump[:]))
	}
}
