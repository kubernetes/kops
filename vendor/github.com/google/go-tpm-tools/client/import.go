package client

import (
	"fmt"

	"github.com/google/go-tpm-tools/internal"
	pb "github.com/google/go-tpm-tools/proto/tpm"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

func loadHandle(k *Key, blob *pb.ImportBlob) (tpmutil.Handle, error) {
	auth, err := k.session.Auth()
	if err != nil {
		return tpm2.HandleNull, err
	}
	private, err := tpm2.Import(k.rw, k.Handle(), auth, blob.PublicArea, blob.Duplicate, blob.EncryptedSeed, nil, nil)
	if err != nil {
		return tpm2.HandleNull, fmt.Errorf("import failed: %w", err)
	}

	auth, err = k.session.Auth()
	if err != nil {
		return tpm2.HandleNull, err
	}
	handle, _, err := tpm2.LoadUsingAuth(k.rw, k.Handle(), auth, blob.PublicArea, private)
	if err != nil {
		return tpm2.HandleNull, fmt.Errorf("load failed: %w", err)
	}
	return handle, nil
}

// Import decrypts the secret contained in an encoded import request.
// The key used must be an encryption key (signing keys cannot be used).
// The req parameter should come from server.CreateImportBlob.
func (k *Key) Import(blob *pb.ImportBlob) ([]byte, error) {
	handle, err := loadHandle(k, blob)
	if err != nil {
		return nil, err
	}
	defer tpm2.FlushContext(k.rw, handle)

	unsealSession, err := NewPCRSession(k.rw, internal.PCRSelection(blob.Pcrs))
	if err != nil {
		return nil, err
	}
	defer unsealSession.Close()

	auth, err := unsealSession.Auth()
	if err != nil {
		return nil, err
	}
	out, err := tpm2.UnsealWithSession(k.rw, auth.Session, handle, "")
	if err != nil {
		return nil, fmt.Errorf("unseal failed: %w", err)
	}
	return out, nil
}

// ImportSigningKey returns the signing key contained in an encoded import request.
// The parent key must be an encryption key (signing keys cannot be used).
// The req parameter should come from server.CreateSigningKeyImportBlob.
func (k *Key) ImportSigningKey(blob *pb.ImportBlob) (key *Key, err error) {
	handle, err := loadHandle(k, blob)
	if err != nil {
		return nil, err
	}
	key = &Key{rw: k.rw, handle: handle}

	defer func() {
		if err != nil {
			key.Close()
		}
	}()

	if key.pubArea, _, _, err = tpm2.ReadPublic(k.rw, handle); err != nil {
		return
	}
	if key.session, err = NewPCRSession(k.rw, internal.PCRSelection(blob.Pcrs)); err != nil {
		return
	}
	return key, key.finish()
}
