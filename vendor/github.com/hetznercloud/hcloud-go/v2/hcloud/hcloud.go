/*
Package hcloud is a library for the Hetzner Cloud API.

The Hetzner Cloud API reference is available at https://docs.hetzner.cloud.

Make sure to follow our API changelog available at https://docs.hetzner.cloud/changelog
(or the RRS feed available at https://docs.hetzner.cloud/changelog/feed.rss) to be
notified about additions, deprecations and removals.

# Example

	package main

	import (
		"context"
		"fmt"
		"log"

		"github.com/hetznercloud/hcloud-go/v2/hcloud"
		"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/actionutil"
	)

	func main() {
		ctx := context.Background()

		client := hcloud.NewClient(
			hcloud.WithToken("token"),
			hcloud.WithApplication("my-tool", "v1.0.0"),
		)

		result, _, err := client.Server.Create(ctx, hcloud.ServerCreateOpts{
			Name:       "Foo",
			Image:      &hcloud.Image{Name: "ubuntu-24.0"},
			ServerType: &hcloud.ServerType{Name: "cpx22"},
			Location:   &hcloud.Location{Name: "hel1"},
		})
		if err != nil {
			log.Fatalf("error creating server: %s\n", err)
		}

		// Always await any returned actions, to make sure the async process is completed before you use the result:
		err = client.Action.WaitFor(ctx, actionutil.AppendNext(result.Action, result.NextActions)...)
		if err != nil {
			log.Fatalf("error creating server: %s\n", err)
		}

		server, _, err := client.Server.GetByID(ctx, result.Server.ID)
		if err != nil {
			log.Fatalf("error retrieving server: %s\n", err)
		}
		if server != nil {
			fmt.Printf("server is called %q\n", server.Name) // prints 'server is called "Foo"'
		} else {
			fmt.Println("server not found")
		}
	}

# Retry mechanism

The [Client.Do] method will retry failed requests that match certain criteria. The
default retry interval is defined by an exponential backoff algorithm truncated to 60s
with jitter. The default maximal number of retries is 5.

The following rules defines when a request can be retried:

When the [http.Client] returned a network timeout error.

When the API returned an HTTP error, with the status code:
  - [http.StatusBadGateway]
  - [http.StatusGatewayTimeout]

When the API returned an application error, with the code:
  - [ErrorCodeConflict]
  - [ErrorCodeRateLimitExceeded]
  - [ErrorCodeBadGateway]
  - [ErrorCodeTimeout]

Changes to the retry policy might occur between releases, and will not be considered
breaking changes.
*/
package hcloud

import "github.com/hetznercloud/hcloud-go/v2/hcloud/internal/version"

// Version is the library's version following Semantic Versioning.
const Version = version.Version
