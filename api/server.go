package api

import "time"

type ServerInfo struct {
	Version *ServerInfoVersion `json:"version"`
	Contact *ServerInfoContact `json:"contact"`
	Auth    *ServerInfoAuth    `json:"auth"`
	Renewal *ServerInfoRenewal `json:"renewal"`
}

type ServerInfoVersion struct {
	Daemon string `json:"daemon"`
	API    string `json:"api"`
}

type ServerInfoContact struct {
	URL   string   `json:"url"`
	EMail []string `json:"email"`
}

type ServerInfoAuth struct {
	GroupsPrefix      string `json:"groups-prefix"`
	GroupsDomainDelim string `json:"groups-domain-delim"`
}

type ServerInfoRenewal struct {
	LastRun    *time.Time `json:"lastRun"`
	Successful uint       `json:"successful"`
	Failed     uint       `json:"failed"`
}

type DNSHandlerInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Feature     []string `json:"feature"`
	ZoneNesting bool     `json:"zoneNesting"`
}

type DNSRootzoneInfo struct {
	Root    string   `json:"root"`
	AutoDNS string   `json:"autodns"`
	AcmeDNS string   `json:"acmedns"`
	CA      []string `json:"ca"`
}

type CAInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"desc"`
	LogoPath    string   `json:"logo"`
	URL         string   `json:"url"`
	Roots       string   `json:"roots"`
	TotalValid  uint     `json:"totalValid"`
	TotalIssued uint     `json:"totalIssued"`
	Type        string   `json:"type"`
	IsAcme      bool     `json:"acme"`
	Rootzones   []string `json:"rtzn"`
	Enabled     bool     `json:"enabled"`
}

type AutoDNSInfo struct {
	IPv4 string `json:"ipv4,omitempty" validate:"required,ipv4"`
}

type ServerStatus struct {
	DNS      []HealthInfo `json:"dns"`
	CA       []HealthInfo `json:"ca"`
	DB       []HealthInfo `json:"db"`
	Auth     []HealthInfo `json:"auth"`
	Resolver []HealthInfo `json:"resolver"`
	Cached   bool         `json:"cached"`
	Age      int32        `json:"age"`
}

type HealthInfo struct {
	ID       string `json:"id"`
	Healthy  bool   `json:"healthy"`
	Endpoint string `json:"endpoint"`
	Message  string `json:"message"`
}
