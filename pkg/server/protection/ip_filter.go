package protection

import (
	"net"
	"sync"
)

type IPFilter struct {
	whitelist map[string]bool
	blacklist map[string]bool
	mu        sync.RWMutex
}

func NewIPFilter(whitelist, blacklist []string) *IPFilter {
	f := &IPFilter{
		whitelist: make(map[string]bool),
		blacklist: make(map[string]bool),
	}

	for _, ip := range whitelist {
		f.whitelist[ip] = true
	}

	for _, ip := range blacklist {
		f.blacklist[ip] = true
	}

	return f
}

func (f *IPFilter) IsAllowed(addr net.Addr) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	ip := addr.String()

	// If whitelist is not empty, only allow whitelisted IPs
	if len(f.whitelist) > 0 {
		return f.whitelist[ip]
	}

	// Otherwise, block blacklisted IPs
	return !f.blacklist[ip]
}
