go-graphql-subscription-example
===============================
    
[![build-status](https://img.shields.io/travis/ccamel/go-graphql-subscription-example.svg?logo=travis&style=flat-square)](https://travis-ci.org/ccamel/go-graphql-subscription-example) [![go-report-card](https://goreportcard.com/badge/github.com/ccamel/go-graphql-subscription-example)](https://goreportcard.com/report/github.com/ccamel/go-graphql-subscription-example)
[![git3moji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg?style=flat-square)](https://gitmoji.carloscuesta.me)
[![License](https://img.shields.io/github/license/ccamel/go-graphql-subscription-example.svg?style=flat-square)]( https://github.com/ccamel/go-graphql-subscription-example/blob/master/LICENSE)

> Project that demonstrates [graphQL] [subscriptions (over Websocket)](https://github.com/apollographql/subscriptions-transport-ws/blob/v0.9.4/PROTOCOL.md) to consume [Apache Kafka](https://kafka.apache.org/) messages.    

## Purpose

This repository implements a simple service allowing clients to consume messages from a kafka topic through a [graphQL](https://graphql.org/) subscription endpoint.

<p align="center">
  <img src="https://github.com/ccamel/go-graphql-subscription-example/raw/master/doc/overview.png" title="overview">
</p>

This particular example demonstrates how to perform basic operations such as:

- serve a [graphiQL](https://github.com/graphql/graphiql) page
- implement a subscription resolver using WebSocket transport
- implement custom [graphQL] _scalars_
- consume Kafka messages
- filter messages using an expression evaluator
- ...

## Pre-requisites
    
 **Requires Go 1.11.x** or above, which support Go modules. Read more about them [here](https://github.com/golang/go/wiki/Modules).    
    
## Build  
  
The project comes with a `Makefile`, so all the main activities can be performed by [make](https://www.gnu.org/software/make/).  
  
:warning: The source code provided is incomplete: it does not contain generated code:  
  
- generated code for embedding the static resources  
  
To build the project, simply invoke the `build` targets:  

```sh  
make build  
```

Alternately, the project can be build by docker:

```sh  
make dockerize  
```

Command will produce the image `ccamel/go-graphql-subscription-example`.

## How to use

### 1. Start Kafka server

At first, kafka must be started. See [official documentation](https://kafka.apache.org/quickstart) for more.

Kafka uses ZooKeeper so you need to first start a ZooKeeper server if you don't already have one.

```sh
> bin/zookeeper-server-start.sh config/zookeeper.properties
```  

Now start the Kafka server:

```sh
> bin/kafka-server-start.sh config/server.properties
```  

### 2. Create topics

For the purpose of the demo, some topics shall be created. So, let's create 2 topics named `topic-a` and `topic-b`,
with a single partition and only one replica:

```sh
> bin/kafka-topics.sh --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic topic-a
> bin/kafka-topics.sh --create --bootstrap-server localhost:9092 --replication-factor 1 --partitions 1 --topic topic-b
```

### 3. Start the GraphQL server

The configuration is pretty straightforward:

```sh
> ./go-graphql-subscription-example --help
Usage:
  go-graphql-subscription-example [flags]

Flags:
      --brokers strings   The list of broker addresses used to connect to the kafka cluster (default [localhost:9092])
  -h, --help              help for go-graphql-subscription-example
      --port uint16       The listening port (default 8000)
      --topics strings    The list of kafka topics that subscribers can consume (default [foo])
```

Run the application which exposes the 2 previously created topics to subscribers: 

```sh
> ./go-graphql-subscription-example --topics topic-a,topic-b 
```

Alternately, if the docker image has been previously built, the container can be started this way:

```sh
> docker run -ti --rm -p 8000:8000 ccamel/go-graphql-subscription-example --topics topic-a,topic-b
``` 

### 4. Subscribe

The application exposes a graphql endpoint through which clients can receive messages coming from a kafka topic.

Navigate to `http://localhost:8000/graphiql` URL and submit the subscription to the topic `topic-a`.

```graphql
subscription {
  event(on: "topic-a")
}
```

The offset id to consume from can also be specified. Negative values have a special meaning:
- `-1`: the most recent offset available for a partition (end)
- `-2`: the least recent offset available for a partition (beginning)

```graphql
subscription {
  event(on: "topic-a", at: -1)
}
```

Additionally, a filter expression can be specified. The events consumed are then only ones matching the given predicate.

```graphql
subscription {
  event(
    on: "topic-a",
    at: -1,
    matching: "value > 8"
  )
}
```

### 5. Push messages

Run the producer and then type a few messages into the console to send to Kafka. Note that messages shall be 
[JSON objects](https://www.json.org/).

```sh
> bin/kafka-console-producer.sh --broker-list localhost:9092 --topic topic-a
{ "message": "hello world !", "value": 14 }
``` 

The message should be displayed on the browser.

## Stack    

### Technical

This application mainly uses:    
    
* **GraphQL**
 
  ↳ [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go)
  
  ↳ [graph-gophers/graphql-transport-ws](https://github.com/graph-gophers/graphql-transport-ws) 
  
  ↳ [graphql/graphiql](https://github.com/graphql/graphiql)       

* **Kafka**
 
  ↳ [segment-integrations/connect-kafka](https://github.com/segment-integrations/connect-kafka)
  
* **Expression language**
 
  ↳ [antonmedv/expr](https://github.com/antonmedv/expr)

* **Design Patterns**

  ↳ [ReactiveX/RxGo v2](https://github.com/ReactiveX/RxGo/tree/v2)

* **CLI**
 
  ↳ [spf13/cobra](https://github.com/spf13/cobra)  

* **Log** 

  ↳ [rs/zerolog](https://github.com/rs/zerolog)  

### Project

* **Build**

  ↳ [make](https://www.gnu.org/software/make/)

* **Linter**  

  ↳ [golangci-lint](https://github.com/golangci/golangci-lint)

[graphQL]: https://graphql.org/
