package rawkafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// KafkaCodec represents an Avro codec to encode the raw HTTP messages info
type KafkaCodec struct {
	ValueSchema   string
	ValueSchemaID int
}

// SchemaDefaultLocation contains the default path to the request schema
var SchemaDefaultLocation = "./request.avsc"

// LoadKafkaCodec loads a new codec from the schema path
func LoadKafkaCodec(schemaPath string) (*KafkaCodec, error) {
	content, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	codec := KafkaCodec{
		ValueSchema: string(content),
	}

	return &codec, nil
}

// Register registers this Kafka codec to the Schema Registry
func (codec *KafkaCodec) Register(endpoint, topic string) error {
	id, err := registerSchema(codec.ValueSchema, "value", endpoint, topic)
	if err != nil {
		return err
	}

	codec.ValueSchemaID = id
	return nil
}

// Restify encodes the data of the raw request into a Kafka-REST compatible
// format
func (codec *KafkaCodec) Restify(req *RequestInfo) ([]byte, error) {
	reqData := map[string]interface{}{
		"value_schema_id": codec.ValueSchemaID,
		"records": []interface{}{
			map[string]interface{}{
				"value": req,
			},
		},
	}

	return json.Marshal(reqData)
}

// registerSchema registers an Avro schema to the Schema registry endpoint
func registerSchema(schema, schemaType, endpoint, topic string) (int, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"schema": schema,
	})
	requestBuffer := bytes.NewBuffer(requestBody)
	if err != nil {
		log.Printf("failed to create a bytes buffer for the request body\n")
		return 0, err
	}
	if !strings.Contains(endpoint, "://") {
		log.Printf("bad endpoint: %#+v\n", endpoint)
		return 0, fmt.Errorf("Bad endpoint value: %#v", endpoint)
	}
	if topic == "" {
		log.Printf("bad topic: %#+v\n", topic)
		return 0, fmt.Errorf("Bad topic value: %#v", topic)
	}
	registerEndpoint := urljoin(endpoint, "subjects", topic+"-"+schemaType, "versions")
	log.Printf("registering schema at %#+v\n", registerEndpoint)

	client := &http.Client{}
	req, err := http.NewRequest("POST", registerEndpoint, requestBuffer)
	if err != nil {
		log.Printf("failed to create the schema registry request\n")
		return 0, err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	response, err := client.Do(req)
	if err != nil {
		log.Printf("schema register request failed\n")
		return 0, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("could't read the schema registry response\n")
		return 0, err
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("bad HTTP response from the schema registry\n")
		return 0, fmt.Errorf("Bad HTTP response while registering schema: %d", response.StatusCode)
	}

	var schemaInfo struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(content, &schemaInfo); err != nil {
		log.Printf("couldn't unmarshal the JSON in the response from the schema registry\n")
		return 0, err
	}

	return schemaInfo.ID, nil
}

func urljoin(baseURL string, parts ...string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	parts = append([]string{u.Path}, parts...)
	u.Path = path.Join(parts...)
	return u.String()
}
