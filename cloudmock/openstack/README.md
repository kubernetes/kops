# Openstack Cloudmock

## Design

Because the gophercloud library does not provide client interfaces whose client-side functions could be mocked like aws-sdk-go, this cloudmock uses a local HTTP server and updates state based on incoming requests from the gophercloud clients.
This is how the [gophercloud library tests](https://github.com/gophercloud/gophercloud/blob/51f8fa152459ae60d3b348023ad79f850db3a931/openstack/compute/v2/servers/testing/fixtures.go#L896-L914) themselves are implemented.

Each package represents one of the Openstack service clients and contains its own `net/http/httptest` server.
Each package defines the endpoints for that client's resources.

## Troubleshooting

One recommended way to troubleshoot requests and responses is with Wireshark or an equivalent, monitoring the loopback interface.