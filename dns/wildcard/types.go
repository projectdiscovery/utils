package wildcard

// LookupFunc resolves a host and returns the address answers used for wildcard matching.
type LookupFunc func(host string) ([]string, error)
