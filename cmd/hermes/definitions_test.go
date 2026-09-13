package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BumpyClock/hermes"
)

func TestDefinitionsFailureBeforeRequests(t *testing.T) {
	old := definitionsDirectory
	t.Cleanup(func() { definitionsDirectory = old })
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	defer server.Close()
	for _, source := range []string{"", filepath.Join(t.TempDir(), "missing"), t.TempDir()} {
		definitionsDirectory = source
		cmd := &cobra.Command{}
		cmd.Flags().String("definitions", "", "")
		if err := cmd.Flags().Set("definitions", source); err != nil {
			t.Fatal(err)
		}
		err := runParse(cmd, []string{server.URL})
		var e *hermes.DefinitionError
		if !errors.As(err, &e) {
			t.Fatalf("expected configuration failure before URL validation: %v", err)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("configuration failure performed requests")
	}
}

func TestDefinitionsCLIProcess(t *testing.T) {
	//nolint:gosec // Re-execute this test binary to verify the CLI's exit and output channels.
	cmd := exec.Command(os.Args[0], "-test.run=^TestDefinitionsCLIEntrypoint$")
	cmd.Env = append(os.Environ(), "HERMES_TEST_DEFINITION_CLI=1")
	out, err := cmd.Output()
	if err == nil {
		t.Fatal("expected nonzero configuration failure")
	}
	if len(out) != 0 {
		t.Fatalf("configuration error contaminated stdout: %s", out)
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || len(exit.Stderr) == 0 {
		t.Fatalf("missing stderr diagnostics: %v", err)
	}
}

func TestDefinitionsCLIEntrypoint(t *testing.T) {
	if os.Getenv("HERMES_TEST_DEFINITION_CLI") != "1" {
		t.Skip("subprocess entrypoint")
	}
	os.Args = []string{os.Args[0], "parse", "--definitions", "", "https://example.com"}
	main()
}
