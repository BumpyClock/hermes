// Package offline is the candidate-runner-only DNS boundary.
// Production Hermes does not install or use it.
package offline

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Installed is set only by the explicit candidate source overlay.
var Installed bool

var state struct {
	sync.Mutex
	hosts map[string]bool
	calls int
}

// Configure restricts the in-memory resolver to this run's validated fixture hosts.
func Configure(hosts []string) {
	state.Lock()
	defer state.Unlock()
	state.hosts = map[string]bool{}
	state.calls = 0
	for _, host := range hosts {
		state.hosts[host] = true
	}
}

// Calls counts in-memory lookups; none opens a socket or consults system DNS.
func Calls() int {
	state.Lock()
	defer state.Unlock()
	return state.calls
}

// Resolver supplies fixed public addresses while retaining production IP validation.
type Resolver struct{}

// LookupIPAddr rejects unplanned names and preserves literal private IPs for denial.
func (Resolver) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	state.Lock()
	defer state.Unlock()
	state.calls++
	if ip := net.ParseIP(host); ip != nil {
		return []net.IPAddr{{IP: ip}}, nil
	}
	if !state.hosts[host] {
		return nil, fmt.Errorf("offline conformance: unplanned hostname %q", host)
	}
	return []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, nil
}
