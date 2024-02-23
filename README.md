# Hyper Text Adaptors

[![https://pkg.go.dev/github.com/dkotik/htadaptor](https://pkg.go.dev/badge/github.com/dkotik/htadaptor.svg)](https://pkg.go.dev/github.com/dkotik/htadaptor)
[![https://github.com/dkotik/htadaptor/actions?query=workflow:test](https://github.com/dkotik/htadaptor/workflows/test/badge.svg?branch=main&event=push)](https://github.com/dkotik/htadaptor/actions?query=workflow:test)
[![https://coveralls.io/github/dkotik/htadaptor](https://coveralls.io/repos/github/dkotik/htadaptor/badge.svg?branch=main)](https://coveralls.io/github/dkotik/htadaptor)
[![https://goreportcard.com/report/github.com/dkotik/htadaptor](https://goreportcard.com/badge/github.com/dkotik/htadaptor)](https://goreportcard.com/report/github.com/dkotik/htadaptor)

Package `htadaptor` provides convenient generic domain logic adaptors for HTTP handlers. It eliminates boiler plate code, increases security by enforcing read limits and struct validation, and reduces bugs by providing a more intuitive request data parsing API than the standard library.

Why do you need this package? An HTTP request contains at least five various sources of input that your HTTP handlers may consider: URL path, URL query, headers, cookies, and the request body. Much of the code that you have to write [manually](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/) is wrestling those inputs into a struct. `htadaptor` can do all of it for you:

```go
myHandler := htadaptor.Must(htadaptor.NewUnaryFuncAdaptor(
  // your domain function call
  func(ctx context.Context, myInputStruct) (myOutputStruct, error) {
    // ... myInputStruct is passed in already validated
    // ... the fields of myInputStruct will be populated with
    // ... the contents of `request.Body` with overrides
    //     from sources the below in their given order:
  },
  htadaptor.WithPathValues("slug"),           // (1) URL routing path
  htadaptor.WithQueryValues("search"),        // (2) URL query
  htadaptor.WithHeaderValues("accessToken"),  // (3) header
  htadaptor.WithCookieValues("sessionID"),    // (4) cookie
  htadaptor.WithSessionValues("role"),        // (5) session
))
```

Adaptors address all common function signatures of domain logic calls that operate on a request struct and return a response struct:

<!-- TODO: add FS adaptor -->

| Struct Adaptor | Parameter Values     | Return Values |
|----------------|----------------------|--------------:|
| [UnaryFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#UnaryFuncAdaptor)      | context, inputStruct |    any, error |
| [NullaryFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#NullaryFuncAdaptor)    | context              |    any, error |
| [VoidFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#VoidFuncAdaptor)       | context, inputStruct |         error |

Each inputStruct must implement `htadaptor.Validatable` for safety. String adaptors are best when only one request value is needed:

| String Adaptor  | Parameter Values     | Return Values |
|-----------------|----------------------|--------------:|
| [UnaryStringFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#UnaryStringFuncAdaptor) | context, string      |    any, error |
| [VoidStringFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#VoidStringFuncAdaptor)  | context, string      |         error |

## Installation

```sh
go get github.com/dkotik/htadaptor@latest
```

## Basic Usage

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
    - [WithExtractors](https://pkg.go.dev/github.com/dkotik/htadaptor#WithExtractors)
- [WithEncoder](https://pkg.go.dev/github.com/dkotik/htadaptor#WithEncoder)
- [WithLogger](https://pkg.go.dev/github.com/dkotik/htadaptor#WithLogger)
- [WithErrorHandler](https://pkg.go.dev/github.com/dkotik/htadaptor#WithErrorHandler)

## Extractors

The order of extractors matters with the latter overriding the former. Request body is always processed first.

- [Path](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithPathValues)
- [Chi Path](https://pkg.go.dev/github.com/dkotik/htadaptor/chivalues#New)
- [Query](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithQueryValues)
- [Header](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithHeaderValues)
- [Cookie](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithCookieValues)
- [Session](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithSessionValues)
- Request properties:
    - `extract.NewHostExtractor`
    - `extract.NewRemoteAddressExtractor`
    - `extract.NewUserAgentExtractor`

## Credits

The core idea was sparked in conversations with members of the Ardan Labs team. Package includes reflection schema decoder from Gorilla toolkit. Similar projects:

- Generic REST Controllers: <https://github.com/dolanor/rip/>
- Baby API: <https://github.com/calvinmclean/babyapi>
