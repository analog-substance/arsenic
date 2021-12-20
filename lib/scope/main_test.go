package scope

import (
	"reflect"
	"testing"
)

func getTestScope() *scope {
	asScopeInit = true
	return &scope{
		domainsExplicitlyInScope:     []string{"www.example.com", "target.subdomain.example.net"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains:       []string{"example.net"},
		blacklistedDomains:           []string{"blog.example.com"},
		explicitDomainsLoaded:        true,

		hostIPsExplicitlyInScope:       []string{"10.10.10.10"},
		hostIPsExplicitlyInScopeLoaded: true,
	}
}

func TestGetRootDomains(t *testing.T) {

	type args struct {
		domains          []string
		pruneBlacklisted bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"Should prune blacklisted",
			args{[]string{"www.example.com", "www.example.net"}, true},
			[]string{"example.com"},
		},
		{
			"Should not prune blacklisted",
			args{[]string{"www.example.com", "www.example.net"}, false},
			[]string{"example.com", "example.net"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asScope = getTestScope()

			if got := GetRootDomains(tt.args.domains, tt.args.pruneBlacklisted); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRootDomains() = %v, want %v", got, tt.want)
			}
		})
	}
}

//func TestGetScope(t *testing.T) {
//	asScope = getTestScope()
//	type args struct {
//		scopeType string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    []string
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got, err := GetScope(tt.args.scopeType)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetScope() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetScope() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

func TestIsInScope(t *testing.T) {
	asScope = getTestScope()
	type args struct {
		ipOrHostname string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Should return true for in scope domain",
			args{"www.example.com"},
			true,
		},
		{
			"Should return true for in scope domain",
			args{"pizza.example.com"},
			true,
		},
		{
			"Should return false for blacklisted domain",
			args{"blog.example.com"},
			false,
		},
		{
			"Should return false for out of scope domain",
			args{"blog.example.info"},
			false,
		},
		{
			"Should return false for out of scope domain",
			args{"fake.example.net"},
			false,
		},
		{
			"Should return true for explicit in scope domain",
			args{"target.subdomain.example.net"},
			true,
		},
		{
			"Should return true for explicit in scope IP",
			args{"10.10.10.10"},
			true,
		},
		{
			"Should return false for out of scope IP",
			args{"11.11.11.11"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInScope(tt.args.ipOrHostname, false); got != tt.want {
				t.Errorf("IsInScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIp(t *testing.T) {
	type args struct {
		ipOrHostname string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Should return true for valid IP",
			args{"11.11.11.11"},
			true,
		},
		{
			"Should return false for invalid IP",
			args{"999.11.320.11"},
			false,
		},
		{
			"Should return false domain",
			args{"example.com"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIp(tt.args.ipOrHostname); got != tt.want {
				t.Errorf("IsIp() = %v, want %v", got, tt.want)
			}
		})
	}
}
