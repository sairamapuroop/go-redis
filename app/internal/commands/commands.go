package commands

import (
	"fmt"
	"redis-go/internal/db"
	"strconv"
	"strings"
	"time"
)

type CommandFunc func(args []string, ttl time.Duration) string

type Registry struct {
	db   *db.DB
	cmds map[string]CommandFunc
}

func (r *Registry) GetDB() *db.DB {
	return r.db
}

func NewRegistry(db *db.DB) *Registry {
	r := &Registry{
		db:   db,
		cmds: make(map[string]CommandFunc),
	}

	r.cmds["PING"] = func(args []string, _ time.Duration) string {
		return "+PONG\r\n"
	}

	r.cmds["SET"] = func(args []string, ttl time.Duration) string {
		if len(args) < 2 {
			return "-ERR wrong number of arguments\r\n"
		}

		r.db.Set(args[0], args[1], ttl)
		return "+OK\r\n"
	}

	r.cmds["GET"] = func(args []string, _ time.Duration) string {
		if len(args) != 1 {
			return "-ERR wrong number of arguments\r\n"
		}

		if val, ok := r.db.Get(args[0]); ok {
			return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
		}
		return "*0\r\n"
	}

	r.cmds["DEL"] = func(args []string, _ time.Duration) string {
		if len(args) != 1 {
			return "-ERR wrong number of arguments\r\n"
		}

		if ok := r.db.Delete(args[0]); ok {
			return "+1\r\n"
		}

		return "-0\r\n"
	}

	r.cmds["LPUSH"] = func(args []string, _ time.Duration) string {
		if len(args) < 2 {
			return "-ERR wrong number of arguments\r\n"
		}

		val := r.db.LPush(args[0], args[1:]...)

		return fmt.Sprintf("+%d\r\n", val)

	}

	r.cmds["RPUSH"] = func(args []string, _ time.Duration) string {
		if len(args) < 2 {
			return "-ERR wrong number of arguments\r\n"
		}

		r.db.LPush(args[0], args[1:]...)

		return fmt.Sprintf("+%d\r\n", len(args[1])-1)

	}

	r.cmds["LRANGE"] = func(args []string, _ time.Duration) string {
		if len(args) < 3 {
			return "-ERR wrong number of arguments\r\n"
		}

		start, err := strconv.Atoi(args[1])
		if err != nil {
			return "-ERR\r\n"
		}

		end, err := strconv.Atoi(args[2])
		if err != nil {
			return "-ERR\r\n"
		}

		arr := r.db.LRange(args[0], start, end)
		if arr == nil {
			return "+0\r\n"
		}

		joinarr := strings.Join(arr, ",")

		return joinarr
	}

	r.cmds["SADD"] = func(args []string, _ time.Duration) string {
		if len(args) < 2 {
			return "-ERR wrong number of arguments\r\n"
		}

		r.db.SAdd(args[0], args[1:]...)
		return "+OK\r\n"
	}

	r.cmds["SMEMBERS"] = func(args []string, _ time.Duration) string {
		if len(args) != 1 {
			return "-ERR wrong number of arguments\r\n"
		}

		arr := r.db.SMembers(args[0])
		if arr == nil {
			return "+0\r\n"
		}

		joinarr := strings.Join(arr, ",")

		return joinarr

	}

	r.cmds["HGET"] = func(args []string, _ time.Duration) string {
		if len(args) != 2 {
			return "-ERR wrong number of arguments\r\n"
		}

		if val, ok := r.db.HGet(args[0], args[1]); ok {
			return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
		}
		return "*0\r\n"
	}

	r.cmds["HSET"] = func(args []string, _ time.Duration) string {
		if len(args) < 3 {
			return "-ERR wrong number of arguments\r\n"
		}

		r.db.HSet(args[0], args[1], args[2])
		return "+OK\r\n"
	}

	r.cmds["HGETALL"] = func(args []string, _ time.Duration) string {
		if len(args) != 1 {
			return "-ERR wrong number of arguments\r\n"
		}

		arr := r.db.HGetAll(args[0])
		if arr == nil {
			return "+0\r\n"
		}

		joinarr := strings.Join(arr, ",")

		return joinarr
	}

	r.cmds["FLUSHALL"] = func(args []string, _ time.Duration) string {
		r.db.Flush()
		return "+OK\r\n"
	}

	return r
}

func (r *Registry) Execute(cmd string, args []string, ttl time.Duration) string {
	if fn, ok := r.cmds[cmd]; ok {
		return fn(args, ttl)
	}

	return "-ERR unknown command\r\n"
}
