package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"
	"time"
)

func BatchIdentities(n int) ([]*Private, *Group) {
	startPort := 8000
	startAddr := "127.0.0.1:"
	privs := make([]*Private, n)
	pubs := make([]*Public, n)
	for i := 0; i < n; i++ {
		port := strconv.Itoa(startPort + i)
		addr := startAddr + port
		privs[i] = NewKeyPair(addr)
		pubs[i] = privs[i].Public
	}
	group := &Group{
		Threshold: defaultThreshold(n),
		Nodes:     toIndexedList(pubs),
	}
	return privs, group
}

func BatchDrands(n int) (*Group, []*Drand) {
	ids, group := BatchIdentities(n)
	drands := make([]*Drand, n)
	var err error
	for i := range ids {
		store := NewTestStore(ids[i], group, nil)
		drands[i], err = newDrand(ids[i], group, store)
		if err != nil {
			panic(err)
		}
	}
	return group, drands
}

func BatchRouters(n int) ([]*Private, []*Router) {
	privs, group := BatchIdentities(n)
	routers := make([]*Router, n)
	for i := 0; i < n; i++ {
		routers[i] = NewRouter(privs[i], group)
		go routers[i].Listen()
	}
	sort.Sort(ByIndex(routers))
	time.Sleep(10 * time.Millisecond)
	return privs, routers
}

func CloseAllDrands(drands []*Drand) {
	for _, d := range drands {
		d.r.Stop()
	}
}

func CloseAllRouters(routers []*Router) {
	for _, r := range routers {
		r.Stop()
	}
}

type TestStore struct {
	*sync.Mutex
	Private      *Private
	Public       *Public
	Group        *Group
	Share        *Share
	DistPublic   *DistPublic
	Signatures   map[int64]*BeaconSignature
	CbSignatures func(*BeaconSignature)
}

// cb can be nil
func NewTestStore(private *Private, g *Group, cb func(*BeaconSignature)) Store {
	fmt.Println("NewTestSTore for ", private.Public.Address)
	return &TestStore{
		Private:      private,
		Public:       private.Public,
		Group:        g,
		Signatures:   make(map[int64]*BeaconSignature),
		Mutex:        new(sync.Mutex),
		CbSignatures: cb,
	}
}

func (t *TestStore) SaveKey(p *Private) error {
	t.Lock()
	defer t.Unlock()
	t.Private = p
	return nil
}

func (t *TestStore) LoadKey() (*Private, error) {
	t.Lock()
	defer t.Unlock()
	return t.Private, nil
}

func (t *TestStore) LoadGroup() (*Group, error) {
	t.Lock()
	defer t.Unlock()
	return t.Group, nil
}

func (t *TestStore) SaveShare(s *Share) error {
	t.Lock()
	defer t.Unlock()
	t.Share = s
	return nil
}

func (t *TestStore) LoadShare() (*Share, error) {
	t.Lock()
	defer t.Unlock()
	return t.Share, nil
}

func (t *TestStore) SaveDistPublic(d *DistPublic) error {
	t.Lock()
	defer t.Unlock()
	t.DistPublic = d
	return nil
}

func (t *TestStore) LoadDistPublic() (*DistPublic, error) {
	t.Lock()
	defer t.Unlock()
	return t.DistPublic, nil
}

func (t *TestStore) SaveBeacon(b *BeaconSignature) error {
	t.Lock()
	defer t.Unlock()
	t.Signatures[b.Request.Timestamp] = b
	fmt.Printf("%p: Save Signatures ! callback ? %p\n", t, t.CbSignatures)
	if t.CbSignatures != nil {
		t.CbSignatures(b)
	}
	return nil
}

func (t *TestStore) LoadSignature(path string) (*BeaconSignature, error) {
	t.Lock()
	defer t.Unlock()
	return nil, errors.New("not implemented now")
}

func (t *TestStore) BeaconExists(ts int64) bool {
	t.Lock()
	defer t.Unlock()
	_, ok := t.Signatures[ts]
	return ok
}

type ByIndex []*Router

func (b ByIndex) Len() int {
	return len(b)
}

func (b ByIndex) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByIndex) Less(i, j int) bool {
	return b[i].index < b[j].index
}

func tempDir() string {
	return os.TempDir()
}

func tempShareFile() string {
	return tempDir()
}

type basicKV struct {
	path string
}

func (b *basicKV) String(key string) string {
	switch key {
	case keyFolderFlagName:
		return path.Join(b.path, defaultKeyFile)
	case groupFileFlagName:
		return path.Join(b.path, groupFileFlagName)
	case sigFolderFlagName:
		return path.Join(b.path, sigFolderFlagName)
	default:
		panic("he")
	}
}
