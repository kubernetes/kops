// Package tpm2 contains TPM 2.0 commands and structures.
package tpm2

import (
	"bytes"
	"encoding/binary"

	"github.com/google/go-tpm/tpm2/transport"
)

// handle represents a TPM handle as comprehended in Part 3: Commands.
// In the context of TPM commands, handles are special parameters for which
// there is a known associated name.
// This is not an exported interface, because the reflection logic has special
// behavior for AuthHandle, due to the fact that referencing Session from this
// interface would break the ability to make TPMHandle implement it.
type handle interface {
	// HandleValue is the numeric concrete handle value in the TPM.
	HandleValue() uint32
	// KnownName is the TPM Name of the associated entity. See Part 1, section 16.
	KnownName() *TPM2BName
}

// NamedHandle represents an associated pairing of TPM handle and known Name.
type NamedHandle struct {
	Handle TPMHandle
	Name   TPM2BName
}

// HandleValue implements the handle interface.
func (h NamedHandle) HandleValue() uint32 {
	return h.Handle.HandleValue()
}

// KnownName implements the handle interface.
func (h NamedHandle) KnownName() *TPM2BName {
	return &h.Name
}

// AuthHandle allows the caller to add an authorization session onto a handle.
type AuthHandle struct {
	Handle TPMHandle
	Name   TPM2BName
	Auth   Session
}

// HandleValue implements the handle interface.
func (h AuthHandle) HandleValue() uint32 {
	return h.Handle.HandleValue()
}

// KnownName implements the handle interface.
// If Name is not provided (i.e., only Auth), then rely on the underlying
// TPMHandle.
func (h AuthHandle) KnownName() *TPM2BName {
	if len(h.Name.Buffer) != 0 {
		return &h.Name
	}
	return h.Handle.KnownName()
}

// invalidHandleValue is the sentinel value used for handles that have been
// reconstructed from unmarshalling.
// This value is intentionally invalid to prevent accidental use with a real TPM.
const invalidHandleValue TPMHandle = 0xFFFFFFFF

// UnmarshalledHandle represents a handle reconstructed from unmarshalling.
// This type is used for audit and inspection purposes where only the Name is
// available, not the actual TPM handle value.
//
// The HandleValue() method returns [invalidHandleValue] to prevent accidental
// use of these handles with a real TPM.
type UnmarshalledHandle struct {
	Name TPM2BName
}

// HandleValue implements the handle interface.
// Returns invalidHandleValue since unmarshalled handles don't have real TPM handle values.
func (h UnmarshalledHandle) HandleValue() uint32 {
	return uint32(invalidHandleValue)
}

// KnownName implements the handle interface.
func (h UnmarshalledHandle) KnownName() *TPM2BName {
	return &h.Name
}

// Command is an interface for any TPM command, parameterized by its response
// type.
type Command[R any, PR *R] interface {
	// The TPM command code associated with this command.
	Command() TPMCC
	// Executes the command and returns the response.
	Execute(t transport.TPM, s ...Session) (PR, error)
}

// PolicyCommand is a TPM command that can be part of a TPM policy.
type PolicyCommand interface {
	// Update updates the given policy hash according to the command
	// parameters.
	Update(policy *PolicyCalculator) error
}

// Shutdown is the input to TPM2_Shutdown.
// See definition in Part 3, Commands, section 9.4.
type Shutdown struct {
	// TPM_SU_CLEAR or TPM_SU_STATE
	ShutdownType TPMSU
}

// Command implements the Command interface.
func (Shutdown) Command() TPMCC { return TPMCCShutdown }

// Execute executes the command and returns the response.
func (cmd Shutdown) Execute(t transport.TPM, s ...Session) (*ShutdownResponse, error) {
	var rsp ShutdownResponse
	err := execute[ShutdownResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ShutdownResponse is the response from TPM2_Shutdown.
type ShutdownResponse struct{}

// Startup is the input to TPM2_Startup.
// See definition in Part 3, Commands, section 9.3.
type Startup struct {
	// TPM_SU_CLEAR or TPM_SU_STATE
	StartupType TPMSU
}

// Command implements the Command interface.
func (Startup) Command() TPMCC { return TPMCCStartup }

// Execute executes the command and returns the response.
func (cmd Startup) Execute(t transport.TPM, s ...Session) (*StartupResponse, error) {
	var rsp StartupResponse
	err := execute[StartupResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// StartupResponse is the response from TPM2_Startup.
type StartupResponse struct{}

// StartAuthSession is the input to TPM2_StartAuthSession.
// See definition in Part 3, Commands, section 11.1
type StartAuthSession struct {
	// handle of a loaded decrypt key used to encrypt salt
	// may be TPM_RH_NULL
	TPMKey handle `gotpm:"handle"`
	// entity providing the authValue
	// may be TPM_RH_NULL
	Bind handle `gotpm:"handle"`
	// initial nonceCaller, sets nonceTPM size for the session
	// shall be at least 16 octets
	NonceCaller TPM2BNonce
	// value encrypted according to the type of tpmKey
	// If tpmKey is TPM_RH_NULL, this shall be the Empty Buffer.
	EncryptedSalt TPM2BEncryptedSecret
	// indicates the type of the session; simple HMAC or policy (including
	// a trial policy)
	SessionType TPMSE
	// the algorithm and key size for parameter encryption
	// may select transport.TPM_ALG_NULL
	Symmetric TPMTSymDef
	// hash algorithm to use for the session
	// Shall be a hash algorithm supported by the TPM and not transport.TPM_ALG_NULL
	AuthHash TPMIAlgHash
}

// Command implements the Command interface.
func (StartAuthSession) Command() TPMCC { return TPMCCStartAuthSession }

// Execute executes the command and returns the response.
func (cmd StartAuthSession) Execute(t transport.TPM, s ...Session) (*StartAuthSessionResponse, error) {
	var rsp StartAuthSessionResponse
	if err := execute[StartAuthSessionResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// StartAuthSessionResponse is the response from TPM2_StartAuthSession.
type StartAuthSessionResponse struct {
	// handle for the newly created session
	SessionHandle TPMISHAuthSession `gotpm:"handle"`
	// the initial nonce from the TPM, used in the computation of the sessionKey
	NonceTPM TPM2BNonce
}

// Create is the input to TPM2_Create.
// See definition in Part 3, Commands, section 12.1
type Create struct {
	// handle of parent for new object
	ParentHandle handle `gotpm:"handle,auth"`
	// the sensitive data
	InSensitive TPM2BSensitiveCreate
	// the public template
	InPublic TPM2BPublic
	// data that will be included in the creation data for this
	// object to provide permanent, verifiable linkage between this
	// object and some object owner data
	OutsideInfo TPM2BData
	// PCR that will be used in creation data
	CreationPCR TPMLPCRSelection
}

// Command implements the Command interface.
func (Create) Command() TPMCC { return TPMCCCreate }

// Execute executes the command and returns the response.
func (cmd Create) Execute(t transport.TPM, s ...Session) (*CreateResponse, error) {
	var rsp CreateResponse
	if err := execute[CreateResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// CreateResponse is the response from TPM2_Create.
type CreateResponse struct {
	// the private portion of the object
	OutPrivate TPM2BPrivate
	// the public portion of the created object
	OutPublic TPM2BPublic
	// contains a TPMS_CREATION_DATA
	CreationData tpm2bCreationData
	// digest of creationData using nameAlg of outPublic
	CreationHash TPM2BDigest
	// ticket used by TPM2_CertifyCreation() to validate that the
	// creation data was produced by the TPM.
	CreationTicket TPMTTKCreation
}

// Load is the input to TPM2_Load.
// See definition in Part 3, Commands, section 12.2
type Load struct {
	// handle of parent for new object
	ParentHandle handle `gotpm:"handle,auth"`
	// the private portion of the object
	InPrivate TPM2BPrivate
	// the public portion of the object
	InPublic TPM2BPublic
}

// Command implements the Command interface.
func (Load) Command() TPMCC { return TPMCCLoad }

// Execute executes the command and returns the response.
func (cmd Load) Execute(t transport.TPM, s ...Session) (*LoadResponse, error) {
	var rsp LoadResponse
	if err := execute[LoadResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// LoadResponse is the response from TPM2_Load.
type LoadResponse struct {
	// handle of type TPM_HT_TRANSIENT for loaded object
	ObjectHandle TPMHandle `gotpm:"handle"`
	// Name of the loaded object
	Name TPM2BName
}

// LoadExternal is the input to TPM2_LoadExternal.
// See definition in Part 3, Commands, section 12.3
type LoadExternal struct {
	// the sensitive portion of the object (optional)
	InPrivate TPM2BSensitive `gotpm:"optional"`
	// the public portion of the object
	InPublic TPM2BPublic
	// hierarchy with which the object area is associated
	Hierarchy TPMIRHHierarchy `gotpm:"nullable"`
}

// Command implements the Command interface.
func (LoadExternal) Command() TPMCC { return TPMCCLoadExternal }

// Execute executes the command and returns the response.
func (cmd LoadExternal) Execute(t transport.TPM, s ...Session) (*LoadExternalResponse, error) {
	var rsp LoadExternalResponse
	if err := execute[LoadExternalResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// LoadExternalResponse is the response from TPM2_LoadExternal.
type LoadExternalResponse struct {
	// handle of type TPM_HT_TRANSIENT for loaded object
	ObjectHandle TPMHandle `gotpm:"handle"`
	// Name of the loaded object
	Name TPM2BName
}

// ReadPublic is the input to TPM2_ReadPublic.
// See definition in Part 3, Commands, section 12.4
type ReadPublic struct {
	// TPM handle of an object
	ObjectHandle handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (ReadPublic) Command() TPMCC { return TPMCCReadPublic }

// Execute executes the command and returns the response.
func (cmd ReadPublic) Execute(t transport.TPM, s ...Session) (*ReadPublicResponse, error) {
	var rsp ReadPublicResponse
	if err := execute[ReadPublicResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ReadPublicResponse is the response from TPM2_ReadPublic.
type ReadPublicResponse struct {
	// structure containing the public area of an object
	OutPublic TPM2BPublic
	// name of object
	Name TPM2BName
	// the Qualified Name of the object
	QualifiedName TPM2BName
}

// ActivateCredential is the input to TPM2_ActivateCredential.
// See definition in Part 3, Commands, section 12.5.
type ActivateCredential struct {
	// handle of the object associated with certificate in credentialBlob
	ActivateHandle handle `gotpm:"handle,auth"`
	// loaded key used to decrypt the TPMS_SENSITIVE in credentialBlob
	KeyHandle handle `gotpm:"handle,auth"`
	// the credential
	CredentialBlob TPM2BIDObject
	// keyHandle algorithm-dependent encrypted seed that protects credentialBlob
	Secret TPM2BEncryptedSecret
}

// Command implements the Command interface.
func (ActivateCredential) Command() TPMCC { return TPMCCActivateCredential }

// Execute executes the command and returns the response.
func (cmd ActivateCredential) Execute(t transport.TPM, s ...Session) (*ActivateCredentialResponse, error) {
	var rsp ActivateCredentialResponse
	if err := execute[ActivateCredentialResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ActivateCredentialResponse is the response from TPM2_ActivateCredential.
type ActivateCredentialResponse struct {
	// the decrypted certificate information
	CertInfo TPM2BDigest
}

// MakeCredential is the input to TPM2_MakeCredential.
// See definition in Part 3, Commands, section 12.6.
type MakeCredential struct {
	// loaded public area, used to encrypt the sensitive area containing the credential key
	Handle TPMIDHObject `gotpm:"handle"`
	// the credential information
	Credential TPM2BDigest
	// Name of the object to which the credential applies
	ObjectName TPM2BName
}

// Command implements the Command interface.
func (MakeCredential) Command() TPMCC { return TPMCCMakeCredential }

// Execute executes the command and returns the response.
func (cmd MakeCredential) Execute(t transport.TPM, s ...Session) (*MakeCredentialResponse, error) {
	var rsp MakeCredentialResponse
	if err := execute[MakeCredentialResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// MakeCredentialResponse is the response from TPM2_MakeCredential.
type MakeCredentialResponse struct {
	// the credential
	CredentialBlob TPM2BIDObject
	// handle algorithm-dependent data that wraps the key that encrypts credentialBlob
	Secret TPM2BEncryptedSecret
}

// Unseal is the input to TPM2_Unseal.
// See definition in Part 3, Commands, section 12.7
type Unseal struct {
	ItemHandle handle `gotpm:"handle,auth"`
}

// Command implements the Command interface.
func (Unseal) Command() TPMCC { return TPMCCUnseal }

// Execute executes the command and returns the response.
func (cmd Unseal) Execute(t transport.TPM, s ...Session) (*UnsealResponse, error) {
	var rsp UnsealResponse
	if err := execute[UnsealResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// UnsealResponse is the response from TPM2_Unseal.
type UnsealResponse struct {
	OutData TPM2BSensitiveData
}

// ObjectChangeAuth is the input to TPM2_ObjectChangeAuth.
// See definition in Part 3, Commands, section 12.8
type ObjectChangeAuth struct {
	// TPM handle of an object
	ObjectHandle handle `gotpm:"handle,auth"`
	// handle of the parent
	ParentHandle handle `gotpm:"handle"`
	// new authorization value
	NewAuth TPM2BAuth
}

// Command implements the Command interface.
func (ObjectChangeAuth) Command() TPMCC { return TPMCCObjectChangeAuth }

// Execute executes the command and returns the response.
func (cmd ObjectChangeAuth) Execute(t transport.TPM, s ...Session) (*ObjectChangeAuthResponse, error) {
	var rsp ObjectChangeAuthResponse
	if err := execute[ObjectChangeAuthResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ObjectChangeAuthResponse the response from TPM2_ObjectChangeAuth.
type ObjectChangeAuthResponse struct {
	// private area containing the new authorization value
	OutPrivate TPM2BPrivate
}

// CreateLoaded is the input to TPM2_CreateLoaded.
// See definition in Part 3, Commands, section 12.9
type CreateLoaded struct {
	// Handle of a transient storage key, a persistent storage key,
	// TPM_RH_ENDORSEMENT, TPM_RH_OWNER, TPM_RH_PLATFORM+{PP}, or TPM_RH_NULL
	ParentHandle handle `gotpm:"handle,auth"`
	// the sensitive data, see TPM 2.0 Part 1 Sensitive Values
	InSensitive TPM2BSensitiveCreate
	// the public template
	InPublic TPM2BTemplate
}

// Command implements the Command interface.
func (CreateLoaded) Command() TPMCC { return TPMCCCreateLoaded }

// Execute executes the command and returns the response.
func (cmd CreateLoaded) Execute(t transport.TPM, s ...Session) (*CreateLoadedResponse, error) {
	var rsp CreateLoadedResponse
	if err := execute[CreateLoadedResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// CreateLoadedResponse is the response from TPM2_CreateLoaded.
type CreateLoadedResponse struct {
	// handle of type TPM_HT_TRANSIENT for loaded object
	ObjectHandle TPMHandle `gotpm:"handle"`
	// the sensitive area of the object (optional)
	OutPrivate TPM2BPrivate `gotpm:"optional"`
	// the public portion of the created object
	OutPublic TPM2BPublic
	// the name of the created object
	Name TPM2BName
}

// EncryptDecrypt2 is the input to TPM2_EncryptDecrypt2
type EncryptDecrypt2 struct {
	// reference to public portion of symmetric key to use for encryption
	KeyHandle handle `gotpm:"handle,auth"`
	Message   TPM2BMaxBuffer
	Decrypt   TPMIYesNo
	Mode      TPMIAlgSymMode `gotpm:"nullable"`
	IV        TPM2BIV
}

// Command implements the Command interface.
func (EncryptDecrypt2) Command() TPMCC { return TPMCCEncryptDecrypt2 }

// Execute executes the command and returns the response.
func (cmd EncryptDecrypt2) Execute(t transport.TPM, s ...Session) (*EncryptDecrypt2Response, error) {
	var rsp EncryptDecrypt2Response
	err := execute[EncryptDecrypt2Response](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// EncryptDecrypt2Response is the response from TPM2_EncryptDecrypt2.
type EncryptDecrypt2Response struct {
	OutData TPM2BMaxBuffer
	IV      TPM2BIV
}

// RSAEncrypt is the input to TPM2_RSA_Encrypt
// See definition in Part 3, Commands, section 14.2.
type RSAEncrypt struct {
	// reference to public portion of RSA key to use for encryption
	KeyHandle handle `gotpm:"handle"`
	// message to be encrypted
	Message TPM2BPublicKeyRSA
	// the padding scheme to use if scheme associated with keyHandle is TPM_ALG_NULL
	InScheme TPMTRSADecrypt `gotpm:"nullable"`
	// optional label L to be associated with the message
	Label TPM2BData `gotpm:"optional"`
}

// Command implements the Command interface.
func (RSAEncrypt) Command() TPMCC { return TPMCCRSAEncrypt }

// Execute executes the command and returns the response.
func (cmd RSAEncrypt) Execute(t transport.TPM, s ...Session) (*RSAEncryptResponse, error) {
	var rsp RSAEncryptResponse
	if err := execute[RSAEncryptResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// RSAEncryptResponse is the response from TPM2_RSA_Encrypt
type RSAEncryptResponse struct {
	// encrypted output
	OutData TPM2BPublicKeyRSA
}

// RSADecrypt is the input to TPM2_RSA_Decrypt
// See definition in Part 3, Commands, section 14.3.
type RSADecrypt struct {
	// RSA key to use for decryption
	KeyHandle handle `gotpm:"handle,auth"`
	// cipher text to be decrypted
	CipherText TPM2BPublicKeyRSA
	// the padding scheme to use if scheme associated with keyHandle is TPM_ALG_NULL
	InScheme TPMTRSADecrypt `gotpm:"nullable"`
	// label whose association with the message is to be verified
	Label TPM2BData `gotpm:"optional"`
}

// Command implements the Command interface.
func (RSADecrypt) Command() TPMCC { return TPMCCRSADecrypt }

// Execute executes the command and returns the response.
func (cmd RSADecrypt) Execute(t transport.TPM, s ...Session) (*RSADecryptResponse, error) {
	var rsp RSADecryptResponse
	if err := execute[RSADecryptResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// RSADecryptResponse is the response from TPM2_RSA_Decrypt
type RSADecryptResponse struct {
	// decrypted output
	Message TPM2BPublicKeyRSA
}

// ECDHZGen is the input to TPM2_ECDHZGen.
// See definition in Part 3, Commands, section 14.5
type ECDHZGen struct {
	// handle of a loaded ECC key
	KeyHandle handle `gotpm:"handle,auth"`
	// a public key
	InPoint TPM2BECCPoint
}

// Command implements the Command interface.
func (ECDHZGen) Command() TPMCC { return TPMCCECDHZGen }

// Execute executes the command and returns the response.
func (cmd ECDHZGen) Execute(t transport.TPM, s ...Session) (*ECDHZGenResponse, error) {
	var rsp ECDHZGenResponse
	if err := execute[ECDHZGenResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ECDHZGenResponse is the response from TPM2_ECDHZGen.
type ECDHZGenResponse struct {
	// X and Y coordinates of the product of the multiplication
	OutPoint TPM2BECCPoint
}

// Hash is the input to TPM2_Hash.
// See definition in Part 3, Commands, section 15.4
type Hash struct {
	//data to be hashed
	Data TPM2BMaxBuffer
	// algorithm for the hash being computed - shall not be TPM_ALH_NULL
	HashAlg TPMIAlgHash
	// hierarchy to use for the ticket (TPM_RH_NULL_allowed)
	Hierarchy TPMIRHHierarchy `gotpm:"nullable"`
}

// Command implements the Command interface.
func (Hash) Command() TPMCC { return TPMCCHash }

// Execute executes the command and returns the response.
func (cmd Hash) Execute(t transport.TPM, s ...Session) (*HashResponse, error) {
	var rsp HashResponse
	if err := execute[HashResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// HashResponse is the response from TPM2_Hash.
type HashResponse struct {
	// results
	OutHash TPM2BDigest
	// ticket indicating that the sequence of octets used to
	// compute outDigest did not start with TPM_GENERATED_VALUE
	Validation TPMTTKHashCheck
}

// Hmac is the input to TPM2_HMAC.
// See definition in Part 3, Commands, section 15.5.
type Hmac struct {
	// HMAC key handle requiring an authorization session for the USER role
	Handle AuthHandle `gotpm:"handle,auth"`
	// HMAC data
	Buffer TPM2BMaxBuffer
	// Algorithm to use for HMAC
	HashAlg TPMIAlgHash
}

// Command implements the Command interface.
func (Hmac) Command() TPMCC { return TPMCCHMAC }

// Execute executes the command and returns the response.
func (cmd Hmac) Execute(t transport.TPM, s ...Session) (*HmacResponse, error) {
	var rsp HmacResponse
	if err := execute[HmacResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// HmacResponse is the response from TPM2_HMAC.
type HmacResponse struct {
	// the returned HMAC in a sized buffer
	OutHMAC TPM2BDigest
}

// GetRandom is the input to TPM2_GetRandom.
// See definition in Part 3, Commands, section 16.1
type GetRandom struct {
	// number of octets to return
	BytesRequested uint16
}

// Command implements the Command interface.
func (GetRandom) Command() TPMCC { return TPMCCGetRandom }

// Execute executes the command and returns the response.
func (cmd GetRandom) Execute(t transport.TPM, s ...Session) (*GetRandomResponse, error) {
	var rsp GetRandomResponse
	if err := execute[GetRandomResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// GetRandomResponse is the response from TPM2_GetRandom.
type GetRandomResponse struct {
	// the random octets
	RandomBytes TPM2BDigest
}

// HashSequenceStart is the input to TPM2_HashSequenceStart.
// See definition in Part 3, Commands, section 17.3
type HashSequenceStart struct {
	// authorization value for subsequent use of the sequence
	Auth TPM2BAuth
	// the hash algorithm to use for the hash sequence
	// An Event Sequence starts if this is TPM_ALG_NULL.
	HashAlg TPMIAlgHash
}

// Command implements the Command interface.
func (HashSequenceStart) Command() TPMCC { return TPMCCHashSequenceStart }

// Execute executes the command and returns the response.
func (cmd HashSequenceStart) Execute(t transport.TPM, s ...Session) (*HashSequenceStartResponse, error) {
	var rsp HashSequenceStartResponse
	if err := execute[HashSequenceStartResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// HashSequenceStartResponse is the response from TPM2_StartHashSequence.
type HashSequenceStartResponse struct {
	// a handle to reference the sequence
	SequenceHandle TPMIDHObject
}

// HmacStart is the input to TPM2_HMAC_Start.
// See definition in Part 3, Commands, section 17.2.2
type HmacStart struct {
	// HMAC key handle
	Handle handle `gotpm:"handle,auth"`
	// authorization value for subsequent use of the sequence
	Auth TPM2BAuth
	// the hash algorithm to use for the hmac sequence
	HashAlg TPMIAlgHash `gotpm:"nullable"`
}

// Command implements the Command interface.
func (HmacStart) Command() TPMCC { return TPMCCHMACStart }

// Execute executes the command and returns the response.
func (cmd HmacStart) Execute(t transport.TPM, s ...Session) (*HmacStartResponse, error) {
	var rsp HmacStartResponse
	if err := execute[HmacStartResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// HmacStartResponse is the response from TPM2_HMAC_Start.
// See definition in Part 3, Commands, section 17.2.2
type HmacStartResponse struct {
	// a handle to reference the sequence
	SequenceHandle TPMIDHObject `gotpm:"handle"`
}

// SequenceUpdate is the input to TPM2_SequenceUpdate.
// See definition in Part 3, Commands, section 17.4
type SequenceUpdate struct {
	// handle for the sequence object
	SequenceHandle handle `gotpm:"handle,auth,anon"`
	// data to be added to hash
	Buffer TPM2BMaxBuffer
}

// Command implements the Command interface.
func (SequenceUpdate) Command() TPMCC { return TPMCCSequenceUpdate }

// Execute executes the command and returns the response.
func (cmd SequenceUpdate) Execute(t transport.TPM, s ...Session) (*SequenceUpdateResponse, error) {
	var rsp SequenceUpdateResponse
	if err := execute[SequenceUpdateResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// SequenceUpdateResponse is the response from TPM2_SequenceUpdate.
type SequenceUpdateResponse struct{}

// SequenceComplete is the input to TPM2_SequenceComplete.
// See definition in Part 3, Commands, section 17.5
type SequenceComplete struct {
	// authorization for the sequence
	SequenceHandle handle `gotpm:"handle,auth,anon"`
	// data to be added to the hash/HMAC
	Buffer TPM2BMaxBuffer
	// hierarchy of the ticket for a hash
	Hierarchy TPMIRHHierarchy `gotpm:"nullable"`
}

// Command implements the Command interface.
func (SequenceComplete) Command() TPMCC { return TPMCCSequenceComplete }

// Execute executes the command and returns the response.
func (cmd SequenceComplete) Execute(t transport.TPM, s ...Session) (*SequenceCompleteResponse, error) {
	var rsp SequenceCompleteResponse
	if err := execute[SequenceCompleteResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// SequenceCompleteResponse is the response from TPM2_SequenceComplete.
type SequenceCompleteResponse struct {
	// the returned HMAC or digest in a sized buffer
	Result TPM2BDigest
	// 	ticket indicating that the sequence of octets used to
	// compute outDigest did not start with TPM_GENERATED_VALUE
	Validation TPMTTKHashCheck
}

// Certify is the input to TPM2_Certify.
// See definition in Part 3, Commands, section 18.2.
type Certify struct {
	// handle of the object to be certified
	ObjectHandle handle `gotpm:"handle,auth"`
	// handle of the key used to sign the attestation structure
	SignHandle handle `gotpm:"handle,auth"`
	// user provided qualifying data
	QualifyingData TPM2BData
	// signing scheme to use if the scheme for signHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme
}

// Command implements the Command interface.
func (Certify) Command() TPMCC { return TPMCCCertify }

// Execute executes the command and returns the response.
func (cmd Certify) Execute(t transport.TPM, s ...Session) (*CertifyResponse, error) {
	var rsp CertifyResponse
	if err := execute[CertifyResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// CertifyResponse is the response from TPM2_Certify.
type CertifyResponse struct {
	// the structure that was signed
	CertifyInfo TPM2BAttest
	// the asymmetric signature over certifyInfo using the key referenced by signHandle
	Signature TPMTSignature
}

// CertifyCreation is the input to TPM2_CertifyCreation.
// See definition in Part 3, Commands, section 18.3.
type CertifyCreation struct {
	// handle of the key that will sign the attestation block
	SignHandle handle `gotpm:"handle,auth"`
	// the object associated with the creation data
	ObjectHandle handle `gotpm:"handle"`
	// user-provided qualifying data
	QualifyingData TPM2BData
	// hash of the creation data produced by TPM2_Create() or TPM2_CreatePrimary()
	CreationHash TPM2BDigest
	// signing scheme to use if the scheme for signHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme
	// ticket produced by TPM2_Create() or TPM2_CreatePrimary()
	CreationTicket TPMTTKCreation
}

// Command implements the Command interface.
func (CertifyCreation) Command() TPMCC { return TPMCCCertifyCreation }

// Execute executes the command and returns the response.
func (cmd CertifyCreation) Execute(t transport.TPM, s ...Session) (*CertifyCreationResponse, error) {
	var rsp CertifyCreationResponse
	if err := execute[CertifyCreationResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// CertifyCreationResponse is the response from TPM2_CertifyCreation.
type CertifyCreationResponse struct {
	// the structure that was signed
	CertifyInfo TPM2BAttest
	// the signature over certifyInfo
	Signature TPMTSignature
}

// Quote is the input to TPM2_Quote.
// See definition in Part 3, Commands, section 18.4
type Quote struct {
	// handle of key that will perform signature
	SignHandle handle `gotpm:"handle,auth"`
	// data supplied by the caller
	QualifyingData TPM2BData
	// signing scheme to use if the scheme for signHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme
	// PCR set to quote
	PCRSelect TPMLPCRSelection
}

// Command implements the Command interface.
func (Quote) Command() TPMCC { return TPMCCQuote }

// Execute executes the command and returns the response.
func (cmd Quote) Execute(t transport.TPM, s ...Session) (*QuoteResponse, error) {
	var rsp QuoteResponse
	if err := execute[QuoteResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// QuoteResponse is the response from TPM2_Quote.
type QuoteResponse struct {
	// the quoted information
	Quoted TPM2BAttest
	// the signature over quoted
	Signature TPMTSignature
}

// GetSessionAuditDigest is the input to TPM2_GetSessionAuditDigest.
// See definition in Part 3, Commands, section 18.5
type GetSessionAuditDigest struct {
	// handle of the privacy administrator (TPM_RH_ENDORSEMENT)
	PrivacyAdminHandle handle `gotpm:"handle,auth"`
	// handle of the signing key
	SignHandle handle `gotpm:"handle,auth"`
	// handle of the audit session
	SessionHandle handle `gotpm:"handle"`
	// user-provided qualifying data – may be zero-length
	QualifyingData TPM2BData
	// signing scheme to use if the scheme for signHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme
}

// Command implements the Command interface.
func (GetSessionAuditDigest) Command() TPMCC { return TPMCCGetSessionAuditDigest }

// Execute executes the command and returns the response.
func (cmd GetSessionAuditDigest) Execute(t transport.TPM, s ...Session) (*GetSessionAuditDigestResponse, error) {
	var rsp GetSessionAuditDigestResponse
	if err := execute[GetSessionAuditDigestResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// GetSessionAuditDigestResponse is the response from
// TPM2_GetSessionAuditDigest.
type GetSessionAuditDigestResponse struct {
	// the audit information that was signed
	AuditInfo TPM2BAttest
	// the signature over auditInfo
	Signature TPMTSignature
}

// Commit is the input to TPM2_Commit.
// See definition in Part 3, Commands, section 19.2.
type Commit struct {
	// handle of the key that will be used in the signing operation
	SignHandle handle `gotpm:"handle,auth"`
	// a point (M) on the curve used by signHandle
	P1 TPM2BECCPoint
	// octet array used to derive x-coordinate of a base point
	S2 TPM2BSensitiveData
	// y coordinate of the point associated with s2
	Y2 TPM2BECCParameter
}

// Command implements the Command interface.
func (Commit) Command() TPMCC { return TPMCCCommit }

// Execute executes the command and returns the response.
func (cmd Commit) Execute(t transport.TPM, s ...Session) (*CommitResponse, error) {
	var rsp CommitResponse
	if err := execute[CommitResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}

	return &rsp, nil
}

// CommitResponse is the response from TPM2_Commit.
type CommitResponse struct {
	// ECC point K ≔ [ds](x2, y2)
	K TPM2BECCPoint
	// ECC point L ≔ [r](x2, y2)
	L TPM2BECCPoint
	// ECC point E ≔ [r]P1
	E TPM2BECCPoint
	// least-significant 16 bits of commitCount
	Counter uint16
}

// VerifySignature is the input to TPM2_VerifySignature.
// See definition in Part 3, Commands, section 20.1
type VerifySignature struct {
	// handle of public key that will be used in the validation
	KeyHandle handle `gotpm:"handle"`
	// digest of the signed message
	Digest TPM2BDigest
	// signature to be tested
	Signature TPMTSignature
}

// Command implements the Command interface.
func (VerifySignature) Command() TPMCC { return TPMCCVerifySignature }

// Execute executes the command and returns the response.
func (cmd VerifySignature) Execute(t transport.TPM, s ...Session) (*VerifySignatureResponse, error) {
	var rsp VerifySignatureResponse
	if err := execute[VerifySignatureResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// VerifySignatureResponse is the response from TPM2_VerifySignature.
type VerifySignatureResponse struct {
	Validation TPMTTKVerified
}

// Sign is the input to TPM2_Sign.
// See definition in Part 3, Commands, section 20.2.
type Sign struct {
	// Handle of key that will perform signing
	KeyHandle handle `gotpm:"handle,auth"`
	// digest to be signed
	Digest TPM2BDigest
	// signing scheme to use if the scheme for keyHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme `gotpm:"nullable"`
	// proof that digest was created by the TPM.
	// If keyHandle is not a restricted signing key, then this
	// may be a NULL Ticket with tag = TPM_ST_CHECKHASH.
	Validation TPMTTKHashCheck
}

// Command implements the Command interface.
func (Sign) Command() TPMCC { return TPMCCSign }

// Execute executes the command and returns the response.
func (cmd Sign) Execute(t transport.TPM, s ...Session) (*SignResponse, error) {
	var rsp SignResponse
	if err := execute[SignResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// SignResponse is the response from TPM2_Sign.
type SignResponse struct {
	// the signature
	Signature TPMTSignature
}

// PCRExtend is the input to TPM2_PCR_Extend.
// See definition in Part 3, Commands, section 22.2
type PCRExtend struct {
	// handle of the PCR
	PCRHandle handle `gotpm:"handle,auth"`
	// list of tagged digest values to be extended
	Digests TPMLDigestValues
}

// Command implements the Command interface.
func (PCRExtend) Command() TPMCC { return TPMCCPCRExtend }

// Execute executes the command and returns the response.
func (cmd PCRExtend) Execute(t transport.TPM, s ...Session) (*PCRExtendResponse, error) {
	var rsp PCRExtendResponse
	err := execute[PCRExtendResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// PCRExtendResponse is the response from TPM2_PCR_Extend.
type PCRExtendResponse struct{}

// PCREvent is the input to TPM2_PCR_Event.
// See definition in Part 3, Commands, section 22.3
type PCREvent struct {
	// Handle of the PCR
	PCRHandle handle `gotpm:"handle,auth"`
	// Event data in sized buffer
	EventData TPM2BEvent
}

// Command implements the Command interface.
func (PCREvent) Command() TPMCC { return TPMCCPCREvent }

// Execute executes the command and returns the response.
func (cmd PCREvent) Execute(t transport.TPM, s ...Session) (*PCREventResponse, error) {
	var rsp PCREventResponse
	err := execute[PCREventResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// PCREventResponse is the response from TPM2_PCR_Event.
type PCREventResponse struct{}

// PCRRead is the input to TPM2_PCR_Read.
// See definition in Part 3, Commands, section 22.4
type PCRRead struct {
	// The selection of PCR to read
	PCRSelectionIn TPMLPCRSelection
}

// Command implements the Command interface.
func (PCRRead) Command() TPMCC { return TPMCCPCRRead }

// Execute executes the command and returns the response.
func (cmd PCRRead) Execute(t transport.TPM, s ...Session) (*PCRReadResponse, error) {
	var rsp PCRReadResponse
	if err := execute[PCRReadResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// PCRReadResponse is the response from TPM2_PCR_Read.
type PCRReadResponse struct {
	// the current value of the PCR update counter
	PCRUpdateCounter uint32
	// the PCR in the returned list
	PCRSelectionOut TPMLPCRSelection
	// the contents of the PCR indicated in pcrSelectOut-> pcrSelection[] as tagged digests
	PCRValues TPMLDigest
}

// PCRReset is the input to TPM2_PCRReset.
// See definition in Part 3, Commands, section 22.8.
type PCRReset struct {
	// the PCR to reset
	PCRHandle handle `gotpm:"handle,auth"`
}

// Command implements the Command interface.
func (PCRReset) Command() TPMCC { return TPMCCPCRReset }

// Execute executes the command and returns the response.
func (cmd PCRReset) Execute(t transport.TPM, s ...Session) (*PCRResetResponse, error) {
	var rsp PCRResetResponse
	if err := execute[PCRResetResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// PCRResetResponse is the response from TPM2_PCRReset.
type PCRResetResponse struct{}

// PolicySigned is the input to TPM2_PolicySigned.
// See definition in Part 3, Commands, section 23.3.
type PolicySigned struct {
	// handle for an entity providing the authorization
	AuthObject handle `gotpm:"handle"`
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// the policy nonce for the session
	NonceTPM TPM2BNonce
	// digest of the command parameters to which this authorization is limited
	CPHashA TPM2BDigest
	// a reference to a policy relating to the authorization – may be the Empty Buffer
	PolicyRef TPM2BNonce
	// time when authorization will expire, measured in seconds from the time
	// that nonceTPM was generated
	Expiration int32
	// signed authorization (not optional)
	Auth TPMTSignature
}

// Command implements the Command interface.
func (PolicySigned) Command() TPMCC { return TPMCCPolicySigned }

// Execute executes the command and returns the response.
func (cmd PolicySigned) Execute(t transport.TPM, s ...Session) (*PolicySignedResponse, error) {
	var rsp PolicySignedResponse
	if err := execute[PolicySignedResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// policyUpdate implements the PolicyUpdate helper for the several TPM policy
// commands as described in Part 3, 23.2.3.
func policyUpdate(policy *PolicyCalculator, cc TPMCC, arg2, arg3 []byte) error {
	if err := policy.Update(cc, arg2); err != nil {
		return err
	}
	return policy.Update(arg3)
}

// Update implements the PolicyCommand interface.
func (cmd PolicySigned) Update(policy *PolicyCalculator) error {
	return policyUpdate(policy, TPMCCPolicySigned, cmd.AuthObject.KnownName().Buffer, cmd.PolicyRef.Buffer)
}

// PolicySignedResponse is the response from TPM2_PolicySigned.
type PolicySignedResponse struct {
	// implementation-specific time value used to indicate to the TPM when the ticket expires
	Timeout TPM2BTimeout
	// produced if the command succeeds and expiration in the command was non-zero
	PolicyTicket TPMTTKAuth
}

// PolicySecret is the input to TPM2_PolicySecret.
// See definition in Part 3, Commands, section 23.4.
type PolicySecret struct {
	// handle for an entity providing the authorization
	AuthHandle handle `gotpm:"handle,auth"`
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// the policy nonce for the session
	NonceTPM TPM2BNonce
	// digest of the command parameters to which this authorization is limited
	CPHashA TPM2BDigest
	// a reference to a policy relating to the authorization – may be the Empty Buffer
	PolicyRef TPM2BNonce
	// time when authorization will expire, measured in seconds from the time
	// that nonceTPM was generated
	Expiration int32
}

// Command implements the Command interface.
func (PolicySecret) Command() TPMCC { return TPMCCPolicySecret }

// Execute executes the command and returns the response.
func (cmd PolicySecret) Execute(t transport.TPM, s ...Session) (*PolicySecretResponse, error) {
	var rsp PolicySecretResponse
	if err := execute[PolicySecretResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicySecret) Update(policy *PolicyCalculator) error {
	return policyUpdate(policy, TPMCCPolicySecret, cmd.AuthHandle.KnownName().Buffer, cmd.PolicyRef.Buffer)
}

// PolicySecretResponse is the response from TPM2_PolicySecret.
type PolicySecretResponse struct {
	// implementation-specific time value used to indicate to the TPM when the ticket expires
	Timeout TPM2BTimeout
	// produced if the command succeeds and expiration in the command was non-zero
	PolicyTicket TPMTTKAuth
}

// PolicyOr is the input to TPM2_PolicyOR.
// See definition in Part 3, Commands, section 23.6.
type PolicyOr struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// the list of hashes to check for a match
	PHashList TPMLDigest
}

// Command implements the Command interface.
func (PolicyOr) Command() TPMCC { return TPMCCPolicyOR }

// Execute executes the command and returns the response.
func (cmd PolicyOr) Execute(t transport.TPM, s ...Session) (*PolicyOrResponse, error) {
	var rsp PolicyOrResponse
	err := execute[PolicyOrResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyOr) Update(policy *PolicyCalculator) error {
	policy.Reset()
	var digests bytes.Buffer
	for _, digest := range cmd.PHashList.Digests {
		digests.Write(digest.Buffer)
	}
	return policy.Update(TPMCCPolicyOR, digests.Bytes())
}

// PolicyOrResponse is the response from TPM2_PolicyOr.
type PolicyOrResponse struct{}

// PolicyPCR is the input to TPM2_PolicyPCR.
// See definition in Part 3, Commands, section 23.7.
type PolicyPCR struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// expected digest value of the selected PCR using the
	// hash algorithm of the session; may be zero length
	PcrDigest TPM2BDigest
	// the PCR to include in the check digest
	Pcrs TPMLPCRSelection
}

// Command implements the Command interface.
func (PolicyPCR) Command() TPMCC { return TPMCCPolicyPCR }

// Execute executes the command and returns the response.
func (cmd PolicyPCR) Execute(t transport.TPM, s ...Session) (*PolicyPCRResponse, error) {
	var rsp PolicyPCRResponse
	err := execute[PolicyPCRResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyPCR) Update(policy *PolicyCalculator) error {
	return policy.Update(TPMCCPolicyPCR, cmd.Pcrs, cmd.PcrDigest.Buffer)
}

// PolicyPCRResponse is the response from TPM2_PolicyPCR.
type PolicyPCRResponse struct{}

// PolicyAuthValue is the input to TPM2_PolicyAuthValue.
// See definition in Part 3, Commands, section 23.17.
type PolicyAuthValue struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (PolicyAuthValue) Command() TPMCC { return TPMCCPolicyAuthValue }

// Execute executes the command and returns the response.
func (cmd PolicyAuthValue) Execute(t transport.TPM, s ...Session) (*PolicyAuthValueResponse, error) {
	var rsp PolicyAuthValueResponse
	err := execute[PolicyAuthValueResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyAuthValue interface.
func (cmd PolicyAuthValue) Update(policy *PolicyCalculator) error {
	return policy.Update(TPMCCPolicyAuthValue)
}

// PolicyAuthValueResponse is the response from TPM2_PolicyAuthValue.
type PolicyAuthValueResponse struct{}

// PolicyDuplicationSelect is the input to TPM2_PolicyDuplicationSelect.
// See definition in Part 3, Commands, section 23.15.
type PolicyDuplicationSelect struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	ObjectName    TPM2BName
	NewParentName TPM2BName
	IncludeObject TPMIYesNo
}

// Command implements the Command interface.
func (PolicyDuplicationSelect) Command() TPMCC { return TPMCCPolicyDuplicationSelect }

// Execute executes the command and returns the response.
func (cmd PolicyDuplicationSelect) Execute(t transport.TPM, s ...Session) (*PolicyDuplicationSelectResponse, error) {
	var rsp PolicyDuplicationSelectResponse
	err := execute[PolicyDuplicationSelectResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyDuplicationSelect interface.
func (cmd PolicyDuplicationSelect) Update(policy *PolicyCalculator) error {
	if cmd.IncludeObject {
		return policy.Update(TPMCCPolicyDuplicationSelect, cmd.ObjectName.Buffer, cmd.NewParentName.Buffer, cmd.IncludeObject)
	}
	return policy.Update(TPMCCPolicyDuplicationSelect, cmd.NewParentName.Buffer, cmd.IncludeObject)
}

// PolicyDuplicationSelectResponse is the response from TPM2_PolicyDuplicationSelect.
type PolicyDuplicationSelectResponse struct{}

// PolicyNV is the input to TPM2_PolicyNV.
// See definition in Part 3, Commands, section 23.9.
type PolicyNV struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV Index of the area to read
	NVIndex handle `gotpm:"handle"`
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// the second operand
	OperandB TPM2BOperand
	// the octet offset in the NV Index for the start of operand A
	Offset uint16
	// the comparison to make
	Operation TPMEO
}

// Command implements the Command interface.
func (PolicyNV) Command() TPMCC { return TPMCCPolicyNV }

// Execute executes the command and returns the response.
func (cmd PolicyNV) Execute(t transport.TPM, s ...Session) (*PolicyNVResponse, error) {
	var rsp PolicyNVResponse
	err := execute[PolicyNVResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyNV) Update(policy *PolicyCalculator) error {
	alg, err := policy.alg.Hash()
	if err != nil {
		return err
	}
	h := alg.New()
	h.Write(cmd.OperandB.Buffer)
	binary.Write(h, binary.BigEndian, cmd.Offset)
	binary.Write(h, binary.BigEndian, cmd.Operation)
	args := h.Sum(nil)
	return policy.Update(TPMCCPolicyNV, args, cmd.NVIndex.KnownName().Buffer)
}

// PolicyNVResponse is the response from TPM2_PolicyPCR.
type PolicyNVResponse struct{}

// PolicyCommandCode is the input to TPM2_PolicyCommandCode.
// See definition in Part 3, Commands, section 23.11.
type PolicyCommandCode struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// the allowed commandCode
	Code TPMCC
}

// Command implements the Command interface.
func (PolicyCommandCode) Command() TPMCC { return TPMCCPolicyCommandCode }

// Execute executes the command and returns the response.
func (cmd PolicyCommandCode) Execute(t transport.TPM, s ...Session) (*PolicyCommandCodeResponse, error) {
	var rsp PolicyCommandCodeResponse
	err := execute[PolicyCommandCodeResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyCommandCode) Update(policy *PolicyCalculator) error {
	return policy.Update(TPMCCPolicyCommandCode, cmd.Code)
}

// PolicyCommandCodeResponse is the response from TPM2_PolicyCommandCode.
type PolicyCommandCodeResponse struct{}

// PolicyCPHash is the input to TPM2_PolicyCpHash.
// See definition in Part 3, Commands, section 23.13.
type PolicyCPHash struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// the cpHash added to the policy
	CPHashA TPM2BDigest
}

// Command implements the Command interface.
func (PolicyCPHash) Command() TPMCC { return TPMCCPolicyCpHash }

// Execute executes the command and returns the response.
func (cmd PolicyCPHash) Execute(t transport.TPM, s ...Session) (*PolicyCPHashResponse, error) {
	var rsp PolicyCPHashResponse
	err := execute[PolicyCPHashResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyCPHash) Update(policy *PolicyCalculator) error {
	return policy.Update(TPMCCPolicyCpHash, cmd.CPHashA.Buffer)
}

// PolicyCPHashResponse is the response from TPM2_PolicyCpHash.
type PolicyCPHashResponse struct{}

// PolicyAuthorize is the input to TPM2_PolicySigned.
// See definition in Part 3, Commands, section 23.16.
type PolicyAuthorize struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// digest of the policy being approved
	ApprovedPolicy TPM2BDigest
	// a policy qualifier
	PolicyRef TPM2BDigest
	// Name of a key that can sign a policy addition
	KeySign TPM2BName
	// ticket validating that approvedPolicy and policyRef were signed by keySign
	CheckTicket TPMTTKVerified
}

// Command implements the Command interface.
func (PolicyAuthorize) Command() TPMCC { return TPMCCPolicyAuthorize }

// Execute executes the command and returns the response.
func (cmd PolicyAuthorize) Execute(t transport.TPM, s ...Session) (*PolicyAuthorizeResponse, error) {
	var rsp PolicyAuthorizeResponse
	err := execute[PolicyAuthorizeResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyAuthorize) Update(policy *PolicyCalculator) error {
	return policyUpdate(policy, TPMCCPolicyAuthorize, cmd.KeySign.Buffer, cmd.PolicyRef.Buffer)
}

// PolicyAuthorizeResponse is the response from TPM2_PolicyAuthorize.
type PolicyAuthorizeResponse struct{}

// PolicyGetDigest is the input to TPM2_PolicyGetDigest.
// See definition in Part 3, Commands, section 23.19.
type PolicyGetDigest struct {
	// handle for the policy session
	PolicySession handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (PolicyGetDigest) Command() TPMCC { return TPMCCPolicyGetDigest }

// Execute executes the command and returns the response.
func (cmd PolicyGetDigest) Execute(t transport.TPM, s ...Session) (*PolicyGetDigestResponse, error) {
	var rsp PolicyGetDigestResponse
	if err := execute[PolicyGetDigestResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// PolicyGetDigestResponse is the response from TPM2_PolicyGetDigest.
type PolicyGetDigestResponse struct {
	// the current value of the policySession→policyDigest
	PolicyDigest TPM2BDigest
}

// PolicyNVWritten is the input to TPM2_PolicyNvWritten.
// See definition in Part 3, Commands, section 23.20.
type PolicyNVWritten struct {
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
	// YES if NV Index is required to have been written
	// NO if NV Index is required not to have been written
	WrittenSet TPMIYesNo
}

// Command implements the Command interface.
func (PolicyNVWritten) Command() TPMCC { return TPMCCPolicyNvWritten }

// Execute executes the command and returns the response.
func (cmd PolicyNVWritten) Execute(t transport.TPM, s ...Session) (*PolicyNVWrittenResponse, error) {
	var rsp PolicyNVWrittenResponse
	if err := execute[PolicyNVWrittenResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyNVWritten) Update(policy *PolicyCalculator) error {
	return policy.Update(TPMCCPolicyNvWritten, cmd.WrittenSet)
}

// PolicyNVWrittenResponse is the response from TPM2_PolicyNvWritten.
type PolicyNVWrittenResponse struct {
}

// PolicyAuthorizeNV is the input to TPM2_PolicyAuthorizeNV.
// See definition in Part 3, Commands, section 23.22.
type PolicyAuthorizeNV struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV Index of the area to read
	NVIndex handle `gotpm:"handle"`
	// handle for the policy session being extended
	PolicySession handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (PolicyAuthorizeNV) Command() TPMCC { return TPMCCPolicyAuthorizeNV }

// Execute executes the command and returns the response.
func (cmd PolicyAuthorizeNV) Execute(t transport.TPM, s ...Session) (*PolicyAuthorizeNVResponse, error) {
	var rsp PolicyAuthorizeNVResponse
	err := execute[PolicyAuthorizeNVResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Update implements the PolicyCommand interface.
func (cmd PolicyAuthorizeNV) Update(policy *PolicyCalculator) error {
	policy.Reset()
	return policy.Update(TPMCCPolicyAuthorizeNV, cmd.NVIndex.KnownName().Buffer)
}

// PolicyAuthorizeNVResponse is the response from TPM2_PolicyAuthorizeNV.
type PolicyAuthorizeNVResponse struct{}

// CreatePrimary is the input to TPM2_CreatePrimary.
// See definition in Part 3, Commands, section 24.1
type CreatePrimary struct {
	// TPM_RH_ENDORSEMENT, TPM_RH_OWNER, TPM_RH_PLATFORM+{PP},
	// TPM_RH_NULL, TPM_RH_FW_ENDORSEMENT, TPM_RH_FW_OWNER
	// TPM_RH_FW_PLATFORM+{PP}  or TPM_RH_FW_NULL
	PrimaryHandle handle `gotpm:"handle,auth"`
	// the sensitive data
	InSensitive TPM2BSensitiveCreate
	// the public template
	InPublic TPM2BPublic
	// data that will be included in the creation data for this
	// object to provide permanent, verifiable linkage between this
	// object and some object owner data
	OutsideInfo TPM2BData
	// PCR that will be used in creation data
	CreationPCR TPMLPCRSelection
}

// Command implements the Command interface.
func (CreatePrimary) Command() TPMCC { return TPMCCCreatePrimary }

// Execute executes the command and returns the response.
func (cmd CreatePrimary) Execute(t transport.TPM, s ...Session) (*CreatePrimaryResponse, error) {
	var rsp CreatePrimaryResponse
	if err := execute[CreatePrimaryResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// CreatePrimaryResponse is the response from TPM2_CreatePrimary.
type CreatePrimaryResponse struct {
	// handle of type TPM_HT_TRANSIENT for created Primary Object
	ObjectHandle TPMHandle `gotpm:"handle"`
	// the public portion of the created object
	OutPublic TPM2BPublic
	// contains a TPMS_CREATION_DATA
	CreationData tpm2bCreationData
	// digest of creationData using nameAlg of outPublic
	CreationHash TPM2BDigest
	// ticket used by TPM2_CertifyCreation() to validate that the
	// creation data was produced by the TPM.
	CreationTicket TPMTTKCreation
	// the name of the created object
	Name TPM2BName
}

// Clear is the input to TPM2_Clear.
// See definition in Part 3, Commands, section 24.6
type Clear struct {
	// TPM_RH_LOCKOUT or TPM_RH_PLATFORM+{PP}
	AuthHandle handle `gotpm:"handle,auth"`
}

// Command implements the Command interface.
func (Clear) Command() TPMCC { return TPMCCClear }

// Execute executes the command and returns the response.
func (cmd Clear) Execute(t transport.TPM, s ...Session) (*ClearResponse, error) {
	var rsp ClearResponse
	err := execute[ClearResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ClearResponse is the response from TPM2_Clear.
type ClearResponse struct{}

// HierarchyChangeAuth is the input to TPM2_HierarchyChangeAuth.
// See definition in Part 3, Commands, section 24.8
type HierarchyChangeAuth struct {
	// TPM_RH_ENDORSEMENT, TPM_RH_LOCKOUT, TPM_RH_OWNER or TPM_RH_PLATFORM+{PP}
	AuthHandle handle `gotpm:"handle,auth"`
	// new authorization value
	NewAuth TPM2BAuth
}

// Command implements the Command interface.
func (HierarchyChangeAuth) Command() TPMCC { return TPMCCHierarchyChanegAuth }

// Execute executes the command and returns the response.
func (cmd HierarchyChangeAuth) Execute(t transport.TPM, s ...Session) (*HierarchyChangeAuthResponse, error) {
	var rsp HierarchyChangeAuthResponse
	if err := execute[HierarchyChangeAuthResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// HierarchyChangeAuthResponse is the response from TPM2_HierarchyChangeAuth.
type HierarchyChangeAuthResponse struct{}

// ContextSave is the input to TPM2_ContextSave.
// See definition in Part 3, Commands, section 28.2
type ContextSave struct {
	// handle of the resource to save
	SaveHandle TPMIDHContext
}

// Command implements the Command interface.
func (ContextSave) Command() TPMCC { return TPMCCContextSave }

// Execute executes the command and returns the response.
func (cmd ContextSave) Execute(t transport.TPM, s ...Session) (*ContextSaveResponse, error) {
	var rsp ContextSaveResponse
	if err := execute[ContextSaveResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ContextSaveResponse is the response from TPM2_ContextSave.
type ContextSaveResponse struct {
	Context TPMSContext
}

// ContextLoad is the input to TPM2_ContextLoad.
// See definition in Part 3, Commands, section 28.3
type ContextLoad struct {
	// the context blob
	Context TPMSContext
}

// Command implements the Command interface.
func (ContextLoad) Command() TPMCC { return TPMCCContextLoad }

// Execute executes the command and returns the response.
func (cmd ContextLoad) Execute(t transport.TPM, s ...Session) (*ContextLoadResponse, error) {
	var rsp ContextLoadResponse
	if err := execute[ContextLoadResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ContextLoadResponse is the response from TPM2_ContextLoad.
type ContextLoadResponse struct {
	// the handle assigned to the resource after it has been successfully loaded
	LoadedHandle TPMIDHContext
}

// FlushContext is the input to TPM2_FlushContext.
// See definition in Part 3, Commands, section 28.4
type FlushContext struct {
	// the handle of the item to flush
	FlushHandle handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (FlushContext) Command() TPMCC { return TPMCCFlushContext }

// Execute executes the command and returns the response.
func (cmd FlushContext) Execute(t transport.TPM, s ...Session) (*FlushContextResponse, error) {
	var rsp FlushContextResponse
	err := execute[FlushContextResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// FlushContextResponse is the response from TPM2_FlushContext.
type FlushContextResponse struct{}

// EvictControl is the input to TPM2_EvictControl.
// See definition in Part 3, Commands, section 28.5
type EvictControl struct {
	// TPM_RH_OWNER or TPM_RH_PLATFORM+{PP}
	Auth             handle `gotpm:"handle,auth"`
	ObjectHandle     handle `gotpm:"handle"`
	PersistentHandle TPMIDHPersistent
}

// EvictControlResponse is the response from TPM2_EvictControl.
type EvictControlResponse struct{}

// Command implements the Command interface.
func (EvictControl) Command() TPMCC { return TPMCCEvictControl }

// Execute executes the command and returns the response.
func (cmd EvictControl) Execute(t transport.TPM, s ...Session) (*EvictControlResponse, error) {
	var rsp EvictControlResponse
	if err := execute[EvictControlResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Duplicate is the input to TPM2_Duplicate.
// See definition in Part 3, Commands, section 13.1
type Duplicate struct {
	// ObjectHandle is the handle of the object to dupliate.
	ObjectHandle handle `gotpm:"handle,auth"`

	// NewParentHandle is the handle of the new parent.
	NewParentHandle handle `gotpm:"handle"`

	// EncryptionKeyIn is the optional symmetric encryption key used as the
	// inner wrapper. If SymmetricAlg is TPM_ALG_NULL, then this parameter
	// shall be the Empty Buffer.
	EncryptionKeyIn TPM2BData

	// Definition of the symmetric algorithm to use for the inner wrapper.
	// It may be TPM_ALG_NULL if no inner wrapper is applied.
	Symmetric TPMTSymDef
}

// DuplicateResponse is the response from TPM2_Duplicate.
type DuplicateResponse struct {
	// EncryptionKeyOut is the symmetric encryption key used as the
	// inner wrapper. If SymmetricAlg is TPM_ALG_NULL, this value
	// shall be the Empty Buffer.
	EncryptionKeyOut TPM2BData

	// Duplicate is the private area of the object. It may be encrypted by
	// EncryptionKeyIn and may be doubly encrypted.
	Duplicate TPM2BPrivate

	// OutSymSeed is the seed protected by the asymmetric algorithms of new
	// parent.
	OutSymSeed TPM2BEncryptedSecret
}

// Command implements the Command interface.
func (Duplicate) Command() TPMCC { return TPMCCDuplicate }

// Execute executes the command and returns the response.
func (cmd Duplicate) Execute(t transport.TPM, s ...Session) (*DuplicateResponse, error) {
	var rsp DuplicateResponse
	if err := execute[DuplicateResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// Import is the input to TPM2_Import.
// See definition in Part 3, Commands, section 13.3
type Import struct {
	// handle of parent for new object
	ParentHandle handle `gotpm:"handle,auth"`

	// The optional symmetric encryption key used as the inner wrapper for duplicate
	// If SymmetricAlg is TPM_ALG_NULL, then this parametert shall be the Empty Buffer
	EncryptionKey TPM2BData

	// The public area of the object to be imported
	ObjectPublic TPM2BPublic

	// The symmetrically encrypted duplicate object that may contain an inner
	// symmetric wrapper
	Duplicate TPM2BPrivate

	// The seed for the symmetric key and HMAC key
	InSymSeed TPM2BEncryptedSecret

	// Definition of the symmetric algorithm to use for the inner wrapper
	Symmetric TPMTSymDef
}

// ImportResponse is the response from TPM2_Import.
type ImportResponse struct {
	// the private portion of the object
	OutPrivate TPM2BPrivate
}

// Command implements the Command interface.
func (Import) Command() TPMCC { return TPMCCImport }

// Execute executes the command and returns the response.
func (cmd Import) Execute(t transport.TPM, s ...Session) (*ImportResponse, error) {
	var rsp ImportResponse
	if err := execute[ImportResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ReadClock is the input to TPM2_ReadClock.
// See definition in Part 3, Commands, section 29.1
type ReadClock struct{}

// Command implements the Command interface.
func (ReadClock) Command() TPMCC { return TPMCCReadClock }

// Execute executes the command and returns the response.
func (cmd ReadClock) Execute(t transport.TPM, s ...Session) (*ReadClockResponse, error) {
	var rsp ReadClockResponse
	if err := execute[ReadClockResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// ReadClockResponse is the response from TPM2_ReadClock.
type ReadClockResponse struct {
	CurrentTime TPMSTimeInfo
}

// GetCapability is the input to TPM2_GetCapability.
// See definition in Part 3, Commands, section 30.2
type GetCapability struct {
	// group selection; determines the format of the response
	Capability TPMCap
	// further definition of information
	Property uint32
	// number of properties of the indicated type to return
	PropertyCount uint32
}

// Command implements the Command interface.
func (GetCapability) Command() TPMCC { return TPMCCGetCapability }

// Execute executes the command and returns the response.
func (cmd GetCapability) Execute(t transport.TPM, s ...Session) (*GetCapabilityResponse, error) {
	var rsp GetCapabilityResponse
	if err := execute[GetCapabilityResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// GetCapabilityResponse is the response from TPM2_GetCapability.
type GetCapabilityResponse struct {
	// flag to indicate if there are more values of this type
	MoreData TPMIYesNo
	// the capability data
	CapabilityData TPMSCapabilityData
}

// TestParms is the input to TPM2_TestParms.
// See definition in Part 3, Commands, section 30.3
type TestParms struct {
	// Algorithms parameters to be validates
	Parameters TPMTPublicParms
}

// Command implements the Command interface.
func (TestParms) Command() TPMCC { return TPMCCTestParms }

// Execute executes the command and returns the response.
func (cmd TestParms) Execute(t transport.TPM, s ...Session) (*TestParmsResponse, error) {
	var rsp TestParmsResponse
	if err := execute[TestParmsResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// TestParmsResponse is the response from TPM2_TestParms.
type TestParmsResponse struct{}

// NVDefineSpace is the input to TPM2_NV_DefineSpace.
// See definition in Part 3, Commands, section 31.3.
type NVDefineSpace struct {
	// TPM_RH_OWNER or TPM_RH_PLATFORM+{PP}
	AuthHandle handle `gotpm:"handle,auth"`
	// the authorization value
	Auth TPM2BAuth
	// the public parameters of the NV area
	PublicInfo TPM2BNVPublic
}

// Command implements the Command interface.
func (NVDefineSpace) Command() TPMCC { return TPMCCNVDefineSpace }

// Execute executes the command and returns the response.
func (cmd NVDefineSpace) Execute(t transport.TPM, s ...Session) (*NVDefineSpaceResponse, error) {
	var rsp NVDefineSpaceResponse
	err := execute[NVDefineSpaceResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVDefineSpaceResponse is the response from TPM2_NV_DefineSpace.
type NVDefineSpaceResponse struct{}

// NVUndefineSpace is the input to TPM2_NV_UndefineSpace.
// See definition in Part 3, Commands, section 31.4.
type NVUndefineSpace struct {
	// TPM_RH_OWNER or TPM_RH_PLATFORM+{PP}
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV Index to remove from NV space
	NVIndex handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (NVUndefineSpace) Command() TPMCC { return TPMCCNVUndefineSpace }

// Execute executes the command and returns the response.
func (cmd NVUndefineSpace) Execute(t transport.TPM, s ...Session) (*NVUndefineSpaceResponse, error) {
	var rsp NVUndefineSpaceResponse
	err := execute[NVUndefineSpaceResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVUndefineSpaceResponse is the response from TPM2_NV_UndefineSpace.
type NVUndefineSpaceResponse struct{}

// NVUndefineSpaceSpecial is the input to TPM2_NV_UndefineSpaceSpecial.
// See definition in Part 3, Commands, section 31.5.
type NVUndefineSpaceSpecial struct {
	// Index to be deleted
	NVIndex handle `gotpm:"handle,auth"`
	// TPM_RH_PLATFORM+{PP}
	Platform handle `gotpm:"handle,auth"`
}

// Command implements the Command interface.
func (NVUndefineSpaceSpecial) Command() TPMCC { return TPMCCNVUndefineSpaceSpecial }

// Execute executes the command and returns the response.
func (cmd NVUndefineSpaceSpecial) Execute(t transport.TPM, s ...Session) (*NVUndefineSpaceSpecialResponse, error) {
	var rsp NVUndefineSpaceSpecialResponse
	err := execute[NVUndefineSpaceSpecialResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVUndefineSpaceSpecialResponse is the response from TPM2_NV_UndefineSpaceSpecial.
type NVUndefineSpaceSpecialResponse struct{}

// NVReadPublic is the input to TPM2_NV_ReadPublic.
// See definition in Part 3, Commands, section 31.6.
type NVReadPublic struct {
	// the NV index
	NVIndex handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (NVReadPublic) Command() TPMCC { return TPMCCNVReadPublic }

// Execute executes the command and returns the response.
func (cmd NVReadPublic) Execute(t transport.TPM, s ...Session) (*NVReadPublicResponse, error) {
	var rsp NVReadPublicResponse
	if err := execute[NVReadPublicResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVReadPublicResponse is the response from TPM2_NV_ReadPublic.
type NVReadPublicResponse struct {
	NVPublic TPM2BNVPublic
	NVName   TPM2BName
}

// NVWrite is the input to TPM2_NV_Write.
// See definition in Part 3, Commands, section 31.7.
type NVWrite struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV index of the area to write
	NVIndex handle `gotpm:"handle"`
	// the data to write
	Data TPM2BMaxNVBuffer
	// the octet offset into the NV Area
	Offset uint16
}

// Command implements the Command interface.
func (NVWrite) Command() TPMCC { return TPMCCNVWrite }

// Execute executes the command and returns the response.
func (cmd NVWrite) Execute(t transport.TPM, s ...Session) (*NVWriteResponse, error) {
	var rsp NVWriteResponse
	err := execute[NVWriteResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVWriteResponse is the response from TPM2_NV_Write.
type NVWriteResponse struct{}

// NVIncrement is the input to TPM2_NV_Increment.
// See definition in Part 3, Commands, section 31.8.
type NVIncrement struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV index of the area to write
	NVIndex handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (NVIncrement) Command() TPMCC { return TPMCCNVIncrement }

// Execute executes the command and returns the response.
func (cmd NVIncrement) Execute(t transport.TPM, s ...Session) (*NVIncrementResponse, error) {
	var rsp NVIncrementResponse
	err := execute[NVIncrementResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVIncrementResponse is the response from TPM2_NV_Increment.
type NVIncrementResponse struct{}

// NVWriteLock is the input to TPM2_NV_WriteLock.
// See definition in Part 3, Commands, section 31.11.
type NVWriteLock struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV index of the area to lock
	NVIndex handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (NVWriteLock) Command() TPMCC { return TPMCCNVWriteLock }

// Execute executes the command and returns the response.
func (cmd NVWriteLock) Execute(t transport.TPM, s ...Session) (*NVWriteLockResponse, error) {
	var rsp NVWriteLockResponse
	err := execute[NVWriteLockResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVWriteLockResponse is the response from TPM2_NV_WriteLock.
type NVWriteLockResponse struct{}

// NVRead is the input to TPM2_NV_Read.
// See definition in Part 3, Commands, section 31.13.
type NVRead struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV index to read
	NVIndex handle `gotpm:"handle"`
	// number of octets to read
	Size uint16
	// octet offset into the NV area
	Offset uint16
}

// Command implements the Command interface.
func (NVRead) Command() TPMCC { return TPMCCNVRead }

// Execute executes the command and returns the response.
func (cmd NVRead) Execute(t transport.TPM, s ...Session) (*NVReadResponse, error) {
	var rsp NVReadResponse
	if err := execute[NVReadResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVReadResponse is the response from TPM2_NV_Read.
type NVReadResponse struct {
	// the data read
	Data TPM2BMaxNVBuffer
}

// NVReadLock is the input to TPM2_NV_NVReadLock.
// See definition in Part 3, Commands, section 31.14.
type NVReadLock struct {
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// the NV index of the area to lock
	NVIndex handle `gotpm:"handle"`
}

// Command implements the Command interface.
func (NVReadLock) Command() TPMCC { return TPMCCNVReadLock }

// Execute executes the command and returns the response.
func (cmd NVReadLock) Execute(t transport.TPM, s ...Session) (*NVReadLockResponse, error) {
	var rsp NVReadLockResponse
	err := execute[NVReadLockResponse](t, cmd, &rsp, s...)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVReadLockResponse is the response from TPM2_NV_ReadLock.
type NVReadLockResponse struct{}

// NVCertify is the input to TPM2_NV_Certify.
// See definition in Part 3, Commands, section 31.16.
type NVCertify struct {
	// handle of the key used to sign the attestation structure
	SignHandle handle `gotpm:"handle,auth"`
	// handle indicating the source of the authorization value
	AuthHandle handle `gotpm:"handle,auth"`
	// Index for the area to be certified
	NVIndex handle `gotpm:"handle"`
	// user-provided qualifying data
	QualifyingData TPM2BData
	// signing scheme to use if the scheme for signHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme `gotpm:"nullable"`
	// number of octets to certify
	Size uint16
	// octet offset into the NV area
	Offset uint16
}

// Command implements the Command interface.
func (NVCertify) Command() TPMCC { return TPMCCNVCertify }

// Execute executes the command and returns the response.
func (cmd NVCertify) Execute(t transport.TPM, s ...Session) (*NVCertifyResponse, error) {
	var rsp NVCertifyResponse
	if err := execute[NVCertifyResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NVCertifyResponse is the response from TPM2_NV_Read.
type NVCertifyResponse struct {
	// the structure that was signed
	CertifyInfo TPM2BAttest
	// the asymmetric signature over certifyInfo using the key referenced by signHandle
	Signature TPMTSignature
}

// GetTime is the input to TPM2_GetTime.
// See definition in Part 3, Commands, section 18.7.
type GetTime struct {
	// handle of the privacy administrator (Must be the value TPM_RH_ENDORSEMENT or command will fail)
	PrivacyAdminHandle TPMIRHEndorsement `gotpm:"handle,auth"`
	// the keyHandle identifier of a loaded key that can perform digital signatures
	SignHandle handle `gotpm:"handle,auth"`
	// data to "tick stamp"
	QualifyingData TPM2BData
	// signing scheme to use if the scheme for signHandle is TPM_ALG_NULL
	InScheme TPMTSigScheme `gotpm:"nullable"`
}

// Command implements the Command interface.
func (GetTime) Command() TPMCC { return TPMCCGetTime }

// Execute executes the command and returns the response.
func (cmd GetTime) Execute(t transport.TPM, s ...Session) (*GetTimeResponse, error) {
	var rsp GetTimeResponse
	if err := execute[GetTimeResponse](t, cmd, &rsp, s...); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// GetTimeResponse is the response from TPM2_GetTime.
type GetTimeResponse struct {
	// standard TPM-generated attestation block
	TimeInfo TPM2BAttest
	// the signature over timeInfo
	Signature TPMTSignature
}
