package helper

import (
	"fmt"
	"strings"
	"time"
)

func ParseCommand(arr []string) (string, []string, time.Duration, error) {
	// 1. Basic Validation: Command must exist (at least one token)
	if len(arr) == 0 {
		return "", nil, 0, fmt.Errorf("error: empty command")
	}

	// 2. Capture and Normalize Command
	cmd := strings.ToUpper(arr[0])
	args := arr[1:] // Initialize args as all tokens after the command

	var key string
	var value string
	var ttl time.Duration
	var err error

	// 3. Command-Specific Parsing using a switch statement
	switch cmd {
	case "COMMAND":
		return cmd, args, 0, nil
	case "PING":
		return cmd, args, 0, nil
	case "SET":
		// Expected format: SET key value duration (e.g., SET foo bar 10s)
		if len(args) < 2 {
			return "", nil, 0, fmt.Errorf("error: SET requires key, value")
		}

		if len(args) == 2 {
			return cmd, args, 0, nil
		}

		key = args[0]
		value = args[1]
		// arr[3] is the duration string
		ttl, err = time.ParseDuration(args[2])
		if err != nil {
			return "", nil, 0, fmt.Errorf("error: invalid duration for SET: %w", err)
		}

		// Return structured args: [key, value]
		return cmd, []string{key, value}, ttl, nil

	case "LPUSH":
		// Expected format: LPUSH key value (e.g., LPUSH foo bar)
		if len(args) < 2 {
			return "", nil, 0, fmt.Errorf("error: LPUSH requires key, value")
		}

		key = args[0]

		if len(args) == 2 {
			return cmd, args, 0, nil
		}

		values := args[1:]

		argvals := append([]string{key}, values...)

		return cmd, argvals, 0, nil

	case "RPUSH":
		// Expected format: LPUSH key value (e.g., LPUSH foo bar)
		if len(args) < 2 {
			return "", nil, 0, fmt.Errorf("error: RPUSH requires key, value")
		}

		key = args[0]

		if len(args) == 2 {
			return cmd, args, 0, nil
		}

		values := args[1:]

		argvals := append([]string{key}, values...)

		return cmd, argvals, 0, nil

	case "LRANGE":
		// Expected format: LPUSH key value (e.g., LPUSH foo bar)
		if len(args) < 3 {
			return "", nil, 0, fmt.Errorf("error: LRANGE requires key, start and end")
		}

		key = args[0]
		start := args[1]
		end := args[2]

		return cmd, []string{key, start, end}, 0, nil
	
	case "SADD":

		if len(args) < 2 {
			return "", nil, 0, fmt.Errorf("error: SADD requires key, value")
		}

		key = args[0]

		if len(args) == 2 {
			return cmd, args, 0, nil
		}

		values := args[1:]

		argvals := append([]string{key}, values...)

		return cmd, argvals, 0, nil

	case "SMEMBERS":
		// Expected format: SMEMBERS key
		if len(args) < 1 {
			return "", nil, 0, fmt.Errorf("error: SMEMBERS requires key")
		}

		key = args[0]

		return cmd, []string{key}, 0, nil

	case "HGET":
		// Expected format: HGET key field
		if len(args) < 2 {
			return "", nil, 0, fmt.Errorf("error: HGET requires key and field")
		}

		key := args[0]
		field := args[1]

		return cmd, []string{key, field}, 0, nil

	case "HSET":
		// Expected format: HSET key field value
		if len(args) < 3 {
			return "", nil, 0, fmt.Errorf("error: HSET requires key, field, and value")
		}

		key := args[0]
		field := args[1]
		value := args[2]

		return cmd, []string{key, field, value}, 0, nil

	case "HGETALL":
		// Expected format: HGETALL key
		if len(args) < 1 {
			return "", nil, 0, fmt.Errorf("error: HGETALL requires key")
		}

		key := args[0]

		return cmd, []string{key}, 0, nil

	case "GET", "DEL":
		// Expected format: GET key (or DEL key)
		if len(args) < 1 {
			return "", nil, 0, fmt.Errorf("error: %s requires a key (e.g., %s foo)", cmd, cmd)
		}

		key = args[0]

		// Return structured args: [key]
		return cmd, []string{key}, 0, nil // TTL is irrelevant, so 0

	case "FLUSHALL":
		return cmd, args, 0, nil

	default:
		return "", nil, 0, fmt.Errorf("error: unknown command: %s", cmd)
	}
}
