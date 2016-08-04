package redis

import (
	"encoding/json"
	"time"

	"github.com/evo-cloud/cloudrt/jobs"
	redis "github.com/garyburd/redigo/redis"
)

// Store is a store implementation backed by Redis
type Store struct {
	Server string
	Auth   string

	pool *redis.Pool
}

// NewStore creates a Store instance
func NewStore(server string) *Store {
	s := &Store{
		Server: server,
	}
	s.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 5 * time.Minute,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if s.Auth != "" {
				_, err = c.Do("AUTH", s.Auth)
				if err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	return s
}

// Bucket implements Store
func (s *Store) Bucket(name string) jobs.PartitionedStore {
	return &bucket{prefix: "b:" + name + ":", store: s}
}

// OrderedList implements Store
func (s *Store) OrderedList(name string) jobs.OrderedList {
	return &orderedList{name: "o:" + name, store: s}
}

// Acquire implements Store
func (s *Store) Acquire(name, ownerID string) (jobs.Acquisition, error) {
	a := &acquisition{name: "a:" + name, owner: ownerID, store: s, ttl: 10 * time.Second}
	return a, a.tryAcquire()
}

func (s *Store) connection() redis.Conn {
	return s.pool.Get()
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
