
go-graphql-subscription-example  
===============================  
  
[![git3moji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg?style=flat-square)](https://gitmoji.carloscuesta.me)  
  
> Project that demonstrates GraphQL [subscriptions (over Websocket)](https://github.com/apollographql/subscriptions-transport-ws/blob/v0.9.4/PROTOCOL.md) to consume [Apache Kafka](https://kafka.apache.org/) messages.  
  
## Project

### Pre-requisites  
  
**Requires Go 1.11.x** or above, which support Go modules. Read more about them [here](https://github.com/golang/go/wiki/Modules).  
  
### Build

The project comes with a `Makefile`, so all the main activities can be performed by [make](https://www.gnu.org/software/make/).

:warning: The source code provided is incomplete: it does not contain generated code:

- generated code for embedding the static resources

To build the project, simply invoke the `build` targets:

```sh
make build
```
