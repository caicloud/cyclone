package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"github.com/docker/notary"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const cannedPassphrase = "passphrase"

var passphraseRetriever = func(keyID string, alias string, createNew bool, numAttempts int) (string, bool, error) {
	if numAttempts > 5 {
		giveup := true
		return "", giveup, errors.New("passPhraseRetriever failed after too many requests")
	}
	return cannedPassphrase, false, nil
}

type TestImportStore struct {
	data map[string][]byte
}

func NewTestImportStore() *TestImportStore {
	return &TestImportStore{
		data: make(map[string][]byte),
	}
}

func (s *TestImportStore) Set(name string, data []byte) error {
	s.data[name] = data
	return nil
}

type TestExportStore struct {
	data map[string][]byte
}

func NewTestExportStore() *TestExportStore {
	return &TestExportStore{
		data: make(map[string][]byte),
	}
}

func (s *TestExportStore) Get(name string) ([]byte, error) {
	if data, ok := s.data[name]; ok {
		return data, nil
	}
	return nil, errors.New("Not Found")
}

func (s *TestExportStore) ListFiles() []string {
	files := make([]string, 0, len(s.data))
	for k := range s.data {
		files = append(files, k)
	}
	return files
}

func TestExportKeys(t *testing.T) {
	s := NewTestExportStore()

	b := &pem.Block{}
	b.Bytes = make([]byte, 1000)
	rand.Read(b.Bytes)

	c := &pem.Block{}
	c.Bytes = make([]byte, 1000)
	rand.Read(c.Bytes)

	bBytes := pem.EncodeToMemory(b)
	cBytes := pem.EncodeToMemory(c)

	s.data["ankh"] = bBytes
	s.data["morpork"] = cBytes

	buf := bytes.NewBuffer(nil)

	err := ExportKeys(buf, s, "ankh")
	require.NoError(t, err)

	err = ExportKeys(buf, s, "morpork")
	require.NoError(t, err)

	out, err := ioutil.ReadAll(buf)
	require.NoError(t, err)

	bFinal, rest := pem.Decode(out)
	require.Equal(t, b.Bytes, bFinal.Bytes)
	require.Equal(t, "ankh", bFinal.Headers["path"])

	cFinal, rest := pem.Decode(rest)
	require.Equal(t, c.Bytes, cFinal.Bytes)
	require.Equal(t, "morpork", cFinal.Headers["path"])
	require.Len(t, rest, 0)
}

func TestExportKeysByGUN(t *testing.T) {
	s := NewTestExportStore()

	b := &pem.Block{}
	b.Bytes = make([]byte, 1000)
	rand.Read(b.Bytes)

	b2 := &pem.Block{}
	b2.Bytes = make([]byte, 1000)
	rand.Read(b2.Bytes)

	c := &pem.Block{}
	c.Bytes = make([]byte, 1000)
	rand.Read(c.Bytes)

	bBytes := pem.EncodeToMemory(b)
	b2Bytes := pem.EncodeToMemory(b2)
	cBytes := pem.EncodeToMemory(c)

	s.data["ankh/one"] = bBytes
	s.data["ankh/two"] = b2Bytes
	s.data["morpork/three"] = cBytes

	buf := bytes.NewBuffer(nil)

	err := ExportKeysByGUN(buf, s, "ankh")
	require.NoError(t, err)

	out, err := ioutil.ReadAll(buf)
	require.NoError(t, err)

	bFinal, rest := pem.Decode(out)
	require.Equal(t, b.Bytes, bFinal.Bytes)
	require.Equal(t, "ankh/one", bFinal.Headers["path"])

	b2Final, rest := pem.Decode(rest)
	require.Equal(t, b2.Bytes, b2Final.Bytes)
	require.Equal(t, "ankh/two", b2Final.Headers["path"])
	require.Len(t, rest, 0)
}

func TestExportKeysByID(t *testing.T) {
	s := NewTestExportStore()

	b := &pem.Block{}
	b.Bytes = make([]byte, 1000)
	rand.Read(b.Bytes)

	c := &pem.Block{}
	c.Bytes = make([]byte, 1000)
	rand.Read(c.Bytes)

	bBytes := pem.EncodeToMemory(b)
	cBytes := pem.EncodeToMemory(c)

	s.data["ankh"] = bBytes
	s.data["morpork/identifier"] = cBytes

	buf := bytes.NewBuffer(nil)

	err := ExportKeysByID(buf, s, []string{"identifier"})
	require.NoError(t, err)

	out, err := ioutil.ReadAll(buf)
	require.NoError(t, err)

	cFinal, rest := pem.Decode(out)
	require.Equal(t, c.Bytes, cFinal.Bytes)
	require.Equal(t, "morpork/identifier", cFinal.Headers["path"])
	require.Len(t, rest, 0)
}

func TestExport2InOneFile(t *testing.T) {
	s := NewTestExportStore()

	b := &pem.Block{}
	b.Bytes = make([]byte, 1000)
	rand.Read(b.Bytes)

	b2 := &pem.Block{}
	b2.Bytes = make([]byte, 1000)
	rand.Read(b2.Bytes)

	c := &pem.Block{}
	c.Bytes = make([]byte, 1000)
	rand.Read(c.Bytes)

	bBytes := pem.EncodeToMemory(b)
	b2Bytes := pem.EncodeToMemory(b2)
	bBytes = append(bBytes, b2Bytes...)
	cBytes := pem.EncodeToMemory(c)

	s.data["ankh"] = bBytes
	s.data["morpork"] = cBytes

	buf := bytes.NewBuffer(nil)

	err := ExportKeys(buf, s, "ankh")
	require.NoError(t, err)

	err = ExportKeys(buf, s, "morpork")
	require.NoError(t, err)

	out, err := ioutil.ReadAll(buf)
	require.NoError(t, err)

	bFinal, rest := pem.Decode(out)
	require.Equal(t, b.Bytes, bFinal.Bytes)
	require.Equal(t, "ankh", bFinal.Headers["path"])

	b2Final, rest := pem.Decode(rest)
	require.Equal(t, b2.Bytes, b2Final.Bytes)
	require.Equal(t, "ankh", b2Final.Headers["path"])

	cFinal, rest := pem.Decode(rest)
	require.Equal(t, c.Bytes, cFinal.Bytes)
	require.Equal(t, "morpork", cFinal.Headers["path"])
	require.Len(t, rest, 0)
}

func TestImportKeys(t *testing.T) {
	s := NewTestImportStore()

	from, _ := os.OpenFile("../fixtures/secure.example.com.key", os.O_RDONLY, notary.PrivKeyPerms)
	b := &pem.Block{
		Headers: make(map[string]string),
	}
	b.Bytes, _ = ioutil.ReadAll(from)
	rand.Read(b.Bytes)
	b.Headers["path"] = "ankh"

	c := &pem.Block{
		Headers: make(map[string]string),
	}
	c.Bytes, _ = ioutil.ReadAll(from)
	rand.Read(c.Bytes)
	c.Headers["path"] = "morpork"
	c.Headers["role"] = data.CanonicalSnapshotRole
	c.Headers["gun"] = "somegun"

	bBytes := pem.EncodeToMemory(b)
	cBytes := pem.EncodeToMemory(c)

	byt := append(bBytes, cBytes...)

	in := bytes.NewBuffer(byt)

	err := ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.NoError(t, err)

	bFinal, bRest := pem.Decode(s.data["ankh"])
	require.Equal(t, b.Bytes, bFinal.Bytes)
	_, ok := bFinal.Headers["path"]
	require.False(t, ok, "expected no path header, should have been removed at import")
	require.Equal(t, notary.DefaultImportRole, bFinal.Headers["role"]) // if no role is specified we assume it is a delegation key
	_, ok = bFinal.Headers["gun"]
	require.False(t, ok, "expected no gun header, should have been removed at import")
	require.Len(t, bRest, 0)

	cFinal, cRest := pem.Decode(s.data["morpork"])
	require.Equal(t, c.Bytes, cFinal.Bytes)
	_, ok = cFinal.Headers["path"]
	require.False(t, ok, "expected no path header, should have been removed at import")
	require.Equal(t, data.CanonicalSnapshotRole, cFinal.Headers["role"])
	require.Equal(t, "somegun", cFinal.Headers["gun"])
	require.Len(t, cRest, 0)
}

func TestImportNoPath(t *testing.T) {
	s := NewTestImportStore()

	from, _ := os.OpenFile("../fixtures/secure.example.com.key", os.O_RDONLY, notary.PrivKeyPerms)
	defer from.Close()
	fromBytes, _ := ioutil.ReadAll(from)

	in := bytes.NewBuffer(fromBytes)

	err := ImportKeys(in, []Importer{s}, data.CanonicalRootRole, "", passphraseRetriever)
	require.NoError(t, err)

	for key := range s.data {
		// no path but role included should work
		require.Equal(t, key, filepath.Join(notary.RootKeysSubdir, "12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497"))
	}

	s = NewTestImportStore()

	err = ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.NoError(t, err)

	require.Len(t, s.data, 0) // no path and no role should not work
}

func TestNonRootPathInference(t *testing.T) {
	s := NewTestImportStore()

	from, _ := os.OpenFile("../fixtures/secure.example.com.key", os.O_RDONLY, notary.PrivKeyPerms)
	defer from.Close()
	fromBytes, _ := ioutil.ReadAll(from)

	in := bytes.NewBuffer(fromBytes)

	err := ImportKeys(in, []Importer{s}, data.CanonicalSnapshotRole, "somegun", passphraseRetriever)
	require.NoError(t, err)

	for key := range s.data {
		// no path but role included should work and 12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497 is the key ID of the fixture
		require.Equal(t, key, filepath.Join(notary.NonRootKeysSubdir, "somegun", "12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497"))
	}
}

func TestBlockHeaderPrecedenceRoleAndGun(t *testing.T) {
	s := NewTestImportStore()

	from, _ := os.OpenFile("../fixtures/secure.example.com.key", os.O_RDONLY, notary.PrivKeyPerms)
	defer from.Close()
	fromBytes, _ := ioutil.ReadAll(from)
	b, _ := pem.Decode(fromBytes)
	b.Headers["role"] = data.CanonicalSnapshotRole
	b.Headers["gun"] = "anothergun"
	bBytes := pem.EncodeToMemory(b)

	in := bytes.NewBuffer(bBytes)

	err := ImportKeys(in, []Importer{s}, "somerole", "somegun", passphraseRetriever)
	require.NoError(t, err)

	require.Len(t, s.data, 1)
	for key := range s.data {
		// block header role= root should take precedence over command line flag
		require.Equal(t, key, filepath.Join(notary.NonRootKeysSubdir, "anothergun", "12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497"))
		final, rest := pem.Decode(s.data[key])
		require.Len(t, rest, 0)
		require.Equal(t, final.Headers["role"], "snapshot")
		require.Equal(t, final.Headers["gun"], "anothergun")
	}
}

func TestBlockHeaderPrecedenceGunFromPath(t *testing.T) {
	s := NewTestImportStore()

	from, _ := os.OpenFile("../fixtures/secure.example.com.key", os.O_RDONLY, notary.PrivKeyPerms)
	defer from.Close()
	fromBytes, _ := ioutil.ReadAll(from)
	b, _ := pem.Decode(fromBytes)
	b.Headers["role"] = data.CanonicalSnapshotRole
	b.Headers["path"] = filepath.Join(notary.NonRootKeysSubdir, "anothergun", "12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497")
	bBytes := pem.EncodeToMemory(b)

	in := bytes.NewBuffer(bBytes)

	err := ImportKeys(in, []Importer{s}, "somerole", "somegun", passphraseRetriever)
	require.NoError(t, err)

	require.Len(t, s.data, 1)
	for key := range s.data {
		// block header role= root should take precedence over command line flag
		require.Equal(t, key, filepath.Join(notary.NonRootKeysSubdir, "anothergun", "12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497"))
		final, rest := pem.Decode(s.data[key])
		require.Len(t, rest, 0)
		require.Equal(t, final.Headers["role"], "snapshot")
		require.Equal(t, final.Headers["gun"], "anothergun")
	}
}

func TestImportKeys2InOneFile(t *testing.T) {
	s := NewTestImportStore()

	b := &pem.Block{
		Headers: make(map[string]string),
	}
	b.Bytes = make([]byte, 1000)
	rand.Read(b.Bytes)
	b.Headers["path"] = "ankh"

	b2 := &pem.Block{
		Headers: make(map[string]string),
	}
	b2.Bytes = make([]byte, 1000)
	rand.Read(b2.Bytes)
	b2.Headers["path"] = "ankh"

	c := &pem.Block{
		Headers: make(map[string]string),
	}
	c.Bytes = make([]byte, 1000)
	rand.Read(c.Bytes)
	c.Headers["path"] = "morpork"

	bBytes := pem.EncodeToMemory(b)
	b2Bytes := pem.EncodeToMemory(b2)
	bBytes = append(bBytes, b2Bytes...)
	cBytes := pem.EncodeToMemory(c)

	byt := append(bBytes, cBytes...)

	in := bytes.NewBuffer(byt)

	err := ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.NoError(t, err)

	bFinal, bRest := pem.Decode(s.data["ankh"])
	require.Equal(t, b.Bytes, bFinal.Bytes)
	_, ok := bFinal.Headers["path"]
	require.False(t, ok, "expected no path header, should have been removed at import")
	role, _ := bFinal.Headers["role"]
	require.Equal(t, notary.DefaultImportRole, role)

	b2Final, b2Rest := pem.Decode(bRest)
	require.Equal(t, b2.Bytes, b2Final.Bytes)
	_, ok = b2Final.Headers["path"]
	require.False(t, ok, "expected no path header, should have been removed at import")
	require.Len(t, b2Rest, 0)

	cFinal, cRest := pem.Decode(s.data["morpork"])
	require.Equal(t, c.Bytes, cFinal.Bytes)
	_, ok = cFinal.Headers["path"]
	require.False(t, ok, "expected no path header, should have been removed at import")
	require.Len(t, cRest, 0)
}

func TestImportKeys2InOneFileNoPath(t *testing.T) {
	s := NewTestImportStore()

	from, _ := os.OpenFile("../fixtures/secure.example.com.key", os.O_RDONLY, notary.PrivKeyPerms)
	defer from.Close()
	fromBytes, _ := ioutil.ReadAll(from)
	b, _ := pem.Decode(fromBytes)
	b.Headers["gun"] = "testgun"
	b.Headers["role"] = data.CanonicalSnapshotRole
	bBytes := pem.EncodeToMemory(b)

	b2, _ := pem.Decode(fromBytes)
	b2.Headers["gun"] = "testgun"
	b2.Headers["role"] = data.CanonicalSnapshotRole
	b2Bytes := pem.EncodeToMemory(b2)

	c := &pem.Block{
		Headers: make(map[string]string),
	}
	c.Bytes = make([]byte, 1000)
	rand.Read(c.Bytes)
	c.Headers["path"] = "morpork"

	bBytes = append(bBytes, b2Bytes...)
	cBytes := pem.EncodeToMemory(c)

	byt := append(bBytes, cBytes...)

	in := bytes.NewBuffer(byt)

	err := ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.NoError(t, err)

	bFinal, bRest := pem.Decode(s.data[filepath.Join(notary.NonRootKeysSubdir, "testgun", "12ba0e0a8e05e177bc2c3489bdb6d28836879469f078e68a4812fc8a2d521497")])
	require.Equal(t, b.Headers["gun"], bFinal.Headers["gun"])
	require.Equal(t, b.Headers["role"], bFinal.Headers["role"])

	b2Final, b2Rest := pem.Decode(bRest)
	require.Equal(t, b2.Headers["gun"], b2Final.Headers["gun"])
	require.Equal(t, b2.Headers["role"], b2Final.Headers["role"])
	require.Len(t, b2Rest, 0)

	cFinal, cRest := pem.Decode(s.data["morpork"])
	require.Equal(t, c.Bytes, cFinal.Bytes)
	_, ok := cFinal.Headers["path"]
	require.False(t, ok, "expected no path header, should have been removed at import")
	require.Len(t, cRest, 0)
}

// no path and encrypted key import should fail
func TestEncryptedKeyImportFail(t *testing.T) {
	s := NewTestImportStore()

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err := utils.EncryptPrivateKey(privKey, data.CanonicalRootRole, "", cannedPassphrase)
	require.NoError(t, err)

	in := bytes.NewBuffer(pemBytes)

	_ = ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.Len(t, s.data, 0)
}

// path and encrypted key should succeed, tests gun inference from path as well
func TestEncryptedKeyImportSuccess(t *testing.T) {
	s := NewTestImportStore()

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	originalKey := privKey.Private()
	require.NoError(t, err)

	pemBytes, err := utils.EncryptPrivateKey(privKey, data.CanonicalSnapshotRole, "", cannedPassphrase)
	require.NoError(t, err)

	b, _ := pem.Decode(pemBytes)
	b.Headers["path"] = filepath.Join(notary.NonRootKeysSubdir, "somegun", "encryptedkey")
	pemBytes = pem.EncodeToMemory(b)

	in := bytes.NewBuffer(pemBytes)

	_ = ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.Len(t, s.data, 1)

	keyBytes := s.data[filepath.Join(notary.NonRootKeysSubdir, "somegun", "encryptedkey")]

	bFinal, bRest := pem.Decode(keyBytes)
	require.Equal(t, "somegun", bFinal.Headers["gun"])
	require.Len(t, bRest, 0)

	// we should fail to parse it without the passphrase
	privKey, err = utils.ParsePEMPrivateKey(keyBytes, "")
	require.Equal(t, err, errors.New("could not decrypt private key"))
	require.Nil(t, privKey)

	// we should succeed to parse it with the passphrase
	privKey, err = utils.ParsePEMPrivateKey(keyBytes, cannedPassphrase)
	require.NoError(t, err)
	require.Equal(t, originalKey, privKey.Private())
}

func TestEncryption(t *testing.T) {
	s := NewTestImportStore()

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	originalKey := privKey.Private()
	require.NoError(t, err)

	pemBytes, err := utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)

	in := bytes.NewBuffer(pemBytes)

	_ = ImportKeys(in, []Importer{s}, "", "", passphraseRetriever)
	require.Len(t, s.data, 1)

	shouldBeEnc, ok := s.data[filepath.Join(notary.NonRootKeysSubdir, privKey.ID())]
	// we should have got a key imported to this location
	require.True(t, ok)

	// we should fail to parse it without the passphrase
	privKey, err = utils.ParsePEMPrivateKey(shouldBeEnc, "")
	require.Equal(t, err, errors.New("could not decrypt private key"))
	require.Nil(t, privKey)

	// we should succeed to parse it with the passphrase
	privKey, err = utils.ParsePEMPrivateKey(shouldBeEnc, cannedPassphrase)
	require.NoError(t, err)
	require.Equal(t, originalKey, privKey.Private())
}
