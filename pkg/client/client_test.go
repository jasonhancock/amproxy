package client

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jasonhancock/amproxy/pkg/amproxy"
	"github.com/jasonhancock/amproxy/pkg/lineserver"
	"github.com/jasonhancock/go-logger"
	helpers "github.com/jasonhancock/go-testhelpers/generic"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	l := logger.Silence()

	t.Run("normal", func(t *testing.T) {
		port := helpers.NewRandomPort(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var messages []string
		var wg sync.WaitGroup
		_, err := lineserver.New(ctx, l, &wg, port, func(str string) {
			messages = append(messages, str)
		})
		require.NoError(t, err)

		c := NewClient("my_pub_key", "my_priv_key", port)
		require.NoError(t, c.Connect())

		require.NoError(t, c.Write(amproxy.Message{Name: "name1", Value: "value1", Timestamp: time.Unix(1234, 0)}))
		require.NoError(t, c.Write(amproxy.Message{Name: "name2", Value: "value2", Timestamp: time.Unix(12345, 0)}))
		require.NoError(t, c.Write(amproxy.Message{Name: "name3", Value: "value3", Timestamp: time.Unix(123456, 0)}))
		require.NoError(t, c.Disconnect())
		time.Sleep(300 * time.Millisecond)

		require.Len(t, messages, 3)
		require.Equal(t, "name1 value1 1234 my_pub_key ANX1Szr2bbcU04m/ZgAQKB3/OZ26pIZeIM2D+NfOGUY=", messages[0])
		require.Equal(t, "name2 value2 12345 my_pub_key J7AEAsGwrGr4SZdZzLnN48GSIgOpDj7IJ8rqKbu4vsU=", messages[1])
		require.Equal(t, "name3 value3 123456 my_pub_key wsLfIF/tpOWaJ6/JPjewLr17wG53LKN368QJagTfDlU=", messages[2])
	})

	// This test seems a bit flaky....figure out the problem.
	t.Run("write error", func(t *testing.T) {
		t.Skip("flaky test...need to debug")
		port := helpers.NewRandomPort(t)
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())
		_, err := lineserver.New(ctx, l, &wg, port, func(str string) {})
		require.NoError(t, err)

		c := NewClient("my_pub_key", "my_priv_key", port)
		require.NoError(t, c.Connect())
		cancel()
		wg.Wait()

		require.Error(t, c.Write(amproxy.Message{Name: "name1", Value: "value1", Timestamp: time.Unix(1234, 0)}))
	})

	t.Run("not connected error", func(t *testing.T) {
		port := helpers.NewRandomPort(t)
		c := NewClient("my_pub_key", "my_priv_key", port)

		require.Equal(t, ErrorNotConnected, c.Write(amproxy.Message{Name: "name1", Value: "value1", Timestamp: time.Unix(1234, 0)}))
	})
}
