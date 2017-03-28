package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/net/context"

	ctxu "github.com/docker/distribution/context"
	"github.com/docker/distribution/registry/api/errcode"
	"github.com/docker/notary/server/errors"
	"github.com/docker/notary/server/storage"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/validation"

	"github.com/docker/notary/tuf/testutils"
	"github.com/docker/notary/utils"
	"github.com/stretchr/testify/require"
)

type handlerState struct {
	// interface{} so we can test invalid values
	store   interface{}
	crypto  interface{}
	keyAlgo interface{}
}

func defaultState() handlerState {
	return handlerState{
		store:   storage.NewMemStorage(),
		crypto:  signed.NewEd25519(),
		keyAlgo: data.ED25519Key,
	}
}

func getContext(h handlerState) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "metaStore", h.store)
	ctx = context.WithValue(ctx, "keyAlgorithm", h.keyAlgo)
	ctx = context.WithValue(ctx, "cryptoService", h.crypto)
	return ctxu.WithLogger(ctx, ctxu.GetRequestLogger(ctx))
}

func TestMainHandlerGet(t *testing.T) {
	hand := utils.RootHandlerFactory(nil, context.Background(), &signed.Ed25519{})
	handler := hand(MainHandler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	_, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Received error on GET /: %s", err.Error())
	}
}

func TestMainHandlerNotGet(t *testing.T) {
	hand := utils.RootHandlerFactory(nil, context.Background(), &signed.Ed25519{})
	handler := hand(MainHandler)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	res, err := http.Head(ts.URL)
	if err != nil {
		t.Fatalf("Received error on GET /: %s", err.Error())
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404, received %d", res.StatusCode)
	}
}

type simplerHandler func(context.Context, http.ResponseWriter, *http.Request, map[string]string) error

// GetKeyHandler and RotateKeyHandler needs to have access to a metadata store and cryptoservice,
// a key algorithm
func TestKeyHandlersInvalidConfiguration(t *testing.T) {
	noStore := defaultState()
	noStore.store = nil

	invalidStore := defaultState()
	invalidStore.store = "not a store"

	noCrypto := defaultState()
	noCrypto.crypto = nil

	invalidCrypto := defaultState()
	invalidCrypto.crypto = "not a cryptoservice"

	noKeyAlgo := defaultState()
	noKeyAlgo.keyAlgo = ""

	invalidKeyAlgo := defaultState()
	invalidKeyAlgo.keyAlgo = 1

	invalidStates := map[string][]handlerState{
		"no storage":       {noStore, invalidStore},
		"no cryptoservice": {noCrypto, invalidCrypto},
		"no keyalgorithm":  {noKeyAlgo, invalidKeyAlgo},
	}

	vars := map[string]string{
		"imageName": "gun",
		"tufRole":   data.CanonicalTimestampRole,
	}
	req := &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(nil))}
	for _, keyHandler := range []simplerHandler{getKeyHandler, rotateKeyHandler} {
		for errString, states := range invalidStates {
			for _, s := range states {
				err := keyHandler(getContext(s), httptest.NewRecorder(), req, vars)
				require.Error(t, err)
				require.Contains(t, err.Error(), errString)
			}
		}
	}
}

// GetKeyHandler and RotateKeyHandler need to be set up such that an imageName and tufRole are both
// provided and non-empty.
func TestKeyHandlersNoRoleOrRepo(t *testing.T) {
	state := defaultState()
	req := &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(nil))}
	for _, keyHandler := range []simplerHandler{getKeyHandler, rotateKeyHandler} {
		for _, key := range []string{"imageName", "tufRole"} {
			vars := map[string]string{
				"imageName": "gun",
				"tufRole":   data.CanonicalTimestampRole,
			}

			// not provided
			delete(vars, key)
			err := keyHandler(getContext(state), httptest.NewRecorder(), req, vars)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unknown")

			// empty
			vars[key] = ""
			err = keyHandler(getContext(state), httptest.NewRecorder(), req, vars)
			require.Error(t, err)
			require.Contains(t, err.Error(), "unknown")
		}
	}
}

// GetKeyHandler and RotateKeyHandler called for a non-supported role results in a 400.
func TestKeyHandlersInvalidRole(t *testing.T) {
	state := defaultState()
	for _, keyHandler := range []simplerHandler{getKeyHandler, rotateKeyHandler} {
		for _, role := range []string{data.CanonicalRootRole, data.CanonicalTargetsRole, "targets/a", "invalidrole"} {
			vars := map[string]string{
				"imageName": "gun",
				"tufRole":   role,
			}
			req := &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

			err := keyHandler(getContext(state), httptest.NewRecorder(), req, vars)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid role")
		}
	}
}

// Getting the key for a valid role and gun succeeds
func TestGetKeyHandlerCreatesOnce(t *testing.T) {
	state := defaultState()
	roles := []string{data.CanonicalTimestampRole, data.CanonicalSnapshotRole}
	req := &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	for _, role := range roles {
		vars := map[string]string{"imageName": "gun", "tufRole": role}
		recorder := httptest.NewRecorder()
		err := getKeyHandler(getContext(state), recorder, req, vars)
		require.NoError(t, err)
		require.True(t, len(recorder.Body.String()) > 0)
	}
}

// Getting or rotating the key fails if we don't pass a valid key algorithm
func TestKeyHandlersInvalidKeyAlgo(t *testing.T) {
	roles := []string{data.CanonicalTimestampRole, data.CanonicalSnapshotRole}
	req := &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(nil))}
	for _, keyHandler := range []simplerHandler{getKeyHandler, rotateKeyHandler} {
		for _, role := range roles {
			vars := map[string]string{"imageName": "gun", "tufRole": role}
			recorder := httptest.NewRecorder()
			invalidKeyAlgoState := defaultState()
			invalidKeyAlgoState.keyAlgo = "notactuallyakeyalgorithm"
			err := keyHandler(getContext(invalidKeyAlgoState), recorder, req, vars)
			require.Error(t, err)
		}
	}
}

// Rotating the key for a valid role and gun succeeds
func TestRotateKeyHandlerSuccessfulRotation(t *testing.T) {
	state := defaultState()
	roles := []string{data.CanonicalTimestampRole, data.CanonicalSnapshotRole}
	req := &http.Request{Body: ioutil.NopCloser(bytes.NewBuffer(nil))}

	for _, role := range roles {
		vars := map[string]string{"imageName": "gun", "tufRole": role}
		recorder := httptest.NewRecorder()
		err := rotateKeyHandler(getContext(state), recorder, req, vars)
		require.NoError(t, err)
		require.True(t, len(recorder.Body.String()) > 0)
	}
}

func TestGetHandlerRoot(t *testing.T) {
	metaStore := storage.NewMemStorage()
	repo, _, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "metaStore", metaStore)

	root, err := repo.SignRoot(data.DefaultExpires("root"))
	require.NoError(t, err)
	rootJSON, err := json.Marshal(root)
	require.NoError(t, err)
	metaStore.UpdateCurrent("gun", storage.MetaUpdate{Role: "root", Version: 1, Data: rootJSON})

	req := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBuffer(nil)),
	}

	vars := map[string]string{
		"imageName": "gun",
		"tufRole":   "root",
	}

	rw := httptest.NewRecorder()

	err = getHandler(ctx, rw, req, vars)
	require.NoError(t, err)
}

func TestGetHandlerTimestamp(t *testing.T) {
	metaStore := storage.NewMemStorage()
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	ctx := getContext(handlerState{store: metaStore, crypto: crypto})

	sn, err := repo.SignSnapshot(data.DefaultExpires("snapshot"))
	require.NoError(t, err)
	snJSON, err := json.Marshal(sn)
	require.NoError(t, err)
	metaStore.UpdateCurrent(
		"gun", storage.MetaUpdate{Role: "snapshot", Version: 1, Data: snJSON})

	ts, err := repo.SignTimestamp(data.DefaultExpires("timestamp"))
	require.NoError(t, err)
	tsJSON, err := json.Marshal(ts)
	require.NoError(t, err)
	metaStore.UpdateCurrent(
		"gun", storage.MetaUpdate{Role: "timestamp", Version: 1, Data: tsJSON})

	req := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBuffer(nil)),
	}

	vars := map[string]string{
		"imageName": "gun",
		"tufRole":   "timestamp",
	}

	rw := httptest.NewRecorder()

	err = getHandler(ctx, rw, req, vars)
	require.NoError(t, err)
}

func TestGetHandlerSnapshot(t *testing.T) {
	metaStore := storage.NewMemStorage()
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	ctx := getContext(handlerState{store: metaStore, crypto: crypto})

	// Need to create a timestamp and snapshot
	sn, err := repo.SignSnapshot(data.DefaultExpires("snapshot"))
	require.NoError(t, err)
	snJSON, err := json.Marshal(sn)
	require.NoError(t, err)
	metaStore.UpdateCurrent(
		"gun", storage.MetaUpdate{Role: "snapshot", Version: 1, Data: snJSON})

	ts, err := repo.SignTimestamp(data.DefaultExpires("timestamp"))
	require.NoError(t, err)
	tsJSON, err := json.Marshal(ts)
	require.NoError(t, err)
	metaStore.UpdateCurrent(
		"gun", storage.MetaUpdate{Role: "timestamp", Version: 1, Data: tsJSON})

	req := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBuffer(nil)),
	}

	vars := map[string]string{
		"imageName": "gun",
		"tufRole":   "snapshot",
	}

	rw := httptest.NewRecorder()

	err = getHandler(ctx, rw, req, vars)
	require.NoError(t, err)
}

func TestGetHandler404(t *testing.T) {
	metaStore := storage.NewMemStorage()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "metaStore", metaStore)

	req := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBuffer(nil)),
	}

	vars := map[string]string{
		"imageName": "gun",
		"tufRole":   "root",
	}

	rw := httptest.NewRecorder()

	err := getHandler(ctx, rw, req, vars)
	require.Error(t, err)
}

func TestGetHandlerNilData(t *testing.T) {
	metaStore := storage.NewMemStorage()
	metaStore.UpdateCurrent("gun", storage.MetaUpdate{Role: "root", Version: 1, Data: nil})

	ctx := context.Background()
	ctx = context.WithValue(ctx, "metaStore", metaStore)

	req := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBuffer(nil)),
	}

	vars := map[string]string{
		"imageName": "gun",
		"tufRole":   "root",
	}

	rw := httptest.NewRecorder()

	err := getHandler(ctx, rw, req, vars)
	require.Error(t, err)
}

func TestGetHandlerNoStorage(t *testing.T) {
	ctx := context.Background()

	req := &http.Request{
		Body: ioutil.NopCloser(bytes.NewBuffer(nil)),
	}

	err := GetHandler(ctx, nil, req)
	require.Error(t, err)
}

// a validation failure, such as a snapshots file being missing, will be
// propagated as a detail in the error (which gets serialized as the body of the
// response)
func TestAtomicUpdateValidationFailurePropagated(t *testing.T) {
	metaStore := storage.NewMemStorage()
	gun := "testGUN"
	vars := map[string]string{"imageName": gun}

	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	state := handlerState{store: metaStore, crypto: testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)}

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	rs, tgs, _, _, err := testutils.Serialize(r, tg, sn, ts)
	require.NoError(t, err)

	req, err := store.NewMultiPartMetaRequest("", map[string][]byte{
		data.CanonicalRootRole:    rs,
		data.CanonicalTargetsRole: tgs,
	})
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	err = atomicUpdateHandler(getContext(state), rw, req, vars)
	require.Error(t, err)
	errorObj, ok := err.(errcode.Error)
	require.True(t, ok, "Expected an errcode.Error, got %v", err)
	require.Equal(t, errors.ErrInvalidUpdate, errorObj.Code)
	serializable, ok := errorObj.Detail.(*validation.SerializableError)
	require.True(t, ok, "Expected a SerializableObject, got %v", errorObj.Detail)
	require.IsType(t, validation.ErrBadHierarchy{}, serializable.Error)
}

type failStore struct {
	storage.MemStorage
}

func (s *failStore) GetCurrent(_, _ string) (*time.Time, []byte, error) {
	return nil, nil, fmt.Errorf("oh no! storage has failed")
}

// a non-validation failure, such as the storage failing, will not be propagated
// as a detail in the error (which gets serialized as the body of the response)
func TestAtomicUpdateNonValidationFailureNotPropagated(t *testing.T) {
	metaStore := storage.NewMemStorage()
	gun := "testGUN"
	vars := map[string]string{"imageName": gun}

	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	state := handlerState{store: &failStore{*metaStore}, crypto: testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)}

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	rs, tgs, sns, _, err := testutils.Serialize(r, tg, sn, ts)
	require.NoError(t, err)

	req, err := store.NewMultiPartMetaRequest("", map[string][]byte{
		data.CanonicalRootRole:     rs,
		data.CanonicalTargetsRole:  tgs,
		data.CanonicalSnapshotRole: sns,
	})
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	err = atomicUpdateHandler(getContext(state), rw, req, vars)
	require.Error(t, err)
	errorObj, ok := err.(errcode.Error)
	require.True(t, ok, "Expected an errcode.Error, got %v", err)
	require.Equal(t, errors.ErrInvalidUpdate, errorObj.Code)
	require.Nil(t, errorObj.Detail)
}

type invalidVersionStore struct {
	storage.MemStorage
}

func (s *invalidVersionStore) UpdateMany(_ string, _ []storage.MetaUpdate) error {
	return storage.ErrOldVersion{}
}

// a non-validation failure, such as the storage failing, will be propagated
// as a detail in the error (which gets serialized as the body of the response)
func TestAtomicUpdateVersionErrorPropagated(t *testing.T) {
	metaStore := storage.NewMemStorage()
	gun := "testGUN"
	vars := map[string]string{"imageName": gun}

	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	state := handlerState{
		store: &invalidVersionStore{*metaStore}, crypto: testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)}

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	rs, tgs, sns, _, err := testutils.Serialize(r, tg, sn, ts)
	require.NoError(t, err)

	req, err := store.NewMultiPartMetaRequest("", map[string][]byte{
		data.CanonicalRootRole:     rs,
		data.CanonicalTargetsRole:  tgs,
		data.CanonicalSnapshotRole: sns,
	})
	require.NoError(t, err)

	rw := httptest.NewRecorder()

	err = atomicUpdateHandler(getContext(state), rw, req, vars)
	require.Error(t, err)
	errorObj, ok := err.(errcode.Error)
	require.True(t, ok, "Expected an errcode.Error, got %v", err)
	require.Equal(t, errors.ErrOldVersion, errorObj.Code)
	require.Equal(t, storage.ErrOldVersion{}, errorObj.Detail)
}
