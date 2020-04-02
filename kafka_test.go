package rawkafka

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"
)

func testRunEchoServer(response string, port int) (string, *http.Server, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	server := &http.Server{Addr: addr}

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		w.Write([]byte(response))
	})

	go func() {
		server.ListenAndServe()
	}()

	start := time.Now()
	client := http.Client{
		Timeout: 50 * time.Millisecond,
	}
	for {
		resp, err := client.Get("http://" + addr)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if time.Now().Sub(start).Milliseconds() > 1000.0 {
			return "", nil, fmt.Errorf("timeout expired")
		}
		time.Sleep(50 * time.Millisecond)
	}

	return addr, server, nil
}

func TestKafkaCodecRegister(t *testing.T) {
	r := require.New(t)
	port, err := freeport.GetFreePort()
	r.NoErrorf(err, "couldn't get a free port")
	addr, server, err := testRunEchoServer(`{"id": 999}`, port)
	r.NoErrorf(err, "failed to start the server")

	defer server.Shutdown(context.Background())
	addr = "http://" + addr

	codec, err := LoadKafkaCodec(SchemaDefaultLocation)
	r.NoErrorf(err, "failed to load the schema")

	err = codec.Register(addr, "AnyTopic")
	r.NoErrorf(err, "Failed to register the schema")
	r.Equalf(999, codec.ValueSchemaID, "Failed to parse the response")
}

func TestKafkaCodecRestify(t *testing.T) {
	r := require.New(t)
	codec, err := LoadKafkaCodec(SchemaDefaultLocation)
	r.NoErrorf(err, "failed to load the schema")

	requestInfo := RequestInfo{}
	r.NoError(faker.FakeData(&requestInfo))

	_, err = codec.Restify(&requestInfo)
	r.NoErrorf(err, "failed to Restify a request info")
}
