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

	case "GET", "DEL":
		// Expected format: GET key (or DEL key)
		if len(args) < 1 {
			return "", nil, 0, fmt.Errorf("error: %s requires a key (e.g., %s foo)", cmd, cmd)
		}

		key = args[0]
        
        // Return structured args: [key]
		return cmd, []string{key}, 0, nil // TTL is irrelevant, so 0

	default:
		return "", nil, 0, fmt.Errorf("error: unknown command: %s", cmd)
	}
}