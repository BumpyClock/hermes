// definition-conformance is an offline candidate test executable, not a production CLI.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync/atomic"
	"time"

	hermes "github.com/BumpyClock/hermes"
	"github.com/BumpyClock/hermes/internal/definitionbundle"
	"github.com/BumpyClock/hermes/internal/definitionbundle/offline"
)

var (
	buildID        string
	sourceRevision string
	sourceSHA256   string
)

type identity struct {
	Protocol         int      `json:"protocol"`
	BuildID          string   `json:"build_id"`
	SourceRevision   string   `json:"source_revision"`
	SourceSHA256     string   `json:"source_sha256"`
	Boundary         string   `json:"boundary"`
	ExecutableSHA256 string   `json:"executable_sha256"`
	Schema           int      `json:"schema"`
	Operations       []string `json:"operations"`
	Algorithms       []string `json:"algorithms"`
}

type caseResult struct {
	ID         string   `json:"id"`
	Definition string   `json:"definition"`
	Fixture    string   `json:"fixture"`
	Format     string   `json:"format"`
	Failures   []string `json:"failures"`
}

type report struct {
	Protocol           int          `json:"protocol"`
	Engine             identity     `json:"engine"`
	ManifestSHA256     string       `json:"manifest_sha256"`
	ArchiveSHA256      string       `json:"archive_sha256"`
	DefinitionsVersion string       `json:"definitions_version"`
	Scope              string       `json:"scope"`
	Passed             bool         `json:"passed"`
	ProductionEligible bool         `json:"production_eligible"`
	HTTPOperations     int64        `json:"http_operations"`
	OfflineLookups     int          `json:"offline_lookups"`
	Cases              []caseResult `json:"cases"`
	Errors             []string     `json:"errors"`
}

type denyTransport struct{ calls atomic.Int64 }

func (d *denyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	d.calls.Add(1)
	return nil, errors.New("offline conformance: HTTP denied")
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func describe() (identity, error) {
	support := hermes.DefinitionCapabilities()
	slices.Sort(support.Capabilities)
	slices.Sort(support.Algorithms)
	result := identity{
		Protocol: definitionbundle.Protocol, BuildID: buildID, SourceRevision: sourceRevision,
		SourceSHA256: sourceSHA256, Schema: support.Schema,
		Operations: support.Capabilities, Algorithms: support.Algorithms,
	}
	if !offline.Installed || len(buildID) != 64 || len(sourceSHA256) != 64 || len(sourceRevision) != 40 {
		return result, errors.New("not a prepared offline candidate; use scripts/bundle/prepare_engine.py with explicit immutable source")
	}
	result.Boundary = "fixture-resolver-v1"
	executable, err := os.Executable()
	if err != nil {
		return result, err
	}
	data, err := definitionbundle.ReadFile(executable, 512<<20)
	if err != nil {
		return result, err
	}
	result.ExecutableSHA256 = definitionbundle.Digest(data)
	return result, nil
}

func run(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("definition-conformance", flag.ContinueOnError)
	describeOnly := flags.Bool("describe", false, "report this explicitly prepared engine identity")
	manifestPath := flags.String("manifest", "", "local manifest path")
	manifestPin := flags.String("manifest-sha256", "", "required pinned manifest digest")
	archivePath := flags.String("archive", "", "local archive path")
	archivePin := flags.String("archive-sha256", "", "required pinned archive digest")
	enginePin := flags.String("engine-sha256", "", "required selected executable digest")
	buildPin := flags.String("engine-build-id", "", "required selected build identity")
	work := flags.String("work", "", "new local directory for validated definitions (must not exist)")
	gatePath := flags.String("release-gate", "", "optional separately approved complete-release gate")
	gatePin := flags.String("release-gate-sha256", "", "required digest when a release gate is supplied")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("positional arguments are unsupported")
	}
	engine, err := describe()
	if err != nil {
		return err
	}
	if *describeOnly {
		if len(args) != 1 {
			return errors.New("-describe cannot be combined with validation arguments")
		}
		return json.NewEncoder(output).Encode(engine)
	}
	result := report{Protocol: definitionbundle.Protocol, Engine: engine,
		ManifestSHA256: *manifestPin, ArchiveSHA256: *archivePin,
		Cases: []caseResult{}, Errors: []string{}}
	err = evaluate(&result, *manifestPath, *archivePath, *enginePin, *buildPin, *work, *gatePath, *gatePin)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
	}
	result.Passed = err == nil
	result.ProductionEligible = result.Passed && result.Scope == "complete"
	if encodeErr := json.NewEncoder(output).Encode(result); encodeErr != nil {
		return encodeErr
	}
	return err
}

func evaluate(result *report, manifestPath, archivePath, enginePin, buildPin, work, gatePath, gatePin string) error {
	if enginePin != result.Engine.ExecutableSHA256 || buildPin != result.Engine.BuildID {
		return errors.New("selected candidate engine digest/build identity mismatch")
	}
	if work == "" || manifestPath == "" || archivePath == "" {
		return errors.New("explicit manifest, archive and new work directory are required")
	}
	manifestData, err := definitionbundle.ReadFile(manifestPath, definitionbundle.MaxManifestSize)
	if err != nil {
		return err
	}
	if definitionbundle.Digest(manifestData) != result.ManifestSHA256 {
		return errors.New("pinned manifest digest mismatch")
	}
	manifest, err := definitionbundle.ParseManifest(manifestData)
	if err != nil {
		return err
	}
	result.DefinitionsVersion, result.Scope = manifest.Version, manifest.Scope
	if err = manifest.CheckSupport(result.Engine.Schema, result.Engine.Operations, result.Engine.Algorithms); err != nil {
		return err
	}
	archiveData, err := definitionbundle.ReadFile(archivePath, definitionbundle.MaxArchiveSize)
	if err != nil {
		return err
	}
	files, err := manifest.ReadArchive(archiveData, result.ArchiveSHA256)
	if err != nil {
		return err
	}
	suite, err := definitionbundle.ParseSuite(files)
	if err != nil {
		return err
	}
	var gate *definitionbundle.ReleaseGate
	if gatePath != "" {
		data, readErr := definitionbundle.ReadFile(gatePath, definitionbundle.MaxManifestSize)
		if readErr != nil {
			return readErr
		}
		if definitionbundle.Digest(data) != gatePin {
			return errors.New("approved release gate digest mismatch")
		}
		gate = &definitionbundle.ReleaseGate{}
		if err = definitionbundle.DecodeJSON(data, gate); err != nil {
			return err
		}
	} else if gatePin != "" {
		return errors.New("release gate pin supplied without gate")
	}
	if err = manifest.CheckCoverage(files, suite, gate); err != nil {
		return err
	}
	if err = definitionbundle.WriteDefinitions(files, work); err != nil {
		return err
	}
	snapshot, err := hermes.LoadDefinitions(work)
	if err != nil {
		return fmt.Errorf("real engine loader: %w", err)
	}
	if err = manifest.CheckDeclaredCapabilities(snapshot.UsedCapabilities()); err != nil {
		return err
	}
	for _, c := range suite.Cases {
		want, siteErr := definitionbundle.DefinitionSite(files[c.Definition])
		if siteErr != nil {
			return fmt.Errorf("case %s definition %q: %w", c.ID, c.Definition, siteErr)
		}
		u, _ := url.Parse(c.URL)
		if got, ok := snapshot.Site(u.Hostname()); !ok || got != want {
			return fmt.Errorf("case %s: URL does not select declared definition %q", c.ID, c.Definition)
		}
	}
	return runCases(result, snapshot, files, suite)
}

func runCases(result *report, snapshot *hermes.Definitions, files map[string][]byte, suite *definitionbundle.Suite) error {
	hosts := make([]string, 0, len(suite.Cases))
	for _, c := range suite.Cases {
		u, _ := url.Parse(c.URL)
		hosts = append(hosts, u.Hostname())
	}
	offline.Configure(hosts)
	transport := &denyTransport{}
	oldTransport := http.DefaultTransport
	http.DefaultTransport = transport
	defer func() { http.DefaultTransport = oldTransport }()
	failed := 0
	for _, c := range suite.Cases {
		for _, format := range c.Formats {
			client := hermes.New(hermes.WithDefinitions(snapshot), hermes.WithTransport(transport), hermes.WithContentType(format))
			row := caseResult{ID: c.ID, Definition: c.Definition, Fixture: c.Fixture, Format: format, Failures: []string{}}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			parsed, err := client.ParseHTML(ctx, string(files[c.Fixture]), c.URL)
			cancel()
			if err != nil {
				row.Failures = append(row.Failures, err.Error())
			} else {
				fields := map[string]string{"url": parsed.URL, "title": parsed.Title, "content": parsed.Content,
					"author": parsed.Author, "lead_image_url": parsed.LeadImageURL, "domain": parsed.Domain}
				if parsed.DatePublished != nil {
					fields["date_published"] = parsed.DatePublished.Format("2006-01-02")
				}
				row.Failures = append(row.Failures, c.Compare(fields, format)...)
			}
			if len(row.Failures) > 0 {
				failed++
			}
			result.Cases = append(result.Cases, row)
		}
	}
	result.HTTPOperations, result.OfflineLookups = transport.calls.Load(), offline.Calls()
	if result.HTTPOperations != 0 {
		return fmt.Errorf("fixture attempted %d HTTP operations", result.HTTPOperations)
	}
	if failed > 0 {
		return fmt.Errorf("%d/%d conformance case/format runs failed", failed, len(result.Cases))
	}
	return nil
}
