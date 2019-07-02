package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/citizendata/datawallet/wallet-api/security"
	"time"
)

const (
	timestampLayout = "2006-01-02T15:04:05.000Z"
)

// ApiRequest is a generic structure for requests
type ApiRequest struct {
	RequestTimeUTC string
	Path           string
	Body           string
	PathParams     map[string]string
	QueryParams    map[string]string
	Headers        map[string]string
	TenantID       string
	Signature      string
}

type ApiResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

type ApiMessageBody struct {
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func addCors(h map[string]string) map[string]string {
	if h == nil {
		h = make(map[string]string)
	}
	h["Access-Control-Allow-Origin"] = "*"
	h["Access-Control-Allow-Credentials"] = "true"
	return h
}

func (a *ApiRequest) RequestTime() time.Time {
	t, _ := time.Parse(timestampLayout, a.RequestTimeUTC)
	return t
}

func (a *ApiRequest) ValidateSignature(publicKeyBase64 string) error {
	now := time.Now().UTC()
	reqTime := a.RequestTime()
	if reqTime.IsZero() {
		return errors.New("missing x-api-timestamp header (format: 2006-01-02T15:04:05.000Z)")
	}
	if now.Sub(reqTime) > (10 * time.Second) {
		return errors.New(fmt.Sprintf("bad signature (request too late, time=%s, elapsed=%d, now=%s)", a.RequestTimeUTC, now.Sub(reqTime), now.Format(timestampLayout)))
	}
	if a.Signature == "" {
		return errors.New("bad signature (empty)")
	}

	pubKey, err := security.PemBase64ToPublicKey(publicKeyBase64)
	if err != nil {
		return err
	}

	payload := []byte(fmt.Sprintf("%s|%s|%s", a.Path, a.Body, a.RequestTimeUTC))

	return security.VerifySignature(payload, a.Signature, pubKey)
}

func ApiRequestFromLambda(req *events.APIGatewayProxyRequest, tenantID string) *ApiRequest {
	ts := req.Headers["x-api-timestamp"]
	return &ApiRequest{
		RequestTimeUTC: ts,
		Path:           req.Path,
		Body:           req.Body,
		Headers:        req.Headers,
		PathParams:     req.PathParameters,
		QueryParams:    req.QueryStringParameters,
		TenantID:       tenantID,
		Signature:      req.Headers["x-api-signature"],
	}
}

func LambdaResponseFromApiResponse(resp *ApiResponse) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		Body:       resp.Body,
		StatusCode: resp.StatusCode,
		Headers:    addCors(resp.Headers),
	}
}

func ApiSuccessMessage(message string) *ApiResponse {
	obj := &ApiMessageBody{
		StatusCode: 200,
		Message: message,
	}
	return ApiResponseObject(obj)
}

func ApiResponseObject(v interface{}) *ApiResponse {
	j, err := json.Marshal(v)
	if err != nil {
		return NewApiError("could not marshal response", ErrorInternalError)
	}

	return &ApiResponse{
		StatusCode: 200,
		Body:       string(j),
	}
}
