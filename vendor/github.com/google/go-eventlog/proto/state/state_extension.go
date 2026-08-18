// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

// Package state contains the definitions and utilities related to extracting
// information from an event log.
package state

import (
	"crypto"

	"github.com/google/go-tpm/legacy/tpm2"
)

// CryptoHash converts the TCG registry hash identifier to a crypto.Hash.
func (ha HashAlgo) CryptoHash() (crypto.Hash, error) {
	tcgHash := tpm2.Algorithm(uint16(ha))
	cryptoHash, err := tcgHash.Hash()
	if err != nil {
		return crypto.Hash(0), err
	}
	return cryptoHash, nil
}
