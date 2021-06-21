### PokeAPI

This is my semi-production ready implementation of the PokeAPI. It currently has 5 main components.

* The CLI configuration to run and configure the server from the command line
* The HTTP server implementation which exposes the endpoints:
    * **/pokemon/{name}** - in which we can retrieve trivial information on a pokemon
    * **/pokemon/{name}/translated** - in which we retrieve the same information but a translation is attempted on the description
    * **/status** - trivial status endpoint which always returns HTTP 200 when the servers running
* A PokeAPI API client which allows us to call the Species resource (the only required resource for this challenge.)
* A Translation API client which uses the fun-translations endpoint to perform different types of translations
  these translations are defined by a set of enums in the package.
* A generic API client which performs the 90% of the generic things required when implementing API's.

##### Generic API Client

Inside of the internal/api package contains a generic API client. In my implementation this handles the following:

* Logging
* Request timeouts.
* Retries on certain failures and applying a backoff strategy
* HTTP request caching (if the same request is seen in a short space of time, a cached version is returned)
* Wrapping and contextualising errors from the API
* Generic encoding / decoding
* Authentication
* Very trivial localization utilising the `Accept-Language` header

For a lot of the different components ive tried to provide multiple examples to demonstrate the flexibility
of each of them:

###### Encoders

Probably the most used component in the code, this allows us to control how requests are decoded
and how responses are encoded. They also allow us to flexibly set applicable headers based
on the encoded response. The current implementations in this example are: JSON and XML.

###### Loggers

I have provided a very simple logging interface and have provided implementations using
the `fmt` package in the standard library as well as a more comprehensive example using
the `zap` library (see: [here](https://github.com/uber-go/zap)).

###### Authentication

I have provided examples on how to authenticate using:

* Basic Authentication
* Bearer Token
* A custom header (`X-Funtranslations-Api-Secret` in the case of the translation API)
* A custom query parameter variable

###### Backoff

I have provided two different retry backoff strategies which are:

* Constant - which applies the same backoff time between retries
* Exponential - which applies a exponential growth formula using the amount of retries to exponentially
  increase the timeout between retries.
  
Because I have made all of these features generic on a low-level client, any API client which utilises it
becomes very small and trivial. For example retrieving the Species from the PokeAPI is done
in 4 lines of code. This is why I took this approach, it means we can add more API providers and handle
more resources on existing API clients with very little effort, and any changes made to the generic client is
automatically reflected in any API client using it.

##### Server

I decided to use the [`mux`](https://github.com/gorilla/mux) for the HTTP router implementation because:

* it is very fast
* works with URL parameters
* works VERY well with the standard libraries `net/http` package
* allows us to add middleware to handle generic tasks

I have added Logging and RequestID middlewares which are linked to each handler
so we can a request ID generated for every request and so we can pass a logger to each of
the handler functions as well as perform access-level logging.


##### Running the server

I have provided two ways of running the server:

###### Running from source

If you want to run directly from source, the only prerequisite is having
go installed, preferably 1.16 as this product utilises go modules.

You can visit [here](https://golang.org/doc/install) on instructions on installing locally or there are implementations
usually installed on the package manager available on your system, however these tend to be out-of-date.

After installing go running the server can be achieved with the following commands:

```shell
cd ../path/to/cloned/repo
go run main.go serve --port=5555
```

You can get help on what options are available to serve using:
```shell
cd ../path/to/cloned/repo
go run main.go serve -h
```

###### Running Via Docker

I have also included a Dockerfile which allows you to run the server in a container

Go [here](https://docs.docker.com/get-docker/) to see how to install Docker locally.

After installing, you can run the server by:

```shell
docker build -t pokeapi:latest .
docker run -d --rm -e PORT=5555 -p 5555:5555 pokeapi:latest
```

##### Testing

Tests are automatically ran when the Docker container is built.

However to run them locally you can run (assuming you have [go](https://golang.org/doc/install) installed):

```shell
cd /path/of/repo
go test ./... -cover
```

##### Future Improvements

Some of the improvements I would like to make would firstly be improving the cache for the API client.
This would definitely in the similar fashion to that of the Logger or Backoff policy in which the cache layer
can be switched out as required with a redis / in memory or memcache implementation. I would also like to add a cache
implementation which allows having caching layers, i.e firstly using a in-memory cache
and then falling back to a redis cache for example.

I would also like to improve the API client to have a simple API to handle pagination, as most of the
time with REST API's this is similar, i.e typically either:
* page based with limit
* cursor based with limit
* have a link to the next resource in the current resource response.

In terms of the HTTP server, i would like to create a middleware which handles a lot of the
headers which handle security, i.e similar to [helmetjs](https://helmetjs.github.io/) for an express server.

I would keep the communication to be over HTTP (i.e no TLS on the running HTTP server itself)
and keep the container private and use AWS ECS or KUBERNETES with a load balancer with TLS enabled
which points to a private container instance (or a set of instances)