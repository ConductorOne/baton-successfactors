package client

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorResponseMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"token shape", `{"errorHttpCode":"401","errorMessage":"Company X does not exist in this data center."}`, "Company X does not exist in this data center."},
		{"token shape, code as number", `{"errorHttpCode":401,"errorMessage":"Unauthorized"}`, "Unauthorized"},
		{"odata shape with code", `{"error":{"code":"BAD_REQUEST","message":{"lang":"en-US","value":"Invalid query"}}}`, "BAD_REQUEST: Invalid query"},
		{"odata shape without code", `{"error":{"message":{"lang":"en-US","value":"Invalid query"}}}`, "Invalid query"},
		{"empty object", `{}`, ""},
		{"odata error without a message", `{"error":{"code":"X"}}`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e ErrorResponse
			require.NoError(t, json.Unmarshal([]byte(tt.body), &e))
			assert.Equal(t, tt.want, e.Message())
		})
	}
}
