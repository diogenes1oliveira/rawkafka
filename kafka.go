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
	if err != nil {
		log.Printf("Failed to marshal the schema\n")
		return 0, err
	}
	if !strings.Contains(endpoint, "://") {
		return 0, fmt.Errorf("Bad endpoint value: %#v", endpoint)
	}
	if topic == "" {
		return 0, fmt.Errorf("Bad topic value: %#v", topic)
	}
	registerEndpoint := urljoin(endpoint, "subjects", topic+"-"+schemaType, "versions")
	log.Printf("Registering a %s schema at %s\n", schemaType, registerEndpoint)

	response, err := http.Post(registerEndpoint, "application/vnd.schemaregistry.v1+json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Printf("Failed to register the schema: %v\n", err)
		return 0, err
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Failed to read the schema registry response")
		return 0, err
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("schema registry error: %s\n", string(content))
		return 0, fmt.Errorf("Error while registering schema: %d", response.StatusCode)
	}

	var schemaInfo struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(content, &schemaInfo); err != nil {
		log.Printf("failed to decode the schema registry response: %s\n", string(content))
		return 0, err
	}

	log.Printf("Parsed SchemaID = %d from the schema registry response: %s\n", schemaInfo.ID, string(content))
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
