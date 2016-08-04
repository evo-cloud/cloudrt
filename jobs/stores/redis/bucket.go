package redis

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/evo-cloud/cloudrt/jobs"
	redis "github.com/garyburd/redigo/redis"
)

type bucket struct {
	prefix string
	store  *Store
}

type bucketEnum struct {
	pattern string
	cursor  int
	count   int
	store   *Store
	end     bool
}

func (b *bucket) Enumerate(opts jobs.EnumOptions) jobs.Enumerator {
	return &bucketEnum{
		pattern: b.prefix + strconv.Itoa(opts.Partition) + ":*",
		cursor:  0,
		count:   opts.PageSize,
		store:   b.store,
	}
}

func (e *bucketEnum) Next() ([]jobs.Value, error) {
	if e.end {
		return nil, nil
	}
	conn := e.store.connection()
	defer conn.Close()
	replies, err := redis.Values(conn.Do(
		"SCAN", e.cursor,
		"MATCH", e.pattern,
		"COUNT", e.count))
	if err != nil {
		return nil, err
	}
	cursor, err := redis.Int(replies[0], nil)
	if err != nil {
		return nil, err
	}
	keys, err := redis.Strings(replies[1], nil)
	if err != nil {
		return nil, err
	}
	e.cursor = cursor
	if cursor == 0 {
		e.end = true
	}
	vals := make([]jobs.Value, 0, len(keys))
	for _, key := range keys {
		reply, err := redis.String(conn.Do("GET", key))
		if err != nil {
			return vals, err
		}
		vals = append(vals, &value{
			data: reply,
			ttl:  ttl2Dur(redis.Int64(conn.Do("PTTL", key))),
		})
	}
	return vals, nil
}

func (b *bucket) Put(key string, value interface{}, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	key = b.mapKey(key)
	conn := b.store.connection()
	defer conn.Close()
	if ttl != jobs.Infinite {
		_, err = conn.Do("PSETEX", key, dur2TTL(ttl), string(encoded))
	} else {
		_, err = conn.Do("SET", key, string(encoded))
	}
	return err
}

func (b *bucket) Get(key string) (jobs.Value, error) {
	key = b.mapKey(key)
	conn := b.store.connection()
	defer conn.Close()
	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return nil, err
	}
	return &value{
		data: reply,
		ttl:  ttl2Dur(redis.Int64(conn.Do("PTTL", key))),
	}, nil
}

func (b *bucket) Expire(key string, ttl time.Duration) (err error) {
	key = b.mapKey(key)
	conn := b.store.connection()
	defer conn.Close()
	if ttl != jobs.NoTTL {
		_, err = conn.Do("PERSIST", key)
	} else {
		_, err = conn.Do("PEXPIRE", key, dur2TTL(ttl))
	}
	return
}

func (b *bucket) Remove(key string) (val jobs.Value, err error) {
	key = b.mapKey(key)
	conn := b.store.connection()
	defer conn.Close()
	if reply, e := redis.String(conn.Do("GET", key)); e == nil {
		val = &value{
			data: reply,
			ttl:  ttl2Dur(redis.Int64(conn.Do("PTTL", key))),
		}
	}
	_, err = conn.Do("DEL", key)
	return
}

func (b *bucket) mapKey(key string) string {
	return b.prefix + strconv.Itoa(jobs.Partition(key)) + ":" + key
}
