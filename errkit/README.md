# errkit

why errkit when we already have errorutil ? 

----

errorutil was introduced a year ago with main goal to capture stack of error to identify underlying deeply nested errors. but it does not follow the go paradigm of error handling and implementation and is counter intuitive to that. i.e in golang looking at any error util/implementation library "errors" "pkg/errors" or "uber.go/multierr" etc, they all follow the same pattern i.e `.Error()` method is never used and instead it is wrapped with helper structs following a particular interface which allows traversing the error chain and using helper functions like `.Cause() error` or `.Unwrap() error` or `errors.Is()` and more. but errorutil marshalls the error to string which does not play well with the go error handling paradigm. Apart from that over time usage of errorutil has been cumbersome because it is not drop in replacement for any error package and it does not allow propogating/traversing error chain in a go idiomatic way.


`errkit` is new error library that is built upon learnings from `errorutil` and has following features:

- drop in replacement for (no syntax change / refactor required)
    - `errors` package
    - `pkg/errors` package (now deprecated)
    - `uber/multierr` package
- is compatible with all known go error handling implementations and can parse errors from any library and is compatible with existing error handling libraries and helper functions like `Is()` , `As()` , `Cause()` and more.
- is go idiomatic and follows the go error handling paradigm
- Has Attributes support (see below)
- Implements and categorizes errors into different classes (see below)
    - `ErrClassNetworkTemporary`
    - `ErrClassNetworkPermanent`
    - `ErrClassDeadline`
    - Custom Classes via `ErrClass` interface
- Supports easy conversion to slog Item for structured logging reatining all error info
- Helper functions to implement public/user facing errors by using error classes


**Attributes Support**

To strictly follow the go error handling paradigm and making it easy to traverse error chain, errkit support adding `Attr(key comparable,value any)` to error instead of wrapping it with a string message. This keeps extra error info minimal without being too verbose a good example of this is following

```go
// normal way of error propogating through nested stack
err := errkit.New("i/o timeout")

// xyz.go
err := errkit.Wrap(err,"failed to connect %s",addr)

// abc.go
err := errkit.Wrap(err,"error occured when downloading %s",xyz)
```

with attributes support you can do following

```go
// normal way of error propogating through nested stack
err := errkit.New("i/o timeout")

// xyz.go
err = errkit.WithAttr(err,&errkit.Resource{},errkit.Resource{type: "network",add: addr})

// abc.go
err = errkit.WithAttr(err,&errkit.Action{},errkit.Action{type: "download"})
```

the good part is that all attributes must implement a interface 'ErrAttr' that way all of them are compatible and can be consolidated into a struct/object of choice. for more example see `attr_test.go`

In case the same attribute types are added they are deduplicated / consolidated into parent attribute if it supports that.