package scope

import (
	"os"
	"reflect"
	"regexp"
	"testing"
)

func getTestScope() *scope {
	return &scope{
		domainsExplicitlyInScope:     []string{"www.example.com", "target.subdomain.example.net"},
		rootDomainsExplicitlyInScope: []string{"example.com"},
		blacklistedRootDomains:       []string{"example.net"},
		blacklistedDomainRegexps:     []*regexp.Regexp{regexp.MustCompile(regexp.QuoteMeta("blog.example.com"))},
		domainInfoLoaded:             true,

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

// func TestGetScope(t *testing.T) {
// 	asScope = getTestScope()
// 	type args struct {
// 		scopeType string
// 	}
// 	tests := []struct {
// 		name    string
// 		setup   func()
// 		args    args
// 		want    []string
// 		wantErr bool
// 	}{
// 		{
// 			name: "",
// 			setup: func() {
// 				os.WriteFile("scope-domains.txt", []byte("www.example.com\ntest.example.com\nex.com"), 0644)
// 				os.WriteFile("scope-domains-1.txt", []byte("1.1.example.com\n1.2.example.com"), 0644)
// 				os.WriteFile("scope-domains-2.txt", []byte("2.1.example.com\nblacklist.ex.com"), 0644)
// 			},
// 			args: args{
// 				scopeType: "domains",
// 			},
// 			want: []string{},
// 		},
// 	}

// 	t.Cleanup(func() {
// 		os.RemoveAll("scope-domains.txt")
// 		os.RemoveAll("scope-domains-1.txt")
// 		os.RemoveAll("scope-domains-2.txt")
// 	})
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.setup != nil {
// 				tt.setup()
// 			}

// 			got, err := GetScope(tt.args.scopeType)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("GetScope() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetScope() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestGetConstScope(t *testing.T) {
	type args struct {
		scopeType string
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "No scope file",
			args: args{
				scopeType: "test",
			},
			wantErr: true,
		},
		{
			name: "Scope file with dups",
			setup: func() {
				os.WriteFile("scope-domains.txt", []byte("*.www.example.com\ntest.example.com\ntest.example.com\nblah.example.com"), 0644)
			},
			args: args{
				scopeType: "domains",
			},
			want: []string{
				"blah.example.com",
				"test.example.com",
				"www.example.com",
			},
		},
	}

	t.Cleanup(func() {
		os.RemoveAll("scope-domains.txt")
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			got, err := GetConstScope(tt.args.scopeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConstScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConstScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
