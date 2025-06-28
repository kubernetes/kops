package instrumentation

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/hetznercloud/hcloud-go/v2/hcloud/exp/ctxutil"
)

type Instrumenter struct {
	subsystemIdentifier     string // will be used as part of the metric name (hcloud_<identifier>_requests_total)
	instrumentationRegistry prometheus.Registerer
}

// New creates a new Instrumenter. The subsystemIdentifier will be used as part of the metric names (e.g. hcloud_<identifier>_requests_total).
func New(subsystemIdentifier string, instrumentationRegistry prometheus.Registerer) *Instrumenter {
	return &Instrumenter{subsystemIdentifier: subsystemIdentifier, instrumentationRegistry: instrumentationRegistry}
}

// InstrumentedRoundTripper returns an instrumented round tripper.
func (i *Instrumenter) InstrumentedRoundTripper(transport http.RoundTripper) http.RoundTripper {
	// By default, http client would use DefaultTransport on nil, but we internally are relying on it being configured
	if transport == nil {
		transport = http.DefaultTransport
	}

	inFlightRequestsGauge := registerOrReuse(
		i.instrumentationRegistry,
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: fmt.Sprintf("hcloud_%s_in_flight_requests", i.subsystemIdentifier),
			Help: fmt.Sprintf("A gauge of in-flight requests to the hcloud %s.", i.subsystemIdentifier),
		}),
	)

	requestsPerEndpointCounter := registerOrReuse(
		i.instrumentationRegistry,
		prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("hcloud_%s_requests_total", i.subsystemIdentifier),
				Help: fmt.Sprintf("A counter for requests to the hcloud %s per endpoint.", i.subsystemIdentifier),
			},
			[]string{"code", "method", "api_endpoint"},
		),
	)

	requestLatencyHistogram := registerOrReuse(
		i.instrumentationRegistry,
		prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    fmt.Sprintf("hcloud_%s_request_duration_seconds", i.subsystemIdentifier),
				Help:    fmt.Sprintf("A histogram of request latencies to the hcloud %s .", i.subsystemIdentifier),
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
	)

	return promhttp.InstrumentRoundTripperInFlight(inFlightRequestsGauge,
		promhttp.InstrumentRoundTripperDuration(requestLatencyHistogram,
			i.instrumentRoundTripperEndpoint(requestsPerEndpointCounter,
				transport,
			),
		),
	)
}

// instrumentRoundTripperEndpoint implements a hcloud specific round tripper to count requests per API endpoint
// numeric IDs are removed from the URI Path.
//
// Sample:
//
//	/volumes/1234/actions/attach --> /volumes/actions/attach
func (i *Instrumenter) instrumentRoundTripperEndpoint(counter *prometheus.CounterVec, next http.RoundTripper) promhttp.RoundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(r)
		if err == nil {
			apiEndpoint := ctxutil.OpPath(r.Context())
			// If the request does not set the operation path, we must construct it. Happens e.g. for
			// user crafted requests.
			if apiEndpoint == "" {
				apiEndpoint = preparePathForLabel(resp.Request.URL.Path)
			}
			counter.WithLabelValues(
				strconv.Itoa(resp.StatusCode),
				strings.ToLower(resp.Request.Method),
				apiEndpoint,
			).Inc()
		}

		return resp, err
	}
}

// registerOrReuse will try to register the passed Collector, but in case a conflicting collector was already registered,
// it will instead return that collector. Make sure to always use the collector return by this method.
// Similar to [Registry.MustRegister] it will panic if any other error occurs.
func registerOrReuse[C prometheus.Collector](registry prometheus.Registerer, collector C) C {
	err := registry.Register(collector)
	if err != nil {
		var arErr prometheus.AlreadyRegisteredError
		// If we get a AlreadyRegisteredError we can return the existing collector
		if errors.As(err, &arErr) {
			if existingCollector, ok := arErr.ExistingCollector.(C); ok {
				collector = existingCollector
			} else {
				panic("received incompatible existing collector")
			}
		} else {
			panic(err)
		}
	}

	return collector
}

var pathLabelRegexp = regexp.MustCompile("[^a-z/_]+")

func preparePathForLabel(path string) string {
	path = strings.ToLower(path)

	// replace the /v1/ that indicated the API version
	path, _ = strings.CutPrefix(path, "/v1")

	// replace all numbers and chars that are not a-z, / or _
	path = pathLabelRegexp.ReplaceAllString(path, "-")

	return path
}
