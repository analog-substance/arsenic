package engine

import (
	"fmt"
	"regexp"

	"github.com/analog-substance/arsenic/lib/scope"
	"github.com/analog-substance/tengo/v2"
	"github.com/analog-substance/tengomod/interop"
)

// ScopeModule represents the 'scope' import module
func (s *Script) ScopeModule() map[string]tengo.Object {
	return map[string]tengo.Object{
		"domains": &tengo.UserFunction{
			Name:  "domains",
			Value: s.domains,
		},
		"root_domains": &interop.AdvFunction{
			Name:    "root_domains",
			NumArgs: interop.MaxArgs(1),
			Args:    []interop.AdvArg{interop.BoolArg("all-root-domains")},
			Value:   s.rootDomains,
		},
		"const_domains": &tengo.UserFunction{
			Name:  "const_domains",
			Value: s.constDomains,
		},
		"const_subdomains": &interop.AdvFunction{
			Name:    "const_subdomains",
			NumArgs: interop.ExactArgs(1),
			Args:    []interop.AdvArg{interop.StrArg("domain")},
			Value:   s.constSubDomains,
		},
		"prune": &interop.AdvFunction{
			Name:    "prune",
			NumArgs: interop.MinArgs(1),
			Args:    []interop.AdvArg{interop.StrSliceArg("items", true)},
			Value:   s.prune,
		},
	}
}

// domains returns the in scope domains
// Represents 'scope.domains() []string'
func (s *Script) domains(args ...tengo.Object) (tengo.Object, error) {
	domains, err := scope.GetScope("domains")
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	return interop.GoStrSliceToTArray(domains), nil
}

// rootDomains returns the in scope root domains
// Represents 'scope.root_domains(all_root_domains bool = false)'
func (s *Script) rootDomains(args map[string]interface{}) (tengo.Object, error) {
	pruneBlacklisted := true
	if value, ok := args["all-root-domains"]; ok {
		pruneBlacklisted = !(value.(bool))
	}

	domains, err := scope.GetScope("domains")
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	rootDomains := scope.GetRootDomains(domains, pruneBlacklisted)
	return interop.GoStrSliceToTArray(rootDomains), nil
}

// constDomains returns the domains which are always in scope contained in the scope-domains.txt file
// Represents 'scope.const_domains'
func (s *Script) constDomains(args ...tengo.Object) (tengo.Object, error) {
	constDomains, err := scope.GetConstScope("domains")
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	return interop.GoStrSliceToTArray(constDomains), nil
}

// constSubDomains returns the sub-domains of the specified domain which are always in scope contained in the scope-domains.txt file
// Represents 'scope.const_domains'
func (s *Script) constSubDomains(args map[string]interface{}) (tengo.Object, error) {
	domain := args["domain"].(string)

	re := regexp.MustCompile(fmt.Sprintf(`(?i)%s$`, regexp.QuoteMeta(domain)))

	constDomains, err := scope.GetConstScope("domains")
	if err != nil {
		return interop.GoErrToTErr(err), nil
	}

	var constSubDomains []string
	for _, d := range constDomains {
		if re.MatchString(d) {
			constSubDomains = append(constSubDomains, d)
		}
	}

	return interop.GoStrSliceToTArray(constSubDomains), nil
}

// prune will only return the in scope ips/domains from the provided items
// Represents 'scope.prune(items ....string)'
func (s *Script) prune(args map[string]interface{}) (tengo.Object, error) {
	allItems := args["items"].([]string)

	var inScope []string
	for _, item := range allItems {
		if scope.IsInScope(item, false) {
			inScope = append(inScope, item)
		}
	}

	return interop.GoStrSliceToTArray(inScope), nil
}
