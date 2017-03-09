package meters

import (
	"github.com/rackspace/gophercloud"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToMeterListQuery() (string, error)
}

// ListOpts allows the filtering and sorting of collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the server attributes you want to see returned.
type ListOpts struct {
	QueryField string `q:"q.field"`
	QueryOp    string `q:"q.op"`
	QueryValue string `q:"q.value"`

	// ID of the last-seen item from the previous response
	Marker string `q:"marker"`

	// Optional, maximum number of results to return
	Limit int `q:"limit"`
}

// ToMeterListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToMeterListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return q.String(), nil
}

// List makes a request against the API to list meters accessible to you.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) ListResult {
	var res ListResult
	url := listURL(client)

	if opts != nil {
		query, err := opts.ToMeterListQuery()
		if err != nil {
			res.Err = err
			return res
		}
		url += query
	}

	_, res.Err = client.Get(url, &res.Body, &gophercloud.RequestOpts{})
	return res
}

// ShowOptsBuilder allows extensions to add additional parameters to the
// Show request.
type ShowOptsBuilder interface {
	ToShowQuery() (string, error)
}

// ShowOpts allows the filtering and sorting of collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the server attributes you want to see returned.
type ShowOpts struct {
	QueryField string `q:"q.field"`
	QueryOp    string `q:"q.op"`
	QueryValue string `q:"q.value"`

	// ID of the last-seen item from the previous response
	Marker string `q:"marker"`

	// Optional, maximum number of results to return
	Limit int `q:"limit"`
}

// ToMeterShowQuery formats a ShowOpts into a query string.
func (opts ShowOpts) ToShowQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return q.String(), nil
}

// Show makes a request against the API to show a specific meter
func Show(client *gophercloud.ServiceClient, meterName string, opts ShowOptsBuilder) ShowResult {
	var res ShowResult
	url := showURL(client, meterName)

	if opts != nil {
		query, err := opts.ToShowQuery()
		if err != nil {
			res.Err = err
			return res
		}
		url += query
	}

	_, res.Err = client.Get(url, &res.Body, &gophercloud.RequestOpts{})
	return res
}

// StatisticsOptsBuilder allows extensions to add additional parameters to the
// List request.
type MeterStatisticsOptsBuilder interface {
	ToMeterStatisticsQuery() (string, error)
}

// StatisticsOpts allows the filtering and sorting of collections through
// the API. Filtering is achieved by passing in struct field values that map to
// the server attributes you want to see returned.
type MeterStatisticsOpts struct {
	QueryField string `q:"q.field"`
	QueryOp    string `q:"q.op"`
	QueryValue string `q:"q.value"`

	// Optional group by
	GroupBy string `q:"groupby"`

	// Optional number of seconds in a period
	Period int `q:"period"`
}

// ToStatisticsQuery formats a StatisticsOpts into a query string.
func (opts MeterStatisticsOpts) ToMeterStatisticsQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return q.String(), nil
}

// List makes a request against the API to list meters accessible to you.
func MeterStatistics(client *gophercloud.ServiceClient, n string, opts MeterStatisticsOptsBuilder) StatisticsResult {
	var res StatisticsResult
	url := statisticsURL(client, n)

	if opts != nil {
		query, err := opts.ToMeterStatisticsQuery()
		if err != nil {
			res.Err = err
			return res
		}
		url += query
	}

	_, res.Err = client.Get(url, &res.Body, &gophercloud.RequestOpts{})
	return res
}
