package redis

import (
	"encoding/json"
	"time"

	"github.com/evo-cloud/cloudrt/jobs"
	redis "github.com/garyburd/redigo/redis"
)

type orderedList struct {
	name  string
	store *Store
}

type orderedListEnum struct {
	name   string
	cursor int
	count  int
	end    bool
	store  *Store
}

func (l *orderedList) Enumerate(opts jobs.EnumOptions) jobs.Enumerator {
	return &orderedListEnum{
		name:   l.name,
		cursor: 0,
		count:  opts.PageSize,
		store:  l.store,
	}
}

func (e *orderedListEnum) Next() ([]jobs.Value, error) {
	if e.end {
		return nil, nil
	}
	conn := e.store.connection()
	defer conn.Close()
	replies, err := redis.Values(conn.Do(
		"ZSCAN", e.name, e.cursor,
		"COUNT", e.count))
	if err != nil {
		return nil, err
	}
	cursor, err := redis.Int(replies[0], nil)
	if err != nil {
		return nil, err
	}
	items, err := redis.Values(replies[1], nil)
	if err != nil {
		return nil, err
	}
	e.cursor = cursor
	if cursor == 0 {
		e.end = true
	}
	vals := make([]jobs.Value, 0, len(items))
	for _, item := range items {
		arr, ok := item.([]interface{})
		if !ok || len(arr) == 0 {
			continue
		}
		key, ok := arr[0].(string)
		if !ok {
			continue
		}
		encoded, _ := json.Marshal(&key)
		vals = append(vals, &value{
			data: string(encoded),
			ttl:  jobs.NoTTL,
		})
	}
	return vals, nil
}

func (l *orderedList) Set(id string, exist bool) (err error) {
	conn := l.store.connection()
	defer conn.Close()
	if exist {
		_, err = conn.Do("ZADD", l.name, float64(time.Now()), id)
	} else {
		_, err = conn.Do("ZREM", l.name, id)
	}
	return
}

func (l *orderedList) Has(id string) (bool, error) {
	conn := l.store.connection()
	defer conn.Close()
	score, err := redis.Float64(conn.Do("ZSCORE", l.name, id))
	return score != 0, err
}
