# rawkafka

Forward raw HTTP requests to a Kafka cluster

![data transformation in rawkafka](https://github.com/diogenes1oliveira/rawkafka/blob/master/diagram.svg "rawkafka - data transformation")

## About

The Kafka REST protocol requires a specific format for its requests, as you can
see in the [API Spec](https://docs.confluent.io/current/kafka-rest/api.html#post--topics-(string-topic_name)).
This service is a simple Go HTTP server that listens for all methods and paths, 
formatting the request properly and sending them to a Kafka REST endpoint.

At startup, the configured schema is automatically registered at the Schema
Registry endpoint.

Link to the current Avro schema: https://github.com/diogenes1oliveira/rawkafka/blob/master/request.avsc

## Usage

``` 
Usage:
  rawkafka [OPTIONS]

Application Options:
      --port=                Port to bind to (default: 9000) [$RAWKAFKA_PORT]
      --host=                Host IP to bind to (default: 0.0.0.0) [$RAWKAFKA_HOST]
      --topic=               Name of the topic to publish the messages to (default: RawRequest)
                             [$RAWKAFKA_TOPIC]
      --rest-endpoint=       Kafka REST endpoint [$RAWKAFKA_REST_ENDPOINT]
      --schema-location=     Avro schema location (default: ./request.avsc) [$RAWKAFKA_SCHEMA_LOCATION]
      --schema-registry-url= Schema registry URL [$RAWKAFKA_SCHEMA_REGISTRY_URL]

Help Options:
  -h, --help                 Show this help message
```

