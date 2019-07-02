package api

import "encoding/json"

const (
	ErrorValidation = "VALIDATION"
	ErrorUnauthorized = "UNAUTHORIZED"
	ErrorInternalError = "INTERNAL"
)

var (
	statusCodeMap = map[string]int {
		ErrorValidation: 400,
		ErrorUnauthorized: 401,
		ErrorInternalError: 500,
	}
)

func getStatusCode(errorType string) int {
	if val, ok := statusCodeMap[errorType]; ok {
		return val
	}
	return 500
}

type ApiErrorBody struct {
	ErrorType  string `json:"errorType"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func NewApiError(message string, errorType string) *ApiResponse {
	errorBody := ApiErrorBody{
		ErrorType:  errorType,
		Message:    message,
		StatusCode: getStatusCode(errorType),
	}
	errorBodyJson, err := json.Marshal(errorBody)

	// shouldn't happen
	if err != nil {
		panic(err)
	}

	return &ApiResponse{
		StatusCode: errorBody.StatusCode,
		Body:       string(errorBodyJson),
	}
}