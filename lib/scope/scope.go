package scope

import (
	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"github.com/spf13/viper"
	"golang.org/x/net/publicsuffix"
	"regexp"
	"strings"
)

type scope struct {
	domainsExplicitlyInScope     []string
	rootDomainsExplicitlyInScope []string
	blacklistedRootDomains       []string
	blacklistedDomains           []string
	blacklistedDomainRegexps     []*regexp.Regexp
	domainInfoLoaded             bool

	hostIPsExplicitlyInScope       []string
	hostIPsExplicitlyInScopeLoaded bool
}

var asScope = &scope{}

func getScope() *scope {
	return asScope
}

func (s *scope) loadDomainInfo() {
	if !s.domainInfoLoaded {
		domainSet := set.NewSet("")
		rootDomainSet := set.NewSet("")
		s.blacklistedRootDomains = viper.GetStringSlice("blacklist.root-domains")
		s.blacklistedDomains = viper.GetStringSlice("blacklist.domains")

		util.ReadLineByLine("scope-domains.txt", func(line string) {
			line = normalizeScope(line, "domain")
			rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(line)

			domainSet.Add(line)
			rootDomainSet.Add(rootDomain)
		})

		s.domainsExplicitlyInScope = domainSet.StringSlice()
		s.rootDomainsExplicitlyInScope = rootDomainSet.StringSlice()
		s.domainInfoLoaded = true
	}

	if len(s.blacklistedDomains) != len(s.blacklistedDomainRegexps) {
		for _, domain := range s.blacklistedDomains {
			s.blacklistedDomainRegexps = append(s.blacklistedDomainRegexps, regexp.MustCompile(regexp.QuoteMeta(domain)))
		}
	}
}

func (s *scope) loadExplicitScopeIPs() {
	if !s.hostIPsExplicitlyInScopeLoaded {
		hostIPMap := map[string]bool{}

		util.ReadLineByLine("scope-ips.txt", func(line string) {
			line = normalizeScope(line, "ip")
			hostIPMap[line] = true
		})

		for hostIP := range hostIPMap {
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
	s.loadDomainInfo()
	for _, re := range s.blacklistedDomainRegexps {
		if re.MatchString(checkDomain) {
			return true
		}
	}

	return false
}

func (s *scope) IsDomainExplicitlyInScope(checkDomain string) bool {
	s.loadDomainInfo()
	for _, domain := range s.domainsExplicitlyInScope {
		if strings.EqualFold(checkDomain, domain) {
			return true
		}
	}

	return false
}

func (s *scope) IsRootDomainInScope(checkRootDomain string) bool {
	s.loadDomainInfo()
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
