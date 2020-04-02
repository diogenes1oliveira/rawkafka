package rawkafka

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
		return 0, err
	}
	if !strings.Contains(endpoint, "://") {
		return 0, fmt.Errorf("Bad endpoint value: %#v", endpoint)
	}
	if topic == "" {
		return 0, fmt.Errorf("Bad topic value: %#v", topic)
	}
	registerEndpoint := urljoin(endpoint, "subjects", topic+"-"+schemaType, "versions")

	response, err := http.Post(registerEndpoint, "application/vnd.schemaregistry.v1+json", bytes.NewBuffer(requestBody))
	if err != nil {
		return 0, err
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Error while registering schema: %d", response.StatusCode)
	}

	var schemaInfo struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(content, &schemaInfo); err != nil {
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
