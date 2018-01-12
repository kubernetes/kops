package option

import (
	"reflect"
	"testing"

	"google.golang.org/api/internal"
	"google.golang.org/grpc"
)

// Check that the slice passed into WithScopes is copied.
func TestCopyScopes(t *testing.T) {
	o := &internal.DialSettings{}

	scopes := []string{"a", "b"}
	WithScopes(scopes...).Apply(o)

	// Modify after using.
	scopes[1] = "c"

	if o.Scopes[0] != "a" || o.Scopes[1] != "b" {
		t.Errorf("want ['a', 'b'], got %+v", o.Scopes)
	}
}

func TestApply(t *testing.T) {
	conn := &grpc.ClientConn{}
	opts := []ClientOption{
		WithEndpoint("https://example.com:443"),
		WithScopes("a"), // the next WithScopes should overwrite this one
		WithScopes("https://example.com/auth/helloworld", "https://example.com/auth/otherthing"),
		WithGRPCConn(conn),
		WithUserAgent("ua"),
		WithServiceAccountFile("service-account.json"),
		WithAPIKey("api-key"),
	}
	var got internal.DialSettings
	for _, opt := range opts {
		opt.Apply(&got)
	}
	want := internal.DialSettings{
		Scopes:                     []string{"https://example.com/auth/helloworld", "https://example.com/auth/otherthing"},
		UserAgent:                  "ua",
		Endpoint:                   "https://example.com:443",
		GRPCConn:                   conn,
		ServiceAccountJSONFilename: "service-account.json",
		APIKey: "api-key",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot  %#v\nwant %#v", got, want)
	}
}
