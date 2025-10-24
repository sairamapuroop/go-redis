package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// *2\r\n$3\r\nGET\r\n$3\r\nkey\r\n

func ReadArray(r *bufio.Reader) ([]string, error) {

	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	if prefix != '*' {
		return nil, fmt.Errorf("expected array prefix, got %q", prefix)
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	count, _ := strconv.Atoi(line[:len(line)-2])

	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		if b, _ := r.ReadByte(); b != '$' {
			return nil, fmt.Errorf("expected bulk string")
		}
		lenLine, _ := r.ReadString('\n')
		length, _ := strconv.Atoi(lenLine[:len(lenLine)-2])
		data := make([]byte, length)

		io.ReadFull(r, data)
		r.ReadString('\n')
		result = append(result, string(data))

	}

	return result, nil
}
