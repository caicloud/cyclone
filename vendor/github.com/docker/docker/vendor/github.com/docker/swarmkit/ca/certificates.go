package ca

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	cfcsr "github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/initca"
	cflog "github.com/cloudflare/cfssl/log"
	cfsigner "github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
	"github.com/docker/distribution/digest"
	"github.com/docker/go-events"
	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/ioutils"
	"github.com/docker/swarmkit/remotes"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

const (
	// Security Strength Equivalence
	//-----------------------------------
	//| ECC  |  DH/DSA/RSA  |
	//| 256  |     3072     |
	//| 384  |     7680     |
	//-----------------------------------

	// RootKeySize is the default size of the root CA key
	// It would be ideal for the root key to use P-384, but in P-384 is not optimized in go yet :(
	RootKeySize = 256
	// RootKeyAlgo defines the default algorithm for the root CA Key
	RootKeyAlgo = "ecdsa"
	// PassphraseENVVar defines the environment variable to look for the
	// root CA private key material encryption key
	PassphraseENVVar = "SWARM_ROOT_CA_PASSPHRASE"
	// PassphraseENVVarPrev defines the alternate environment variable to look for the
	// root CA private key material encryption key. It can be used for seamless
	// KEK rotations.
	PassphraseENVVarPrev = "SWARM_ROOT_CA_PASSPHRASE_PREV"
	// RootCAExpiration represents the expiration for the root CA in seconds (20 years)
	RootCAExpiration = "630720000s"
	// DefaultNodeCertExpiration represents the default expiration for node certificates (3 months)
	DefaultNodeCertExpiration = 2160 * time.Hour
	// CertBackdate represents the amount of time each certificate is backdated to try to avoid
	// clock drift issues.
	CertBackdate = 1 * time.Hour
	// CertLowerRotationRange represents the minimum fraction of time that we will wait when randomly
	// choosing our next certificate rotation
	CertLowerRotationRange = 0.5
	// CertUpperRotationRange represents the maximum fraction of time that we will wait when randomly
	// choosing our next certificate rotation
	CertUpperRotationRange = 0.8
	// MinNodeCertExpiration represents the minimum expiration for node certificates
	MinNodeCertExpiration = 1 * time.Hour
)

// A recoverableErr is an non-fatal error encountered signing a certificate,
// which means that the certificate issuance may be retried at a later time.
type recoverableErr struct {
	err error
}

func (r recoverableErr) Error() string {
	return r.err.Error()
}

// ErrNoLocalRootCA is an error type used to indicate that the local root CA
// certificate file does not exist.
var ErrNoLocalRootCA = errors.New("local root CA certificate does not exist")

// ErrNoValidSigner is an error type used to indicate that our RootCA doesn't have the ability to
// sign certificates.
var ErrNoValidSigner = recoverableErr{err: errors.New("no valid signer found")}

func init() {
	cflog.Level = 5
}

// CertPaths is a helper struct that keeps track of the paths of a
// Cert and corresponding Key
type CertPaths struct {
	Cert, Key string
}

// RootCA is the representation of everything we need to sign certificates
type RootCA struct {
	// Key will only be used by the original manager to put the private
	// key-material in raft, no signing operations depend on it.
	Key []byte
	// Cert includes the PEM encoded Certificate bundle for the Root CA
	Cert []byte
	Pool *x509.CertPool
	// Digest of the serialized bytes of the certificate
	Digest digest.Digest
	// This signer will be nil if the node doesn't have the appropriate key material
	Signer cfsigner.Signer
	// Path stores the location on disk where the RootCA lives
	Path CertPaths
}

// CanSign ensures that the signer has all three necessary elements needed to operate
func (rca *RootCA) CanSign() bool {
	if rca.Cert == nil || rca.Pool == nil || rca.Signer == nil {
		return false
	}

	return true
}

// IssueAndSaveNewCertificates generates a new key-pair, signs it with the local root-ca, and returns a
// tls certificate
func (rca *RootCA) IssueAndSaveNewCertificates(kw KeyWriter, cn, ou, org string) (*tls.Certificate, error) {
	csr, key, err := GenerateNewCSR()
	if err != nil {
		return nil, errors.Wrap(err, "error when generating new node certs")
	}

	if !rca.CanSign() {
		return nil, ErrNoValidSigner
	}

	// Obtain a signed Certificate
	certChain, err := rca.ParseValidateAndSignCSR(csr, cn, ou, org)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign node certificate")
	}

	// Create a valid TLSKeyPair out of the PEM encoded private key and certificate
	tlsKeyPair, err := tls.X509KeyPair(certChain, key)
	if err != nil {
		return nil, err
	}

	if err := kw.Write(certChain, key, nil); err != nil {
		return nil, err
	}

	return &tlsKeyPair, nil
}

// RequestAndSaveNewCertificates gets new certificates issued, either by signing them locally if a signer is
// available, or by requesting them from the remote server at remoteAddr.
func (rca *RootCA) RequestAndSaveNewCertificates(ctx context.Context, kw KeyWriter, token string, r remotes.Remotes, transport credentials.TransportCredentials, nodeInfo chan<- api.IssueNodeCertificateResponse) (*tls.Certificate, error) {
	// Create a new key/pair and CSR
	csr, key, err := GenerateNewCSR()
	if err != nil {
		return nil, errors.Wrap(err, "error when generating new node certs")
	}

	// Get the remote manager to issue a CA signed certificate for this node
	// Retry up to 5 times in case the manager we first try to contact isn't
	// responding properly (for example, it may have just been demoted).
	var response *api.NodeCertificateStatusResponse
	for i := 0; i != 5; i++ {
		response, err = GetRemoteSignedCertificate(ctx, csr, token, rca.Pool, r, transport, nodeInfo)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// Доверяй, но проверяй.
	// Before we overwrite our local key + certificate, let's make sure the server gave us one that is valid
	// Create an X509Cert so we can .Verify()
	certBlock, _ := pem.Decode(response.Certificate.Certificate)
	if certBlock == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}
	X509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, err
	}
	// We retrieve the certificate with the current root pool, so we know this was issued by a legitimate manager.
	// However, there might have been a server-side root rotation, so we verify this cert with a new pool.
	// If we got a valid response.RootCABundle, turn it into a Pool, and verify the newly minted certificate using it.
	var (
		newRootErr error
		newRootCA  RootCA
	)
	rootCAPool := rca.Pool
	if response.RootCABundle != nil {
		newRootCA, newRootErr = NewRootCA(response.RootCABundle, nil, rca.Path, time.Minute)
		if newRootErr == nil {
			// The response.RootCABundle we got from the remote server seems to be good, use it
			rootCAPool = newRootCA.Pool
		}
	}

	// Create VerifyOptions with either the new certificate bundle, or the old pool
	opts := x509.VerifyOptions{
		Roots: rootCAPool,
	}

	// Check to see if this certificate was signed by one of the CAs, and isn't expired
	if _, err := X509Cert.Verify(opts); err != nil {
		return nil, err
	}

	// Create a valid TLSKeyPair out of the PEM encoded private key and certificate
	tlsKeyPair, err := tls.X509KeyPair(response.Certificate.Certificate, key)
	if err != nil {
		return nil, err
	}

	var kekUpdate *KEKData
	for i := 0; i < 5; i++ {
		kekUpdate, err = rca.getKEKUpdate(ctx, X509Cert, tlsKeyPair, r)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	// If a CA certificate bundle exists it has been validated before. If it's different, let's write it to disk.
	// Root rotation should always happen by appending a new CA cert, and later removing the old one,
	// so it's safer to do it in this order of operations (write root, write certificate)
	if newRootErr == nil && !bytes.Equal(rca.Cert, response.RootCABundle) {
		if err := newRootCA.saveCertificate(); err != nil {
			return nil, err
		}
	}

	if err := kw.Write(response.Certificate.Certificate, key, kekUpdate); err != nil {
		return nil, err
	}

	return &tlsKeyPair, nil
}

func (rca *RootCA) getKEKUpdate(ctx context.Context, cert *x509.Certificate, keypair tls.Certificate, r remotes.Remotes) (*KEKData, error) {
	var managerRole bool
	for _, ou := range cert.Subject.OrganizationalUnit {
		if ou == ManagerRole {
			managerRole = true
			break
		}
	}

	if managerRole {
		mtlsCreds := credentials.NewTLS(&tls.Config{ServerName: CARole, RootCAs: rca.Pool, Certificates: []tls.Certificate{keypair}})
		conn, peer, err := getGRPCConnection(mtlsCreds, r)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		client := api.NewCAClient(conn)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		response, err := client.GetUnlockKey(ctx, &api.GetUnlockKeyRequest{})
		if err != nil {
			if grpc.Code(err) == codes.Unimplemented { // if the server does not support keks, return as if no encryption key was specified
				return &KEKData{}, nil
			}

			r.Observe(peer, -remotes.DefaultObservationWeight)
			return nil, err
		}
		r.Observe(peer, remotes.DefaultObservationWeight)
		return &KEKData{KEK: response.UnlockKey, Version: response.Version.Index}, nil
	}

	// If this is a worker, set to never encrypt. We always want to set to the lock key to nil,
	// in case this was a manager that was demoted to a worker.
	return &KEKData{}, nil
}

// PrepareCSR creates a CFSSL Sign Request based on the given raw CSR and
// overrides the Subject and Hosts with the given extra args.
func PrepareCSR(csrBytes []byte, cn, ou, org string) cfsigner.SignRequest {
	// All managers get added the subject-alt-name of CA, so they can be
	// used for cert issuance.
	hosts := []string{ou, cn}
	if ou == ManagerRole {
		hosts = append(hosts, CARole)
	}

	return cfsigner.SignRequest{
		Request: string(csrBytes),
		// OU is used for Authentication of the node type. The CN has the random
		// node ID.
		Subject: &cfsigner.Subject{CN: cn, Names: []cfcsr.Name{{OU: ou, O: org}}},
		// Adding ou as DNS alt name, so clients can connect to ManagerRole and CARole
		Hosts: hosts,
	}
}

// ParseValidateAndSignCSR returns a signed certificate from a particular rootCA and a CSR.
func (rca *RootCA) ParseValidateAndSignCSR(csrBytes []byte, cn, ou, org string) ([]byte, error) {
	if !rca.CanSign() {
		return nil, ErrNoValidSigner
	}

	signRequest := PrepareCSR(csrBytes, cn, ou, org)

	cert, err := rca.Signer.Sign(signRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign node certificate")
	}

	return rca.AppendFirstRootPEM(cert)
}

// AppendFirstRootPEM appends the first certificate from this RootCA's cert
// bundle to the given cert bundle (which should already be encoded as a series
// of PEM-encoded certificate blocks).
func (rca *RootCA) AppendFirstRootPEM(cert []byte) ([]byte, error) {
	// Append the first root CA Cert to the certificate, to create a valid chain
	// Get the first Root CA Cert on the bundle
	firstRootCA, _, err := helpers.ParseOneCertificateFromPEM(rca.Cert)
	if err != nil {
		return nil, err
	}
	if len(firstRootCA) < 1 {
		return nil, errors.New("no valid Root CA certificates found")
	}
	// Convert the first root CA back to PEM
	firstRootCAPEM := helpers.EncodeCertificatePEM(firstRootCA[0])
	if firstRootCAPEM == nil {
		return nil, errors.New("error while encoding the Root CA certificate")
	}
	// Append this Root CA to the certificate to make [Cert PEM]\n[Root PEM][EOF]
	certChain := append(cert, firstRootCAPEM...)

	return certChain, nil
}

func (rca *RootCA) saveCertificate() error {
	if rca.Cert == nil {
		return errors.New("no valid certificate bundle found")

	}
	if rca.Path.Cert == "" {
		return errors.New("no path found for this root CA")
	}

	// Make sure the necessary dirs exist and they are writable
	err := os.MkdirAll(filepath.Dir(rca.Path.Cert), 0755)
	if err != nil {
		return err
	}

	return ioutils.AtomicWriteFile(rca.Path.Cert, rca.Cert, 0644)
}

// NewRootCA creates a new RootCA object from unparsed PEM cert bundle and key byte
// slices. key may be nil, and in this case NewRootCA will return a RootCA
// without a signer.
func NewRootCA(certBytes, keyBytes []byte, paths CertPaths, certExpiry time.Duration) (RootCA, error) {
	// Parse all the certificates in the cert bundle
	parsedCerts, err := helpers.ParseCertificatesPEM(certBytes)
	if err != nil {
		return RootCA{}, err
	}
	// Check to see if we have at least one valid cert
	if len(parsedCerts) < 1 {
		return RootCA{}, errors.New("no valid Root CA certificates found")
	}

	// Create a Pool with all of the certificates found
	pool := x509.NewCertPool()
	for _, cert := range parsedCerts {
		// Check to see if all of the certificates are valid, self-signed root CA certs
		if err := cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature); err != nil {
			return RootCA{}, errors.Wrap(err, "error while validating Root CA Certificate")
		}
		pool.AddCert(cert)
	}

	// Calculate the digest for our Root CA bundle
	digest := digest.FromBytes(certBytes)

	if len(keyBytes) == 0 {
		// This RootCA does not have a valid signer.
		return RootCA{Cert: certBytes, Digest: digest, Pool: pool, Path: paths}, nil
	}

	var (
		passphraseStr              string
		passphrase, passphrasePrev []byte
		priv                       crypto.Signer
	)

	// Attempt two distinct passphrases, so we can do a hitless passphrase rotation
	if passphraseStr = os.Getenv(PassphraseENVVar); passphraseStr != "" {
		passphrase = []byte(passphraseStr)
	}

	if p := os.Getenv(PassphraseENVVarPrev); p != "" {
		passphrasePrev = []byte(p)
	}

	// Attempt to decrypt the current private-key with the passphrases provided
	priv, err = helpers.ParsePrivateKeyPEMWithPassword(keyBytes, passphrase)
	if err != nil {
		priv, err = helpers.ParsePrivateKeyPEMWithPassword(keyBytes, passphrasePrev)
		if err != nil {
			return RootCA{}, errors.Wrap(err, "malformed private key")
		}
	}

	// We will always use the first certificate inside of the root bundle as the active one
	if err := ensureCertKeyMatch(parsedCerts[0], priv.Public()); err != nil {
		return RootCA{}, err
	}

	signer, err := local.NewSigner(priv, parsedCerts[0], cfsigner.DefaultSigAlgo(priv), SigningPolicy(certExpiry))
	if err != nil {
		return RootCA{}, err
	}

	// If the key was loaded from disk unencrypted, but there is a passphrase set,
	// ensure it is encrypted, so it doesn't hit raft in plain-text
	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		// This RootCA does not have a valid signer.
		return RootCA{Cert: certBytes, Digest: digest, Pool: pool, Path: paths}, nil
	}
	if passphraseStr != "" && !x509.IsEncryptedPEMBlock(keyBlock) {
		keyBytes, err = EncryptECPrivateKey(keyBytes, passphraseStr)
		if err != nil {
			return RootCA{}, err
		}
	}

	return RootCA{Signer: signer, Key: keyBytes, Digest: digest, Cert: certBytes, Pool: pool, Path: paths}, nil
}

func ensureCertKeyMatch(cert *x509.Certificate, key crypto.PublicKey) error {
	switch certPub := cert.PublicKey.(type) {
	// TODO: Handle RSA keys.
	case *ecdsa.PublicKey:
		ecKey, ok := key.(*ecdsa.PublicKey)
		if ok && certPub.X.Cmp(ecKey.X) == 0 && certPub.Y.Cmp(ecKey.Y) == 0 {
			return nil
		}
	default:
		return errors.New("unknown or unsupported certificate public key algorithm")
	}

	return errors.New("certificate key mismatch")
}

// GetLocalRootCA returns the PEM-encoded root CA Certificate if it exists
func GetLocalRootCA(paths CertPaths) (RootCA, error) {
	// Check if we have a Certificate file
	cert, err := ioutil.ReadFile(paths.Cert)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNoLocalRootCA
		}

		return RootCA{}, err
	}

	return NewRootCA(cert, nil, paths, DefaultNodeCertExpiration)
}

func getGRPCConnection(creds credentials.TransportCredentials, r remotes.Remotes) (*grpc.ClientConn, api.Peer, error) {
	peer, err := r.Select()
	if err != nil {
		return nil, api.Peer{}, err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithTimeout(5 * time.Second),
		grpc.WithBackoffMaxDelay(5 * time.Second),
	}

	conn, err := grpc.Dial(peer.Addr, opts...)
	if err != nil {
		return nil, api.Peer{}, err
	}
	return conn, peer, nil
}

// GetRemoteCA returns the remote endpoint's CA certificate
func GetRemoteCA(ctx context.Context, d digest.Digest, r remotes.Remotes) (RootCA, error) {
	// This TLS Config is intentionally using InsecureSkipVerify. We use the
	// digest instead to check the integrity of the CA certificate.
	insecureCreds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, peer, err := getGRPCConnection(insecureCreds, r)
	if err != nil {
		return RootCA{}, err
	}
	defer conn.Close()

	client := api.NewCAClient(conn)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	defer func() {
		if err != nil {
			r.Observe(peer, -remotes.DefaultObservationWeight)
			return
		}
		r.Observe(peer, remotes.DefaultObservationWeight)
	}()
	response, err := client.GetRootCACertificate(ctx, &api.GetRootCACertificateRequest{})
	if err != nil {
		return RootCA{}, err
	}

	if d != "" {
		verifier, err := digest.NewDigestVerifier(d)
		if err != nil {
			return RootCA{}, errors.Wrap(err, "unexpected error getting digest verifier")
		}

		io.Copy(verifier, bytes.NewReader(response.Certificate))

		if !verifier.Verified() {
			return RootCA{}, errors.Errorf("remote CA does not match fingerprint. Expected: %s", d.Hex())

		}
	}

	// Check the validity of the remote Cert
	_, err = helpers.ParseCertificatePEM(response.Certificate)
	if err != nil {
		return RootCA{}, err
	}

	// Create a Pool with our RootCACertificate
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(response.Certificate) {
		return RootCA{}, errors.New("failed to append certificate to cert pool")
	}

	return RootCA{Cert: response.Certificate, Digest: digest.FromBytes(response.Certificate), Pool: pool}, nil
}

// CreateRootCA creates a Certificate authority for a new Swarm Cluster, potentially
// overwriting any existing CAs.
func CreateRootCA(rootCN string, paths CertPaths) (RootCA, error) {
	// Create a simple CSR for the CA using the default CA validator and policy
	req := cfcsr.CertificateRequest{
		CN:         rootCN,
		KeyRequest: &cfcsr.BasicKeyRequest{A: RootKeyAlgo, S: RootKeySize},
		CA:         &cfcsr.CAConfig{Expiry: RootCAExpiration},
	}

	// Generate the CA and get the certificate and private key
	cert, _, key, err := initca.New(&req)
	if err != nil {
		return RootCA{}, err
	}

	rootCA, err := NewRootCA(cert, key, paths, DefaultNodeCertExpiration)
	if err != nil {
		return RootCA{}, err
	}

	// save the cert to disk
	if err := rootCA.saveCertificate(); err != nil {
		return RootCA{}, err
	}

	return rootCA, nil
}

// GetRemoteSignedCertificate submits a CSR to a remote CA server address,
// and that is part of a CA identified by a specific certificate pool.
func GetRemoteSignedCertificate(ctx context.Context, csr []byte, token string, rootCAPool *x509.CertPool, r remotes.Remotes, creds credentials.TransportCredentials, nodeInfo chan<- api.IssueNodeCertificateResponse) (*api.NodeCertificateStatusResponse, error) {
	if rootCAPool == nil {
		return nil, errors.New("valid root CA pool required")
	}

	if creds == nil {
		// This is our only non-MTLS request, and it happens when we are boostraping our TLS certs
		// We're using CARole as server name, so an external CA doesn't also have to have ManagerRole in the cert SANs
		creds = credentials.NewTLS(&tls.Config{ServerName: CARole, RootCAs: rootCAPool})
	}

	conn, peer, err := getGRPCConnection(creds, r)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Create a CAClient to retrieve a new Certificate
	caClient := api.NewNodeCAClient(conn)

	// Send the Request and retrieve the request token
	issueRequest := &api.IssueNodeCertificateRequest{CSR: csr, Token: token}
	issueResponse, err := caClient.IssueNodeCertificate(ctx, issueRequest)
	if err != nil {
		r.Observe(peer, -remotes.DefaultObservationWeight)
		return nil, err
	}

	// Send back the NodeID on the nodeInfo, so the caller can know what ID was assigned by the CA
	if nodeInfo != nil {
		nodeInfo <- *issueResponse
	}

	statusRequest := &api.NodeCertificateStatusRequest{NodeID: issueResponse.NodeID}
	expBackoff := events.NewExponentialBackoff(events.ExponentialBackoffConfig{
		Base:   time.Second,
		Factor: time.Second,
		Max:    30 * time.Second,
	})

	// Exponential backoff with Max of 30 seconds to wait for a new retry
	for {
		// Send the Request and retrieve the certificate
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		statusResponse, err := caClient.NodeCertificateStatus(ctx, statusRequest)
		if err != nil {
			r.Observe(peer, -remotes.DefaultObservationWeight)
			return nil, err
		}

		// If the certificate was issued, return
		if statusResponse.Status != nil && statusResponse.Status.State == api.IssuanceStateIssued {
			if statusResponse.Certificate == nil {
				return nil, errors.New("no certificate in CertificateStatus response")
			}

			// The certificate in the response must match the CSR
			// we submitted. If we are getting a response for a
			// certificate that was previously issued, we need to
			// retry until the certificate gets updated per our
			// current request.
			if bytes.Equal(statusResponse.Certificate.CSR, csr) {
				r.Observe(peer, remotes.DefaultObservationWeight)
				return statusResponse, nil
			}
		}

		// If we're still pending, the issuance failed, or the state is unknown
		// let's continue trying.
		expBackoff.Failure(nil, nil)
		time.Sleep(expBackoff.Proceed(nil))
	}
}

// readCertValidity returns the certificate issue and expiration time
func readCertValidity(kr KeyReader) (time.Time, time.Time, error) {
	var zeroTime time.Time
	// Read the Cert
	cert, _, err := kr.Read()
	if err != nil {
		return zeroTime, zeroTime, err
	}

	// Create an x509 certificate out of the contents on disk
	certBlock, _ := pem.Decode(cert)
	if certBlock == nil {
		return zeroTime, zeroTime, errors.New("failed to decode certificate block")
	}
	X509Cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return zeroTime, zeroTime, err
	}

	return X509Cert.NotBefore, X509Cert.NotAfter, nil

}

// GenerateNewCSR returns a newly generated key and CSR signed with said key
func GenerateNewCSR() (csr, key []byte, err error) {
	req := &cfcsr.CertificateRequest{
		KeyRequest: cfcsr.NewBasicKeyRequest(),
	}

	csr, key, err = cfcsr.ParseRequest(req)
	if err != nil {
		return
	}

	return
}

// EncryptECPrivateKey receives a PEM encoded private key and returns an encrypted
// AES256 version using a passphrase
// TODO: Make this method generic to handle RSA keys
func EncryptECPrivateKey(key []byte, passphraseStr string) ([]byte, error) {
	passphrase := []byte(passphraseStr)
	cipherType := x509.PEMCipherAES256

	keyBlock, _ := pem.Decode(key)
	if keyBlock == nil {
		// This RootCA does not have a valid signer.
		return nil, errors.New("error while decoding PEM key")
	}

	encryptedPEMBlock, err := x509.EncryptPEMBlock(rand.Reader,
		"EC PRIVATE KEY",
		keyBlock.Bytes,
		passphrase,
		cipherType)
	if err != nil {
		return nil, err
	}

	if encryptedPEMBlock.Headers == nil {
		return nil, errors.New("unable to encrypt key - invalid PEM file produced")
	}

	return pem.EncodeToMemory(encryptedPEMBlock), nil
}
