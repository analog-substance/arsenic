package models

type ReviewHost struct {
	Host     string `json:"host"`
	Reviewer string `json:"reviewer"`
}

type HostContentDiscovery struct {
	Host string `json:"host"`
}
