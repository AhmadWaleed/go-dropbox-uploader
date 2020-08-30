package dropbox

import "encoding/json"

// ClientErr bad request
type ClientErr struct {
	Status     string
	StatusCode int
	Summary    string `json:"error_summary"`
}

// Error string.
func (e *ClientErr) Error() string {
	body, err := json.MarshalIndent(&e, "", " ")
	if err != nil {
		return e.Summary
	}

	return string(body)
}
