package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BumpyClock/hermes"
)

func TestManagedDefinitionFlagsRejectConflictingSourcesBeforeRequests(t *testing.T) {
	oldLocal, oldManaged, oldAutomatic, oldCache := definitionsDirectory, managedDefinitionsVersion, managedDefinitionsAuto, definitionsCacheDirectory
	t.Cleanup(func() {
		definitionsDirectory, managedDefinitionsVersion, managedDefinitionsAuto, definitionsCacheDirectory = oldLocal, oldManaged, oldAutomatic, oldCache
	})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests.Add(1) }))
	defer server.Close()

	definitionsDirectory, managedDefinitionsVersion = "local", "synthetic-1"
	cmd := managedCommand(t, map[string]string{"definitions": "local", "managed-definitions": "synthetic-1"})
	err := runParse(cmd, []string{server.URL})
	if err == nil || err.Error() != "--definitions and --managed-definitions cannot be used together" {
		t.Fatalf("conflicting source error = %v", err)
	}

	definitionsDirectory, managedDefinitionsVersion, definitionsCacheDirectory = "", "", "cache"
	cmd = managedCommand(t, map[string]string{"definitions-cache": "cache"})
	err = runParse(cmd, []string{server.URL})
	if err == nil || err.Error() != "--definitions-cache requires --managed-definitions" {
		t.Fatalf("cache-only error = %v", err)
	}
	managedDefinitionsVersion, managedDefinitionsAuto, definitionsCacheDirectory = "synthetic-1", true, ""
	cmd = managedCommand(t, map[string]string{"managed-definitions": "synthetic-1", "managed-definitions-auto": "true"})
	err = runParse(cmd, []string{server.URL})
	if err == nil || err.Error() != "--managed-definitions and --managed-definitions-auto cannot be used together" {
		t.Fatalf("conflicting managed mode error = %v", err)
	}
	if requests.Load() != 0 {
		t.Fatal("invalid managed CLI flags made article requests")
	}
}

func TestManagedDefinitionsFailureAndWarningPrecedeArticleWorkers(t *testing.T) {
	oldLocal, oldManaged, oldAutomatic, oldCache, oldLoader, oldStderr, oldConcurrency := definitionsDirectory, managedDefinitionsVersion, managedDefinitionsAuto, definitionsCacheDirectory, loadManagedDefinitions, os.Stderr, concurrency
	t.Cleanup(func() {
		definitionsDirectory, managedDefinitionsVersion, managedDefinitionsAuto, definitionsCacheDirectory, loadManagedDefinitions, os.Stderr, concurrency = oldLocal, oldManaged, oldAutomatic, oldCache, oldLoader, oldStderr, oldConcurrency
	})
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests.Add(1) }))
	defer server.Close()
	managedDefinitionsVersion, managedDefinitionsAuto, definitionsCacheDirectory = "synthetic-1", false, "cache"
	concurrency = 1
	cmd := managedCommand(t, map[string]string{"managed-definitions": "synthetic-1", "definitions-cache": "cache"})

	loadManagedDefinitions = func(context.Context, hermes.ManagedDefinitionsOptions) (*hermes.ManagedDefinitions, error) {
		return nil, errors.New("managed setup failed")
	}
	if err := runParse(cmd, []string{server.URL}); err == nil || err.Error() != "managed setup failed" {
		t.Fatalf("managed setup error = %v", err)
	}
	if requests.Load() != 0 {
		t.Fatal("managed setup failure made article requests")
	}

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = write
	loadManagedDefinitions = func(context.Context, hermes.ManagedDefinitionsOptions) (*hermes.ManagedDefinitions, error) {
		return &hermes.ManagedDefinitions{
			Snapshot: &hermes.Definitions{}, Version: "synthetic-1",
			Warning: errors.New("unsafe managed definitions redirect to untrusted.example/private"),
		}, nil
	}
	err = runParse(cmd, []string{""})
	_ = write.Close()
	stderr, readErr := io.ReadAll(read)
	_ = read.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	const secret = "signed-token-must-not-leak"
	if err == nil || !strings.Contains(string(stderr), "Warning: unsafe managed definitions redirect to untrusted.example/private") ||
		strings.Contains(string(stderr), secret) || requests.Load() != 0 {
		t.Fatalf("warning stderr=%q err=%v requests=%d", stderr, err, requests.Load())
	}
}

func managedCommand(t *testing.T, values map[string]string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("definitions", "", "")
	cmd.Flags().String("managed-definitions", "", "")
	cmd.Flags().Bool("managed-definitions-auto", false, "")
	cmd.Flags().String("definitions-cache", "", "")
	for name, value := range values {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	return cmd
}
