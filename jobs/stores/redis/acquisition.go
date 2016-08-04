package redis

import (
	"fmt"
	"time"

	redis "github.com/garyburd/redigo/redis"
)

type acquisition struct {
	name    string
	owner   string
	ownedBy string
	ttl     time.Duration
	store   *Store
}

func (a *acquisition) Acquired() bool {
	return a.owner == a.ownedBy
}

func (a *acquisition) Owner() string {
	return a.ownedBy
}

func (a *acquisition) TTL() time.Duration {
	return a.ttl
}

func (a *acquisition) Refresh(ttl time.Duration) error {
	return a.mustOwned(func(conn redis.Conn) error {
		conn.Send("MULTI")
		conn.Send("PEXPIRE", a.key(), dur2TTL(a.ttl))
		_, err := conn.Do("EXEC")
		return err
	})
}

func (a *acquisition) Release() error {
	return a.mustOwned(func(conn redis.Conn) error {
		conn.Send("MULTI")
		conn.Send("DEL", a.key())
		_, err := conn.Do("EXEC")
		return err
	})
}

func (a *acquisition) tryAcquire() error {
	key := a.key()
	conn := a.store.connection()
	defer conn.Close()
	_, err := conn.Do("WATCH", key)
	if err != nil {
		return err
	}
	val, err := redis.Values(conn.Do("HMGET", key, "owner", "refs"))
	if err != nil {
		return err
	}
	if val != nil && len(val) == 2 {
		if ownedBy, e := redis.String(val[0], nil); e == nil {
			a.ownedBy = ownedBy
		}
	}
	if a.ownedBy != "" && a.ownedBy != a.owner {
		_, err = conn.Do("DISCARD")
		return err
	}
	conn.Send("MULTI")
	conn.Send("HSET", key, "owner", a.owner)
	conn.Send("HINCRBY", key, "refs", 1)
	conn.Send("PEXPIRE", key, dur2TTL(a.ttl))
	if _, err = conn.Do("EXEC"); err != nil {
		return err
	}
	ownedBy, err := redis.String(conn.Do("HGET", key, "owner"))
	if err != nil {
		return err
	}
	a.ownedBy = ownedBy
	return nil
}

func (a *acquisition) key() string {
	return a.name + ":" + a.owner
}

func (a *acquisition) mustOwned(fn func(conn redis.Conn) error) error {
	key := a.key()
	conn := a.store.connection()
	defer conn.Close()
	_, err := conn.Do("WATCH", key)
	if err != nil {
		return err
	}
	ownedBy, err := redis.String(conn.Do("HGET", key, "owner"))
	if err != nil {
		return err
	}
	if ownedBy != a.owner {
		conn.Do("DISCARD")
		return fmt.Errorf("not owned by %s (owned by %s)", a.owner, ownedBy)
	}
	return fn(conn)
}
