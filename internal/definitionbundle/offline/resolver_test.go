package offline

import (
	"context"
	"testing"
)

func TestResolverHasNoNetworkFallback(t *testing.T) {
	Configure([]string{"fixture.example"})
	resolver := Resolver{}
	addresses, err := resolver.LookupIPAddr(context.Background(), "fixture.example")
	if err != nil || len(addresses) != 1 || addresses[0].IP.String() != "93.184.216.34" {
		t.Fatalf("allowlisted fixture failed: %v %v", addresses, err)
	}
	if _, err = resolver.LookupIPAddr(context.Background(), "unplanned.example"); err == nil {
		t.Fatal("unplanned DNS name accepted")
	}
	addresses, err = resolver.LookupIPAddr(context.Background(), "127.0.0.1")
	if err != nil || addresses[0].IP.String() != "127.0.0.1" {
		t.Fatal("private literal rewritten, bypassing production SSRF checks")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := resolver.LookupIPAddr(ctx, "fixture.example"); err != context.Canceled {
		t.Fatalf("context not preserved: %v", err)
	}
	if Calls() != 3 {
		t.Fatalf("unexpected offline lookup count: %d", Calls())
	}
}
