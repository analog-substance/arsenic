package scope

import (
	"testing"
)

func Test_scope_IsBlacklistedDomain(t *testing.T) {
	type fields struct {
		domainsExplicitlyInScope       []string
		rootDomainsExplicitlyInScope   []string
		blacklistedRootDomains         []string
		blacklistedDomains             []string
		explicitDomainsLoaded          bool
		hostIPsExplicitlyInScope       []string
		hostIPsExplicitlyInScopeLoaded bool
	}
	defaultFields := fields {
		domainsExplicitlyInScope: []string{"www.example.com"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains: []string{"example.net"},
		blacklistedDomains: []string{"blog.example.com"},
		explicitDomainsLoaded: true,

		hostIPsExplicitlyInScope: []string{},
		hostIPsExplicitlyInScopeLoaded: true,
	}
	type args struct {
		checkDomain string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Non blacklisted domain should return false",
			defaultFields,
			args{"example.com"},
			false,
		},
		{
			"Non blacklisted domain should return false",
			defaultFields,
			args{"www.example.com"},
			false,
		},
		{
			"blacklisted domain should return true",
			defaultFields,
			args{"blog.example.com"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scope{
				domainsExplicitlyInScope:       tt.fields.domainsExplicitlyInScope,
				rootDomainsExplicitlyInScope:   tt.fields.rootDomainsExplicitlyInScope,
				blacklistedRootDomains:         tt.fields.blacklistedRootDomains,
				blacklistedDomains:             tt.fields.blacklistedDomains,
				explicitDomainsLoaded:          tt.fields.explicitDomainsLoaded,
				hostIPsExplicitlyInScope:       tt.fields.hostIPsExplicitlyInScope,
				hostIPsExplicitlyInScopeLoaded: tt.fields.hostIPsExplicitlyInScopeLoaded,
			}
			if got := s.IsBlacklistedDomain(tt.args.checkDomain); got != tt.want {
				t.Errorf("IsBlacklistedDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scope_IsBlacklistedRootDomain(t *testing.T) {
	type fields struct {
		domainsExplicitlyInScope       []string
		rootDomainsExplicitlyInScope   []string
		blacklistedRootDomains         []string
		blacklistedDomains             []string
		explicitDomainsLoaded          bool
		hostIPsExplicitlyInScope       []string
		hostIPsExplicitlyInScopeLoaded bool
	}

	defaultFields := fields {
		domainsExplicitlyInScope: []string{"www.example.com"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains: []string{"example.net"},
		blacklistedDomains: []string{"blog.example.com"},
		explicitDomainsLoaded: true,

		hostIPsExplicitlyInScope: []string{},
		hostIPsExplicitlyInScopeLoaded: true,
	}

	type args struct {
		rootDomain string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Non blacklisted root domain should return false",
			defaultFields,
			args{"example.com"},
			false,
		},
		{
			"Non blacklisted root domain should return false",
			defaultFields,
			args{"example.info"},
			false,
		},
		{
			"blacklisted root domain should return true",
			defaultFields,
			args{"example.net"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scope{
				domainsExplicitlyInScope:       tt.fields.domainsExplicitlyInScope,
				rootDomainsExplicitlyInScope:   tt.fields.rootDomainsExplicitlyInScope,
				blacklistedRootDomains:         tt.fields.blacklistedRootDomains,
				blacklistedDomains:             tt.fields.blacklistedDomains,
				explicitDomainsLoaded:          tt.fields.explicitDomainsLoaded,
				hostIPsExplicitlyInScope:       tt.fields.hostIPsExplicitlyInScope,
				hostIPsExplicitlyInScopeLoaded: tt.fields.hostIPsExplicitlyInScopeLoaded,
			}
			if got := s.IsBlacklistedRootDomain(tt.args.rootDomain); got != tt.want {
				t.Errorf("IsBlacklistedRootDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scope_IsDomainExplicitlyInScope(t *testing.T) {
	type fields struct {
		domainsExplicitlyInScope       []string
		rootDomainsExplicitlyInScope   []string
		blacklistedRootDomains         []string
		blacklistedDomains             []string
		explicitDomainsLoaded          bool
		hostIPsExplicitlyInScope       []string
		hostIPsExplicitlyInScopeLoaded bool
	}
	defaultFields := fields {
		domainsExplicitlyInScope: []string{"www.example.com"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains: []string{"example.net"},
		blacklistedDomains: []string{"blog.example.com"},
		explicitDomainsLoaded: true,

		hostIPsExplicitlyInScope: []string{},
		hostIPsExplicitlyInScopeLoaded: true,
	}
	type args struct {
		checkDomain string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Explicit domain should return true",
			defaultFields,
			args{"www.example.com"},
			true,
		},
		{
			"Non explicit domain should return false",
			defaultFields,
			args{"garbage.example.com"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scope{
				domainsExplicitlyInScope:       tt.fields.domainsExplicitlyInScope,
				rootDomainsExplicitlyInScope:   tt.fields.rootDomainsExplicitlyInScope,
				blacklistedRootDomains:         tt.fields.blacklistedRootDomains,
				blacklistedDomains:             tt.fields.blacklistedDomains,
				explicitDomainsLoaded:          tt.fields.explicitDomainsLoaded,
				hostIPsExplicitlyInScope:       tt.fields.hostIPsExplicitlyInScope,
				hostIPsExplicitlyInScopeLoaded: tt.fields.hostIPsExplicitlyInScopeLoaded,
			}
			if got := s.IsDomainExplicitlyInScope(tt.args.checkDomain); got != tt.want {
				t.Errorf("IsDomainExplicitlyInScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scope_IsIPInScope(t *testing.T) {
	type fields struct {
		domainsExplicitlyInScope       []string
		rootDomainsExplicitlyInScope   []string
		blacklistedRootDomains         []string
		blacklistedDomains             []string
		explicitDomainsLoaded          bool
		hostIPsExplicitlyInScope       []string
		hostIPsExplicitlyInScopeLoaded bool
	}
	defaultFields := fields {
		domainsExplicitlyInScope: []string{"www.example.com"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains: []string{"example.net"},
		blacklistedDomains: []string{"blog.example.com"},
		explicitDomainsLoaded: true,

		hostIPsExplicitlyInScope: []string{"10.10.10.10"},
		hostIPsExplicitlyInScopeLoaded: true,
	}
	type args struct {
		checkHostIP string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Explicit IP should return true",
			defaultFields,
			args{"10.10.10.10"},
			true,
		},
		{
			"Non explicit IP should return false",
			defaultFields,
			args{"11.11.11.11"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scope{
				domainsExplicitlyInScope:       tt.fields.domainsExplicitlyInScope,
				rootDomainsExplicitlyInScope:   tt.fields.rootDomainsExplicitlyInScope,
				blacklistedRootDomains:         tt.fields.blacklistedRootDomains,
				blacklistedDomains:             tt.fields.blacklistedDomains,
				explicitDomainsLoaded:          tt.fields.explicitDomainsLoaded,
				hostIPsExplicitlyInScope:       tt.fields.hostIPsExplicitlyInScope,
				hostIPsExplicitlyInScopeLoaded: tt.fields.hostIPsExplicitlyInScopeLoaded,
			}
			if got := s.IsIPInScope(tt.args.checkHostIP); got != tt.want {
				t.Errorf("IsIPInScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scope_IsRootDomainInScope(t *testing.T) {
	type fields struct {
		domainsExplicitlyInScope       []string
		rootDomainsExplicitlyInScope   []string
		blacklistedRootDomains         []string
		blacklistedDomains             []string
		explicitDomainsLoaded          bool
		hostIPsExplicitlyInScope       []string
		hostIPsExplicitlyInScopeLoaded bool
	}
	defaultFields := fields {
		domainsExplicitlyInScope: []string{"www.example.com"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains: []string{"example.net"},
		blacklistedDomains: []string{"blog.example.com"},
		explicitDomainsLoaded: true,

		hostIPsExplicitlyInScope: []string{},
		hostIPsExplicitlyInScopeLoaded: true,
	}
	type args struct {
		checkRootDomain string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Explicit domain should return true",
			defaultFields,
			args{"example.com"},
			true,
		},
		{
			"Non explicit domain should return false",
			defaultFields,
			args{"example.info"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scope{
				domainsExplicitlyInScope:       tt.fields.domainsExplicitlyInScope,
				rootDomainsExplicitlyInScope:   tt.fields.rootDomainsExplicitlyInScope,
				blacklistedRootDomains:         tt.fields.blacklistedRootDomains,
				blacklistedDomains:             tt.fields.blacklistedDomains,
				explicitDomainsLoaded:          tt.fields.explicitDomainsLoaded,
				hostIPsExplicitlyInScope:       tt.fields.hostIPsExplicitlyInScope,
				hostIPsExplicitlyInScopeLoaded: tt.fields.hostIPsExplicitlyInScopeLoaded,
			}
			if got := s.IsRootDomainInScope(tt.args.checkRootDomain); got != tt.want {
				t.Errorf("IsRootDomainInScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scope_IsDomainInScope(t *testing.T) {
	type fields struct {
		domainsExplicitlyInScope       []string
		rootDomainsExplicitlyInScope   []string
		blacklistedRootDomains         []string
		blacklistedDomains             []string
		explicitDomainsLoaded          bool
		hostIPsExplicitlyInScope       []string
		hostIPsExplicitlyInScopeLoaded bool
	}
	defaultFields := fields {
		domainsExplicitlyInScope: []string{"www.example.com","target.subdomain.example.net"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains: []string{"example.net"},
		blacklistedDomains: []string{"blog.example.com"},
		explicitDomainsLoaded: true,

		hostIPsExplicitlyInScope: []string{},
		hostIPsExplicitlyInScopeLoaded: true,
	}
	type args struct {
		domain string
		forceBlacklistPrecedence bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"Should return true for in scope domain",
			defaultFields,
			args{"www.example.com", false},
			true,
		},
		{
			"Should return true for in scope domain",
			defaultFields,
			args{"pizza.example.com", false},
			true,
		},
		{
			"Should return false for blacklisted domain",
			defaultFields,
			args{"blog.example.com", false},
			false,
		},
		{
			"Should return false for out of scope domain",
			defaultFields,
			args{"blog.example.info", false},
			false,
		},
		{
			"Should return false for out of scope domain",
			defaultFields,
			args{"fake.example.net", false},
			false,
		},
		{
			"Should return true for explicit in scope domain",
			defaultFields,
			args{"target.subdomain.example.net", false},
			true,
		},
		{
			"Should return false for subdomain of blacklisted domain when preferred",
			defaultFields,
			args{"pizza-123.us-west-2.elb.amazonaws.com", true},
			false,
		},
		{
			"Should return true for subdomain of blacklisted domain when preferred",
			defaultFields,
			args{"pizza-123.us-west-2.elb.amazonaws.com", false},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := scope{
				domainsExplicitlyInScope:       tt.fields.domainsExplicitlyInScope,
				rootDomainsExplicitlyInScope:   tt.fields.rootDomainsExplicitlyInScope,
				blacklistedRootDomains:         tt.fields.blacklistedRootDomains,
				blacklistedDomains:             tt.fields.blacklistedDomains,
				explicitDomainsLoaded:          tt.fields.explicitDomainsLoaded,
				hostIPsExplicitlyInScope:       tt.fields.hostIPsExplicitlyInScope,
				hostIPsExplicitlyInScopeLoaded: tt.fields.hostIPsExplicitlyInScopeLoaded,
			}
			if got := s.IsDomainInScope(tt.args.domain, tt.args.forceBlacklistPrecedence); got != tt.want {
				t.Errorf("IsDomainInScope() = %v, want %v", got, tt.want)
			}
		})
	}
}