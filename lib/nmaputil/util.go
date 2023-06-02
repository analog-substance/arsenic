package nmaputil

import (
	"path/filepath"
	"strings"

	"github.com/analog-substance/arsenic/lib/host"
	"github.com/analog-substance/fileutil"
	"github.com/analog-substance/nmap/v3"
)

func getHost(hostnames []string, ips []string) (*host.Host, error) {
	var err error

	currentHost := host.GetFirst(append(hostnames, ips...)...)
	if currentHost == nil {
		currentHost, err = host.AddHost(hostnames, ips)
		if err != nil {
			return nil, err
		}
	}
	return currentHost, nil
}

func writeToFile(h *host.Host, name string, data []byte) error {
	path := filepath.Join(h.Dir, "recon", name)
	err := fileutil.WriteString(path, string(data))
	if err != nil {
		return err
	}
	return nil
}

func hasOpenPorts(h nmap.Host) bool {
	for _, p := range h.Ports {
		if strings.Contains(p.State.State, "open") {
			return true
		}
	}
	return false
}
