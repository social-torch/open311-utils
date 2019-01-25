package open311

// A Service is offered by a city and defines what requests a citizen can make
// Single service (type) offered via Open311
// see https://wiki.open311.org/GeoReport_v2/#get-service-list

type Service struct {
	ServiceCode string   `json:"service_code"`
	ServiceName string   `json:"service_name"`
	Description string   `json:"description"`
	Metadata    bool     `json:"metadata"`
	Type        string   `json:"type"`
	Keywords    []string `json:"keywords"`
	Group       string   `json:"group"`
}
