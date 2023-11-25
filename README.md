# Hyper Text Adaptors

[![https://pkg.go.dev/github.com/dkotik/htadaptor](https://pkg.go.dev/badge/github.com/dkotik/htadaptor.svg)](https://pkg.go.dev/github.com/dkotik/htadaptor)
[![https://github.com/dkotik/htadaptor/actions?query=workflow:test](https://github.com/dkotik/htadaptor/workflows/test/badge.svg?branch=main&event=push)](https://github.com/dkotik/htadaptor/actions?query=workflow:test)
[![https://coveralls.io/github/dkotik/htadaptor](https://coveralls.io/repos/github/dkotik/htadaptor/badge.svg?branch=main)](https://coveralls.io/github/dkotik/htadaptor)
[![https://goreportcard.com/report/github.com/dkotik/htadaptor](https://goreportcard.com/badge/github.com/dkotik/htadaptor)](https://goreportcard.com/report/github.com/dkotik/htadaptor)

Package `htadaptor` provides generic domain logic adaptors for HTTP handlers. Available adaptors cover almost every possible combination of domain call shapes:

<!-- TODO: link adaptors to GoDoc -->
<!-- TODO: add FS adaptor -->

| Struct Adaptor | Parameter Values     | Return Values |
|----------------|----------------------|--------------:|
| [UnaryFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#UnaryFuncAdaptor)      | context, inputStruct |    any, error |
| [NullaryFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#NullaryFuncAdaptor)    | context              |    any, error |
| [VoidFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#VoidFuncAdaptor)       | context, inputStruct |         error |

Each inputStruct must implement `htadaptor.Validatable` for safety.

| String Adaptor  | Parameter Values     | Return Values |
|-----------------|----------------------|--------------:|
| [UnaryStringFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#UnaryStringFuncAdaptor) | context, string      |    any, error |
| [VoidStringFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#VoidStringFuncAdaptor)  | context, string      |         error |

## Installation

```sh
go get github.com/dkotik/htadaptor@latest
```

## Usage

```go
mux := http.NewServeMux()
mux.Handle("/api/v1/order", htadaptor.Must(
  htadaptor.NewUnaryFuncAdaptor(myService.Order),
))
```

See `examples` folder for most common project uses.

## Adaptor Options

- [WithDecoder](https://pkg.go.dev/github.com/dkotik/htadaptor#WithDecoder)
    - [WithReadLimit](https://pkg.go.dev/github.com/dkotik/htadaptor#WithReadLimit)
    - [WithMemoryLimit](https://pkg.go.dev/github.com/dkotik/htadaptor#WithMemoryLimit)
    - [WithQueryValue](https://pkg.go.dev/github.com/dkotik/htadaptor#WithQueryValue)
    - [WithHeaderValue](https://pkg.go.dev/github.com/dkotik/htadaptor#WithHeaderValue)
    - [WithExtractors](https://pkg.go.dev/github.com/dkotik/htadaptor#WithExtractors)
- [WithEncoder](https://pkg.go.dev/github.com/dkotik/htadaptor#WithEncoder)
- [WithLogger](https://pkg.go.dev/github.com/dkotik/htadaptor#WithLogger)
    - [WithSlogLogger](https://pkg.go.dev/github.com/dkotik/htadaptor#WithSlogLogger)
- [WithErrorHandler](https://pkg.go.dev/github.com/dkotik/htadaptor#WithErrorHandler)

## Extractors

- [Query](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithQueryValues)
- [Header](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithHeaderValues)
- Path (pending with 1.22)
- [Chi](https://pkg.go.dev/github.com/dkotik/htadaptor/chivalues#New)

## Credits

Package includes reflection schema decoder from Gorilla toolkit.
