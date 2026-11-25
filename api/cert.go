package api

type CertClaimInfo struct {
	Name            string         `json:"name" validate:"required,fqdn"`
	Wildcard        bool           `json:"wildcard"`
	SubjectAltNames []string       `json:"san,omitempty" validate:"dive,required,fqdn|fqdnWildcard"`
	AutoDNS         *AutoDNSInfo   `json:"autodns,omitempty"`
	Hints           CertClaimHints `json:"hints"`
}

type CertClaimHints struct {
	TTL uint16 `json:"ttl,omitempty"` // Time from now until cert expiry in days.
}

type CertResources struct {
	Certificate string `json:"cert"`
	Key         string `json:"key"`
	Root        string `json:"root"`
	RootChain   string `json:"rootchain"`
	Chain       string `json:"chain"`
	FullChain   string `json:"fullchain"`
}

type CertInfo struct {
	Name      string `json:"name"`
	ClaimedBy struct {
		Name  string `json:"name"`
		EMail string `json:"email"`
	} `json:"claimedBy"`
	ClaimedOn   string `json:"claimedOn"`
	ValidTo     string `json:"validTo"`
	Valid       bool   `json:"valid"`
	LastAccess  string `json:"lastAccess"`
	NextRenewal string `json:"nextRenewal"`
	RenewCount  uint   `json:"renewCount"`
	AccessCount uint   `json:"accessCount"`
	Wildcard    bool   `json:"wildcard"`
	SubjectCN   string `json:"subjectCN"`
	IssuerCN    string `json:"issuerCN"`
	Serial      string `json:"serial"`
}

type CertInfoWithCA struct {
	CertInfo
	CA string `json:"ca"`
}
