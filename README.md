# Hyper Text Adaptors

[![https://pkg.go.dev/github.com/dkotik/htadaptor](https://pkg.go.dev/badge/github.com/dkotik/htadaptor.svg)](https://pkg.go.dev/github.com/dkotik/htadaptor)

Package `htadaptor` provides convenient generic domain logic adaptors for HTTP handlers. It eliminates boiler plate code, increases security by enforcing read limits and struct validation, and reduces bugs by providing a more intuitive request data parsing API than the standard library.

<details>
  <summary>Planned features for <strong>v1.0.0</strong> release. ↩</summary>

- [ ] unpanic should be based on ErrorHandler
- [ ] Add file system adaptor with decoder that can stream files.
- [ ] Trim dependencies to zero.
</details>

## Why do you need this package?

An HTTP request contains at least five various sources of input that your HTTP handlers may consider: URL path, URL query, headers, cookies, and the request body. Much of the code that you have to write manually is wrestling those inputs into a struct. Willem Schots wrote an excellent [explanation here](https://www.willem.dev/articles/generic-http-handlers). `htadaptor` can do all of it for you:

```go
myHandler := htadaptor.Must(htadaptor.New().AdaptFunc(
  // your domain function call
  func(context.Context, *myInputStruct) (*myOutputStruct, error) {
    // ... myInputStruct is passed in already validated
    // ... the fields of myInputStruct will be populated with
    // ... the contents of `request.Body` with overrides
    //     from sources below in their given order:
  },
  htadaptor.WithPathValues("slug"),           // (1) URL routing path
  htadaptor.WithQueryValues("search"),        // (2) URL query
  htadaptor.WithHeaderValues("accessToken"),  // (3) header
  htadaptor.WithCookieValues("sessionID"),    // (4) cookie
  htadaptor.WithSessionValues("role"),        // (5) session
))
```

The adaptors address common function signatures of domain logic calls that operate on a request struct and return a response struct with **contextual awareness** all the way through the call stack:

| Struct Adaptor | Parameter Values     | Return Values |
|----------------|----------------------|--------------:|
| [AdaptFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#Adaptor.AdaptFunc)      | context, inputStruct |    any, error |
| [AdaptNullaryFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#Adaptor.AdaptNullaryFunc)    | context              |    any, error |
| [AdaptVoidFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#Adaptor.AdaptVoidFunc)       | context, inputStruct |         error |

String adaptors are best when only one request value is needed:

| String Adaptor  | Parameter Values     | Return Values |
|-----------------|----------------------|--------------:|
| [AdaptStringFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#Adaptor.AdaptStringFunc) | context, string      |    any, error |
| [AdaptVoidStringFunc](https://pkg.go.dev/github.com/dkotik/htadaptor#Adaptor.AdaptVoidStringFunc)  | context, string      |         error |

Many claim that the standard library interface `http.ResponseWriter` is entirely sufficient for writing HTTP applications. While true in principle, the general interface is a slope for many vulnerabilities. For example, developers frequently write directly to http.ResponseWriter, which sniffs the MIME content type from early response bytes, which can disable browser protections against cross-site scripting.

## Installation

```sh
go get github.com/dkotik/htadaptor@latest
```

## Basic Usage

```go
type order struct {
	ID string
	Items []string
}

func (o *order) Validate(ctx context.Context) error {
	if len(o.Items) == 0 {
		return errors.New("empty order")
	}
	return nil
}

type orderConfirmation struct {
	ID string
}

mux := http.NewServeMux()
mux.Handle("/api/v1/order", htadaptor.Must(
  htadaptor.New().AdaptFunc(
  	func(ctx context.Context, *order) (*orderConfirmation, error) {
   		// order is provided already validated
   		// ... process order here
      return &orderConfirmation{
      	ID: order.ID,
      }, nil
   	},
  ),
))
```

See [examples](https://github.com/dkotik/htadaptor/tree/main/examples) folder for common project uses.

## Adaptor Options

- [WithReadLimit](https://pkg.go.dev/github.com/dkotik/htadaptor#WithReadLimit)
- [WithMemoryLimit](https://pkg.go.dev/github.com/dkotik/htadaptor#WithMemoryLimit)
- [WithExtractors](https://pkg.go.dev/github.com/dkotik/htadaptor#WithExtractors)
- [WithEncoder](https://pkg.go.dev/github.com/dkotik/htadaptor#WithEncoder)
- [WithErrorHandler](https://pkg.go.dev/github.com/dkotik/htadaptor#WithErrorHandler)
- [WithUnsafeDecoder](https://pkg.go.dev/github.com/dkotik/htadaptor#WithUnsafeDecoder)

## Extractors

The order of extractors matters with the latter overriding the former. Request body is always processed first.

- [Path](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithPathValues)
- [Chi Path](https://pkg.go.dev/github.com/dkotik/htadaptor/extract/chivalues#New)
- [Query](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithQueryValues)
- [Header](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithHeaderValues)
- [Cookie](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithCookieValues)
- [Session](https://pkg.go.dev/github.com/dkotik/htadaptor/reflectd#WithSessionValues)
- Request properties can also be included into deserialization:
    - `extract.NewMethodExtractor`
    - `extract.NewHostExtractor`
    - `extract.NewRemoteAddressExtractor`
    - `extract.NewUserAgentExtractor`
- Or, make your own by implementing [Extractor](https://pkg.go.dev/github.com/dkotik/htadaptor/extract#Extractor) interface.

## Credits

The core idea was sparked in conversations with members of the Ardan Labs team. Package includes reflection schema decoder from Gorilla toolkit. Similar projects:


- [danielgtaylor/huma](https://github.com/danielgtaylor/huma) with REST and RPC
- [dolanor/rip](https://github.com/dolanor/rip/) with REST
- [go-fuego/fuego](https://github.com/go-fuego/fuego) with OpenAPI
- [matt1484/chimera](https://github.com/matt1484/chimera) for Chi
- [calvinmclean/babyapi](https://github.com/calvinmclean/babyapi)

How is `htadaptor` different from the other generic HTTP adaptors? It is more terse due to focusing on wrapping `http.Handlers` from the standard library. It is expected that the REST interface will be handled separately by either an `http.Mux` or a [helper](https://pkg.go.dev/github.com/dkotik/htadaptor#NewMethodMux).
