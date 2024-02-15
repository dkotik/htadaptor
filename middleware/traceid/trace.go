/*
Package traceid provides [htadaptor.Middleware] that injects
trace identifiers into request [context.Context] and a matching
[slog.Handler] for populating the log records with the identifier.
*/
package traceid

// https://lukas.zapletalovi.com/posts/2023/about-structured-logging-in-go121/?utm_source=pocket_saves

type contextKey struct{}
