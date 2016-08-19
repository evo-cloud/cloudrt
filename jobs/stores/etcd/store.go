package etcd

import (
	"encoding/json"
	etcd "github.com/coreos/etcd/client"
	"github.com/evo-cloud/cloudrt/jobs"
	"log"
	"time"
)

// Store is a store implementation backed by etcd
type Store struct {
	Endpoints []string
	Client    *etcd.Client
}

// NewStore creates a Store instance
func NewStore(endpoints []string) *Store {
	s := &Store{
		Endpoints: endpoints,
	}

	config := etcd.Config{
		Endpoints:               s.Endpoints,
		Transport:               etcd.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := etcd.New(config)
	if err != nil {
		log.Fatal(err)
	}
	s.Client = &c

	return s
}

// Bucket implements Store
func (s *Store) Bucket(name string) jobs.PartitionedStore {
	return &bucket{
		prefix: name,
		s:      s,
	}
}

// OrderedList implements Store
func (s *Store) OrderedList(name string) jobs.OrderedList {
	return &orderedList{
		name: name,
		s:    s,
	}
}

// Acquire implements Store
func (s *Store) Acquire(name, ownerID string) (jobs.Acquisition, error) {
	a := &acquisition{
		name:  name,
		owner: ownerID,
		ttl:   10 * time.Second,
		s:     s,
	}
	return a, a.tryAcquire()
}

type value struct {
	data string
	ttl  time.Duration
}

func (v *value) TTL() time.Duration {
	return v.ttl
}

func (v *value) Unmarshal(out interface{}) error {
	return json.Unmarshal([]byte(v.data), out)
}

func ttl2Dur(ttl int64, err error) time.Duration {
	if err != nil || ttl < 0 {
		return jobs.NoTTL
	}
	return time.Duration(ttl) * time.Millisecond
}

func dur2TTL(dur time.Duration) int64 {
	return int64(dur / time.Millisecond)
}
