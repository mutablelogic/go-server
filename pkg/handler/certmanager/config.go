package certmanager

type X509Name struct {
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizational_unit,omitempty"`
	Country            string `json:"country,omitempty"`
	Province           string `json:"province,omitempty"`
	Locality           string `json:"locality,omitempty"`
	StreetAddress      string `json:"street_address,omitempty"`
	PostalCode         string `json:"postal_code,omitempty"`
}

type Config struct {
	X509Name `json:"x509_name"`
}
