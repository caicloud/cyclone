// A CryptoService client wrapper around a remote wrapper service.

package client

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	pb "github.com/docker/notary/proto"
	"github.com/docker/notary/tuf/data"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

// RemotePrivateKey is a key that is on a remote service, so no private
// key bytes are available
type RemotePrivateKey struct {
	data.PublicKey
	sClient pb.SignerClient
}

// RemoteSigner wraps a RemotePrivateKey and implements the crypto.Signer
// interface
type RemoteSigner struct {
	RemotePrivateKey
}

// Public method of a crypto.Signer needs to return a crypto public key.
func (rs *RemoteSigner) Public() crypto.PublicKey {
	publicKey, err := x509.ParsePKIXPublicKey(rs.RemotePrivateKey.Public())
	if err != nil {
		return nil
	}

	return publicKey
}

// NewRemotePrivateKey returns RemotePrivateKey, a data.PrivateKey that is only
// good for signing. (You can't get the private bytes out for instance.)
func NewRemotePrivateKey(pubKey data.PublicKey, sClient pb.SignerClient) *RemotePrivateKey {
	return &RemotePrivateKey{
		PublicKey: pubKey,
		sClient:   sClient,
	}
}

// Private returns nil bytes
func (pk *RemotePrivateKey) Private() []byte {
	return nil
}

// Sign calls a remote service to sign a message.
func (pk *RemotePrivateKey) Sign(rand io.Reader, msg []byte,
	opts crypto.SignerOpts) ([]byte, error) {

	keyID := pb.KeyID{ID: pk.ID()}
	sr := &pb.SignatureRequest{
		Content: msg,
		KeyID:   &keyID,
	}
	sig, err := pk.sClient.Sign(context.Background(), sr)
	if err != nil {
		return nil, err
	}
	return sig.Content, nil
}

// SignatureAlgorithm returns the signing algorithm based on the type of
// PublicKey algorithm.
func (pk *RemotePrivateKey) SignatureAlgorithm() data.SigAlgorithm {
	switch pk.PublicKey.Algorithm() {
	case data.ECDSAKey, data.ECDSAx509Key:
		return data.ECDSASignature
	case data.RSAKey, data.RSAx509Key:
		return data.RSAPSSSignature
	case data.ED25519Key:
		return data.EDDSASignature
	default: // unknown
		return ""
	}
}

// CryptoSigner returns a crypto.Signer tha wraps the RemotePrivateKey. Needed
// for implementing the interface.
func (pk *RemotePrivateKey) CryptoSigner() crypto.Signer {
	return &RemoteSigner{RemotePrivateKey: *pk}
}

// NotarySigner implements a RPC based Trust service that calls the Notary-signer Service
type NotarySigner struct {
	kmClient pb.KeyManagementClient
	sClient  pb.SignerClient
}

// NewGRPCConnection is a convenience method that returns GRPC Client Connection given a hostname, endpoint, and TLS options
func NewGRPCConnection(hostname string, port string, tlsConfig *tls.Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	netAddr := net.JoinHostPort(hostname, port)
	creds := credentials.NewTLS(tlsConfig)
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return grpc.Dial(netAddr, opts...)
}

// NewNotarySigner is a convenience method that returns NotarySigner given a GRPC connection
func NewNotarySigner(conn *grpc.ClientConn) *NotarySigner {
	kmClient := pb.NewKeyManagementClient(conn)
	sClient := pb.NewSignerClient(conn)
	return &NotarySigner{
		kmClient: kmClient,
		sClient:  sClient,
	}
}

// Create creates a remote key and returns the PublicKey associated with the remote private key
func (trust *NotarySigner) Create(role, gun, algorithm string) (data.PublicKey, error) {
	publicKey, err := trust.kmClient.CreateKey(context.Background(),
		&pb.CreateKeyRequest{Algorithm: algorithm, Role: role, Gun: gun})
	if err != nil {
		return nil, err
	}
	public := data.NewPublicKey(publicKey.KeyInfo.Algorithm.Algorithm, publicKey.PublicKey)
	return public, nil
}

// AddKey adds a key
func (trust *NotarySigner) AddKey(role, gun string, k data.PrivateKey) error {
	return errors.New("Adding a key to NotarySigner is not supported")
}

// RemoveKey deletes a key by ID - if the key didn't exist, succeed anyway
func (trust *NotarySigner) RemoveKey(keyid string) error {
	_, err := trust.kmClient.DeleteKey(context.Background(), &pb.KeyID{ID: keyid})
	return err
}

// GetKey retrieves a key by ID - returns nil if the key doesn't exist
func (trust *NotarySigner) GetKey(keyid string) data.PublicKey {
	pubKey, _, err := trust.getKeyInfo(keyid)
	if err != nil {
		return nil
	}
	return pubKey
}

func (trust *NotarySigner) getKeyInfo(keyid string) (data.PublicKey, string, error) {
	keyInfo, err := trust.kmClient.GetKeyInfo(context.Background(), &pb.KeyID{ID: keyid})
	if err != nil {
		return nil, "", err
	}
	return data.NewPublicKey(keyInfo.KeyInfo.Algorithm.Algorithm, keyInfo.PublicKey), keyInfo.Role, nil
}

// GetPrivateKey retrieves by ID an object that can be used to sign, but that does
// not contain any private bytes.  If the key doesn't exist, returns an error.
func (trust *NotarySigner) GetPrivateKey(keyid string) (data.PrivateKey, string, error) {
	pubKey, role, err := trust.getKeyInfo(keyid)
	if err != nil {
		return nil, "", err
	}
	return NewRemotePrivateKey(pubKey, trust.sClient), role, nil
}

// ListKeys not supported for NotarySigner
func (trust *NotarySigner) ListKeys(role string) []string {
	return []string{}
}

// ListAllKeys not supported for NotarySigner
func (trust *NotarySigner) ListAllKeys() map[string]string {
	return map[string]string{}
}

// CheckHealth checks the health of one of the clients, since both clients run
// from the same GRPC server.
func (trust *NotarySigner) CheckHealth(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	status, err := trust.kmClient.CheckHealth(ctx, &pb.Void{})
	defer cancel()
	if err == nil && len(status.Status) > 0 {
		var stats string
		for k, v := range status.Status {
			stats += k + ":" + v + "; "
		}
		return fmt.Errorf("Trust is not healthy: %s", stats)
	}
	if err != nil && grpc.Code(err) == codes.DeadlineExceeded {
		return fmt.Errorf("Timed out reaching trust service after %s.", timeout)
	}

	return err
}
