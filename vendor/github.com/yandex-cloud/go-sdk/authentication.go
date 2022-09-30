// Copyright (c) 2021 YANDEX LLC.

package ycsdk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
	"github.com/yandex-cloud/go-sdk/pkg/singleflight"
)

type Authenticator interface {
	CreateIAMToken(ctx context.Context) (*iam.CreateIamTokenResponse, error)
	CreateIAMTokenForServiceAccount(ctx context.Context, serviceAccountID string) (*iam.CreateIamTokenResponse, error)
}

var _ Authenticator = &SDK{}

func NewIAMTokenMiddleware(authenticator Authenticator, now func() time.Time) *IamTokenMiddleware {
	return &IamTokenMiddleware{
		now:            now,
		authenticator:  authenticator,
		subjectToState: map[authSubject]iamTokenState{},
	}
}

type IamTokenMiddleware struct {
	authenticator Authenticator
	// now may be replaced in tests
	now func() time.Time

	singleFlight singleflight.Group

	// mutex guards conn and currentState, and excludes multiple simultaneous token updates
	mutex          sync.RWMutex
	subjectToState map[authSubject]iamTokenState
}

type iamTokenState struct {
	token     string
	expiresAt time.Time
	version   int
}

func WithAuthAsServiceAccount(serviceAccountID string) grpc.CallOption {
	return &withServiceAccountID{serviceAccountIDGet: func(ctx context.Context) (string, error) {
		return serviceAccountID, nil
	}}
}

type SAGetter func(ctx context.Context) (string, error)

func WithAuthAsServiceAccounts(saGetter SAGetter) grpc.CallOption {
	return &withServiceAccountID{serviceAccountIDGet: saGetter}
}

func (c *IamTokenMiddleware) InterceptUnary(ctx context.Context, method string, req, reply interface{}, conn *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, err := c.contextWithAuthMetadata(ctx, method, opts)
	if err != nil {
		return err
	}
	return invoker(ctx, method, req, reply, conn, opts...)
}

func (c *IamTokenMiddleware) InterceptStream(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx, err := c.contextWithAuthMetadata(ctx, method, opts)
	if err != nil {
		return nil, err
	}
	return streamer(ctx, desc, conn, method, opts...)
}

func (c *IamTokenMiddleware) contextWithAuthMetadata(ctx context.Context, method string, opts []grpc.CallOption) (context.Context, error) {
	if !needAuth(method) {
		return ctx, nil
	}
	// User can add WithAuthAsServiceAccount to default call options and we will
	// always try to issue token for service account. That results in a deadlock.
	// Here we check for methods that always require original authentication and
	// not delegated mode.
	needOriginalSubject := method == "/yandex.cloud.iam.v1.IamTokenService/CreateForServiceAccount"
	grpclog.Infof("Getting IAM Token for %s", method)
	token, err := c.GetIAMToken(ctx, needOriginalSubject, opts...)
	if err != nil {
		return nil, err
	}
	grpclog.Infof("Got IAM Token, set 'authorization' header.")
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token), nil
}

func needAuth(method string) bool {
	switch method {
	case "/yandex.cloud.endpoint.ApiEndpointService/List", "/yandex.cloud.iam.v1.IamTokenService/Create":
		return false
	default:
		return true
	}
}

func (c *IamTokenMiddleware) GetIAMToken(ctx context.Context, originalSubject bool, opts ...grpc.CallOption) (string, error) {
	subject, err := callAuthSubject(ctx, originalSubject, opts)
	if err != nil {
		return "", err
	}
	if subject, ok := subject.(serviceAccountSubject); ok {
		grpclog.Infof("Getting IAM Token for Service Account: %s. ", subject.serviceAccountID)
	}
	c.mutex.RLock()
	state := c.subjectToState[subject]
	c.mutex.RUnlock()

	token := state.token
	expiresIn := state.expiresAt.Sub(c.now())
	if expiresIn > 0 {
		grpclog.Infof("IAM Token Cached. Expires in: %s. ", expiresIn)
		return token, nil
	}
	if token == "" {
		grpclog.Infof("No IAM token cached. Creating.")
	} else {
		grpclog.Infof("IAM Token expired at: %s. Updating. ", state.expiresAt)
	}
	token, err = c.updateTokenSingleFlight(ctx, subject, state)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			return "", err
		}
		return "", status.Errorf(codes.Unauthenticated, "%v", err)
	}
	return token, nil
}

func (c *IamTokenMiddleware) updateTokenSingleFlight(ctx context.Context, subject authSubject, state iamTokenState) (string, error) {
	type updateTokenResponse struct {
		token string
		err   error
	}
	resp := c.singleFlight.Do(subject, func() interface{} {
		token, err := c.updateToken(ctx, subject, state.version)
		return updateTokenResponse{token, err}
	}).(updateTokenResponse)
	return resp.token, resp.err
}

func (c *IamTokenMiddleware) updateToken(ctx context.Context, subject authSubject, currentVersion int) (string, error) {
	c.mutex.RLock()
	state := c.subjectToState[subject]
	c.mutex.RUnlock()
	if state.version != currentVersion {
		// someone have already updated it
		return state.token, nil
	}

	resp, err := subject.createIAMToken(ctx, c.authenticator)
	if err != nil {
		return "", sdkerrors.WithMessage(err, "iam token create failed")
	}
	expiresAt, expiresAtErr := resp.ExpiresAt.AsTime(), resp.ExpiresAt.CheckValid()
	if expiresAtErr != nil {
		grpclog.Warningf("invalid IAM Token expires_at: %s", expiresAtErr)
		// Fallback to short term caching.
		expiresAt = c.now().Add(time.Minute)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.subjectToState[subject] = iamTokenState{
		token:     resp.IamToken,
		expiresAt: expiresAt,
		version:   currentVersion + 1,
	}
	return resp.IamToken, nil
}

type authSubject interface {
	createIAMToken(ctx context.Context, a Authenticator) (*iam.CreateIamTokenResponse, error)
}

var _ authSubject
var _ serviceAccountSubject

type mainSubject struct{}
type serviceAccountSubject struct{ serviceAccountID string }

func (s mainSubject) createIAMToken(ctx context.Context, a Authenticator) (*iam.CreateIamTokenResponse, error) {
	return a.CreateIAMToken(ctx)
}
func (s serviceAccountSubject) createIAMToken(ctx context.Context, a Authenticator) (*iam.CreateIamTokenResponse, error) {
	return a.CreateIAMTokenForServiceAccount(ctx, s.serviceAccountID)
}

type withServiceAccountID struct {
	grpc.EmptyCallOption
	serviceAccountIDGet SAGetter
}

func callAuthSubject(ctx context.Context, originalSubject bool, os []grpc.CallOption) (authSubject, error) {
	if originalSubject {
		return mainSubject{}, nil
	}
	var saOpt *withServiceAccountID
	for _, o := range os {
		o, ok := o.(*withServiceAccountID)
		if ok {
			saOpt = o
		}
	}
	var subject authSubject = mainSubject{}
	if saOpt != nil {
		sa, err := saOpt.serviceAccountIDGet(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting SA for delegation: %v", err)
		}
		subject = serviceAccountSubject{
			serviceAccountID: sa,
		}
	}
	return subject, nil
}
