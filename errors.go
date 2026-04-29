package openservice

import "errors"

var (
	ErrInvalidConfig       = errors.New("openservice: invalid config")
	ErrMissingBaseURL      = errors.New("openservice: baseURL is required")
	ErrMissingMID          = errors.New("openservice: mid is required")
	ErrMissingSecret       = errors.New("openservice: secret is required")
	ErrInvalidTimeout      = errors.New("openservice: timeout must be greater than zero")
	ErrInvalidBaseURL      = errors.New("openservice: baseURL must be a valid http or https URL")
	ErrInvalidRequest      = errors.New("openservice: invalid request")
	ErrInvalidResponse     = errors.New("openservice: invalid response")
	ErrHTTPTransport       = errors.New("openservice: http transport failed")
	ErrUnexpectedStatus    = errors.New("openservice: unexpected http status")
	ErrResponseCodeNonZero = errors.New("openservice: response code is not zero")
	ErrMissingData         = errors.New("openservice: response data is missing")
)
