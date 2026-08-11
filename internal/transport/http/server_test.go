package httpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"os"

	httpserver "github.com/marketing-digest/gateway/internal/transport/http"
)

func TestLiveness(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	srv := httpserver.New(":0", nil, log, "*", "marketing-digest-dev")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("%v", body)
	}
}
