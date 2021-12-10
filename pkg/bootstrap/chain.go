package bootstrap

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"
)

func NewChainVerifier(chain ...Verifier) Verifier {
	return &ChainVerifier{chain: chain}
}

type ChainVerifier struct {
	chain []Verifier
}

func (v *ChainVerifier) VerifyToken(ctx context.Context, token string, body []byte, useInstanceIDForNodeName bool) (*VerifyResult, error) {
	for _, verifier := range v.chain {
		// TODO: Check prefix?
		result, err := verifier.VerifyToken(ctx, token, body, useInstanceIDForNodeName)
		if err == nil {
			return result, nil
		}
		klog.Infof("failed to verify token: %v", err)
	}
	return nil, fmt.Errorf("unable to verify token")
}
