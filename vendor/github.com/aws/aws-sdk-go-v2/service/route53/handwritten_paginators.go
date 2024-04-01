package route53

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// ListResourceRecordSetsAPIClient is a client that implements the ListResourceRecordSets
// operation
type ListResourceRecordSetsAPIClient interface {
	ListResourceRecordSets(context.Context, *ListResourceRecordSetsInput, ...func(*Options)) (*ListResourceRecordSetsOutput, error)
}

var _ ListResourceRecordSetsAPIClient = (*Client)(nil)

// ListResourceRecordSetsPaginatorOptions is the paginator options for ListResourceRecordSets
type ListResourceRecordSetsPaginatorOptions struct {
	// (Optional) The maximum number of ResourceRecordSets that you want Amazon Route 53 to
	// return.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListResourceRecordSetsPaginator is a paginator for ListResourceRecordSets
type ListResourceRecordSetsPaginator struct {
	options               ListResourceRecordSetsPaginatorOptions
	client                ListResourceRecordSetsAPIClient
	params                *ListResourceRecordSetsInput
	firstPage             bool
	startRecordName       *string
	startRecordType       types.RRType
	startRecordIdentifier *string
	isTruncated           bool
}

// NewListResourceRecordSetsPaginator returns a new ListResourceRecordSetsPaginator
func NewListResourceRecordSetsPaginator(client ListResourceRecordSetsAPIClient, params *ListResourceRecordSetsInput, optFns ...func(*ListResourceRecordSetsPaginatorOptions)) *ListResourceRecordSetsPaginator {
	if params == nil {
		params = &ListResourceRecordSetsInput{}
	}

	options := ListResourceRecordSetsPaginatorOptions{}
	if params.MaxItems != nil {
		options.Limit = *params.MaxItems
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListResourceRecordSetsPaginator{
		options:               options,
		client:                client,
		params:                params,
		firstPage:             true,
		startRecordName:       params.StartRecordName,
		startRecordType:       params.StartRecordType,
		startRecordIdentifier: params.StartRecordIdentifier,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListResourceRecordSetsPaginator) HasMorePages() bool {
	return p.firstPage || p.isTruncated
}

// NextPage retrieves the next ListResourceRecordSets page.
func (p *ListResourceRecordSetsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListResourceRecordSetsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.StartRecordName = p.startRecordName
	params.StartRecordIdentifier = p.startRecordIdentifier
	params.StartRecordType = p.startRecordType

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxItems = limit

	result, err := p.client.ListResourceRecordSets(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.startRecordName
	p.isTruncated = result.IsTruncated
	p.startRecordName = nil
	p.startRecordIdentifier = nil
	p.startRecordType = ""
	if result.IsTruncated {
		p.startRecordName = result.NextRecordName
		p.startRecordIdentifier = result.NextRecordIdentifier
		p.startRecordType = result.NextRecordType
	}

	if p.options.StopOnDuplicateToken &&
		prevToken != nil &&
		p.startRecordName != nil &&
		*prevToken == *p.startRecordName {
		p.isTruncated = false
	}

	return result, nil
}
