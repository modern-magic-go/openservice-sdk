package openservice

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOfficialAccountAccessToken_PostsSignedQueryAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/wechat/getAccessToken", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "1900001", query.Get("mid"))
		assert.NotEmpty(t, query.Get("nonce_str"))
		assert.NotEmpty(t, query.Get("timestamp"))
		assert.NotEmpty(t, query.Get("sign"))

		payload := map[string]any{}
		for key := range query {
			payload[key] = query.Get(key)
		}
		assert.True(t, NewSigner("merchant-secret").VerifySign(payload))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Result[map[string]string]{
			Code:    0,
			Message: "success",
			Data:    map[string]string{"access_token": "ACCESS_TOKEN"},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL: server.URL,
		MID:     "1900001",
		Secret:  "merchant-secret",
		Timeout: time.Second,
	})
	require.NoError(t, err)

	resp, err := client.OfficialAccount().AccessToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "ACCESS_TOKEN", resp.AccessToken)
}
