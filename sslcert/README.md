# `sslcert`: Util functions for the builtin `autocert` package.

## sqlcertcache

SQL cache for [acme/autocert](https://godoc.org/golang.org/x/crypto/acme/autocert) written in Go.
It's a fork / copy of [https://github.com/goenning/sqlcertcache](goenning/sqlcertcache).

### Example

```go
conn, err := sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
if err != nil {
  // Handle error
}

cache, err := sqlcertcache.New(conn, "autocert_cache")
if err != nil {
  // Handle error
}

m := autocert.Manager{
    Cache:      cache,
    // ... omit other options here ...
}
```

### Performance

`autocert` has an internal in-memory cache that is used before quering this
long-term cache, so you don't need to worry about your SQL database being hit
many times just to get a certificate. It should only do once per process+key.

## HTTP to HTTPS redirect

`Manager.HTTPHandler(fallback)` returns an `http.Handler` that answers ACME
"http-01" challenges and passes the other requests to `fallback` (which
redirects to HTTPS when `fallback` is nil). It must be served on port 80 by the
caller:

```go
m, err := sslcert.New(
    sslcert.WithCertDomain("example.com"),
    sslcert.WithSQLiteCache(conn),
)
if err != nil {
    // Handle error
}

// Only redirect requests for the configured domain, so a spoofed Host header
// can not turn this into an open redirect.
fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    host := strings.TrimSuffix(strings.ToLower(r.Host), ".")
    if h, _, err := net.SplitHostPort(host); err == nil {
        host = h
    }

    if host != "example.com" {
        http.NotFound(w, r)
        return
    }

    if r.Method != http.MethodGet && r.Method != http.MethodHead {
        http.Error(w, "Use HTTPS", http.StatusBadRequest)
        return
    }

    http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusMovedPermanently)
})

srv := &http.Server{
    Addr:              ":80",
    Handler:           m.HTTPHandler(fallback),
    ReadHeaderTimeout: 5 * time.Second,
}
log.Fatal(srv.ListenAndServe())
```

Requests under `/.well-known/acme-challenge/` are always handled by `autocert`
(and usually return 404 for a fixed-certificate manager); all other requests go
to `fallback`.

When `fallback` is nil, the redirect logic built into `autocert` is used:
`GET`/`HEAD` requests get a `302` redirect to
`https://<Host><path>?<query>` (an explicit port is rewritten to `:443`), other
methods get `400`. That built-in logic does not apply `HostPolicy` to requests,
and a custom fallback is not protected by `HostPolicy` either; validate the
`Host` header yourself as shown above. Do not serve a nil fallback on the public
internet.

Calling `HTTPHandler` also enables the `http-01` challenge type for `autocert`;
`tls-alpn-01` remains available. If it is never called, domains are verified
with `tls-alpn-01` only and no HTTP listener on port 80 is required, provided
the TLS config advertises `acme-tls/1` (for example via `Manager.Listener()`,
or by adding `acme.ALPNProto` to `NextProtos`).
