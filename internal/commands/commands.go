package commands

import (
	"fmt"
	"redis-go/internal/db"
	"time"
)

type CommandFunc func(args []string, ttl time.Duration) string

type Registry struct {
	db   *db.DB
	cmds map[string]CommandFunc
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

	return r
}

func (r *Registry) Execute(cmd string, args []string, ttl time.Duration) string {
	if fn, ok := r.cmds[cmd]; ok {
		return fn(args,ttl)
	}

	return "-ERR unknown command\r\n"
}
