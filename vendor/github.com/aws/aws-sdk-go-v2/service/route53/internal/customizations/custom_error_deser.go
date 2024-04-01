package customizations

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/aws/smithy-go"
	smithyxml "github.com/aws/smithy-go/encoding/xml"
	"github.com/aws/smithy-go/middleware"
	"github.com/aws/smithy-go/ptr"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	awsmiddle "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// HandleCustomErrorDeserialization check if Route53 response is an error and needs
// custom error deserialization.
func HandleCustomErrorDeserialization(stack *middleware.Stack) error {
	return stack.Deserialize.Insert(&processResponse{}, "OperationDeserializer", middleware.After)
}

// middleware to process raw response and look for error response with InvalidChangeBatch error tag
type processResponse struct{}

// ID returns the middleware ID.
func (*processResponse) ID() string {
	return "Route53:ProcessResponseForCustomErrorResponse"
}

func (m *processResponse) HandleDeserialize(
	ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler) (
	out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
) {
	out, metadata, err = next.HandleDeserialize(ctx, in)
	if err != nil {
		return out, metadata, err
	}

	response, ok := out.RawResponse.(*smithyhttp.Response)
	if !ok {
		return out, metadata, &smithy.DeserializationError{Err: fmt.Errorf("unknown transport type %T", out.RawResponse)}
	}

	// check if success response
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return
	}

	var readBuff bytes.Buffer
	body := io.TeeReader(response.Body, &readBuff)

	rootDecoder := xml.NewDecoder(body)
	t, err := smithyxml.FetchRootElement(rootDecoder)
	if err == io.EOF {
		return out, metadata, nil
	}

	// rewind response body
	response.Body = ioutil.NopCloser(io.MultiReader(&readBuff, response.Body))

	// if start tag is "InvalidChangeBatch", the error response needs custom unmarshaling.
	if strings.EqualFold(t.Name.Local, "InvalidChangeBatch") {
		return out, metadata, route53CustomErrorDeser(&metadata, response)
	}

	return out, metadata, err
}

// error type for invalidChangeBatchError
type invalidChangeBatchError struct {
	Messages  []string `xml:"Messages>Message"`
	RequestID string   `xml:"RequestId"`
}

func route53CustomErrorDeser(metadata *middleware.Metadata, response *smithyhttp.Response) error {
	err := invalidChangeBatchError{}
	xml.NewDecoder(response.Body).Decode(&err)

	// set request id in metadata
	if len(err.RequestID) != 0 {
		awsmiddle.SetRequestIDMetadata(metadata, err.RequestID)
	}

	return &types.InvalidChangeBatch{
		Message:  ptr.String("ChangeBatch errors occurred"),
		Messages: err.Messages,
	}
}
