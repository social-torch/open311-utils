package open311

// Issues that have been reported as service requests.  Location
// is submitted via lat/long or address
type Request struct {
	ServiceRequestId  string  `json:service_request_id`
	ServiceCode       string  `json:service_code`
	ServiceName       string  `json:service_name`
	Description       string  `json:description`
	Address           string  `json:address`
	AddressId         string  `json:address_id`
	ZipCode           int32   `json:zipcode`
	Latitude          float32 `json:lat`
	Longitude         float32 `json:lon`
	MediaUrl          string  `json:media_url`
	AgencyResponsible string  `json:agency_responsible`
	ServiceNotice     string  `json:service_notice`
	Status            string  `json:"name"`
	StatusNotes       string  `json:status_notes`
	RequestedDateTime string  `json:requested_datetime`
	UpdatedDateTime   string  `json:update_datetime`
	ExpectedDateTime  string  `json:expected_datetime`
}

