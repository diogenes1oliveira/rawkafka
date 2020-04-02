package rawkafka

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	r := require.New(t)

	testCases := []map[string]string{
		{
			"method": "GET",
			"body":   "",
		},
		{
			"method": "GET",
			"body":   "asdf",
		},
		{
			"method": "POST",
			"body":   "",
		},
		{
			"method": "POST",
			"body":   "asdf",
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("--> Test case %d: method = %s, body = %s <--\n", i+1, testCase["method"], testCase["body"])
		testSampleRequest(r, testCase["method"], []byte(testCase["body"]))
	}
}

func testSampleRequest(r *require.Assertions, method string, body []byte) {
	req := RequestInfo{}
	r.NoError(faker.SetRandomMapAndSliceSize(5))
	r.NoError(faker.SetRandomStringLength(5))

	for {
		r.NoError(faker.FakeData(&req))

		if len(req.Headers) == 0 || len(req.IP) == 0 || len(req.URL) == 0 {
			continue
		}

		break
	}

	req.Method = method
	req.Body = body
	reader := bytes.NewReader(req.Body)

	ip := fmt.Sprintf("%s:8080", req.IP)

	reqHTTP, err := http.NewRequest(req.Method, req.URL, reader)
	reqHTTP.Header = req.Headers
	r.NoErrorf(err, "failed to create a request")
	reqHTTP.RemoteAddr = ip

	parsedReq := RequestInfo{}
	parsedReq.Parse(reqHTTP)

	r.EqualValuesf(req.Headers, parsedReq.Headers, "headers don't match")
	r.EqualValuesf(req.IP, parsedReq.IP, "IP doesn't match")
	r.EqualValuesf(req.Method, parsedReq.Method, "method doesn't match")
	r.EqualValuesf(req.URL, parsedReq.URL, "params don't match")
	r.EqualValuesf(req.Body, parsedReq.Body, "body doesn't match")
}

func TestGetIPFromRequest(t *testing.T) {
	r := require.New(t)

	var createTestRequest = func(ip string, url string) *http.Request {
		reader := bytes.NewReader([]byte{})
		if url == "" {
			url = "http://example.com/"
		}
		req, err := http.NewRequest("GET", url, reader)
		r.NoError(err, "Failed to create the request")
		if ip != "" {
			for headerName := range HeadersIPForwarding {
				req.Header.Set(headerName, ip)
				break
			}
		}
		return req
	}

	parsedIP, err := GetIPFromRequest(createTestRequest("", ""))
	r.Errorf(err, "Should fail with no IP")

	i := 0

	for headerName := range HeadersIPForwarding {
		i++
		ip := fmt.Sprintf("%d.%d.%d.%d", i+1, i+1, i+1, i+1)
		req := createTestRequest("", "")
		req.Header.Set(headerName, ip)

		parsedIP, err = GetIPFromRequest(req)
		r.NoErrorf(err, "Failed to parse the IP")
		r.Equal(ip, parsedIP)
	}

}

func TestGetHeaderListFromEnv(t *testing.T) {
	r := require.New(t)
	envName := "RANDOM_ENV_" + strings.Replace(uuid.New().String(), "-", "", -1)
	defaultValue := "header-a,header-b,header-c"
	defaultMap := map[string]bool{
		"Header-A": true,
		"Header-B": true,
		"Header-C": true,
	}

	fmt.Printf("--> Test case 1: unset variable <--\n")
	os.Unsetenv(envName)
	r.EqualValues(defaultMap, getHeaderListFromEnv(envName, defaultValue))

	fmt.Printf("--> Test case 2: empty variable <--\n")
	os.Setenv(envName, "")
	r.EqualValues(defaultMap, getHeaderListFromEnv(envName, defaultValue))

	fmt.Printf("--> Test case 3: set variable <--\n")
	os.Setenv(envName, "header-d")
	r.EqualValues(map[string]bool{"Header-D": true}, getHeaderListFromEnv(envName, defaultValue))
}
