# go-graphql-subscription-example

[![lint](https://img.shields.io/github/actions/workflow/status/ccamel/go-graphql-subscription-example/lint.yml?label=lint&logo=github)](https://github.com/ccamel/go-graphql-subscription-example/actions/workflows/lint.yml)
[![build](https://img.shields.io/github/actions/workflow/status/ccamel/go-graphql-subscription-example/build.yml?label=build&logo=github)](https://github.com/ccamel/go-graphql-subscription-example/actions/workflows/build.yml)
![go](https://img.shields.io/github/go-mod/go-version/ccamel/go-graphql-subscription-example?logo=go)
[![go-report-card](https://goreportcard.com/badge/github.com/ccamel/go-graphql-subscription-example/master)](https://goreportcard.com/report/github.com/ccamel/go-graphql-subscription-example)
[![maintainability](https://api.codeclimate.com/v1/badges/67162ec92b2fb97bdb3e/maintainability)](https://codeclimate.com/github/ccamel/go-graphql-subscription-example/maintainability)
[![quality-gate-status](https://sonarcloud.io/api/project_badges/measure?project=ccamel_go-graphql-subscription-example&metric=alert_status)](https://sonarcloud.io/dashboard?id=ccamel_go-graphql-subscription-example)
[![lines-of-code](https://sonarcloud.io/api/project_badges/measure?project=ccamel_go-graphql-subscription-example&metric=ncloc)](https://sonarcloud.io/dashboard?id=ccamel_go-graphql-subscription-example)
[![stackshare](https://img.shields.io/badge/Stackshare-%23ffffff.svg?logo=stackshare&logoColor=0690FA)](https://stackshare.io/ccamel/go-graphql-subscription-example)
[![git3moji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg?style=flat-square)](https://gitmoji.carloscuesta.me)
[![magefile](https://magefile.org/badge.svg)](https://magefile.org)
[![license](https://img.shields.io/github/license/ccamel/go-graphql-subscription-example.svg?style=flat-square)](https://github.com/ccamel/go-graphql-subscription-example/blob/master/LICENSE)
[![fossa-status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fccamel%2Fgo-graphql-subscription-example.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fccamel%2Fgo-graphql-subscription-example?ref=badge_shield)

> Project that demonstrates [graphQL] [subscriptions (over Websocket)](https://github.com/apollographql/subscriptions-transport-ws/blob/v0.9.4/PROTOCOL.md) to consume pre-configured topics from different kinds of
> stream sources like [Apache Kafka](https://kafka.apache.org/), [redis](https://redis.io/), [NSQ](https://nsq.io)...

## Purpose

This repository implements a simple service allowing clients to consume messages from a topics/channels through a [graphQL](https://graphql.org/) subscription endpoint.

<p align="center">
  <img src="doc/overview.png" title="overview">
</p>

This particular example demonstrates how to perform basic operations such as:

- serve a [graphiQL](https://github.com/graphql/graphiql) page
- expose a [Prometheus](https://prometheus.io/) endpoint
- implement a subscription resolver using WebSocket transport (compliant with [Apollo v0.9.16 protocol](https://github.com/apollographql/subscriptions-transport-ws/blob/v0.9.16/PROTOCOL.md))
- implement custom [graphQL] _scalars_
- consumer following kind of stream sources:
  - [Apache Kafka](https://kafka.apache.org/) -
        an open-source stream-processing software which aims to provide a unified,
        high-throughput, low-latency platform for handling real-time data feeds.
  - [Redis Streams](https://redis.io/) -
        an open source, in-memory data structure store, message broker with [streaming](https://redis.io/topics/streams-intro) capabilities.
  - [NSQ](https://nsq.io) -
        a realtime distributed messaging platform designed to operate at scale.

- process messages using [reactive streams](http://reactivex.io/)
- filter messages using an expression evaluator
- ...

## Pre-requisites

 **Requires Go 1.14.x** or above, which support Go modules. Read more about them [here](https://github.com/golang/go/wiki/Modules).

## Build

The project comes with a `magefile.go`, so all the main activities can be performed by [mage](https://magefile.org).

:warning: The source code provided is incomplete - build needs a code generation phase, especially for the embedding of the static resources.

To build the project, simply invoke the `build` target:

```sh
mage build
```

Alternately, the project can be build by [docker](https://www.docker.com/):

```sh
mage docker
```

Command will produce the image `ccamel/go-graphql-subscription-example`.

## How to play with it using Kafka?

### 1. Start Kafka server

At first, kafka must be started. See [official documentation](https://kafka.apache.org/quickstart) for more.

Kafka uses [ZooKeeper](https://zookeeper.apache.org/) so you need to first start a ZooKeeper server if you don't already have one.

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
  -h, --help             help for go-graphql-subscription-example
      --port uint16      The listening port (default 8000)
      --source string    The URI of the source to connect to
      --topics strings   The list of topics/stream names that subscribers can consume (default [foo])
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
You can refer to [antonmedv/expr] for an overview of the syntax to use to write predicates.

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

## How to play with it using Redis?

⚠️ Redis implementation does not support offsets (i.e. the capability to resume at some point in time).

### 1. Start Redis

At first, a redis server (at least v5.0) must be started. See [official documentation](https://redis.io/download) for more.

### 2. Start the GraphQL server

Run the application which exposes the 2 previously created topics to subscribers:

```sh
> ./go-graphql-subscription-example --source redis://6379?name=foo --topics topic-a,topic-b
```

Alternately, if the docker image has been previously built, the container can be started this way:

```sh
> docker run -ti --rm -p 8000:8000 ccamel/go-graphql-subscription-example --source redis://6379?name=foo --topics topic-a,topic-b
```

### 3. Subscribe

The application exposes a graphql endpoint through which clients can receive messages coming from a redis stream.

Navigate to `http://localhost:8000/graphiql` URL and submit the subscription to the topic `topic-a`.

```graphql
subscription {
  event(on: "topic-a")
}
```

Additionally, a filter expression can be specified. The events consumed are then only ones matching the given predicate.
You can refer to [antonmedv/expr] for an overview of the syntax to use to write predicates.

```graphql
subscription {
  event(
    on: "topic-a",
    matching: "message contains \"hello\""
  )
}
```

### 4. Push messages

Start the `redis-cli` and then use the `XADD` command to send the messages to the Redis stream.

```sh
> redis-cli
127.0.0.1:6379> XADD topic-a * message "hello world !" "value" "14"
```

The message should be displayed on the browser.

## How to play with it using NSQ?

⚠️ NSQ implementation does not support offsets (i.e. the capability to resume at some point in time).

### 1. Start NSQ

At first, NSQ must be started. See [official documentation](https://nsq.io/overview/quick_start.html) for more.

```sh
> nsqlookupd
> nsqd --lookupd-tcp-address=127.0.0.1:4160
> nsqadmin --lookupd-http-address=127.0.0.1:4161
```

### 2. Start the GraphQL server

Run the application which exposes the 2 previously created topics to subscribers:

```sh
> ./go-graphql-subscription-example --source nsq: --topics topic-a,topic-b
```

Alternately, if the docker image has been previously built, the container can be started this way:

```sh
> docker run -ti --rm -p 8000:8000 ccamel/go-graphql-subscription-example --source nsq: --topics topic-a,topic-b
```

### 3. Subscribe

The application exposes a graphql endpoint through which clients can receive messages coming from a redis stream.

Navigate to `http://localhost:8000/graphiql` URL and submit the subscription to the topic `topic-a`.

```graphql
subscription {
  event(on: "topic-a")
}
```

Additionally, a filter expression can be specified. The events consumed are then only ones matching the given predicate.
You can refer to [antonmedv/expr] for an overview of the syntax to use to write predicates.

```graphql
subscription {
  event(
    on: "topic-a",
    matching: "message contains \"hello\""
  )
}
```

### 4. Push messages

Publish a message to the topic `topic-a` by using the command line below:

```sh
> curl -d '{ "message": "hello world !", "value": 14 }' 'http://127.0.0.1:4151/pub?topic=topic-a'
```

The message should be displayed on the browser.

## Stack

### Technical

This application mainly uses:

- **GraphQL**

    ↳ [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go)

    ↳ [graph-gophers/graphql-transport-ws](https://github.com/graph-gophers/graphql-transport-ws)

    ↳ [graphql/graphiql](https://github.com/graphql/graphiql)

- **Prometheus**

    ↳ [prometheus/client_golang](https://github.com/prometheus/client_golang)

- **Kafka**

    ↳ [segment-integrations/connect-kafka](https://github.com/segment-integrations/connect-kafka)

- **Redis**

    ↳ [robinjoseph08/redisqueue](https://github.com/robinjoseph08/redisqueue)

- **NSQ**

    ↳ [nsqio/go-nsq](https://github.com/nsqio/go-nsq)

- **Expression language**

    ↳ [antonmedv/expr]

- **Design Patterns**

    ↳ [ReactiveX/RxGo v2](https://github.com/ReactiveX/RxGo/tree/v2.0.0)

- **CLI**

    ↳ [spf13/cobra](https://github.com/spf13/cobra)

- **Log**

    ↳ [rs/zerolog](https://github.com/rs/zerolog)

### Project

- **Build**

    ↳ [mage](https://magefile.org)

- **Linter**

    ↳ [golangci-lint](https://github.com/golangci/golangci-lint)

## License

[MIT] © [Chris Camel]

[![fossa-status-large](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fccamel%2Fgo-graphql-subscription-example.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fccamel%2Fgo-graphql-subscription-example?ref=badge_large)

[antonmedv/expr]: https://github.com/antonmedv/expr

[graphQL]: https://graphql.org/
