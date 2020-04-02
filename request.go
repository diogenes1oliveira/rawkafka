package rawkafka

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// HeadersExcluded contains a map with the HTTP headers that should not be
// stored or processed. It is extracted from the environment variable
// KAFKA_RAW_HTTP_EXCLUDED_HEADERS and by default excludes the 'Cookie' header
var HeadersExcluded = getHeaderListFromEnv("KAFKA_RAW_HEADERS_HTTP_EXCLUDED", "Cookie")

// HeadersIPForwarding contains a map with the HTTP headers that should be
// considered for IP forwarding. It is extracted from the environment variable
// KAFKA_RAW_HTTP_HEADERS_IP_FORWARDING and by default considers the headers
// X-Forwarded-For and X-Real-Ip in this order
var HeadersIPForwarding = getHeaderListFromEnv("KAFKA_RAW_HTTP_HEADERS_IP_FORWARDING", "X-Forwarded-For,X-Real-Ip")

// RequestInfo contains the information for a raw HTTP request
type RequestInfo struct {
	Headers     http.Header `json:"headers"`
	IP          string      `json:"ip" faker:"ipv4"`
	Method      string      `json:"method"`
	ServerTime  time.Time   `json:"server_time"`
	URL         string      `json:"url" faker:"url"`
	Body        []byte      `json:"body"`
	ParseErrors []string
}

// Parse fills up the raw request struct with info parsed from the HTTP request
func (reqInfo *RequestInfo) Parse(req *http.Request) {
	var err error

	reqInfo.Headers = req.Header
	reqInfo.ParseErrors = []string{}

	if reqInfo.IP, err = GetIPFromRequest(req); err != nil {
		reqInfo.ParseErrors = append(reqInfo.ParseErrors, err.Error())
		reqInfo.IP = ""
	}
	reqInfo.Method = req.Method
	reqInfo.ServerTime = time.Now()
	reqInfo.URL = req.URL.String()
	if reqInfo.Body, err = ioutil.ReadAll(req.Body); err != nil {
		reqInfo.ParseErrors = append(reqInfo.ParseErrors, err.Error())
	}

}

// GetIPFromRequest extracts the user IP from the headers, falling back to the
// IP from the request itself
func GetIPFromRequest(req *http.Request) (string, error) {
	ip := ""

	for headerName := range HeadersIPForwarding {
		ip = req.Header.Get(headerName)
		if net.ParseIP(ip) != nil {
			return ip, nil
		}
	}

	userIP, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	if net.ParseIP(userIP) == nil {
		return "", fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	return userIP, nil
}

func getHeaderListFromEnv(envName string, defaultValue string) map[string]bool {
	envValue := os.Getenv(envName)
	if envValue == "" {
		envValue = defaultValue
	}

	headers := map[string]bool{}

	for _, header := range strings.Split(envValue, ",") {
		canonicalName := http.CanonicalHeaderKey(header)
		headers[canonicalName] = true
	}

	return headers
}
