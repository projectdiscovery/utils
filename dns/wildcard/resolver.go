package wildcard

import (
	"errors"
	"strings"
	"sync"

	mapsutil "github.com/projectdiscovery/utils/maps"
	sliceutil "github.com/projectdiscovery/utils/slice"
	stringsutil "github.com/projectdiscovery/utils/strings"
	"github.com/rs/xid"
)

const DefaultWildcardProbeCount = 3
const reProbeCount = 2

var errDomainFound = errors.New("domain found")

type probeState uint8

const (
	probeStateError probeState = iota
	probeStateNoAnswers
	probeStateResolved
)

// Resolver represents a wildcard resolver extracted from shuffledns.
type Resolver struct {
	Domains *sliceutil.SyncSlice[string]
	lookup  LookupFunc

	levelAnswersNormalCache *mapsutil.SyncLockMap[string, struct{}]
	wildcardAnswersCache    *mapsutil.SyncLockMap[string, wildcardAnswerCacheValue]

	probeCount int
}

type wildcardAnswerCacheValue struct {
	IPS *mapsutil.SyncLockMap[string, struct{}]
}

func mapValues(m *mapsutil.SyncLockMap[string, struct{}]) map[string]struct{} {
	values := make(map[string]struct{})
	if m == nil {
		return values
	}

	_ = m.Iterate(func(key string, value struct{}) error {
		values[key] = value
		return nil
	})

	return values
}

// NewResolver initializes and creates a new resolver to find wildcards.
func NewResolver(domains []string, lookup LookupFunc) *Resolver {
	fqdns := sliceutil.NewSyncSlice[string]()
	fqdns.Append(domains...)
	return NewResolverWithDomains(fqdns, lookup)
}

// NewResolverWithDomains initializes a resolver with a pre-built domain slice.
func NewResolverWithDomains(domains *sliceutil.SyncSlice[string], lookup LookupFunc) *Resolver {
	if domains == nil {
		domains = sliceutil.NewSyncSlice[string]()
	}

	return &Resolver{
		Domains:                 domains,
		lookup:                  lookup,
		levelAnswersNormalCache: mapsutil.NewSyncLockMap[string, struct{}](),
		wildcardAnswersCache:    mapsutil.NewSyncLockMap[string, wildcardAnswerCacheValue](),
		probeCount:              DefaultWildcardProbeCount,
	}
}

// SetProbeCount sets the number of probes to use for wildcard detection.
// Higher values improve detection of wildcards using DNS round-robin.
func (w *Resolver) SetProbeCount(count int) {
	if count > 0 {
		w.probeCount = count
	}
}

// probeWildcardIPs probes the given wildcard pattern multiple times concurrently and returns all IPs found.
// If the first probe returns no answers, the level is treated as a normal level.
// Transport or resolver errors are returned separately so callers do not cache them as normal answers.
// First query is executed sequentially for early exit, remaining queries run in parallel.
func (w *Resolver) probeWildcardIPs(pattern string, count int) ([]string, probeState) {
	if count <= 0 {
		return nil, probeStateNoAnswers
	}

	ips := sliceutil.NewSyncSlice[string]()

	probe := func() ([]string, probeState) {
		probeHost := strings.ReplaceAll(pattern, "*.", xid.New().String()+".")
		answers, err := w.lookup(probeHost)
		if err != nil {
			return nil, probeStateError
		}
		if len(answers) == 0 {
			return nil, probeStateNoAnswers
		}
		return answers, probeStateResolved
	}

	resultIPs, state := probe()
	if state != probeStateResolved {
		return nil, state
	}
	if len(resultIPs) > 0 {
		ips.Append(resultIPs...)
	}

	if count == 1 {
		if ips.Len() == 0 {
			return nil, probeStateNoAnswers
		}
		return sliceutil.Dedupe(ips.Slice), probeStateResolved
	}

	var wg sync.WaitGroup
	for i := 1; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resultIPs, state := probe()
			if state == probeStateResolved && len(resultIPs) > 0 {
				ips.Append(resultIPs...)
			}
		}()
	}

	wg.Wait()

	if ips.Len() == 0 {
		return nil, probeStateNoAnswers
	}

	return sliceutil.Dedupe(ips.Slice), probeStateResolved
}

// generateWildcardPermutations generates wildcard permutations for a given subdomain
// and domain. It generates permutations for each level of the subdomain
// in reverse order.
func generateWildcardPermutations(subdomain, domain string) []string {
	var hosts []string
	subdomainTokens := strings.Split(subdomain, ".")

	var builder strings.Builder
	builder.Grow(len(subdomain) + len(domain) + 5)

	// Iterate from the reverse order. This way we generate the roots
	// first and allows us to do filtering faster, by trying out the root
	// like *.example.com first, and *.child.example.com in that order.
	// If we get matches for the root, we can skip the child and rest.
	builder.WriteString("*.")
	builder.WriteString(domain)
	hosts = append(hosts, builder.String())
	builder.Reset()

	for i := len(subdomainTokens); i > 1; i-- {
		_, _ = builder.WriteString("*.")
		_, _ = builder.WriteString(strings.Join(subdomainTokens[i-1:], "."))
		_, _ = builder.WriteRune('.')
		_, _ = builder.WriteString(domain)
		hosts = append(hosts, builder.String())
		builder.Reset()
	}
	return hosts
}

// LookupHost returns wildcard IP addresses of a wildcard if it's a wildcard.
// To determine this, we split the target host by dots, generate wildcard
// permutations for each level of the matched domain, and probe those levels.
// If any of the host IPs overlap with wildcard answers collected for those
// levels, the host is treated as wildcard-backed.
func (w *Resolver) LookupHost(host string, knownIPs []string) (bool, map[string]struct{}) {
	wildcards := make(map[string]struct{})

	var domain string
	w.Domains.Each(func(i int, domainCandidate string) error {
		if stringsutil.HasSuffixAny(host, "."+domainCandidate) {
			domain = domainCandidate
			return errDomainFound
		}
		return nil
	})

	// Ignore records without a matching domain. This may be interesting for
	// dangling-domain detection later, but wildcard matching intentionally skips it.
	if domain == "" {
		return false, nil
	}

	subdomainPart := strings.TrimSuffix(host, "."+domain)

	// create the wildcard generation prefix.
	// We use a rand prefix at the beginning like %rand%.domain.tld
	// A permutation is generated for each level of the subdomain.
	hosts := generateWildcardPermutations(subdomainPart, domain)

	// Iterate over all the hosts generated for rand.
	for _, h := range hosts {
		h = strings.TrimSuffix(h, ".")

		original := h

		// Check if we have already resolved this host level successfully
		// and if so, use the cached answer
		//
		// ex. *.campaigns.google.com is a wildcard so we cache it
		// and it is used always for resolutions in future.
		cachedValue, cachedValueOk := w.wildcardAnswersCache.Get(original)
		if cachedValueOk {
			for _, knownIP := range knownIPs {
				if _, ipExists := cachedValue.IPS.Get(knownIP); ipExists {
					return true, mapValues(cachedValue.IPS)
				}
			}
			if extraIPs, state := w.probeWildcardIPs(original, reProbeCount); state == probeStateResolved && len(extraIPs) > 0 {
				for _, record := range extraIPs {
					wildcards[record] = struct{}{}
					_ = cachedValue.IPS.Set(record, struct{}{})
				}
				_ = w.wildcardAnswersCache.Set(original, cachedValue)
				for _, knownIP := range knownIPs {
					if _, ipExists := cachedValue.IPS.Get(knownIP); ipExists {
						return true, mapValues(cachedValue.IPS)
					}
				}
			}
		}

		// Check if this level already produced a normal response with no wildcard answers.
		// Example: *.google.com is not a wildcard and returns NXDOMAIN,
		// so future checks at that level can be skipped.
		if _, ok := w.levelAnswersNormalCache.Get(original); ok {
			continue
		}

		probeIPs, state := w.probeWildcardIPs(original, w.probeCount)
		if state == probeStateNoAnswers {
			_ = w.levelAnswersNormalCache.Set(original, struct{}{})
			continue
		}
		if state != probeStateResolved {
			continue
		}

		if len(probeIPs) > 0 {
			if !cachedValueOk {
				cachedValue.IPS = mapsutil.NewSyncLockMap[string, struct{}]()
			}
			for _, record := range probeIPs {
				wildcards[record] = struct{}{}
				_ = cachedValue.IPS.Set(record, struct{}{})
			}
			_ = w.wildcardAnswersCache.Set(original, cachedValue)
			for _, knownIP := range knownIPs {
				if _, ipExists := cachedValue.IPS.Get(knownIP); ipExists {
					return true, mapValues(cachedValue.IPS)
				}
			}

			for i := 0; i < w.probeCount; i++ {
				answers, err := w.lookup(host)
				if err == nil {
					for _, record := range answers {
						if _, ipExists := cachedValue.IPS.Get(record); ipExists {
							return true, mapValues(cachedValue.IPS)
						}
					}
				}
			}
		}
	}

	for _, knownIP := range knownIPs {
		if _, ok := wildcards[knownIP]; ok {
			return true, wildcards
		}
	}

	return false, wildcards
}

func (w *Resolver) GetAllWildcardIPs() map[string]struct{} {
	ips := make(map[string]struct{})

	_ = w.wildcardAnswersCache.Iterate(func(key string, value wildcardAnswerCacheValue) error {
		for ip := range mapValues(value.IPS) {
			if _, ok := ips[ip]; !ok {
				ips[ip] = struct{}{}
			}
		}
		return nil
	})
	return ips
}
