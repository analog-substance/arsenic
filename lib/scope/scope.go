package scope

import (
	"regexp"
	"strings"

	"net"

	"github.com/analog-substance/arsenic/lib/config"
	"github.com/analog-substance/arsenic/lib/set"
	"golang.org/x/net/publicsuffix"
)

type scope struct {
	domainsExplicitlyInScope     []string
	rootDomainsExplicitlyInScope []string
	blacklistedRootDomains       []string
	blacklistedDomainRegexps     []*regexp.Regexp
	domainInfoLoaded             bool

	hostIPsExplicitlyInScope       []string
	hostIPsExplicitlyInScopeLoaded bool
}

var asScope = &scope{}

func getScope() *scope {
	asScope.loadDomainInfo()
	asScope.loadExplicitScopeIPs()
	return asScope
}

func (s *scope) loadDomainInfo() {
	if s.domainInfoLoaded {
		return
	}

	blacklistConfig := config.Get().Blacklist

	rootDomainSet := set.NewStringSet()
	s.blacklistedRootDomains = blacklistConfig.RootDomains

	var err error
	s.domainsExplicitlyInScope, err = GetConstScope("domains")
	if err != nil {
		return
	}

	for _, domain := range s.domainsExplicitlyInScope {
		rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(domain)
		rootDomainSet.Add(rootDomain)
	}

	s.rootDomainsExplicitlyInScope = rootDomainSet.StringSlice()

	for _, domain := range blacklistConfig.Domains {
		s.blacklistedDomainRegexps = append(s.blacklistedDomainRegexps, regexp.MustCompile(regexp.QuoteMeta(domain)))
	}

	s.domainInfoLoaded = true
}

func (s *scope) loadExplicitScopeIPs() {
	if !s.hostIPsExplicitlyInScopeLoaded {
		var err error

		s.hostIPsExplicitlyInScope, err = GetConstScope("ips")
		if err != nil {
			return
		}

		s.hostIPsExplicitlyInScopeLoaded = true
	}
}

func (s *scope) IsBlacklistedRootDomain(rootDomain string) bool {
	for _, badRootDomain := range s.blacklistedRootDomains {
		if strings.EqualFold(badRootDomain, rootDomain) {
			return true
		}
	}

	return false
}

func (s *scope) IsBlacklistedDomain(checkDomain string) bool {
	for _, re := range s.blacklistedDomainRegexps {
		if re.MatchString(checkDomain) {
			return true
		}
	}

	return false
}

func (s *scope) IsDomainExplicitlyInScope(checkDomain string) bool {
	for _, domain := range s.domainsExplicitlyInScope {
		if strings.EqualFold(checkDomain, domain) {
			return true
		}
	}

	return false
}

func (s *scope) IsRootDomainInScope(checkRootDomain string) bool {
	for _, domain := range s.rootDomainsExplicitlyInScope {
		if strings.EqualFold(checkRootDomain, domain) {
			return true
		}
	}

	return false
}

func (s *scope) IsIPInScope(checkHostIP string) bool {
	parsedIP := net.ParseIP(checkHostIP)
	for _, hostIP := range s.hostIPsExplicitlyInScope {
		if strings.Contains(hostIP, "/") {
			_, inScopeIPNet, _ := net.ParseCIDR(hostIP)
			if inScopeIPNet.Contains(parsedIP) {
				return true
			}
		} else if checkHostIP == hostIP {
			return true
		}
	}

	return false
}

func (s *scope) IsDomainInScope(domain string, forceRootDomainBlacklistPrecedence bool) bool {
	rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(domain)
	if rootDomain == domain {
		rootDomain, _ = publicsuffix.PublicSuffix(domain)
	}

	if forceRootDomainBlacklistPrecedence {
		if s.IsBlacklistedRootDomain(rootDomain) {
			return false
		}
	}

	if s.IsDomainExplicitlyInScope(domain) {
		return true
	}

	if len(rootDomain) > 0 {
		if s.IsRootDomainInScope(rootDomain) {
			if !s.IsBlacklistedDomain(domain) {
				return true
			}
		}
	}

	return false
}
