package lineserver

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/jasonhancock/go-logger"
	"github.com/stretchr/testify/require"
)

func TestLineProtocolServer(t *testing.T) {
	var lines []string
	fn := func(line string) {
		lines = append(lines, line)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	_, err := New(ctx, logger.Silence(), &wg, "localhost:8095", fn)
	require.NoError(t, err)
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", "localhost:8095")
	require.NoError(t, err)
	fmt.Fprintf(conn, "this is line 1\n")
	fmt.Fprintf(conn, "this is line 2\n")
	fmt.Fprintf(conn, "\n")
	conn.Close()

	time.Sleep(200 * time.Millisecond)
	cancel()
	wg.Wait()

	require.Len(t, lines, 2)
	require.Equal(t, "this is line 1", lines[0])
	require.Equal(t, "this is line 2", lines[1])
}
