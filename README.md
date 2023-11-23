# Hyper Text Adaptors

Package `htadaptor` provides generic domain logic adaptors for HTTP handlers. Available adaptors cover almost every possible combination of domain call shapes:

| Struct Adaptor | Parameter Values     | Return Values |
|----------------|----------------------|--------------:|
| UnaryFunc      | context, inputStruct |    any, error |
| NullaryFunc    | context              |    any, error |
| VoidFunc       | context, inputStruct |         error |

Each inputStruct must implement `htadaptor.Validatable` for safety.

| String Adaptor  | Parameter Values     | Return Values |
|-----------------|----------------------|--------------:|
| UnaryStringFunc | context, string      |    any, error |
| VoidStringFunc  | context, string      |         error |

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

<!-- TODO: link to GoDoc -->

- WithDecoder
    - WithReadLimit
    - WithMemoryLimit
    - WithQueryValue
    - WithHeaderValue
    - WithExtractor
- WithEncoder
- WithLogger
    - WithSlogLogger
- WithErrorHandler
