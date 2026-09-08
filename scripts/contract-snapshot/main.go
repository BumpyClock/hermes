// Command contract-snapshot records public parser results for refactor comparisons.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	hermes "github.com/BumpyClock/hermes"
)

type transport func(*http.Request) (*http.Response, error)

func (f transport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type observation struct {
	ID           string            `json:"id"`
	Result       *hermes.Result    `json:"result"`
	Error        string            `json:"error,omitempty"`
	Code         *hermes.ErrorCode `json:"code,omitempty"`
	Operation    string            `json:"operation,omitempty"`
	URL          string            `json:"error_url,omitempty"`
	Cause        string            `json:"cause,omitempty"`
	ErrorChain   []string          `json:"error_chain,omitempty"`
	ErrorHelpers []bool            `json:"error_helpers,omitempty"`
	Markdown     string            `json:"markdown,omitempty"`
	Helpers      []bool            `json:"helpers,omitempty"`
}

func record(enc *json.Encoder, id string, result *hermes.Result, err error) {
	row := observation{ID: id, Result: result}
	if err != nil {
		row.Error = err.Error()
		for cause := err; cause != nil; cause = errors.Unwrap(cause) {
			row.ErrorChain = append(row.ErrorChain, fmt.Sprintf("%T: %s", cause, cause))
		}
		var parseError *hermes.ParseError
		if errors.As(err, &parseError) {
			row.Code, row.Operation, row.URL = &parseError.Code, parseError.Op, parseError.URL
			row.ErrorHelpers = []bool{parseError.IsTimeout(), parseError.IsSSRF(), parseError.IsFetch(), parseError.IsExtract(), parseError.IsInvalidURL(), parseError.IsContext(), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded), errors.Is(err, &hermes.ParseError{Code: parseError.Code})}
			if parseError.Err != nil {
				row.Cause = parseError.Err.Error()
			}
		}
	}
	if result != nil {
		row.Markdown = result.FormatMarkdown()
		row.Helpers = []bool{result.IsEmpty(), result.HasAuthor(), result.HasDate(), result.HasImage()}
	}
	if err := enc.Encode(row); err != nil {
		panic(err)
	}
}

func response(body, contentType string, status int) transport {
	return func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)), Header: http.Header{"Content-Type": {contentType}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	}
}

type failedBody struct{}

func (failedBody) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
func (failedBody) Close() error             { return nil }

func main() {
	fixtures := flag.String("fixtures", "", "fixture directory")
	pattern := flag.String("match", "*.html", "fixture filename glob")
	contracts := flag.Bool("contracts", false, "record deterministic error contracts")
	flag.Parse()
	enc := json.NewEncoder(os.Stdout)
	if *contracts {
		snapshotContracts(enc)
		return
	}
	paths, err := filepath.Glob(filepath.Join(*fixtures, *pattern))
	if err != nil || len(paths) == 0 {
		panic("no HTML fixtures")
	}
	for _, path := range paths {
		body, err := os.ReadFile(path) //nolint:gosec // The caller selects the local fixture directory.
		if err != nil {
			panic(err)
		}
		name := filepath.Base(path)
		host := strings.SplitN(strings.TrimSuffix(name, ".html"), "--", 2)[0]
		for _, route := range []string{host, "93.184.216.34"} {
			for _, format := range []string{"html", "markdown", "text"} {
				client := hermes.New(hermes.WithContentType(format), hermes.WithTransport(response(string(body), "text/html; charset=utf-8", http.StatusOK)))
				url := "https://" + route + "/contract/article"
				result, err := client.ParseHTML(context.Background(), string(body), url)
				record(enc, name+"/"+route+"/"+format+"/html", result, err)
				result, err = client.Parse(context.Background(), url)
				record(enc, name+"/"+route+"/"+format+"/fetch", result, err)
			}
		}
	}
}

func snapshotContracts(enc *json.Encoder) {
	const body = `<html><head><title>A contract example article</title></head><body><article><h1>A contract example article</h1><p>This article has enough words to test the public parser output and its result helpers.</p></article></body></html>`
	const url = "https://93.184.216.34/article"
	for _, target := range []string{"", "://bad", "ftp://93.184.216.34/a", "https://127.0.0.1/a", "https://10.0.0.1/a", "https://[::1]/a", "https://"} {
		client := hermes.New(hermes.WithTransport(response(body, "text/html", 200)))
		r, e := client.Parse(context.Background(), target)
		record(enc, "url/fetch/"+target, r, e)
		r, e = client.ParseHTML(context.Background(), body, target)
		record(enc, "url/html/"+target, r, e)
	}
	for _, format := range []string{"", "html", "markdown", "text", "md", "txt", "text/plain", "text/markdown", "text/html", " HTML ", "unknown"} {
		client := hermes.New(hermes.WithContentType(format), hermes.WithTransport(response(body, "text/html", 200)))
		r, e := client.ParseHTML(context.Background(), body, url)
		record(enc, "format/"+format, r, e)
	}
	for _, status := range []int{200, 204, 400, 404} {
		client := hermes.New(hermes.WithTransport(response(body, "text/html", status)))
		r, e := client.Parse(context.Background(), url)
		record(enc, fmt.Sprintf("status/%d", status), r, e)
	}
	for _, mime := range []string{"text/html", "text/plain", "application/json", "image/png"} {
		client := hermes.New(hermes.WithTransport(response(body, mime, 200)))
		r, e := client.Parse(context.Background(), url)
		record(enc, "mime/"+mime, r, e)
	}
	for _, failure := range []string{"transport", "body"} {
		attempts := 0
		client := hermes.New(hermes.WithTransport(transport(func(req *http.Request) (*http.Response, error) {
			attempts++
			if failure == "transport" {
				return nil, io.ErrUnexpectedEOF
			}
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/html"}}, Body: failedBody{}, Request: req}, nil
		})))
		r, e := client.Parse(context.Background(), url)
		record(enc, fmt.Sprintf("retry/%s/attempts=%d", failure, attempts), r, e)
		if attempts != 4 {
			panic("retry attempt contract changed")
		}
	}
	client := hermes.New(hermes.WithTransport(response(body, "text/html", 200)))
	r, e := client.ParseHTML(context.Background(), "", url)
	record(enc, "empty-html", r, e)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, e = client.Parse(ctx, url)
	record(enc, "canceled-fetch", r, e)
	r, e = client.ParseHTML(ctx, body, url)
	record(enc, "canceled-html", r, e)
}
