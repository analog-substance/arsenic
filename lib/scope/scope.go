package scope

import (
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
	"strings"
)

type scope struct {
	domainsExplicitlyInScope     []string
	rootDomainsExplicitlyInScope []string
	blacklistedRootDomains       []string
	blacklistedDomains           []string
	explicitDomainsLoaded        bool

	hostIPsExplicitlyInScope       []string
	hostIPsExplicitlyInScopeLoaded bool
}

var asScope = &scope{}
var asScopeInit = false

func getScope() *scope {
	return asScope
}

func (s *scope) loadExplicitDomains() {
	if !s.explicitDomainsLoaded {
		domainsMap := map[string]bool{}
		rootDomainsMap := map[string]bool{}
		s.blacklistedRootDomains = viper.GetStringSlice("blacklist.root-domains")
		s.blacklistedDomains = viper.GetStringSlice("blacklist.domains")

		util.ReadLineByLine("scope-domains.txt", func(line string) {
			line = normalizeScope(line, "domain")
			rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(line)

			domainsMap[line] = true
			rootDomainsMap[rootDomain] = true
		})

		for domain, _ := range domainsMap {
			s.domainsExplicitlyInScope = append(s.domainsExplicitlyInScope, domain)
		}

		for rootDomain, _ := range rootDomainsMap {
			if !s.IsBlacklistedRootDomain(rootDomain) {
				s.rootDomainsExplicitlyInScope = append(s.rootDomainsExplicitlyInScope, rootDomain)
			}
		}

		s.explicitDomainsLoaded = true
	}
}

func (s *scope) loadExplicitScopeIPs() {
	if !s.hostIPsExplicitlyInScopeLoaded {
		hostIPMap := map[string]bool{}

		util.ReadLineByLine("scope-ips.txt", func(line string) {
			line = normalizeScope(line, "ip")
			hostIPMap[line] = true
		})

		for hostIP, _ := range hostIPMap {
			s.hostIPsExplicitlyInScope = append(s.hostIPsExplicitlyInScope, hostIP)
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
	s.loadExplicitDomains()
	for _, badDomain := range s.blacklistedDomains {
		if strings.EqualFold(badDomain, checkDomain) {
			return true
		}
	}

	return false
}

func (s *scope) IsDomainExplicitlyInScope(checkDomain string) bool {
	s.loadExplicitDomains()
	for _, domain := range s.domainsExplicitlyInScope {
		if strings.EqualFold(checkDomain, domain) {
			return true
		}
	}

	return false
}

func (s *scope) IsRootDomainInScope(checkRootDomain string) bool {
	s.loadExplicitDomains()
	for _, domain := range s.rootDomainsExplicitlyInScope {
		if strings.EqualFold(checkRootDomain, domain) {
			return true
		}
	}

	return false
}

func (s *scope) IsIPInScope(checkHostIP string) bool {
	s.loadExplicitScopeIPs()
	for _, hostIP := range s.hostIPsExplicitlyInScope {
		if checkHostIP == hostIP {
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
