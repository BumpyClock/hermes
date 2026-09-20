package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	hermes "github.com/BumpyClock/hermes"
	"github.com/BumpyClock/hermes/internal/definitionbundle/offline"
)

func TestUnpreparedBuildRefusesConformance(t *testing.T) {
	if offline.Installed {
		t.Skip("prepared candidate uses the explicit offline boundary")
	}
	var output bytes.Buffer
	if err := run([]string{"-describe"}, &output); err == nil || output.Len() != 0 {
		t.Fatalf("ordinary executable claimed candidate identity: %v %s", err, output.String())
	}
}

func TestPinnedSelectionCheckedBeforeBundleInputs(t *testing.T) {
	result := &report{Engine: identity{BuildID: "selected-build", ExecutableSHA256: "selected-binary"}}
	for _, pins := range [][2]string{{"different-binary", "selected-build"}, {"selected-binary", "different-build"}} {
		err := evaluate(result, "must-not-read-manifest", "must-not-read-archive", pins[0], pins[1], "unused", "", "")
		if err == nil || !strings.Contains(err.Error(), "engine digest/build identity mismatch") {
			t.Fatalf("wrong engine selection reached bundle inputs: %v", err)
		}
	}
}

func TestOfflineBoundarySecurity(t *testing.T) {
	if !offline.Installed {
		t.Skip("requires explicitly prepared candidate snapshot; no live DNS fallback")
	}
	offline.Configure([]string{"fixture.example"})
	transport := &denyTransport{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = transport
	defer func() { http.DefaultTransport = oldTransport }()
	client := hermes.New(hermes.WithDefinitions(nil), hermes.WithTransport(transport))
	const body = `<html><head><title>Offline security fixture</title></head><body><article><p>This original article validates the public ParseHTML network boundary without resolving any real domain.</p></article></body></html>`
	if _, err := client.ParseHTML(context.Background(), body, "https://fixture.example/article"); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"https://127.0.0.1/a", "https://10.0.0.1/a", "https://[::1]/a", "https://unplanned.invalid/a"} {
		_, err := client.ParseHTML(context.Background(), body, target)
		var parseError *hermes.ParseError
		if !errors.As(err, &parseError) || parseError.Code != hermes.ErrInvalidURL {
			t.Fatalf("expected ErrInvalidURL for %s: %v", target, err)
		}
	}
	if transport.calls.Load() != 0 || offline.Calls() == 0 {
		t.Fatalf("unexpected network boundary evidence: HTTP=%d offline lookups=%d", transport.calls.Load(), offline.Calls())
	}
}
