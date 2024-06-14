// Copyright (c) 2018, Google LLC All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tpm2 supports direct communication with a TPM 2.0 device under Linux.
package tpm2

import (
	"bytes"
	"crypto"
	"fmt"
	"io"

	"github.com/google/go-tpm/tpmutil"
)

// GetRandom gets random bytes from the TPM.
func GetRandom(rw io.ReadWriter, size uint16) ([]byte, error) {
	resp, err := runCommand(rw, TagNoSessions, CmdGetRandom, size)
	if err != nil {
		return nil, err
	}

	var randBytes tpmutil.U16Bytes
	if _, err := tpmutil.Unpack(resp, &randBytes); err != nil {
		return nil, err
	}
	return randBytes, nil
}

// FlushContext removes an object or session under handle to be removed from
// the TPM. This must be called for any loaded handle to avoid out-of-memory
// errors in TPM.
func FlushContext(rw io.ReadWriter, handle tpmutil.Handle) error {
	_, err := runCommand(rw, TagNoSessions, CmdFlushContext, handle)
	return err
}

func encodeTPMLPCRSelection(sel ...PCRSelection) ([]byte, error) {
	if len(sel) == 0 {
		return tpmutil.Pack(uint32(0))
	}

	// PCR selection is a variable-size bitmask, where position of a set bit is
	// the selected PCR index.
	// Size of the bitmask in bytes is pre-pended. It should be at least
	// sizeOfPCRSelect.
	//
	// For example, selecting PCRs 3 and 9 looks like:
	// size(3)  mask     mask     mask
	// 00000011 00000000 00000001 00000100
	var retBytes []byte
	for _, s := range sel {
		if len(s.PCRs) == 0 {
			return tpmutil.Pack(uint32(0))
		}

		ts := tpmsPCRSelection{
			Hash: s.Hash,
			Size: sizeOfPCRSelect,
			PCRs: make(tpmutil.RawBytes, sizeOfPCRSelect),
		}

		// s[i].PCRs parameter is indexes of PCRs, convert that to set bits.
		for _, n := range s.PCRs {
			if n >= 8*sizeOfPCRSelect {
				return nil, fmt.Errorf("PCR index %d is out of range (exceeds maximum value %d)", n, 8*sizeOfPCRSelect-1)
			}
			byteNum := n / 8
			bytePos := byte(1 << byte(n%8))
			ts.PCRs[byteNum] |= bytePos
		}

		tmpBytes, err := tpmutil.Pack(ts)
		if err != nil {
			return nil, err
		}

		retBytes = append(retBytes, tmpBytes...)
	}
	tmpSize, err := tpmutil.Pack(uint32(len(sel)))
	if err != nil {
		return nil, err
	}
	retBytes = append(tmpSize, retBytes...)

	return retBytes, nil
}

func decodeTPMLPCRSelection(buf *bytes.Buffer) ([]PCRSelection, error) {
	var count uint32
	var sel []PCRSelection

	// This unpacks buffer which is of type TPMLPCRSelection
	// and returns the count of TPMSPCRSelections.
	if err := tpmutil.UnpackBuf(buf, &count); err != nil {
		return sel, err
	}

	var ts tpmsPCRSelection
	for i := 0; i < int(count); i++ {
		var s PCRSelection
		if err := tpmutil.UnpackBuf(buf, &ts.Hash, &ts.Size); err != nil {
			return sel, err
		}
		ts.PCRs = make(tpmutil.RawBytes, ts.Size)
		if _, err := buf.Read(ts.PCRs); err != nil {
			return sel, err
		}
		s.Hash = ts.Hash
		for j := 0; j < int(ts.Size); j++ {
			for k := 0; k < 8; k++ {
				set := ts.PCRs[j] & byte(1<<byte(k))
				if set == 0 {
					continue
				}
				s.PCRs = append(s.PCRs, 8*j+k)
			}
		}
		sel = append(sel, s)
	}
	if len(sel) == 0 {
		sel = append(sel, PCRSelection{
			Hash: AlgUnknown,
		})
	}
	return sel, nil
}

func decodeOneTPMLPCRSelection(buf *bytes.Buffer) (PCRSelection, error) {
	sels, err := decodeTPMLPCRSelection(buf)
	if err != nil {
		return PCRSelection{}, err
	}
	if len(sels) != 1 {
		return PCRSelection{}, fmt.Errorf("got %d TPMS_PCR_SELECTION items in TPML_PCR_SELECTION, expected 1", len(sels))
	}
	return sels[0], nil
}

func decodeReadPCRs(in []byte) (map[int][]byte, error) {
	buf := bytes.NewBuffer(in)
	var updateCounter uint32
	if err := tpmutil.UnpackBuf(buf, &updateCounter); err != nil {
		return nil, err
	}

	sel, err := decodeOneTPMLPCRSelection(buf)
	if err != nil {
		return nil, err
	}

	var digestCount uint32
	if err = tpmutil.UnpackBuf(buf, &digestCount); err != nil {
		return nil, fmt.Errorf("decoding TPML_DIGEST length: %v", err)
	}
	if int(digestCount) != len(sel.PCRs) {
		return nil, fmt.Errorf("received %d PCRs but %d digests", len(sel.PCRs), digestCount)
	}

	vals := make(map[int][]byte)
	for _, pcr := range sel.PCRs {
		var val tpmutil.U16Bytes
		if err = tpmutil.UnpackBuf(buf, &val); err != nil {
			return nil, fmt.Errorf("decoding TPML_DIGEST item: %v", err)
		}
		vals[pcr] = val
	}
	return vals, nil
}

// ReadPCRs reads PCR values from the TPM.
// This is only a wrapper over TPM2_PCR_Read() call, thus can only return
// at most 8 PCRs digests.
func ReadPCRs(rw io.ReadWriter, sel PCRSelection) (map[int][]byte, error) {
	Cmd, err := encodeTPMLPCRSelection(sel)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdPCRRead, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}

	return decodeReadPCRs(resp)
}

func decodeReadClock(in []byte) (uint64, uint64, error) {
	var curTime, curClock uint64

	if _, err := tpmutil.Unpack(in, &curTime, &curClock); err != nil {
		return 0, 0, err
	}
	return curTime, curClock, nil
}

// ReadClock returns current clock values from the TPM.
//
// First return value is time in milliseconds since TPM was initialized (since
// system startup).
//
// Second return value is time in milliseconds since TPM reset (since Storage
// Primary Seed is changed).
func ReadClock(rw io.ReadWriter) (uint64, uint64, error) {
	resp, err := runCommand(rw, TagNoSessions, CmdReadClock)
	if err != nil {
		return 0, 0, err
	}
	return decodeReadClock(resp)
}

func decodeGetCapability(in []byte) ([]interface{}, bool, error) {
	var moreData byte
	var capReported Capability

	buf := bytes.NewBuffer(in)
	if err := tpmutil.UnpackBuf(buf, &moreData, &capReported); err != nil {
		return nil, false, err
	}

	switch capReported {
	case CapabilityHandles:
		var numHandles uint32
		if err := tpmutil.UnpackBuf(buf, &numHandles); err != nil {
			return nil, false, fmt.Errorf("could not unpack handle count: %v", err)
		}

		var handles []interface{}
		for i := 0; i < int(numHandles); i++ {
			var handle tpmutil.Handle
			if err := tpmutil.UnpackBuf(buf, &handle); err != nil {
				return nil, false, fmt.Errorf("could not unpack handle: %v", err)
			}
			handles = append(handles, handle)
		}
		return handles, moreData > 0, nil
	case CapabilityAlgs:
		var numAlgs uint32
		if err := tpmutil.UnpackBuf(buf, &numAlgs); err != nil {
			return nil, false, fmt.Errorf("could not unpack algorithm count: %v", err)
		}

		var algs []interface{}
		for i := 0; i < int(numAlgs); i++ {
			var alg AlgorithmDescription
			if err := tpmutil.UnpackBuf(buf, &alg); err != nil {
				return nil, false, fmt.Errorf("could not unpack algorithm description: %v", err)
			}
			algs = append(algs, alg)
		}
		return algs, moreData > 0, nil
	case CapabilityTPMProperties:
		var numProps uint32
		if err := tpmutil.UnpackBuf(buf, &numProps); err != nil {
			return nil, false, fmt.Errorf("could not unpack fixed properties count: %v", err)
		}

		var props []interface{}
		for i := 0; i < int(numProps); i++ {
			var prop TaggedProperty
			if err := tpmutil.UnpackBuf(buf, &prop); err != nil {
				return nil, false, fmt.Errorf("could not unpack tagged property: %v", err)
			}
			props = append(props, prop)
		}
		return props, moreData > 0, nil

	case CapabilityPCRs:
		var pcrss []interface{}
		pcrs, err := decodeTPMLPCRSelection(buf)
		if err != nil {
			return nil, false, fmt.Errorf("could not unpack pcr selection: %v", err)
		}
		for i := 0; i < len(pcrs); i++ {
			pcrss = append(pcrss, pcrs[i])
		}

		return pcrss, moreData > 0, nil

	default:
		return nil, false, fmt.Errorf("unsupported capability %v", capReported)
	}
}

// GetCapability returns various information about the TPM state.
//
// Currently only CapabilityHandles (list active handles) and CapabilityAlgs
// (list supported algorithms) are supported. CapabilityHandles will return
// a []tpmutil.Handle for vals, CapabilityAlgs will return
// []AlgorithmDescription.
//
// moreData is true if the TPM indicated that more data is available. Follow
// the spec for the capability in question on how to query for more data.
func GetCapability(rw io.ReadWriter, capa Capability, count, property uint32) (vals []interface{}, moreData bool, err error) {
	resp, err := runCommand(rw, TagNoSessions, CmdGetCapability, capa, property, count)
	if err != nil {
		return nil, false, err
	}
	return decodeGetCapability(resp)
}

// GetManufacturer returns the manufacturer ID
func GetManufacturer(rw io.ReadWriter) ([]byte, error) {
	caps, _, err := GetCapability(rw, CapabilityTPMProperties, 1, uint32(Manufacturer))
	if err != nil {
		return nil, err
	}

	prop := caps[0].(TaggedProperty)
	return tpmutil.Pack(prop.Value)
}

func encodeAuthArea(sections ...AuthCommand) ([]byte, error) {
	var res tpmutil.RawBytes
	for _, s := range sections {
		buf, err := tpmutil.Pack(s)
		if err != nil {
			return nil, err
		}
		res = append(res, buf...)
	}

	size, err := tpmutil.Pack(uint32(len(res)))
	if err != nil {
		return nil, err
	}

	return concat(size, res)
}

func encodePCREvent(pcr tpmutil.Handle, eventData []byte) ([]byte, error) {
	ha, err := tpmutil.Pack(pcr)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: EmptyAuth})
	if err != nil {
		return nil, err
	}
	event, err := tpmutil.Pack(tpmutil.U16Bytes(eventData))
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, event)
}

// PCREvent writes an update to the specified PCR.
func PCREvent(rw io.ReadWriter, pcr tpmutil.Handle, eventData []byte) error {
	Cmd, err := encodePCREvent(pcr, eventData)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdPCREvent, tpmutil.RawBytes(Cmd))
	return err
}

func encodeSensitiveArea(s tpmsSensitiveCreate) ([]byte, error) {
	// TPMS_SENSITIVE_CREATE
	buf, err := tpmutil.Pack(s)
	if err != nil {
		return nil, err
	}
	// TPM2B_SENSITIVE_CREATE
	return tpmutil.Pack(tpmutil.U16Bytes(buf))
}

// encodeCreate works for both TPM2_Create and TPM2_CreatePrimary.
func encodeCreate(owner tpmutil.Handle, sel PCRSelection, auth AuthCommand, ownerPassword string, sensitiveData []byte, pub Public, outsideInfo []byte) ([]byte, error) {
	parent, err := tpmutil.Pack(owner)
	if err != nil {
		return nil, err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return nil, err
	}
	inSensitive, err := encodeSensitiveArea(tpmsSensitiveCreate{
		UserAuth: []byte(ownerPassword),
		Data:     sensitiveData,
	})
	if err != nil {
		return nil, err
	}
	inPublic, err := pub.Encode()
	if err != nil {
		return nil, err
	}
	publicBlob, err := tpmutil.Pack(tpmutil.U16Bytes(inPublic))
	if err != nil {
		return nil, err
	}
	outsideInfoBlob, err := tpmutil.Pack(tpmutil.U16Bytes(outsideInfo))
	if err != nil {
		return nil, err
	}
	creationPCR, err := encodeTPMLPCRSelection(sel)
	if err != nil {
		return nil, err
	}
	return concat(
		parent,
		encodedAuth,
		inSensitive,
		publicBlob,
		outsideInfoBlob,
		creationPCR,
	)
}

func decodeCreatePrimary(in []byte) (handle tpmutil.Handle, public, creationData, creationHash tpmutil.U16Bytes, ticket Ticket, creationName tpmutil.U16Bytes, err error) {
	var paramSize uint32

	buf := bytes.NewBuffer(in)
	// Handle and auth data.
	if err := tpmutil.UnpackBuf(buf, &handle, &paramSize); err != nil {
		return 0, nil, nil, nil, Ticket{}, nil, fmt.Errorf("decoding handle, paramSize: %v", err)
	}

	if err := tpmutil.UnpackBuf(buf, &public, &creationData, &creationHash, &ticket, &creationName); err != nil {
		return 0, nil, nil, nil, Ticket{}, nil, fmt.Errorf("decoding public, creationData, creationHash, ticket, creationName: %v", err)
	}

	if _, err := DecodeCreationData(creationData); err != nil {
		return 0, nil, nil, nil, Ticket{}, nil, fmt.Errorf("parsing CreationData: %v", err)
	}
	return handle, public, creationData, creationHash, ticket, creationName, err
}

// CreatePrimary initializes the primary key in a given hierarchy.
// The second return value is the public part of the generated key.
func CreatePrimary(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, parentPassword, ownerPassword string, p Public) (tpmutil.Handle, crypto.PublicKey, error) {
	hnd, public, _, _, _, _, err := CreatePrimaryEx(rw, owner, sel, parentPassword, ownerPassword, p)
	if err != nil {
		return 0, nil, err
	}

	pub, err := DecodePublic(public)
	if err != nil {
		return 0, nil, fmt.Errorf("parsing public: %v", err)
	}

	pubKey, err := pub.Key()
	if err != nil {
		return 0, nil, fmt.Errorf("extracting cryto.PublicKey from Public part of primary key: %v", err)
	}

	return hnd, pubKey, err
}

// CreatePrimaryEx initializes the primary key in a given hierarchy.
// This function differs from CreatePrimary in that all response elements
// are returned, and they are returned in relatively raw form.
func CreatePrimaryEx(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, parentPassword, ownerPassword string, pub Public) (keyHandle tpmutil.Handle, public, creationData, creationHash []byte, ticket Ticket, creationName []byte, err error) {
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(parentPassword)}
	Cmd, err := encodeCreate(owner, sel, auth, ownerPassword, nil /*inSensitive*/, pub, nil /*OutsideInfo*/)
	if err != nil {
		return 0, nil, nil, nil, Ticket{}, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdCreatePrimary, tpmutil.RawBytes(Cmd))
	if err != nil {
		return 0, nil, nil, nil, Ticket{}, nil, err
	}

	return decodeCreatePrimary(resp)
}

// CreatePrimaryRawTemplate is CreatePrimary, but with the public template
// (TPMT_PUBLIC) provided pre-encoded. This is commonly used with key templates
// stored in NV RAM.
func CreatePrimaryRawTemplate(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, parentPassword, ownerPassword string, public []byte) (tpmutil.Handle, crypto.PublicKey, error) {
	pub, err := DecodePublic(public)
	if err != nil {
		return 0, nil, fmt.Errorf("parsing input template: %v", err)
	}
	return CreatePrimary(rw, owner, sel, parentPassword, ownerPassword, pub)
}

func decodeReadPublic(in []byte) (Public, []byte, []byte, error) {
	var resp struct {
		Public        tpmutil.U16Bytes
		Name          tpmutil.U16Bytes
		QualifiedName tpmutil.U16Bytes
	}
	if _, err := tpmutil.Unpack(in, &resp); err != nil {
		return Public{}, nil, nil, err
	}
	pub, err := DecodePublic(resp.Public)
	if err != nil {
		return Public{}, nil, nil, err
	}
	return pub, resp.Name, resp.QualifiedName, nil
}

// ReadPublic reads the public part of the object under handle.
// Returns the public data, name and qualified name.
func ReadPublic(rw io.ReadWriter, handle tpmutil.Handle) (Public, []byte, []byte, error) {
	resp, err := runCommand(rw, TagNoSessions, CmdReadPublic, handle)
	if err != nil {
		return Public{}, nil, nil, err
	}

	return decodeReadPublic(resp)
}

func decodeCreate(in []byte) (private, public, creationData, creationHash tpmutil.U16Bytes, creationTicket Ticket, err error) {
	buf := bytes.NewBuffer(in)
	var paramSize uint32
	if err := tpmutil.UnpackBuf(buf, &paramSize, &private, &public, &creationData, &creationHash, &creationTicket); err != nil {
		return nil, nil, nil, nil, Ticket{}, fmt.Errorf("decoding Handle, Private, Public, CreationData, CreationHash, CreationTicket: %v", err)
	}
	if err != nil {
		return nil, nil, nil, nil, Ticket{}, fmt.Errorf("decoding CreationTicket: %v", err)
	}
	if _, err := DecodeCreationData(creationData); err != nil {
		return nil, nil, nil, nil, Ticket{}, fmt.Errorf("decoding CreationData: %v", err)
	}
	return private, public, creationData, creationHash, creationTicket, nil
}

func create(rw io.ReadWriter, parentHandle tpmutil.Handle, auth AuthCommand, objectPassword string, sensitiveData []byte, pub Public, pcrSelection PCRSelection, outsideInfo []byte) (private, public, creationData, creationHash []byte, creationTicket Ticket, err error) {
	cmd, err := encodeCreate(parentHandle, pcrSelection, auth, objectPassword, sensitiveData, pub, outsideInfo)
	if err != nil {
		return nil, nil, nil, nil, Ticket{}, err
	}
	resp, err := runCommand(rw, TagSessions, CmdCreate, tpmutil.RawBytes(cmd))
	if err != nil {
		return nil, nil, nil, nil, Ticket{}, err
	}
	return decodeCreate(resp)
}

// CreateKey creates a new key pair under the owner handle.
// Returns private key and public key blobs as well as the
// creation data, a hash of said data and the creation ticket.
func CreateKey(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, parentPassword, ownerPassword string, pub Public) (private, public, creationData, creationHash []byte, creationTicket Ticket, err error) {
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(parentPassword)}
	return create(rw, owner, auth, ownerPassword, nil /*inSensitive*/, pub, sel, nil /*OutsideInfo*/)
}

// CreateKeyUsingAuth creates a new key pair under the owner handle using the
// provided AuthCommand. Returns private key and public key blobs as well as
// the creation data, a hash of said data, and the creation ticket.
func CreateKeyUsingAuth(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, auth AuthCommand, ownerPassword string, pub Public) (private, public, creationData, creationHash []byte, creationTicket Ticket, err error) {
	return create(rw, owner, auth, ownerPassword, nil /*inSensitive*/, pub, sel, nil /*OutsideInfo*/)
}

// CreateKeyWithSensitive is very similar to CreateKey, except
// that it can take in a piece of sensitive data.
func CreateKeyWithSensitive(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, parentPassword, ownerPassword string, pub Public, sensitive []byte) (private, public, creationData, creationHash []byte, creationTicket Ticket, err error) {
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(parentPassword)}
	return create(rw, owner, auth, ownerPassword, sensitive, pub, sel, nil /*OutsideInfo*/)
}

// CreateKeyWithOutsideInfo is very similar to CreateKey, except
// that it returns the outside information.
func CreateKeyWithOutsideInfo(rw io.ReadWriter, owner tpmutil.Handle, sel PCRSelection, parentPassword, ownerPassword string, pub Public, outsideInfo []byte) (private, public, creationData, creationHash []byte, creationTicket Ticket, err error) {
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(parentPassword)}
	return create(rw, owner, auth, ownerPassword, nil /*inSensitive*/, pub, sel, outsideInfo)
}

// Seal creates a data blob object that seals the sensitive data under a parent and with a
// password and auth policy. Access to the parent must be available with a simple password.
// Returns private and public portions of the created object.
func Seal(rw io.ReadWriter, parentHandle tpmutil.Handle, parentPassword, objectPassword string, objectAuthPolicy []byte, sensitiveData []byte) ([]byte, []byte, error) {
	inPublic := Public{
		Type:       AlgKeyedHash,
		NameAlg:    AlgSHA256,
		Attributes: FlagFixedTPM | FlagFixedParent,
		AuthPolicy: objectAuthPolicy,
	}
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(parentPassword)}
	private, public, _, _, _, err := create(rw, parentHandle, auth, objectPassword, sensitiveData, inPublic, PCRSelection{}, nil /*OutsideInfo*/)
	if err != nil {
		return nil, nil, err
	}
	return private, public, nil
}

func encodeImport(parentHandle tpmutil.Handle, auth AuthCommand, publicBlob, privateBlob, symSeed, encryptionKey tpmutil.U16Bytes, sym *SymScheme) ([]byte, error) {
	ph, err := tpmutil.Pack(parentHandle)
	if err != nil {
		return nil, err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return nil, err
	}
	data, err := tpmutil.Pack(encryptionKey, publicBlob, privateBlob, symSeed)
	if err != nil {
		return nil, err
	}
	encodedScheme, err := sym.encode()
	if err != nil {
		return nil, err
	}

	return concat(ph, encodedAuth, data, encodedScheme)
}

func decodeImport(resp []byte) ([]byte, error) {
	var paramSize uint32
	var outPrivate tpmutil.U16Bytes
	_, err := tpmutil.Unpack(resp, &paramSize, &outPrivate)
	return outPrivate, err
}

// Import allows a user to import a key created on a different computer
// or in a different TPM. The publicBlob and privateBlob must always be
// provided. symSeed should be non-nil iff an "outer wrapper" is used. Both of
// encryptionKey and sym should be non-nil iff an "inner wrapper" is used.
func Import(rw io.ReadWriter, parentHandle tpmutil.Handle, auth AuthCommand, publicBlob, privateBlob, symSeed, encryptionKey []byte, sym *SymScheme) ([]byte, error) {
	Cmd, err := encodeImport(parentHandle, auth, publicBlob, privateBlob, symSeed, encryptionKey, sym)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdImport, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeImport(resp)
}

func encodeLoad(parentHandle tpmutil.Handle, auth AuthCommand, publicBlob, privateBlob tpmutil.U16Bytes) ([]byte, error) {
	ah, err := tpmutil.Pack(parentHandle)
	if err != nil {
		return nil, err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(privateBlob, publicBlob)
	if err != nil {
		return nil, err
	}
	return concat(ah, encodedAuth, params)
}

func decodeLoad(in []byte) (tpmutil.Handle, []byte, error) {
	var handle tpmutil.Handle
	var paramSize uint32
	var name tpmutil.U16Bytes

	if _, err := tpmutil.Unpack(in, &handle, &paramSize, &name); err != nil {
		return 0, nil, err
	}

	// Re-encode the name as a TPM2B_NAME so it can be parsed by DecodeName().
	b := &bytes.Buffer{}
	if err := name.TPMMarshal(b); err != nil {
		return 0, nil, err
	}
	return handle, b.Bytes(), nil
}

// Load loads public/private blobs into an object in the TPM.
// Returns loaded object handle and its name.
func Load(rw io.ReadWriter, parentHandle tpmutil.Handle, parentAuth string, publicBlob, privateBlob []byte) (tpmutil.Handle, []byte, error) {
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(parentAuth)}
	return LoadUsingAuth(rw, parentHandle, auth, publicBlob, privateBlob)
}

// LoadUsingAuth loads public/private blobs into an object in the TPM using the
// provided AuthCommand. Returns loaded object handle and its name.
func LoadUsingAuth(rw io.ReadWriter, parentHandle tpmutil.Handle, auth AuthCommand, publicBlob, privateBlob []byte) (tpmutil.Handle, []byte, error) {
	Cmd, err := encodeLoad(parentHandle, auth, publicBlob, privateBlob)
	if err != nil {
		return 0, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdLoad, tpmutil.RawBytes(Cmd))
	if err != nil {
		return 0, nil, err
	}
	return decodeLoad(resp)
}

func encodeLoadExternal(pub Public, private Private, hierarchy tpmutil.Handle) ([]byte, error) {
	privateBlob, err := private.Encode()
	if err != nil {
		return nil, err
	}
	publicBlob, err := pub.Encode()
	if err != nil {
		return nil, err
	}

	return tpmutil.Pack(tpmutil.U16Bytes(privateBlob), tpmutil.U16Bytes(publicBlob), hierarchy)
}

func decodeLoadExternal(in []byte) (tpmutil.Handle, []byte, error) {
	var handle tpmutil.Handle
	var name tpmutil.U16Bytes

	if _, err := tpmutil.Unpack(in, &handle, &name); err != nil {
		return 0, nil, err
	}
	return handle, name, nil
}

// LoadExternal loads a public (and optionally a private) key into an object in
// the TPM. Returns loaded object handle and its name.
func LoadExternal(rw io.ReadWriter, pub Public, private Private, hierarchy tpmutil.Handle) (tpmutil.Handle, []byte, error) {
	Cmd, err := encodeLoadExternal(pub, private, hierarchy)
	if err != nil {
		return 0, nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdLoadExternal, tpmutil.RawBytes(Cmd))
	if err != nil {
		return 0, nil, err
	}
	handle, name, err := decodeLoadExternal(resp)
	if err != nil {
		return 0, nil, err
	}
	return handle, name, nil
}

// PolicyPassword sets password authorization requirement on the object.
func PolicyPassword(rw io.ReadWriter, handle tpmutil.Handle) error {
	_, err := runCommand(rw, TagNoSessions, CmdPolicyPassword, handle)
	return err
}

func encodePolicySecret(entityHandle tpmutil.Handle, entityAuth AuthCommand, policyHandle tpmutil.Handle, policyNonce, cpHash, policyRef tpmutil.U16Bytes, expiry int32) ([]byte, error) {
	auth, err := encodeAuthArea(entityAuth)
	if err != nil {
		return nil, err
	}
	handles, err := tpmutil.Pack(entityHandle, policyHandle)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(policyNonce, cpHash, policyRef, expiry)
	if err != nil {
		return nil, err
	}
	return concat(handles, auth, params)
}

func decodePolicySecret(in []byte) ([]byte, *Ticket, error) {
	buf := bytes.NewBuffer(in)

	var paramSize uint32
	var timeout tpmutil.U16Bytes
	if err := tpmutil.UnpackBuf(buf, &paramSize, &timeout); err != nil {
		return nil, nil, fmt.Errorf("decoding timeout: %v", err)
	}
	var t Ticket
	if err := tpmutil.UnpackBuf(buf, &t); err != nil {
		return nil, nil, fmt.Errorf("decoding ticket: %v", err)
	}
	return timeout, &t, nil
}

// PolicySecret sets a secret authorization requirement on the provided entity.
func PolicySecret(rw io.ReadWriter, entityHandle tpmutil.Handle, entityAuth AuthCommand, policyHandle tpmutil.Handle, policyNonce, cpHash, policyRef []byte, expiry int32) ([]byte, *Ticket, error) {
	Cmd, err := encodePolicySecret(entityHandle, entityAuth, policyHandle, policyNonce, cpHash, policyRef, expiry)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdPolicySecret, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodePolicySecret(resp)
}

func encodePolicySigned(validationKeyHandle tpmutil.Handle, policyHandle tpmutil.Handle, policyNonce, cpHash, policyRef tpmutil.U16Bytes, expiry int32, auth []byte) ([]byte, error) {
	handles, err := tpmutil.Pack(validationKeyHandle, policyHandle)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(policyNonce, cpHash, policyRef, expiry, auth)
	if err != nil {
		return nil, err
	}
	return concat(handles, params)
}

func decodePolicySigned(in []byte) ([]byte, *Ticket, error) {
	buf := bytes.NewBuffer(in)

	var timeout tpmutil.U16Bytes
	if err := tpmutil.UnpackBuf(buf, &timeout); err != nil {
		return nil, nil, fmt.Errorf("decoding timeout: %v", err)
	}
	var t Ticket
	if err := tpmutil.UnpackBuf(buf, &t); err != nil {
		return nil, nil, fmt.Errorf("decoding ticket: %v", err)
	}
	return timeout, &t, nil
}

// PolicySigned sets a signed authorization requirement on the provided policy.
func PolicySigned(rw io.ReadWriter, validationKeyHandle tpmutil.Handle, policyHandle tpmutil.Handle, policyNonce, cpHash, policyRef []byte, expiry int32, signedAuth []byte) ([]byte, *Ticket, error) {
	Cmd, err := encodePolicySigned(validationKeyHandle, policyHandle, policyNonce, cpHash, policyRef, expiry, signedAuth)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdPolicySigned, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodePolicySigned(resp)
}

func encodePolicyPCR(session tpmutil.Handle, expectedDigest tpmutil.U16Bytes, sel PCRSelection) ([]byte, error) {
	params, err := tpmutil.Pack(session, expectedDigest)
	if err != nil {
		return nil, err
	}
	pcrs, err := encodeTPMLPCRSelection(sel)
	if err != nil {
		return nil, err
	}
	return concat(params, pcrs)
}

// PolicyPCR sets PCR state binding for authorization on a session.
//
// expectedDigest is optional. When specified, it's compared against the digest
// of PCRs matched by sel.
//
// Note that expectedDigest must be a *digest* of the expected PCR value. You
// must compute the digest manually. ReadPCR returns raw PCR values, not their
// digests.
// If you wish to select multiple PCRs, concatenate their values before
// computing the digest. See "TPM 2.0 Part 1, Selecting Multiple PCR".
func PolicyPCR(rw io.ReadWriter, session tpmutil.Handle, expectedDigest []byte, sel PCRSelection) error {
	Cmd, err := encodePolicyPCR(session, expectedDigest, sel)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagNoSessions, CmdPolicyPCR, tpmutil.RawBytes(Cmd))
	return err
}

// PolicyOr compares PolicySession→Digest against the list of provided values.
// If the current Session→Digest does not match any value in the list,
// the TPM shall return TPM_RC_VALUE. Otherwise, the TPM will reset policySession→Digest
// to a Zero Digest. Then policySession→Digest is extended by the concatenation of
// TPM_CC_PolicyOR and the concatenation of all of the digests.
func PolicyOr(rw io.ReadWriter, session tpmutil.Handle, digests TPMLDigest) error {
	d, err := digests.Encode()
	if err != nil {
		return err
	}
	data, err := tpmutil.Pack(session, d)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagNoSessions, CmdPolicyOr, data)
	return err
}

// PolicyGetDigest returns the current policyDigest of the session.
func PolicyGetDigest(rw io.ReadWriter, handle tpmutil.Handle) ([]byte, error) {
	resp, err := runCommand(rw, TagNoSessions, CmdPolicyGetDigest, handle)
	if err != nil {
		return nil, err
	}

	var digest tpmutil.U16Bytes
	_, err = tpmutil.Unpack(resp, &digest)
	return digest, err
}

func encodeStartAuthSession(tpmKey, bindKey tpmutil.Handle, nonceCaller, secret tpmutil.U16Bytes, se SessionType, sym, hashAlg Algorithm) ([]byte, error) {
	ha, err := tpmutil.Pack(tpmKey, bindKey)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(nonceCaller, secret, se, sym, hashAlg)
	if err != nil {
		return nil, err
	}
	return concat(ha, params)
}

func decodeStartAuthSession(in []byte) (tpmutil.Handle, []byte, error) {
	var handle tpmutil.Handle
	var nonce tpmutil.U16Bytes
	if _, err := tpmutil.Unpack(in, &handle, &nonce); err != nil {
		return 0, nil, err
	}
	return handle, nonce, nil
}

// StartAuthSession initializes a session object.
// Returns session handle and the initial nonce from the TPM.
func StartAuthSession(rw io.ReadWriter, tpmKey, bindKey tpmutil.Handle, nonceCaller, secret []byte, se SessionType, sym, hashAlg Algorithm) (tpmutil.Handle, []byte, error) {
	Cmd, err := encodeStartAuthSession(tpmKey, bindKey, nonceCaller, secret, se, sym, hashAlg)
	if err != nil {
		return 0, nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdStartAuthSession, tpmutil.RawBytes(Cmd))
	if err != nil {
		return 0, nil, err
	}
	return decodeStartAuthSession(resp)
}

func encodeUnseal(sessionHandle, itemHandle tpmutil.Handle, password string) ([]byte, error) {
	ha, err := tpmutil.Pack(itemHandle)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: sessionHandle, Attributes: AttrContinueSession, Auth: []byte(password)})
	if err != nil {
		return nil, err
	}
	return concat(ha, auth)
}

func decodeUnseal(in []byte) ([]byte, error) {
	var paramSize uint32
	var unsealed tpmutil.U16Bytes

	if _, err := tpmutil.Unpack(in, &paramSize, &unsealed); err != nil {
		return nil, err
	}
	return unsealed, nil
}

// Unseal returns the data for a loaded sealed object.
func Unseal(rw io.ReadWriter, itemHandle tpmutil.Handle, password string) ([]byte, error) {
	return UnsealWithSession(rw, HandlePasswordSession, itemHandle, password)
}

// UnsealWithSession returns the data for a loaded sealed object.
func UnsealWithSession(rw io.ReadWriter, sessionHandle, itemHandle tpmutil.Handle, password string) ([]byte, error) {
	Cmd, err := encodeUnseal(sessionHandle, itemHandle, password)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdUnseal, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeUnseal(resp)
}

func encodeQuote(signingHandle tpmutil.Handle, signerAuth string, toQuote tpmutil.U16Bytes, sel PCRSelection, sigAlg Algorithm) ([]byte, error) {
	ha, err := tpmutil.Pack(signingHandle)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(signerAuth)})
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(toQuote, sigAlg)
	if err != nil {
		return nil, err
	}
	pcrs, err := encodeTPMLPCRSelection(sel)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params, pcrs)
}

func decodeQuote(in []byte) ([]byte, []byte, error) {
	buf := bytes.NewBuffer(in)
	var paramSize uint32
	if err := tpmutil.UnpackBuf(buf, &paramSize); err != nil {
		return nil, nil, err
	}
	buf.Truncate(int(paramSize))
	var attest tpmutil.U16Bytes
	if err := tpmutil.UnpackBuf(buf, &attest); err != nil {
		return nil, nil, err
	}
	return attest, buf.Bytes(), nil
}

// Quote returns a quote of PCR values. A quote is a signature of the PCR
// values, created using a signing TPM key.
//
// Returns attestation data and the decoded signature.
func Quote(rw io.ReadWriter, signingHandle tpmutil.Handle, signerAuth, unused string, toQuote []byte, sel PCRSelection, sigAlg Algorithm) ([]byte, *Signature, error) {
	// TODO: Remove "unused" parameter on next breaking change.
	attest, sigRaw, err := QuoteRaw(rw, signingHandle, signerAuth, unused, toQuote, sel, sigAlg)
	if err != nil {
		return nil, nil, err
	}
	sig, err := DecodeSignature(bytes.NewBuffer(sigRaw))
	if err != nil {
		return nil, nil, err
	}
	return attest, sig, nil
}

// QuoteRaw is very similar to Quote, except that it will return
// the raw signature in a byte array without decoding.
func QuoteRaw(rw io.ReadWriter, signingHandle tpmutil.Handle, signerAuth, _ string, toQuote []byte, sel PCRSelection, sigAlg Algorithm) ([]byte, []byte, error) {
	// TODO: Remove "unused" parameter on next breaking change.
	Cmd, err := encodeQuote(signingHandle, signerAuth, toQuote, sel, sigAlg)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdQuote, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeQuote(resp)
}

func encodeActivateCredential(auth []AuthCommand, activeHandle tpmutil.Handle, keyHandle tpmutil.Handle, credBlob, secret tpmutil.U16Bytes) ([]byte, error) {
	ha, err := tpmutil.Pack(activeHandle, keyHandle)
	if err != nil {
		return nil, err
	}
	a, err := encodeAuthArea(auth...)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(credBlob, secret)
	if err != nil {
		return nil, err
	}
	return concat(ha, a, params)
}

func decodeActivateCredential(in []byte) ([]byte, error) {
	var paramSize uint32
	var certInfo tpmutil.U16Bytes

	if _, err := tpmutil.Unpack(in, &paramSize, &certInfo); err != nil {
		return nil, err
	}
	return certInfo, nil
}

// ActivateCredential associates an object with a credential.
// Returns decrypted certificate information.
func ActivateCredential(rw io.ReadWriter, activeHandle, keyHandle tpmutil.Handle, activePassword, protectorPassword string, credBlob, secret []byte) ([]byte, error) {
	return ActivateCredentialUsingAuth(rw, []AuthCommand{
		{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(activePassword)},
		{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(protectorPassword)},
	}, activeHandle, keyHandle, credBlob, secret)
}

// ActivateCredentialUsingAuth associates an object with a credential, using the
// given set of authorizations. Two authorization must be provided.
// Returns decrypted certificate information.
func ActivateCredentialUsingAuth(rw io.ReadWriter, auth []AuthCommand, activeHandle, keyHandle tpmutil.Handle, credBlob, secret []byte) ([]byte, error) {
	if len(auth) != 2 {
		return nil, fmt.Errorf("len(auth) = %d, want 2", len(auth))
	}

	Cmd, err := encodeActivateCredential(auth, activeHandle, keyHandle, credBlob, secret)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdActivateCredential, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeActivateCredential(resp)
}

func encodeMakeCredential(protectorHandle tpmutil.Handle, credential, activeName tpmutil.U16Bytes) ([]byte, error) {
	ha, err := tpmutil.Pack(protectorHandle)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(credential, activeName)
	if err != nil {
		return nil, err
	}
	return concat(ha, params)
}

func decodeMakeCredential(in []byte) ([]byte, []byte, error) {
	var credBlob, encryptedSecret tpmutil.U16Bytes

	if _, err := tpmutil.Unpack(in, &credBlob, &encryptedSecret); err != nil {
		return nil, nil, err
	}
	return credBlob, encryptedSecret, nil
}

// MakeCredential creates an encrypted credential for use in MakeCredential.
// Returns encrypted credential and wrapped secret used to encrypt it.
func MakeCredential(rw io.ReadWriter, protectorHandle tpmutil.Handle, credential, activeName []byte) ([]byte, []byte, error) {
	Cmd, err := encodeMakeCredential(protectorHandle, credential, activeName)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdMakeCredential, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeMakeCredential(resp)
}

func encodeEvictControl(ownerAuth string, owner, objectHandle, persistentHandle tpmutil.Handle) ([]byte, error) {
	ha, err := tpmutil.Pack(owner, objectHandle)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(ownerAuth)})
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(persistentHandle)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params)
}

// EvictControl toggles persistence of an object within the TPM.
func EvictControl(rw io.ReadWriter, ownerAuth string, owner, objectHandle, persistentHandle tpmutil.Handle) error {
	Cmd, err := encodeEvictControl(ownerAuth, owner, objectHandle, persistentHandle)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdEvictControl, tpmutil.RawBytes(Cmd))
	return err
}

func encodeClear(handle tpmutil.Handle, auth AuthCommand) ([]byte, error) {
	ah, err := tpmutil.Pack(handle)
	if err != nil {
		return nil, err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return nil, err
	}
	return concat(ah, encodedAuth)
}

// Clear clears lockout, endorsement and owner hierarchy authorization values
func Clear(rw io.ReadWriter, handle tpmutil.Handle, auth AuthCommand) error {
	Cmd, err := encodeClear(handle, auth)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdClear, tpmutil.RawBytes(Cmd))
	return err
}

func encodeHierarchyChangeAuth(handle tpmutil.Handle, auth AuthCommand, newAuth string) ([]byte, error) {
	ah, err := tpmutil.Pack(handle)
	if err != nil {
		return nil, err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return nil, err
	}
	param, err := tpmutil.Pack(tpmutil.U16Bytes(newAuth))
	if err != nil {
		return nil, err
	}
	return concat(ah, encodedAuth, param)
}

// HierarchyChangeAuth changes the authorization values for a hierarchy or for the lockout authority
func HierarchyChangeAuth(rw io.ReadWriter, handle tpmutil.Handle, auth AuthCommand, newAuth string) error {
	Cmd, err := encodeHierarchyChangeAuth(handle, auth, newAuth)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdHierarchyChangeAuth, tpmutil.RawBytes(Cmd))
	return err
}

// ContextSave returns an encrypted version of the session, object or sequence
// context for storage outside of the TPM. The handle references context to
// store.
func ContextSave(rw io.ReadWriter, handle tpmutil.Handle) ([]byte, error) {
	return runCommand(rw, TagNoSessions, CmdContextSave, handle)
}

// ContextLoad reloads context data created by ContextSave.
func ContextLoad(rw io.ReadWriter, saveArea []byte) (tpmutil.Handle, error) {
	resp, err := runCommand(rw, TagNoSessions, CmdContextLoad, tpmutil.RawBytes(saveArea))
	if err != nil {
		return 0, err
	}
	var handle tpmutil.Handle
	_, err = tpmutil.Unpack(resp, &handle)
	return handle, err
}

func encodeIncrementNV(handle tpmutil.Handle, authString string) ([]byte, error) {
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(authString)})
	if err != nil {
		return nil, err
	}
	out, err := tpmutil.Pack(handle, handle)
	if err != nil {
		return nil, err
	}
	return concat(out, auth)
}

// NVIncrement increments a counter in NVRAM.
func NVIncrement(rw io.ReadWriter, handle tpmutil.Handle, authString string) error {
	Cmd, err := encodeIncrementNV(handle, authString)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdIncrementNVCounter, tpmutil.RawBytes(Cmd))
	return err
}

// NVUndefineSpace removes an index from TPM's NV storage.
func NVUndefineSpace(rw io.ReadWriter, ownerAuth string, owner, index tpmutil.Handle) error {
	authArea := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(ownerAuth)}
	return NVUndefineSpaceEx(rw, owner, index, authArea)
}

// NVUndefineSpaceEx removes an index from NVRAM. Unlike, NVUndefineSpace(), custom command
// authorization can be provided.
func NVUndefineSpaceEx(rw io.ReadWriter, owner, index tpmutil.Handle, authArea AuthCommand) error {
	out, err := tpmutil.Pack(owner, index)
	if err != nil {
		return err
	}
	auth, err := encodeAuthArea(authArea)
	if err != nil {
		return err
	}
	cmd, err := concat(out, auth)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdUndefineSpace, tpmutil.RawBytes(cmd))
	return err
}

// NVUndefineSpaceSpecial This command allows removal of a platform-created NV Index that has TPMA_NV_POLICY_DELETE SET.
// The policy to authorize NV index access needs to be created with PolicyCommandCode(rw, sessionHandle, CmdNVUndefineSpaceSpecial) function
// nvAuthCmd takes the session handle for the policy and the AuthValue (which can be emptyAuth) for the authorization.
// platformAuth takes either a sessionHandle for the platform policy or HandlePasswordSession and the platformAuth value for authorization.
func NVUndefineSpaceSpecial(rw io.ReadWriter, nvIndex tpmutil.Handle, nvAuth, platformAuth AuthCommand) error {
	authBytes, err := encodeAuthArea(nvAuth, platformAuth)
	if err != nil {
		return err
	}
	auth, err := tpmutil.Pack(authBytes)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdNVUndefineSpaceSpecial, nvIndex, HandlePlatform, tpmutil.RawBytes(auth))
	return err
}

// NVDefineSpace creates an index in TPM's NV storage.
func NVDefineSpace(rw io.ReadWriter, owner, handle tpmutil.Handle, ownerAuth, authString string, policy []byte, attributes NVAttr, dataSize uint16) error {
	nvPub := NVPublic{
		NVIndex:    handle,
		NameAlg:    AlgSHA1,
		Attributes: attributes,
		AuthPolicy: policy,
		DataSize:   dataSize,
	}
	authArea := AuthCommand{
		Session:    HandlePasswordSession,
		Attributes: AttrContinueSession,
		Auth:       []byte(ownerAuth),
	}
	return NVDefineSpaceEx(rw, owner, authString, nvPub, authArea)
}

// NVDefineSpaceEx accepts NVPublic structure and AuthCommand, allowing more flexibility.
func NVDefineSpaceEx(rw io.ReadWriter, owner tpmutil.Handle, authVal string, pubInfo NVPublic, authArea AuthCommand) error {
	ha, err := tpmutil.Pack(owner)
	if err != nil {
		return err
	}
	auth, err := encodeAuthArea(authArea)
	if err != nil {
		return err
	}
	publicInfo, err := tpmutil.Pack(pubInfo)
	if err != nil {
		return err
	}
	params, err := tpmutil.Pack(tpmutil.U16Bytes(authVal), tpmutil.U16Bytes(publicInfo))
	if err != nil {
		return err
	}
	cmd, err := concat(ha, auth, params)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdDefineSpace, tpmutil.RawBytes(cmd))
	return err
}

// NVWrite writes data into the TPM's NV storage.
func NVWrite(rw io.ReadWriter, authHandle, nvIndex tpmutil.Handle, authString string, data tpmutil.U16Bytes, offset uint16) error {
	auth := AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(authString)}
	return NVWriteEx(rw, authHandle, nvIndex, auth, data, offset)
}

// NVWriteEx does the same as NVWrite with the exception of letting the user take care of the AuthCommand before calling the function.
// This allows more flexibility and does not limit the AuthCommand to PasswordSession.
func NVWriteEx(rw io.ReadWriter, authHandle, nvIndex tpmutil.Handle, authArea AuthCommand, data tpmutil.U16Bytes, offset uint16) error {
	h, err := tpmutil.Pack(authHandle, nvIndex)
	if err != nil {
		return err
	}
	authEnc, err := encodeAuthArea(authArea)
	if err != nil {
		return err
	}

	d, err := tpmutil.Pack(data, offset)
	if err != nil {
		return err
	}

	b, err := concat(h, authEnc, d)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdWriteNV, tpmutil.RawBytes(b))
	return err
}

func encodeLockNV(owner, handle tpmutil.Handle, authString string) ([]byte, error) {
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(authString)})
	if err != nil {
		return nil, err
	}
	out, err := tpmutil.Pack(owner, handle)
	if err != nil {
		return nil, err
	}
	return concat(out, auth)
}

// NVWriteLock inhibits further writes on the given NV index if at least one of
// the AttrWriteSTClear or AttrWriteDefine bits is set.
//
// AttrWriteSTClear causes the index to be locked until the TPM is restarted
// (see the Startup function).
//
// AttrWriteDefine causes the index to be locked permanently if data has been
// written to the index; otherwise the lock is removed on startup.
//
// NVWriteLock returns an error if neither bit is set.
//
// It is not an error to call NVWriteLock for an index that is already locked
// for writing.
func NVWriteLock(rw io.ReadWriter, owner, handle tpmutil.Handle, authString string) error {
	Cmd, err := encodeLockNV(owner, handle, authString)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdWriteLockNV, tpmutil.RawBytes(Cmd))
	return err
}

func decodeNVReadPublic(in []byte) (NVPublic, error) {
	var pub NVPublic
	var buf tpmutil.U16Bytes
	if _, err := tpmutil.Unpack(in, &buf); err != nil {
		return pub, err
	}
	_, err := tpmutil.Unpack(buf, &pub)
	return pub, err
}

// NVReadPublic reads the public data of an NV index.
func NVReadPublic(rw io.ReadWriter, index tpmutil.Handle) (NVPublic, error) {
	// Read public area to determine data size.
	resp, err := runCommand(rw, TagNoSessions, CmdReadPublicNV, index)
	if err != nil {
		return NVPublic{}, err
	}
	return decodeNVReadPublic(resp)
}

func decodeNVRead(in []byte) ([]byte, error) {
	var paramSize uint32
	var data tpmutil.U16Bytes
	if _, err := tpmutil.Unpack(in, &paramSize, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func encodeNVRead(nvIndex, authHandle tpmutil.Handle, password string, offset, dataSize uint16) ([]byte, error) {
	handles, err := tpmutil.Pack(authHandle, nvIndex)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(password)})
	if err != nil {
		return nil, err
	}

	params, err := tpmutil.Pack(dataSize, offset)
	if err != nil {
		return nil, err
	}

	return concat(handles, auth, params)
}

// NVRead reads a full data blob from an NV index. This function is
// deprecated; use NVReadEx instead.
func NVRead(rw io.ReadWriter, index tpmutil.Handle) ([]byte, error) {
	return NVReadEx(rw, index, index, "", 0)
}

// NVReadEx reads a full data blob from an NV index, using the given
// authorization handle. NVRead commands are done in blocks of blockSize.
// If blockSize is 0, the TPM is queried for TPM_PT_NV_BUFFER_MAX, and that
// value is used.
func NVReadEx(rw io.ReadWriter, index, authHandle tpmutil.Handle, password string, blockSize int) ([]byte, error) {
	if blockSize == 0 {
		readBuff, _, err := GetCapability(rw, CapabilityTPMProperties, 1, uint32(NVMaxBufferSize))
		if err != nil {
			return nil, fmt.Errorf("GetCapability for TPM_PT_NV_BUFFER_MAX failed: %v", err)
		}
		if len(readBuff) != 1 {
			return nil, fmt.Errorf("could not determine NVRAM read/write buffer size")
		}
		rb, ok := readBuff[0].(TaggedProperty)
		if !ok {
			return nil, fmt.Errorf("GetCapability returned unexpected type: %T, expected TaggedProperty", readBuff[0])
		}
		blockSize = int(rb.Value)
	}

	// Read public area to determine data size.
	pub, err := NVReadPublic(rw, index)
	if err != nil {
		return nil, fmt.Errorf("decoding NV_ReadPublic response: %v", err)
	}

	// Read the NVRAM area in blocks.
	outBuff := make([]byte, 0, int(pub.DataSize))
	for len(outBuff) < int(pub.DataSize) {
		readSize := blockSize
		if readSize > (int(pub.DataSize) - len(outBuff)) {
			readSize = int(pub.DataSize) - len(outBuff)
		}

		Cmd, err := encodeNVRead(index, authHandle, password, uint16(len(outBuff)), uint16(readSize))
		if err != nil {
			return nil, fmt.Errorf("building NV_Read command: %v", err)
		}
		resp, err := runCommand(rw, TagSessions, CmdReadNV, tpmutil.RawBytes(Cmd))
		if err != nil {
			return nil, fmt.Errorf("running NV_Read command (cursor=%d,size=%d): %v", len(outBuff), readSize, err)
		}
		data, err := decodeNVRead(resp)
		if err != nil {
			return nil, fmt.Errorf("decoding NV_Read command: %v", err)
		}
		outBuff = append(outBuff, data...)
	}
	return outBuff, nil
}

// NVReadLock inhibits further reads of the given NV index if AttrReadSTClear
// is set. After the TPM is restarted the index can be read again (see the
// Startup function).
//
// NVReadLock returns an error if the AttrReadSTClear bit is not set.
//
// It is not an error to call NVReadLock for an index that is already locked
// for reading.
func NVReadLock(rw io.ReadWriter, owner, handle tpmutil.Handle, authString string) error {
	Cmd, err := encodeLockNV(owner, handle, authString)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdReadLockNV, tpmutil.RawBytes(Cmd))
	return err
}

// decodeHash unpacks a successful response to TPM2_Hash, returning the computed digest and
// validation ticket.
func decodeHash(resp []byte) ([]byte, *Ticket, error) {
	var digest tpmutil.U16Bytes
	var validation Ticket

	buf := bytes.NewBuffer(resp)
	if err := tpmutil.UnpackBuf(buf, &digest, &validation); err != nil {
		return nil, nil, err
	}
	return digest, &validation, nil
}

// Hash computes a hash of data in buf using TPM2_Hash, returning the computed
// digest and validation ticket. The validation ticket serves as confirmation
// from the TPM that the data in buf did not begin with TPM_GENERATED_VALUE.
// NOTE: TPM2_Hash can only accept data up to MAX_DIGEST_BUFFER in size, which
// is implementation-dependent, but guaranteed to be at least 1024 octets.
func Hash(rw io.ReadWriter, alg Algorithm, buf tpmutil.U16Bytes, hierarchy tpmutil.Handle) (digest []byte, validation *Ticket, err error) {
	resp, err := runCommand(rw, TagNoSessions, CmdHash, buf, alg, hierarchy)
	if err != nil {
		return nil, nil, err
	}
	return decodeHash(resp)
}

// HashSequenceStart starts a hash or an event sequence. If hashAlg is an
// implemented hash, then a hash sequence is started. If hashAlg is
// TPM_ALG_NULL, then an event sequence is started.
func HashSequenceStart(rw io.ReadWriter, sequenceAuth string, hashAlg Algorithm) (seqHandle tpmutil.Handle, err error) {
	resp, err := runCommand(rw, TagNoSessions, CmdHashSequenceStart, tpmutil.U16Bytes(sequenceAuth), hashAlg)
	if err != nil {
		return 0, err
	}
	var handle tpmutil.Handle
	_, err = tpmutil.Unpack(resp, &handle)
	return handle, err
}

func encodeSequenceUpdate(sequenceAuth string, seqHandle tpmutil.Handle, buf tpmutil.U16Bytes) ([]byte, error) {
	ha, err := tpmutil.Pack(seqHandle)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(sequenceAuth)})
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(buf)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params)
}

// SequenceUpdate is used to add data to a hash or HMAC sequence.
func SequenceUpdate(rw io.ReadWriter, sequenceAuth string, seqHandle tpmutil.Handle, buffer []byte) error {
	cmd, err := encodeSequenceUpdate(sequenceAuth, seqHandle, buffer)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdSequenceUpdate, tpmutil.RawBytes(cmd))
	return err
}

func decodeSequenceComplete(resp []byte) ([]byte, *Ticket, error) {
	var digest tpmutil.U16Bytes
	var validation Ticket
	var paramSize uint32

	if _, err := tpmutil.Unpack(resp, &paramSize, &digest, &validation); err != nil {
		return nil, nil, err
	}
	return digest, &validation, nil
}

func encodeSequenceComplete(sequenceAuth string, seqHandle, hierarchy tpmutil.Handle, buf tpmutil.U16Bytes) ([]byte, error) {
	ha, err := tpmutil.Pack(seqHandle)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(sequenceAuth)})
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(buf, hierarchy)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params)
}

// SequenceComplete adds the last part of data, if any, to a hash/HMAC sequence
// and returns the result.
func SequenceComplete(rw io.ReadWriter, sequenceAuth string, seqHandle, hierarchy tpmutil.Handle, buffer []byte) (digest []byte, validation *Ticket, err error) {
	cmd, err := encodeSequenceComplete(sequenceAuth, seqHandle, hierarchy, buffer)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdSequenceComplete, tpmutil.RawBytes(cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeSequenceComplete(resp)
}

func encodeEventSequenceComplete(auths []AuthCommand, pcrHandle, seqHandle tpmutil.Handle, buf tpmutil.U16Bytes) ([]byte, error) {
	ha, err := tpmutil.Pack(pcrHandle, seqHandle)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(auths...)
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(buf)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params)
}

func decodeEventSequenceComplete(resp []byte) ([]*HashValue, error) {
	var paramSize uint32
	var hashCount uint32
	var err error

	buf := bytes.NewBuffer(resp)
	if err := tpmutil.UnpackBuf(buf, &paramSize, &hashCount); err != nil {
		return nil, err
	}

	buf.Truncate(int(paramSize))
	digests := make([]*HashValue, hashCount)
	for i := uint32(0); i < hashCount; i++ {
		if digests[i], err = decodeHashValue(buf); err != nil {
			return nil, err
		}
	}

	return digests, nil
}

// EventSequenceComplete adds the last part of data, if any, to an Event
// Sequence and returns the result in a digest list. If pcrHandle references a
// PCR and not AlgNull, then the returned digest list is processed in the same
// manner as the digest list input parameter to PCRExtend() with the pcrHandle
// in each bank extended with the associated digest value.
func EventSequenceComplete(rw io.ReadWriter, pcrAuth, sequenceAuth string, pcrHandle, seqHandle tpmutil.Handle, buffer []byte) (digests []*HashValue, err error) {
	auth := []AuthCommand{
		{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(pcrAuth)},
		{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(sequenceAuth)},
	}
	cmd, err := encodeEventSequenceComplete(auth, pcrHandle, seqHandle, buffer)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdEventSequenceComplete, tpmutil.RawBytes(cmd))
	if err != nil {
		return nil, err
	}
	return decodeEventSequenceComplete(resp)
}

// Startup initializes a TPM (usually done by the OS).
func Startup(rw io.ReadWriter, typ StartupType) error {
	_, err := runCommand(rw, TagNoSessions, CmdStartup, typ)
	return err
}

// Shutdown shuts down a TPM (usually done by the OS).
func Shutdown(rw io.ReadWriter, typ StartupType) error {
	_, err := runCommand(rw, TagNoSessions, CmdShutdown, typ)
	return err
}

// nullTicket is a hard-coded null ticket of type TPMT_TK_HASHCHECK.
// It is for Sign commands that do not require the TPM to verify that the digest
// is not from data that started with TPM_GENERATED_VALUE.
var nullTicket = Ticket{
	Type:      TagHashCheck,
	Hierarchy: HandleNull,
	Digest:    tpmutil.U16Bytes{},
}

func encodeSign(sessionHandle, key tpmutil.Handle, password string, digest tpmutil.U16Bytes, sigScheme *SigScheme, validation *Ticket) ([]byte, error) {
	ha, err := tpmutil.Pack(key)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: sessionHandle, Attributes: AttrContinueSession, Auth: []byte(password)})
	if err != nil {
		return nil, err
	}
	d, err := tpmutil.Pack(digest)
	if err != nil {
		return nil, err
	}
	s, err := sigScheme.encode()
	if err != nil {
		return nil, err
	}
	if validation == nil {
		validation = &nullTicket
	}
	v, err := tpmutil.Pack(validation)
	if err != nil {
		return nil, err
	}

	return concat(ha, auth, d, s, v)
}

func decodeSign(buf []byte) (*Signature, error) {
	in := bytes.NewBuffer(buf)
	var paramSize uint32
	if err := tpmutil.UnpackBuf(in, &paramSize); err != nil {
		return nil, err
	}
	return DecodeSignature(in)
}

// SignWithSession computes a signature for digest using a given loaded key. Signature
// algorithm depends on the key type. Used for keys with non-password authorization policies.
// If 'key' references a Restricted Decryption key, 'validation' must be a valid hash verification
// ticket from the TPM, which can be obtained by using Hash() to hash the data with the TPM.
// If 'validation' is nil, a NULL ticket is passed to TPM2_Sign.
func SignWithSession(rw io.ReadWriter, sessionHandle, key tpmutil.Handle, password string, digest []byte, validation *Ticket, sigScheme *SigScheme) (*Signature, error) {
	Cmd, err := encodeSign(sessionHandle, key, password, digest, sigScheme, validation)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdSign, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeSign(resp)
}

// Sign computes a signature for digest using a given loaded key. Signature
// algorithm depends on the key type.
// If 'key' references a Restricted Decryption key, 'validation' must be a valid hash verification
// ticket from the TPM, which can be obtained by using Hash() to hash the data with the TPM.
// If 'validation' is nil, a NULL ticket is passed to TPM2_Sign.
func Sign(rw io.ReadWriter, key tpmutil.Handle, password string, digest []byte, validation *Ticket, sigScheme *SigScheme) (*Signature, error) {
	return SignWithSession(rw, HandlePasswordSession, key, password, digest, validation, sigScheme)
}

func encodeCertify(objectAuth, signerAuth string, object, signer tpmutil.Handle, qualifyingData tpmutil.U16Bytes) ([]byte, error) {
	ha, err := tpmutil.Pack(object, signer)
	if err != nil {
		return nil, err
	}

	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(objectAuth)}, AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(signerAuth)})
	if err != nil {
		return nil, err
	}

	scheme := SigScheme{Alg: AlgRSASSA, Hash: AlgSHA256}
	// Use signing key's scheme.
	s, err := scheme.encode()
	if err != nil {
		return nil, err
	}
	data, err := tpmutil.Pack(qualifyingData)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, data, s)
}

// This function differs from encodeCertify in that it takes the scheme to be used as an additional argument.
func encodeCertifyEx(objectAuth, signerAuth string, object, signer tpmutil.Handle, qualifyingData tpmutil.U16Bytes, scheme SigScheme) ([]byte, error) {
	ha, err := tpmutil.Pack(object, signer)
	if err != nil {
		return nil, err
	}

	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(objectAuth)}, AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(signerAuth)})
	if err != nil {
		return nil, err
	}

	s, err := scheme.encode()
	if err != nil {
		return nil, err
	}
	data, err := tpmutil.Pack(qualifyingData)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, data, s)
}

func decodeCertify(resp []byte) ([]byte, []byte, error) {
	var paramSize uint32
	var attest tpmutil.U16Bytes

	buf := bytes.NewBuffer(resp)
	if err := tpmutil.UnpackBuf(buf, &paramSize); err != nil {
		return nil, nil, err
	}
	buf.Truncate(int(paramSize))
	if err := tpmutil.UnpackBuf(buf, &attest); err != nil {
		return nil, nil, err
	}
	return attest, buf.Bytes(), nil
}

// Certify generates a signature of a loaded TPM object with a signing key
// signer. This function calls encodeCertify which makes use of the hardcoded
// signing scheme {AlgRSASSA, AlgSHA256}. Returned values are: attestation data (TPMS_ATTEST),
// signature and error, if any.
func Certify(rw io.ReadWriter, objectAuth, signerAuth string, object, signer tpmutil.Handle, qualifyingData []byte) ([]byte, []byte, error) {
	cmd, err := encodeCertify(objectAuth, signerAuth, object, signer, qualifyingData)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdCertify, tpmutil.RawBytes(cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeCertify(resp)
}

// CertifyEx generates a signature of a loaded TPM object with a signing key
// signer. This function differs from Certify in that it takes the scheme
// to be used as an additional argument and calls encodeCertifyEx instead
// of encodeCertify. Returned values are: attestation data (TPMS_ATTEST),
// signature and error, if any.
func CertifyEx(rw io.ReadWriter, objectAuth, signerAuth string, object, signer tpmutil.Handle, qualifyingData []byte, scheme SigScheme) ([]byte, []byte, error) {
	cmd, err := encodeCertifyEx(objectAuth, signerAuth, object, signer, qualifyingData, scheme)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdCertify, tpmutil.RawBytes(cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeCertify(resp)
}

func encodeCertifyCreation(objectAuth string, object, signer tpmutil.Handle, qualifyingData, creationHash tpmutil.U16Bytes, scheme SigScheme, ticket Ticket) ([]byte, error) {
	handles, err := tpmutil.Pack(signer, object)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(objectAuth)})
	if err != nil {
		return nil, err
	}
	s, err := scheme.encode()
	if err != nil {
		return nil, err
	}
	params, err := tpmutil.Pack(qualifyingData, creationHash, tpmutil.RawBytes(s), ticket)
	if err != nil {
		return nil, err
	}
	return concat(handles, auth, params)
}

// CertifyCreation generates a signature of a newly-created &
// loaded TPM object, using signer as the signing key.
func CertifyCreation(rw io.ReadWriter, objectAuth string, object, signer tpmutil.Handle, qualifyingData, creationHash []byte, sigScheme SigScheme, creationTicket Ticket) (attestation, signature []byte, err error) {
	Cmd, err := encodeCertifyCreation(objectAuth, object, signer, qualifyingData, creationHash, sigScheme, creationTicket)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdCertifyCreation, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeCertify(resp)
}

func runCommand(rw io.ReadWriter, tag tpmutil.Tag, Cmd tpmutil.Command, in ...interface{}) ([]byte, error) {
	resp, code, err := tpmutil.RunCommand(rw, tag, Cmd, in...)
	if err != nil {
		return nil, err
	}
	if code != tpmutil.RCSuccess {
		return nil, decodeResponse(code)
	}
	return resp, decodeResponse(code)
}

// concat is a helper for encoding functions that separately encode handle,
// auth and param areas. A nil error is always returned, so that callers can
// simply return concat(a, b, c).
func concat(chunks ...[]byte) ([]byte, error) {
	return bytes.Join(chunks, nil), nil
}

func encodePCRExtend(pcr tpmutil.Handle, hashAlg Algorithm, hash tpmutil.RawBytes, password string) ([]byte, error) {
	ha, err := tpmutil.Pack(pcr)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(password)})
	if err != nil {
		return nil, err
	}
	pcrCount := uint32(1)
	extend, err := tpmutil.Pack(pcrCount, hashAlg, hash)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, extend)
}

// PCRExtend extends a value into the selected PCR
func PCRExtend(rw io.ReadWriter, pcr tpmutil.Handle, hashAlg Algorithm, hash []byte, password string) error {
	Cmd, err := encodePCRExtend(pcr, hashAlg, hash, password)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdPCRExtend, tpmutil.RawBytes(Cmd))
	return err
}

// ReadPCR reads the value of the given PCR.
func ReadPCR(rw io.ReadWriter, pcr int, hashAlg Algorithm) ([]byte, error) {
	pcrSelection := PCRSelection{
		Hash: hashAlg,
		PCRs: []int{pcr},
	}
	pcrVals, err := ReadPCRs(rw, pcrSelection)
	if err != nil {
		return nil, fmt.Errorf("unable to read PCRs from TPM: %v", err)
	}
	pcrVal, present := pcrVals[pcr]
	if !present {
		return nil, fmt.Errorf("PCR %d value missing from response", pcr)
	}
	return pcrVal, nil
}

func encodePCRReset(pcr tpmutil.Handle) ([]byte, error) {
	ha, err := tpmutil.Pack(pcr)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: EmptyAuth})
	if err != nil {
		return nil, err
	}
	return concat(ha, auth)
}

// PCRReset resets the value of the given PCR. Usually, only PCR 16 (Debug) and
// PCR 23 (Application) are resettable on the default locality.
func PCRReset(rw io.ReadWriter, pcr tpmutil.Handle) error {
	Cmd, err := encodePCRReset(pcr)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdPCRReset, tpmutil.RawBytes(Cmd))
	return err
}

// EncryptSymmetric encrypts data using a symmetric key.
//
// WARNING: This command performs low-level cryptographic operations.
// Secure use of this command is subtle and requires careful analysis.
// Please consult with experts in cryptography for how to use it securely.
//
// The iv is the initialization vector. The iv must not be empty and its size depends on the
// details of the symmetric encryption scheme.
//
// The data may be longer than block size, EncryptSymmetric will chain
// multiple TPM calls to encrypt the entire blob.
//
// Key handle should point at SymCipher object which is a child of the key (and
// not e.g. RSA key itself).
func EncryptSymmetric(rw io.ReadWriteCloser, keyAuth string, key tpmutil.Handle, iv, data []byte) ([]byte, error) {
	return encryptDecryptSymmetric(rw, keyAuth, key, iv, data, false)
}

// DecryptSymmetric decrypts data using a symmetric key.
//
// WARNING: This command performs low-level cryptographic operations.
// Secure use of this command is subtle and requires careful analysis.
// Please consult with experts in cryptography for how to use it securely.
//
// The iv is the initialization vector. The iv must not be empty and its size
// depends on the details of the symmetric encryption scheme.
//
// The data may be longer than block size, DecryptSymmetric will chain multiple
// TPM calls to decrypt the entire blob.
//
// Key handle should point at SymCipher object which is a child of the key (and
// not e.g. RSA key itself).
func DecryptSymmetric(rw io.ReadWriteCloser, keyAuth string, key tpmutil.Handle, iv, data []byte) ([]byte, error) {
	return encryptDecryptSymmetric(rw, keyAuth, key, iv, data, true)
}

func encodeEncryptDecrypt(keyAuth string, key tpmutil.Handle, iv, data tpmutil.U16Bytes, decrypt bool) ([]byte, error) {
	ha, err := tpmutil.Pack(key)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(keyAuth)})
	if err != nil {
		return nil, err
	}
	// Use encryption key's mode.
	params, err := tpmutil.Pack(decrypt, AlgNull, iv, data)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params)
}

func encodeEncryptDecrypt2(keyAuth string, key tpmutil.Handle, iv, data tpmutil.U16Bytes, decrypt bool) ([]byte, error) {
	ha, err := tpmutil.Pack(key)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(keyAuth)})
	if err != nil {
		return nil, err
	}
	// Use encryption key's mode.
	params, err := tpmutil.Pack(data, decrypt, AlgNull, iv)
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, params)
}

func decodeEncryptDecrypt(resp []byte) ([]byte, []byte, error) {
	var paramSize uint32
	var out, nextIV tpmutil.U16Bytes
	if _, err := tpmutil.Unpack(resp, &paramSize, &out, &nextIV); err != nil {
		return nil, nil, err
	}
	return out, nextIV, nil
}

func encryptDecryptBlockSymmetric(rw io.ReadWriteCloser, keyAuth string, key tpmutil.Handle, iv, data []byte, decrypt bool) ([]byte, []byte, error) {
	Cmd, err := encodeEncryptDecrypt2(keyAuth, key, iv, data, decrypt)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdEncryptDecrypt2, tpmutil.RawBytes(Cmd))
	if err != nil {
		fmt0Err, ok := err.(Error)
		if ok && fmt0Err.Code == RCCommandCode {
			// If TPM2_EncryptDecrypt2 is not supported, fall back to
			// TPM2_EncryptDecrypt.
			Cmd, _ := encodeEncryptDecrypt(keyAuth, key, iv, data, decrypt)
			resp, err = runCommand(rw, TagSessions, CmdEncryptDecrypt, tpmutil.RawBytes(Cmd))
			if err != nil {
				return nil, nil, err
			}
		}
	}
	if err != nil {
		return nil, nil, err
	}
	return decodeEncryptDecrypt(resp)
}

func encryptDecryptSymmetric(rw io.ReadWriteCloser, keyAuth string, key tpmutil.Handle, iv, data []byte, decrypt bool) ([]byte, error) {
	var out, block []byte
	var err error

	for rest := data; len(rest) > 0; {
		if len(rest) > maxDigestBuffer {
			block, rest = rest[:maxDigestBuffer], rest[maxDigestBuffer:]
		} else {
			block, rest = rest, nil
		}
		block, iv, err = encryptDecryptBlockSymmetric(rw, keyAuth, key, iv, block, decrypt)
		if err != nil {
			return nil, err
		}
		out = append(out, block...)
	}

	return out, nil
}

func encodeRSAEncrypt(key tpmutil.Handle, message tpmutil.U16Bytes, scheme *AsymScheme, label string) ([]byte, error) {
	ha, err := tpmutil.Pack(key)
	if err != nil {
		return nil, err
	}
	m, err := tpmutil.Pack(message)
	if err != nil {
		return nil, err
	}
	s, err := scheme.encode()
	if err != nil {
		return nil, err
	}
	if label != "" {
		label += "\x00"
	}
	l, err := tpmutil.Pack(tpmutil.U16Bytes(label))
	if err != nil {
		return nil, err
	}
	return concat(ha, m, s, l)
}

func decodeRSAEncrypt(resp []byte) ([]byte, error) {
	var out tpmutil.U16Bytes
	_, err := tpmutil.Unpack(resp, &out)
	return out, err
}

// RSAEncrypt performs RSA encryption in the TPM according to RFC 3447. The key must be
// a (public) key loaded into the TPM beforehand. Note that when using OAEP with a label,
// a null byte is appended to the label and the null byte is included in the padding
// scheme.
func RSAEncrypt(rw io.ReadWriter, key tpmutil.Handle, message []byte, scheme *AsymScheme, label string) ([]byte, error) {
	Cmd, err := encodeRSAEncrypt(key, message, scheme, label)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdRSAEncrypt, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeRSAEncrypt(resp)
}

func encodeRSADecrypt(sessionHandle, key tpmutil.Handle, password string, message tpmutil.U16Bytes, scheme *AsymScheme, label string) ([]byte, error) {
	ha, err := tpmutil.Pack(key)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: sessionHandle, Attributes: AttrContinueSession, Auth: []byte(password)})
	if err != nil {
		return nil, err
	}
	m, err := tpmutil.Pack(message)
	if err != nil {
		return nil, err
	}
	s, err := scheme.encode()
	if err != nil {
		return nil, err
	}
	if label != "" {
		label += "\x00"
	}
	l, err := tpmutil.Pack(tpmutil.U16Bytes(label))
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, m, s, l)
}

func decodeRSADecrypt(resp []byte) ([]byte, error) {
	var out tpmutil.U16Bytes
	var paramSize uint32
	_, err := tpmutil.Unpack(resp, &paramSize, &out)
	return out, err
}

// RSADecrypt performs RSA decryption in the TPM according to RFC 3447. The key must be
// a private RSA key in the TPM with FlagDecrypt set. Note that when using OAEP with a
// label, a null byte is appended to the label and the null byte is included in the
// padding scheme.
func RSADecrypt(rw io.ReadWriter, key tpmutil.Handle, password string, message []byte, scheme *AsymScheme, label string) ([]byte, error) {
	return RSADecryptWithSession(rw, HandlePasswordSession, key, password, message, scheme, label)
}

// RSADecryptWithSession performs RSA decryption in the TPM according to RFC 3447. The key must be
// a private RSA key in the TPM with FlagDecrypt set. Note that when using OAEP with a
// label, a null byte is appended to the label and the null byte is included in the
// padding scheme.
func RSADecryptWithSession(rw io.ReadWriter, sessionHandle, key tpmutil.Handle, password string, message []byte, scheme *AsymScheme, label string) ([]byte, error) {
	Cmd, err := encodeRSADecrypt(sessionHandle, key, password, message, scheme, label)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdRSADecrypt, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeRSADecrypt(resp)
}

func encodeECDHKeyGen(key tpmutil.Handle) ([]byte, error) {
	return tpmutil.Pack(key)
}

func decodeECDHKeyGen(resp []byte) (*ECPoint, *ECPoint, error) {
	// Unpack z and pub as TPM2B_ECC_POINT, which is a TPMS_ECC_POINT with a total size prepended.
	var z2B, pub2B tpmutil.U16Bytes
	_, err := tpmutil.Unpack(resp, &z2B, &pub2B)
	if err != nil {
		return nil, nil, err
	}
	var zPoint, pubPoint ECPoint
	_, err = tpmutil.Unpack(z2B, &zPoint.XRaw, &zPoint.YRaw)
	if err != nil {
		return nil, nil, err
	}
	_, err = tpmutil.Unpack(pub2B, &pubPoint.XRaw, &pubPoint.YRaw)
	if err != nil {
		return nil, nil, err
	}
	return &zPoint, &pubPoint, nil
}

// ECDHKeyGen generates an ephemeral ECC key, calculates the ECDH point multiplcation of the
// ephemeral private key and a loaded public key, and returns the public ephemeral point along with
// the coordinates of the resulting point.
func ECDHKeyGen(rw io.ReadWriter, key tpmutil.Handle) (zPoint, pubPoint *ECPoint, err error) {
	Cmd, err := encodeECDHKeyGen(key)
	if err != nil {
		return nil, nil, err
	}
	resp, err := runCommand(rw, TagNoSessions, CmdECDHKeyGen, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, nil, err
	}
	return decodeECDHKeyGen(resp)
}

func encodeECDHZGen(key tpmutil.Handle, password string, inPoint ECPoint) ([]byte, error) {
	ha, err := tpmutil.Pack(key)
	if err != nil {
		return nil, err
	}
	auth, err := encodeAuthArea(AuthCommand{Session: HandlePasswordSession, Attributes: AttrContinueSession, Auth: []byte(password)})
	if err != nil {
		return nil, err
	}
	p, err := tpmutil.Pack(inPoint)
	if err != nil {
		return nil, err
	}
	// Pack the TPMS_ECC_POINT as a TPM2B_ECC_POINT.
	p2B, err := tpmutil.Pack(tpmutil.U16Bytes(p))
	if err != nil {
		return nil, err
	}
	return concat(ha, auth, p2B)
}

func decodeECDHZGen(resp []byte) (*ECPoint, error) {
	var paramSize uint32
	// Unpack a TPM2B_ECC_POINT, which is a TPMS_ECC_POINT with a total size prepended.
	var z2B tpmutil.U16Bytes
	_, err := tpmutil.Unpack(resp, &paramSize, &z2B)
	if err != nil {
		return nil, err
	}
	var zPoint ECPoint
	_, err = tpmutil.Unpack(z2B, &zPoint.XRaw, &zPoint.YRaw)
	if err != nil {
		return nil, err
	}
	return &zPoint, nil
}

// ECDHZGen performs ECDH point multiplication between a private key held in the TPM and a given
// public point, returning the coordinates of the resulting point. The key must have FlagDecrypt
// set.
func ECDHZGen(rw io.ReadWriter, key tpmutil.Handle, password string, inPoint ECPoint) (zPoint *ECPoint, err error) {
	Cmd, err := encodeECDHZGen(key, password, inPoint)
	if err != nil {
		return nil, err
	}
	resp, err := runCommand(rw, TagSessions, CmdECDHZGen, tpmutil.RawBytes(Cmd))
	if err != nil {
		return nil, err
	}
	return decodeECDHZGen(resp)
}

// DictionaryAttackLockReset cancels the effect of a TPM lockout due to a number
// of successive authorization failures, by setting the lockout counter to zero.
// The command requires Lockout Authorization and only one lockoutAuth authorization
// failure is allowed for this command during a lockoutRecovery interval.
// Lockout Authorization value by default is empty and can be changed via
// a call to HierarchyChangeAuth(HandleLockout).
func DictionaryAttackLockReset(rw io.ReadWriter, auth AuthCommand) error {
	ha, err := tpmutil.Pack(HandleLockout)
	if err != nil {
		return err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return err
	}
	Cmd, err := concat(ha, encodedAuth)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdDictionaryAttackLockReset, tpmutil.RawBytes(Cmd))
	return err
}

// DictionaryAttackParameters changes the lockout parameters.
// The command requires Lockout Authorization and has same authorization policy
// as in DictionaryAttackLockReset.
func DictionaryAttackParameters(rw io.ReadWriter, auth AuthCommand, maxTries, recoveryTime, lockoutRecovery uint32) error {
	ha, err := tpmutil.Pack(HandleLockout)
	if err != nil {
		return err
	}
	encodedAuth, err := encodeAuthArea(auth)
	if err != nil {
		return err
	}
	params, err := tpmutil.Pack(maxTries, recoveryTime, lockoutRecovery)
	if err != nil {
		return err
	}
	Cmd, err := concat(ha, encodedAuth, params)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagSessions, CmdDictionaryAttackParameters, tpmutil.RawBytes(Cmd))
	return err
}

// PolicyCommandCode indicates that the authorization will be limited to a specific command code
func PolicyCommandCode(rw io.ReadWriter, session tpmutil.Handle, cc tpmutil.Command) error {
	data, err := tpmutil.Pack(session, cc)
	if err != nil {
		return err
	}
	_, err = runCommand(rw, TagNoSessions, CmdPolicyCommandCode, data)
	return err
}
