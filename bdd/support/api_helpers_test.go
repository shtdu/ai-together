package support

import (
	"testing"
)

func TestHandleAPIResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "success status code",
			statusCode: 200,
			body:       []byte(`{"status":"ok"}`),
			wantErr:    false,
		},
		{
			name:       "created status code",
			statusCode: 201,
			body:       []byte(`{"id":"123"}`),
			wantErr:    false,
		},
		{
			name:       "bad request with body",
			statusCode: 400,
			body:       []byte(`{"error":"invalid input"}`),
			wantErr:    true,
			errMsg:     "API returned error 400: {\"error\":\"invalid input\"}",
		},
		{
			name:       "not found without body",
			statusCode: 404,
			body:       []byte{},
			wantErr:    true,
			errMsg:     "API returned error 404: Not Found",
		},
		{
			name:       "internal server error",
			statusCode: 500,
			body:       []byte(`{"error":"database connection failed"}`),
			wantErr:    true,
			errMsg:     "API returned error 500: {\"error\":\"database connection failed\"}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handleAPIResponse(tt.statusCode, tt.body)

			if tt.wantErr {
				if err == nil {
					t.Errorf("handleAPIResponse() expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("handleAPIResponse() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("handleAPIResponse() unexpected error = %v", err)
				}
			}
		})
	}
}
