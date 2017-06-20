package utils

import (
	"chess_alg_jx/config"
	"github.com/garyburd/redigo/redis"
	"time"
)

var (
	RedisClient    *Redis
	RedisHost      string
	RedisPrefixKey string
)

func InitRedis() {
	RedisHost = config.Config.RedisAddr
	RedisPrefixKey = config.Config.RedisPrefix
	RedisClient = NewRedis(NewPool(RedisHost), RedisPrefixKey)
}

type Redis struct {
	pool      *redis.Pool
	keyPrefix string
}

//NewRedisPoll get new redis poll
func NewRedisPool(serverIP string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", serverIP)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func NewPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func NewRedis(pool *redis.Pool, keyPrefix string) *Redis {
	return &Redis{pool, keyPrefix}
}

func (this *Redis) Get(key string) (s string, err error) {
	conn := this.pool.Get()
	s, err = redis.String(conn.Do("GET", this.keyPrefix+key))
	if err != nil {
		return
	}
	err = conn.Close()
	return
}

func (this *Redis) Set(key string, value interface{}) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("SET", this.keyPrefix+key, value)
	if err != nil {
		return
	}

	err = conn.Close()

	return
}

func (this *Redis) IsKeyExit(key string) (ok bool, err error) {
	conn := this.pool.Get()
	ok, err = redis.Bool(conn.Do("EXISTS", this.keyPrefix+key))
	if err != nil {
		return
	}
	err = conn.Close()

	return
}

func (this *Redis) Del(key string) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("DEL", this.keyPrefix+key)
	if err != nil {
		return
	}

	err = conn.Close()

	return
}

func (this *Redis) Incr(key string) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("Incr", this.keyPrefix+key)
	if err != nil {
		return
	}

	err = conn.Close()

	return
}

func (this *Redis) LRange(key string, start, end int) (value [][]byte, err error) {
	conn := this.pool.Get()
	v, err := redis.Values(conn.Do("LRANGE", this.keyPrefix+key, start, end))
	if err != nil {
		return
	}
	if err = redis.ScanSlice(v, &value); err != nil {
		return
	}

	err = conn.Close()

	return
}

func (this *Redis) LPush(key string, value interface{}) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("LPUSH", redis.Args{}.Add(this.keyPrefix+key).AddFlat(value)...)
	if err != nil {
		return
	}

	err = conn.Close()

	return
}

func (this *Redis) LPop(key string) (value []byte, err error) {
	conn := this.pool.Get()
	value, err = redis.Bytes(conn.Do("LPOP", this.keyPrefix+key))
	if err != nil {
		return
	}

	err = conn.Close()
	return
}

func (this *Redis) BLPop(key string) (value []byte, err error) {
	conn := this.pool.Get()
	v, err := redis.Values(conn.Do("BLPOP", this.keyPrefix+key, 0))
	if err != nil {
		return
	}
	values := make([][]byte, 2)
	if err = redis.ScanSlice(v, &values); err != nil {
		return
	}
	value = values[1]
	err = conn.Close()
	return
}

func (this *Redis) RPush(key string, value interface{}) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("RPUSH", redis.Args{}.Add(this.keyPrefix+key).AddFlat(value)...)
	if err != nil {
		return
	}

	err = conn.Close()
	return
}

func (this *Redis) RPop(key string) (value []byte, err error) {
	conn := this.pool.Get()
	value, err = redis.Bytes(conn.Do("RPOP", this.keyPrefix+key))
	if err != nil {
		return
	}

	err = conn.Close()
	return
}

func (this *Redis) BRPop(key string) (value []byte, err error) {
	conn := this.pool.Get()
	v, err := redis.Values(conn.Do("BRPOP", this.keyPrefix+key, 0))
	if err != nil {
		return
	}
	values := make([][]byte, 2)
	if err = redis.ScanSlice(v, &values); err != nil {
		return
	}
	value = values[1]

	err = conn.Close()
	return
}

func (this *Redis) SMembers(key string, values interface{}) (err error) {
	conn := this.pool.Get()
	v, err := redis.Values(conn.Do("SMEMBERS", this.keyPrefix+key))
	if err != nil {
		return
	}
	if err = redis.ScanSlice(v, values); err != nil {
		return
	}

	err = conn.Close()
	return
}

func (this *Redis) SAdd(key string, value interface{}) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("SADD", redis.Args{}.Add(this.keyPrefix+key).AddFlat(value)...)
	if err != nil {
		return
	}
	err = conn.Close()
	return
}

func (this *Redis) SRem(key string, value interface{}) (err error) {
	conn := this.pool.Get()
	_, err = conn.Do("SREM", redis.Args{}.Add(this.keyPrefix+key).AddFlat(value)...)
	if err != nil {
		return
	}
	err = conn.Close()
	return
}

func (this *Redis) SIsMember(key string, member interface{}) (ok bool, err error) {
	conn := this.pool.Get()
	ok, err = redis.Bool(conn.Do("SISMEMBER", this.keyPrefix+key, member))
	if err != nil {
		return
	}
	err = conn.Close()
	return
}
