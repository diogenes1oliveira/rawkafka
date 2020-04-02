package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/diogenes1oliveira/rawkafka"
	"github.com/jessevdk/go-flags"
)

var requestContentType = "application/vnd.kafka.avro.v2+json"
var responseContentType = "application/vnd.kafka.v1+json"

var defaultError = map[string]interface{}{
	"error_code": 500,
	"message":    "rawkafka: internal server error",
}

// RequestLoggerFunc is a function that acts as a logger for HTTP requests
type RequestLoggerFunc func(statusCode int, message string, args ...interface{})

// HandleRequest sends the request data to a Kafka REST endpoint
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	t0 := time.Now()
	reqInfo := rawkafka.RequestInfo{}

	log := func(statusCode int, message string, args ...interface{}) {
		timeSpent := float64(time.Since(t0).Microseconds()) / 1000.0
		prefix := fmt.Sprintf("%s - %s %s - [%d %.2f ms] - ", reqInfo.IP, r.Method, r.URL.String(), statusCode, timeSpent)
		log.Printf(prefix+message, args...)
	}

	reqInfo.Parse(r)

	reqBody, err := codec.Restify(&reqInfo)
	if err != nil {
		respondWithError(w, err, log)
		log(500, "ERROR: %v\n", err.Error())
		return
	}

	response, err := http.Post(cmdFlags.RestEndpoint, requestContentType, bytes.NewBuffer(reqBody))
	if err != nil {
		respondWithError(w, err, log)
		log(500, "ERROR: %v\n", err.Error())
		return
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		respondWithError(w, err, log)
		log(500, "ERROR: %v\n", err.Error())
		return
	}

	for header, values := range response.Header {
		for _, value := range values {
			w.Header().Add(header, value)
		}
	}

	w.WriteHeader(response.StatusCode)
	if _, err := w.Write(responseBody); err != nil {
		log(0, "ERROR: Failed to write the response body - %v\n", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		log(response.StatusCode, "Bad response from Kafka REST: %s\n", string(responseBody))
	} else {
		log(response.StatusCode, "%s\n", string(responseBody))
	}
}

// codec contains the encoder of requests to Kafka REST
var codec *rawkafka.KafkaCodec

// cmdFlags contains the user-defined config parameters
var cmdFlags = struct {
	Port              int    `long:"port" env:"RAWKAFKA_PORT" default:"7000" description:"Port to bind to"`
	Host              string `long:"host" env:"RAWKAFKA_HOST" default:"0.0.0.0" description:"Host IP to bind to"`
	Topic             string `long:"topic" env:"RAWKAFKA_TOPIC" default:"RawRequest" description:"Name of the topic to publish the messages to"`
	RestEndpoint      string `long:"rest-endpoint" env:"RAWKAFKA_REST_ENDPOINT" description:"Kafka REST endpoint"`
	SchemaLocation    string `long:"schema-location" env:"RAWKAFKA_SCHEMA_LOCATION" default:"./request.avsc" description:"Avro schema location"`
	SchemaRegistryURL string `long:"schema-registry-url" env:"RAWKAFKA_SCHEMA_REGISTRY_URL" description:"Schema registry URL"`
}{}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error

	if _, err = flags.Parse(&cmdFlags); err != nil {
		os.Exit(0)
	}

	codec, err = rawkafka.LoadKafkaCodec(cmdFlags.SchemaLocation)
	check(err)

	err = codec.Register(cmdFlags.SchemaRegistryURL, cmdFlags.Topic)
	check(err)

	addr := fmt.Sprintf("%s:%d", cmdFlags.Host, cmdFlags.Port)
	http.HandleFunc("/", HandleRequest)
	http.HandleFunc("/ping", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "pong")
	})

	log.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func check(err error) {
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

var defaultErrorMessage = func() []byte {
	content, err := json.MarshalIndent(defaultError, "", "  ")
	if err != nil {
		panic(err)
	}
	return content
}()

func respondWithError(w http.ResponseWriter, err error, log RequestLoggerFunc) {
	errMessage := err.Error()
	reqBody, err := json.MarshalIndent(map[string]interface{}{
		"error_code": 500,
		"message":    "rawkafka: " + errMessage,
	}, "", "  ")

	if err != nil {
		reqBody = defaultErrorMessage
	}

	w.Header().Set("Content-type", responseContentType)
	w.WriteHeader(http.StatusInternalServerError)
	if _, err = w.Write(reqBody); err != nil {
		log(0, "ERROR: Failed to write the error response body - %v\n", err)
	}
}
