package open311

// Issues that have been reported as service requests.
// Location is submitted via lat/long or address
// see https://wiki.open311.org/GeoReport_v2/#get-service-request

type Request struct {
	ServiceRequestId  string  `json:"service_request_id"` // The unique ID of the service request created.
	Status            string  `json:"status"`             // The current status of the service request.
	StatusNotes       string  `json:"status_notes"`       // Explanation of why status was changed to current state or more details on current status than conveyed with status alone.
	ServiceName       string  `json:"service_name"`       // The human readable name of the service request type
	ServiceCode       string  `json:"service_code"`       // The unique identifier for the service request type
	Description       string  `json:"description"`        // A full description of the request or report submitted.
	AgencyResponsible string  `json:"agency_responsible"` // The agency responsible for fulfilling or otherwise addressing the service request.
	ServiceNotice     string  `json:"service_notice"`     // Information about the action expected to fulfill the request or otherwise address the information reported.
	RequestedDateTime string  `json:"requested_datetime"` // The date and time when the service request was made.
	UpdatedDateTime   string  `json:"update_datetime"`    // The date and time when the service request was last modified. For requests with status=closed, this will be the date the request was closed.
	ExpectedDateTime  string  `json:"expected_datetime"`  // The date and time when the service request can be expected to be fulfilled. This may be based on a service-specific service level agreement.
	Address           string  `json:"address"`            // Human readable address or description of location.
	AddressId         string  `json:"address_id"`         // The internal address ID used by a jurisdictions master address repository or other addressing system.
	ZipCode           int32   `json:"zipcode"`            // The postal code for the location of the service request.
	Latitude          float32 `json:"lat"`                // latitude using the (WGS84) projection.
	Longitude         float32 `json:"lon"`                // longitude using the (WGS84) projection.
	MediaUrl          string  `json:"media_url"`          // A URL to media associated with the request, eg an image.
}
