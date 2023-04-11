package scope

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/analog-substance/arsenic/lib/set"
	"github.com/analog-substance/arsenic/lib/util"
	"golang.org/x/net/publicsuffix"
)

func GetRootDomains(domains []string, pruneBlacklisted bool) []string {
	rootDomainset := set.NewStringSet()
	var rootDomains []string
	for _, domain := range domains {
		rootDomain, _ := publicsuffix.EffectiveTLDPlusOne(domain)

		if len(rootDomain) > 0 {
			rootDomainset.Add(rootDomain)
		}
	}

	sorted := rootDomainset.SortedStringSlice()
	for _, rootDomain := range sorted {
		if pruneBlacklisted && getScope().IsBlacklistedRootDomain(rootDomain) {
			continue
		}

		rootDomains = append(rootDomains, rootDomain)
	}
	sort.Strings(rootDomains)
	return rootDomains
}

func IsInScope(ipOrHostname string, forceRootDomainBlacklistPrecedence bool) bool {
	if util.IsIp(ipOrHostname) {
		return getScope().IsIPInScope(ipOrHostname)
	}
	return getScope().IsDomainInScope(ipOrHostname, forceRootDomainBlacklistPrecedence)
}

func GetScope(scopeType string) ([]string, error) {
	glob := fmt.Sprintf("scope-%s-*", scopeType)

	files, _ := filepath.Glob(glob)
	scope := set.NewStringSet()

	for _, filename := range files {
		c, err := util.ReadFileLineByLine(filename)
		if err != nil {
			return nil, err
		}

		for line := range c {
			line = normalizeScope(line, scopeType)
			if getScope().IsDomainInScope(line, false) {
				scope.Add(line)
			}
		}
	}

	// now lets open the actual scope file and add those. since they can't be blacklisted
	constScope, err := GetConstScope(scopeType)
	if err != nil {
		return nil, err
	}
	scope.AddRange(constScope)

	return scope.SortedStringSlice(), nil
}

func GetConstScope(scopeType string) ([]string, error) {
	file := fmt.Sprintf("scope-%s.txt", scopeType)
	scope := set.NewStringSet()

	c, err := util.ReadFileLineByLine(file)
	if err != nil {
		return nil, err
	}

	for line := range c {
		scope.Add(normalizeScope(line, scopeType))
	}

	return scope.SortedStringSlice(), nil
}

func normalizeScope(scopeItem, scopeType string) string {

	if scopeType == "domains" {
		re := regexp.MustCompile(`^\*\.`)
		scopeItem = re.ReplaceAllString(scopeItem, "")
		scopeItem = strings.ToLower(scopeItem)
	}

	return scopeItem
}
