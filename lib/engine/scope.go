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
		"root_domains": &tengo.UserFunction{
			Name:  "root_domains",
			Value: s.rootDomains,
		},
		"const_domains": &tengo.UserFunction{
			Name:  "const_domains",
			Value: s.constDomains,
		},
		"const_subdomains": &tengo.UserFunction{
			Name:  "const_subdomains",
			Value: s.constSubDomains,
		},
		"prune": &tengo.UserFunction{
			Name:  "prune",
			Value: s.prune,
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
func (s *Script) rootDomains(args ...tengo.Object) (tengo.Object, error) {
	pruneBlacklisted := true
	if len(args) > 0 {
		value, ok := tengo.ToBool(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "all_root_domains",
				Expected: "bool",
				Found:    args[0].TypeName(),
			}
		}

		pruneBlacklisted = !value
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
func (s *Script) constSubDomains(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	domain, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "domain",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

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
// Represents 'scope.prune(items ....(string|[]string))'
func (s *Script) prune(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, tengo.ErrWrongNumArguments
	}

	var allItems []string
	for _, arg := range args {
		switch o := arg.(type) {
		case *tengo.String:
			allItems = append(allItems, o.Value)
		case *tengo.Array:
			items, err := interop.GoTSliceToGoStrSlice(o.Value, "items")
			if err != nil {
				return nil, err
			}

			allItems = append(allItems, items...)
		case *tengo.ImmutableArray:
			items, err := interop.GoTSliceToGoStrSlice(o.Value, "items")
			if err != nil {
				return nil, err
			}

			allItems = append(allItems, items...)
		default:
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "item(s)",
				Expected: "array|string",
				Found:    arg.TypeName(),
			}
		}
	}

	var inScope []string
	for _, item := range allItems {
		if scope.IsInScope(item, false) {
			inScope = append(inScope, item)
		}
	}

	return interop.GoStrSliceToTArray(inScope), nil
}
