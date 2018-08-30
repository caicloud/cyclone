package token

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/caicloud/cargo-admin/pkg/api/token/types"
	"github.com/caicloud/cargo-admin/pkg/cauth"
	"github.com/caicloud/cargo-admin/pkg/models"

	"github.com/caicloud/nirvana/log"

	"github.com/caicloud/cargo-admin/pkg/env"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
)

const (
	harborRegistry = "harbor-registry"
	issuer         = "harbor-token-issuer"
)

func Get(domain, user, pwd string, scopes []string) (*types.Token, error) {
	regInfo, err := models.Registry.FindByDomain(domain)
	if err != nil {
		log.Errorf("find registry by domain: %s, error:%v", domain, err)
		return nil, err
	}

	auth := &basicAuth{Username: user, Password: pwd, Scope: scopes}
	if user != "" {
		if err = auth.ValidateUser(regInfo); err != nil {
			// If scopes is empty, it's login request, we should return error
			if len(scopes) == 0 {
				return nil, err
			}

			// If the provided user/password is invalid, fallback to anonymous access by setting
			// user information to empty, with this way, pull public project would be permitted.
			log.Warningf("validate user '%s' failed, fallback to anonymous access", user)
			auth.Username = ""
			auth.Password = ""
		}
	}

	accesses := filterActions(GetActions(scopes))
	cauthCli := cauth.NewClient(env.CauthAddress, regInfo.Name)
	err = CheckPerm(&permManager{auth, regInfo, cauthCli}, accesses)
	if err != nil {
		return nil, err
	}

	return makeToken(user, harborRegistry, accesses)
}

// When push image to one project, registry may mount layers from other projects. We should allow
// pull requests from these other projects in this case. When this happened, the scopes look like:
// [repository:devops_gdfd/easybox:push,pull repository:devops_time/easybox:pull]
func filterActions(actions []*token.ResourceActions) []*token.ResourceActions {
	if len(actions) <= 1 {
		return actions
	}

	filtered := make([]*token.ResourceActions, 0)
	for i, s := range actions {
		if i > 0 && len(s.Actions) == 1 && s.Actions[0] == "pull" {
			log.Infof("referring scope '%s:%s:pull' skipped", s.Type, s.Name)
			continue
		}
		filtered = append(filtered, s)
	}

	return filtered
}

// MakeToken makes a valid jwt token based on parms
func makeToken(username string, service string, accesses []*token.ResourceActions) (*types.Token, error) {
	pk, err := libtrust.LoadKeyFile(env.PrivateKeyFile)
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	tk, expiresIn, issuedAt, err := makeTokenCore(issuer, username, service, env.TokenExpiration, accesses, pk)
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}
	rs := fmt.Sprintf("%s.%s", tk.Raw, base64UrlEncode(tk.Signature))
	return &types.Token{
		Token:     rs,
		ExpiresIn: expiresIn,
		IssuedAt:  issuedAt.Format(time.RFC3339),
	}, nil
}

//make token core
func makeTokenCore(issuer, subject, audience string, expiration int,
	access []*token.ResourceActions, signingKey libtrust.PrivateKey) (t *token.Token, expiresIn int, issuedAt *time.Time, err error) {

	joseHeader := &token.Header{
		Type:       "JWT",
		SigningAlg: "RS256",
		KeyID:      signingKey.KeyID(),
	}

	jwtID, err := randString(16)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Error to generate jwt id: %s", err)
	}

	now := time.Now().UTC()
	issuedAt = &now
	expiresIn = expiration * 60

	claimSet := &token.ClaimSet{
		Issuer:     issuer,
		Subject:    subject,
		Audience:   audience,
		Expiration: now.Add(time.Duration(expiration) * time.Minute).Unix(),
		NotBefore:  now.Unix(),
		IssuedAt:   now.Unix(),
		JWTID:      jwtID,
		Access:     access,
	}

	var joseHeaderBytes, claimSetBytes []byte

	if joseHeaderBytes, err = json.Marshal(joseHeader); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to marshal claim set: %s", err)
	}

	encodedJoseHeader := base64UrlEncode(joseHeaderBytes)
	encodedClaimSet := base64UrlEncode(claimSetBytes)
	payload := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)

	var signatureBytes []byte
	if signatureBytes, _, err = signingKey.Sign(strings.NewReader(payload), crypto.SHA256); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to sign jwt payload: %s", err)
	}

	signature := base64UrlEncode(signatureBytes)
	tokenString := fmt.Sprintf("%s.%s", payload, signature)
	t, err = token.NewToken(tokenString)
	return
}

func randString(length int) (string, error) {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	for i, b := range rb {
		rb[i] = alphanum[int(b)%len(alphanum)]
	}
	return string(rb), nil
}

func base64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
