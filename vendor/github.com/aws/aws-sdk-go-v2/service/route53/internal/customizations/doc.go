// Package customizations provides customizations for the Amazon Route53 API client.
//
// This package provides support for following customizations
//
//	 Process Response Middleware: used for custom error deserializing
//		Sanitize URL Middleware: used for sanitizing url with HostedZoneID member
//
// # Process Response Middleware
//
// Route53 operation "ChangeResourceRecordSets" can have an error response returned in
// a slightly different format. This customization is only applicable to
// ChangeResourceRecordSets operation of Route53.
//
// Here's a sample error response:
//
//	<?xml version="1.0" encoding="UTF-8"?>
//	<InvalidChangeBatch xmlns="https://route53.amazonaws.com/doc/2013-04-01/">
//	    <Messages>
//	        <Message>Tried to create resource record set duplicate.example.com. type A, but it already exists</Message>
//	    </Messages>
//	</InvalidChangeBatch>
//
// The processResponse middleware customizations enables SDK to check for an error
// response starting with "InvalidChangeBatch" tag prior to deserialization.
//
// As this check in error response needs to be performed earlier than response
// deserialization. Since the behavior of Deserialization is in
// reverse order to the other stack steps its easier to consider that "after" means
// "before".
//
//	Middleware layering:
//
//	HTTP Response -> process response error -> deserialize
//
// In case the returned error response has `InvalidChangeBatch` format, the error is
// deserialized and returned. The operation deserializer does not attempt to deserialize
// as an error is returned by the process response error middleware.
//
// # Sanitize URL Middleware
//
// Route53 operations may return a response containing an id member value appended with
// a string, for example. an id 1234 may be returned as 'foo/1234'. While round-tripping such response
// id value into another operation request, SDK must strip out the additional prefix if any.
// The Sanitize URL Middleware strips out such additionally prepended string to the id.
//
// The Id member with such prepended strings target shape 'ResourceId' or 'DelegationSetId'.
// This customization thus is applied only for operations with id's targeting those target shapes.
// This customization has to be applied before the input is serialized.
//
//	Middleware layering:
//
//	Input -> Sanitize URL Middleware -> serialize -> next
package customizations
