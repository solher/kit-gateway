package common

import "fmt"

// APIError defines a standard format for API errors.
type APIError struct {
	// The status code.
	Status int `json:"status"`
	// The description of the API error.
	Description string `json:"description"`
	// The token uniquely identifying the API error.
	ErrorCode string `json:"errorCode"`
	// Additional infos.
	Params map[string]interface{} `json:"params,omitempty"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s : %s", e.ErrorCode, e.Description)
}

var (
	ErrInternal = APIError{
		Status:      500,
		Description: "An internal error occured. Please retry later.",
		ErrorCode:   "INTERNAL_ERROR",
	}
	ErrBodyDecoding = APIError{
		Status:      400,
		Description: "Could not decode the JSON request.",
		ErrorCode:   "BODY_DECODING_ERROR",
	}
	ErrUnauthorized = APIError{
		Status:      401,
		Description: "Authorization Required.",
		ErrorCode:   "AUTHORIZATION_REQUIRED",
	}
	ErrForbidden = APIError{
		Status:      403,
		Description: "The specified resource was not found or you don't have sufficient permissions.",
		ErrorCode:   "FORBIDDEN",
	}
	ErrFilterDecoding = APIError{
		Status:      422,
		Description: "Could not decode the given filter.",
		ErrorCode:   "FILTER_DECODING_ERROR",
	}
	ErrValidation = APIError{
		Status:      422,
		Description: "The model validation failed.",
		ErrorCode:   "VALIDATION_ERROR",
	}
)
