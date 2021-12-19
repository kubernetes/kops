package pkibootstrap

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/pki"
	gcetpm "k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type verifier struct {
	opt    Options
	client client.Client
}

// NewVerifier constructs a new verifier.
func NewVerifier(opt *Options, client client.Client) (bootstrap.Verifier, error) {
	return &verifier{
		opt:    *opt,
		client: client,
	}, nil
}

var _ bootstrap.Verifier = &verifier{}

// TODO: Dedup with gce
func (v *verifier) parseTokenData(tokenPrefix string, authToken string, body []byte) (*gcetpm.AuthToken, *gcetpm.AuthTokenData, error) {
	if !strings.HasPrefix(authToken, tokenPrefix) {
		return nil, nil, fmt.Errorf("incorrect authorization type")
	}
	authToken = strings.TrimPrefix(authToken, tokenPrefix)

	tokenBytes, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, nil, fmt.Errorf("decoding authorization token: %w", err)
	}

	token := &gcetpm.AuthToken{}
	if err = json.Unmarshal(tokenBytes, token); err != nil {
		return nil, nil, fmt.Errorf("unmarshalling authorization token: %w", err)
	}

	tokenData := &gcetpm.AuthTokenData{}
	if err := json.Unmarshal(token.Data, tokenData); err != nil {
		return nil, nil, fmt.Errorf("unmarshalling authorization token data: %w", err)
	}

	// Guard against replay attacks
	if tokenData.Audience != gcetpm.AudienceNodeAuthentication {
		return nil, nil, fmt.Errorf("incorrect Audience")
	}
	timeSkew := math.Abs(time.Since(time.Unix(tokenData.Timestamp, 0)).Seconds())
	if timeSkew > float64(v.opt.MaxTimeSkew) {
		return nil, nil, fmt.Errorf("incorrect Timestamp %v", tokenData.Timestamp)
	}

	// Verify the token has signed the body content.
	requestHash := sha256.Sum256(body)
	if !bytes.Equal(requestHash[:], tokenData.RequestHash) {
		return nil, nil, fmt.Errorf("incorrect RequestHash")
	}

	return token, tokenData, nil
}

// Can generate keys with
// openssl ecparam -name prime256v1 -genkey -noout -out ec-priv-key.pem
// openssl ec -in ec-priv-key.pem -pubout > ec-pub-key.pem
// Note that golang doesn't support secp256k1: https://groups.google.com/g/golang-nuts/c/Mbkug5t3ZYA

func (v *verifier) VerifyToken(ctx context.Context, authToken string, body []byte, useInstanceIDForNodeName bool) (*bootstrap.VerifyResult, error) {
	// Reminder: we shouldn't trust any data we get from the client until we've checked the signature (and even then...)
	// Thankfully the GCE SDK does seem to escape the parameters correctly, for example.

	token, tokenData, err := v.parseTokenData(AuthenticationTokenPrefix, authToken, body)
	if err != nil {
		return nil, err
	}

	// Verify the token has a valid signature.
	result, signingKey, err := v.getSigningKey(ctx, tokenData, useInstanceIDForNodeName)
	if err != nil {
		return nil, err
	}

	if !verifySignature(signingKey, token.Data, token.Signature) {
		return nil, fmt.Errorf("failed to verify claim signature for node")
	}

	return result, nil
}

func (v *verifier) getSigningKey(ctx context.Context, tokenData *gcetpm.AuthTokenData, useInstanceIDForNodeName bool) (*bootstrap.VerifyResult, crypto.PublicKey, error) {
	nodeName := tokenData.Instance
	id := types.NamespacedName{
		Namespace: "kops-system",
		Name:      nodeName,
	}
	var secret corev1.Secret
	if err := v.client.Get(ctx, id, &secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil, fmt.Errorf("secret not found for %v", id)
		}
		return nil, nil, fmt.Errorf("error getting secret %v: %w", id, err)
	}

	// TODO: Check instance-group matches request (does it matter?)

	pubKeyBytes := secret.Data["public-key"]
	if pubKeyBytes == nil {
		return nil, nil, fmt.Errorf("secret %v did not have public-key", id)
	}
	instanceGroupBytes := secret.Data["instance-group"]
	if instanceGroupBytes == nil {
		return nil, nil, fmt.Errorf("secret %v did not have instance-group", id)
	}
	pubKey, err := pki.ParsePEMPublicKey([]byte(pubKeyBytes))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	var sans []string
	// sans, err := v.GetInstanceCertificateAlternateNames(instance)
	// if err != nil {
	// 	return nil, err
	// }

	result := &bootstrap.VerifyResult{
		NodeName:          nodeName,
		InstanceGroupName: string(instanceGroupBytes),
		CertificateNames:  sans,
	}

	return result, pubKey.Key, nil
}

func verifySignature(signingKey crypto.PublicKey, payload []byte, signature []byte) bool {
	attestHash := sha256.Sum256(payload)
	switch signingKey := signingKey.(type) {
	case *ecdsa.PublicKey:
		klog.Infof("attestHash %x", attestHash)
		klog.Infof("sig %x", signature)
		return ecdsa.VerifyASN1(signingKey, attestHash[:], signature)

	default:
		klog.Warningf("key type %T not supported", signingKey)
		return false
	}
}
