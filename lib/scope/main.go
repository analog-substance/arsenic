package scope

import (
	"fmt"
	"github.com/analog-substance/arsenic/lib/util"
	"golang.org/x/net/publicsuffix"
	"net"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

func GetRootDomains(domains []string, pruneBlacklisted bool) []string {
	rootDomainMap := map[string]int{}
	var rootDomains []string
	for _, domain := range domains {
		rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(domain)

		if len(rootDomain) > 0 {
			rootDomainMap[rootDomain] = 1
		}
	}

	for rootDomain := range rootDomainMap {
		addRootDomain := true
		if pruneBlacklisted {
			if getScope().IsBlacklistedRootDomain(rootDomain) {
				addRootDomain = false
				break
			}
		}

		if addRootDomain {
			rootDomains = append(rootDomains, rootDomain)
		}
	}
	sort.Strings(rootDomains)
	return rootDomains
}

func IsIp(ipOrHostname string) bool {
	if net.ParseIP(ipOrHostname) == nil {
		return false
	} else {
		return true
	}
}

func IsInScope(ipOrHostname string, forceRootDomainBlacklistPrecedence bool) bool {
	if IsIp(ipOrHostname) {
		return getScope().IsIPInScope(ipOrHostname)
	}
	return getScope().IsDomainInScope(ipOrHostname, forceRootDomainBlacklistPrecedence)
}

func GetScope(scopeType string) ([]string, error) {

	glob := fmt.Sprintf("scope-%s-*", scopeType)
	actualFile := fmt.Sprintf("scope-%s.txt", scopeType)

	files, _ := filepath.Glob(glob)
	scope := make(map[string]bool)

	for _, filename := range files {
		err := util.ReadLineByLine(filename, func(line string) {
			line = normalizeScope(line, scopeType)
			if getScope().IsDomainInScope(line, false) {
				scope[line] = true
			}
		})
		if err != nil {
			return nil, err
		}
	}

	// now lets open the actual scope file and add those. since they can't be blacklisted
	err := util.ReadLineByLine(actualFile, func(line string) {
		line = normalizeScope(line, scopeType)
		scope[line] = true
	})

	if err != nil {
		return nil, err
	}

	var scopeAr []string
	for scopeItem := range scope {
		scopeAr = append(scopeAr, scopeItem)
	}

	sort.Strings(scopeAr)
	return scopeAr, nil
}

func normalizeScope(scopeItem, scopeType string) string {

	if scopeType == "domains" {
		re := regexp.MustCompile(`^\*\.`)
		scopeItem = re.ReplaceAllString(scopeItem, "")
		scopeItem = strings.ToLower(scopeItem)
	}

	return scopeItem
}
