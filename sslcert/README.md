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