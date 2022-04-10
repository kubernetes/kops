package instrumentation

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Instrumenter struct {
	subsystemIdentifier     string // will be used as part of the metric name (hcloud_<identifier>_requests_total)
	instrumentationRegistry *prometheus.Registry
}

// New creates a new Instrumenter. The subsystemIdentifier will be used as part of the metric names (e.g. hcloud_<identifier>_requests_total)
func New(subsystemIdentifier string, instrumentationRegistry *prometheus.Registry) *Instrumenter {
	return &Instrumenter{subsystemIdentifier: subsystemIdentifier, instrumentationRegistry: instrumentationRegistry}
}

// InstrumentedRoundTripper returns an instrumented round tripper.
func (i *Instrumenter) InstrumentedRoundTripper() http.RoundTripper {
	inFlightRequestsGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: fmt.Sprintf("hcloud_%s_in_flight_requests", i.subsystemIdentifier),
		Help: fmt.Sprintf("A gauge of in-flight requests to the hcloud %s.", i.subsystemIdentifier),
	})

	requestsPerEndpointCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: fmt.Sprintf("hcloud_%s_requests_total", i.subsystemIdentifier),
			Help: fmt.Sprintf("A counter for requests to the hcloud %s per endpoint.", i.subsystemIdentifier),
		},
		[]string{"code", "method", "api_endpoint"},
	)

	requestLatencyHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    fmt.Sprintf("hcloud_%s_request_duration_seconds", i.subsystemIdentifier),
			Help:    fmt.Sprintf("A histogram of request latencies to the hcloud %s .", i.subsystemIdentifier),
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	i.instrumentationRegistry.MustRegister(requestsPerEndpointCounter, requestLatencyHistogram, inFlightRequestsGauge)

	return promhttp.InstrumentRoundTripperInFlight(inFlightRequestsGauge,
		promhttp.InstrumentRoundTripperDuration(requestLatencyHistogram,
			i.instrumentRoundTripperEndpoint(requestsPerEndpointCounter,
				http.DefaultTransport,
			),
		),
	)
}

// instrumentRoundTripperEndpoint implements a hcloud specific round tripper to count requests per API endpoint
// numeric IDs are removed from the URI Path.
// Sample:
// /volumes/1234/actions/attach --> /volumes/actions/attach
func (i *Instrumenter) instrumentRoundTripperEndpoint(counter *prometheus.CounterVec, next http.RoundTripper) promhttp.RoundTripperFunc {
	return func(r *http.Request) (*http.Response, error) {
		resp, err := next.RoundTrip(r)
		if err == nil {
			statusCode := strconv.Itoa(resp.StatusCode)
			counter.WithLabelValues(statusCode, strings.ToLower(resp.Request.Method), preparePathForLabel(resp.Request.URL.Path)).Inc()
		}
		return resp, err
	}
}

func preparePathForLabel(path string) string {
	path = strings.ToLower(path)

	// replace all numbers and chars that are not a-z, / or _
	reg := regexp.MustCompile("[^a-z/_]+")
	path = reg.ReplaceAllString(path, "")

	// replace all artifacts of number replacement (//)
	path = strings.ReplaceAll(path, "//", "/")

	// replace the /v/ that indicated the API version
	return strings.Replace(path, "/v/", "/", 1)
}
