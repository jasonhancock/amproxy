package server

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cheekybits/is"
	"github.com/go-kit/kit/log"
)

func TestLineProtocolServer(t *testing.T) {
	is := is.New(t)

	logger := testLogger()

	var lines []string

	fn := func(line string) {
		lines = append(lines, line)
	}

	s, err := NewLineProtocolServer(log.With(logger, "component", "line_protocol_server"), ":8095", fn)
	is.NoErr(s.Run())
	defer s.Stop()

	conn, err := net.Dial("tcp", ":8095")
	is.NoErr(err)
	fmt.Fprintf(conn, "this is line 1\n")
	fmt.Fprintf(conn, "this is line 2\n")
	fmt.Fprintf(conn, "\n")
	conn.Close()

	time.Sleep(300 * time.Millisecond)

	is.Equal(len(lines), 2)
	is.Equal("this is line 1", lines[0])
	is.Equal("this is line 2", lines[1])
}
