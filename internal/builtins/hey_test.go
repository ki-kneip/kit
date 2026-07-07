package builtins

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"mvdan.cc/sh/v3/interp"
)

func TestHey(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	var out bytes.Buffer
	hc := interp.HandlerContext{Stdout: &out, Stderr: &out}

	if err := hey(context.Background(), hc, []string{"-n", "30", "-c", "5", srv.URL}); err != nil {
		t.Fatal(err)
	}
	if got := hits.Load(); got != 30 {
		t.Errorf("server got %d hits, want 30", got)
	}
	if !strings.Contains(out.String(), "30 done, 0 failed") {
		t.Errorf("unexpected report:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "status 200") {
		t.Errorf("missing status line:\n%s", out.String())
	}
}

func TestHeyUsage(t *testing.T) {
	hc := interp.HandlerContext{Stdout: &bytes.Buffer{}}
	if err := hey(context.Background(), hc, nil); err == nil {
		t.Fatal("expected usage error without URL")
	}
}
