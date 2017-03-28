package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type key struct {
	algorithm string
	public    []byte
}

type ver struct {
	version      int
	data         []byte
	createupdate time.Time
}

// we want to keep these sorted by version so that it's in increasing version
// order
type verList []ver

func (k verList) Len() int      { return len(k) }
func (k verList) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k verList) Less(i, j int) bool {
	return k[i].version < k[j].version
}

// MemStorage is really just designed for dev and testing. It is very
// inefficient in many scenarios
type MemStorage struct {
	lock      sync.Mutex
	tufMeta   map[string]verList
	keys      map[string]map[string]*key
	checksums map[string]map[string]ver
}

// NewMemStorage instantiates a memStorage instance
func NewMemStorage() *MemStorage {
	return &MemStorage{
		tufMeta:   make(map[string]verList),
		keys:      make(map[string]map[string]*key),
		checksums: make(map[string]map[string]ver),
	}
}

// UpdateCurrent updates the meta data for a specific role
func (st *MemStorage) UpdateCurrent(gun string, update MetaUpdate) error {
	id := entryKey(gun, update.Role)
	st.lock.Lock()
	defer st.lock.Unlock()
	if space, ok := st.tufMeta[id]; ok {
		for _, v := range space {
			if v.version >= update.Version {
				return ErrOldVersion{}
			}
		}
	}
	version := ver{version: update.Version, data: update.Data, createupdate: time.Now()}
	st.tufMeta[id] = append(st.tufMeta[id], version)
	checksumBytes := sha256.Sum256(update.Data)
	checksum := hex.EncodeToString(checksumBytes[:])

	_, ok := st.checksums[gun]
	if !ok {
		st.checksums[gun] = make(map[string]ver)
	}
	st.checksums[gun][checksum] = version
	return nil
}

// UpdateMany updates multiple TUF records
func (st *MemStorage) UpdateMany(gun string, updates []MetaUpdate) error {
	st.lock.Lock()
	defer st.lock.Unlock()

	versioner := make(map[string]map[int]struct{})
	constant := struct{}{}

	// ensure that we only update in one transaction
	for _, u := range updates {
		id := entryKey(gun, u.Role)

		// prevent duplicate versions of the same role
		if _, ok := versioner[u.Role][u.Version]; ok {
			return ErrOldVersion{}
		}
		if _, ok := versioner[u.Role]; !ok {
			versioner[u.Role] = make(map[int]struct{})
		}
		versioner[u.Role][u.Version] = constant

		if space, ok := st.tufMeta[id]; ok {
			for _, v := range space {
				if v.version >= u.Version {
					return ErrOldVersion{}
				}
			}
		}
	}

	for _, u := range updates {
		id := entryKey(gun, u.Role)

		version := ver{version: u.Version, data: u.Data, createupdate: time.Now()}
		st.tufMeta[id] = append(st.tufMeta[id], version)
		sort.Sort(st.tufMeta[id]) // ensure that it's sorted
		checksumBytes := sha256.Sum256(u.Data)
		checksum := hex.EncodeToString(checksumBytes[:])

		_, ok := st.checksums[gun]
		if !ok {
			st.checksums[gun] = make(map[string]ver)
		}
		st.checksums[gun][checksum] = version
	}
	return nil
}

// GetCurrent returns the createupdate date metadata for a given role, under a GUN.
func (st *MemStorage) GetCurrent(gun, role string) (*time.Time, []byte, error) {
	id := entryKey(gun, role)
	st.lock.Lock()
	defer st.lock.Unlock()
	space, ok := st.tufMeta[id]
	if !ok || len(space) == 0 {
		return nil, nil, ErrNotFound{}
	}
	return &(space[len(space)-1].createupdate), space[len(space)-1].data, nil
}

// GetChecksum returns the createupdate date and metadata for a given role, under a GUN.
func (st *MemStorage) GetChecksum(gun, role, checksum string) (*time.Time, []byte, error) {
	st.lock.Lock()
	defer st.lock.Unlock()
	space, ok := st.checksums[gun][checksum]
	if !ok || len(space.data) == 0 {
		return nil, nil, ErrNotFound{}
	}
	return &(space.createupdate), space.data, nil
}

// Delete deletes all the metadata for a given GUN
func (st *MemStorage) Delete(gun string) error {
	st.lock.Lock()
	defer st.lock.Unlock()
	for k := range st.tufMeta {
		if strings.HasPrefix(k, gun) {
			delete(st.tufMeta, k)
		}
	}
	delete(st.checksums, gun)
	return nil
}

func entryKey(gun, role string) string {
	return fmt.Sprintf("%s.%s", gun, role)
}
