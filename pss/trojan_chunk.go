// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package pss

import (
	"crypto/rand"
	"encoding/json"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/pss/message"
	"github.com/ethersphere/swarm/storage"
)

type pssEnvelope struct {
	// headers ? missing
	message []byte
}

type trojanHeaders struct {
	span           []byte
	nonce          []byte
	decryptionHint []byte
}

type trojanMessage struct {
	trojanHeaders
	pssMsgCyphertext message.Message
}

// creates a new trojan message structure for the given address
func newTrojanMessage(address chunk.Address, pssMessage message.Message) (*trojanMessage, error) {
	// create initial trojan headers
	trojanHeaders := newTrojanHeaders()
	// find nonce for headers and address
	if err := findMessageNonce(address, trojanHeaders); err != nil {
		return nil, err
	}
	// cypher pss message, plain for now
	pssMsgCyphertext := pssMessage
	return &trojanMessage{
		trojanHeaders:    *trojanHeaders,
		pssMsgCyphertext: pssMsgCyphertext,
	}, nil
}

// creates empty trojan headers struct
func newTrojanHeaders() *trojanHeaders {
	// create span, empty for now
	span := make([]byte, 8)
	// create initial nonce
	nonce := make([]byte, 32)
	// create decryption hint, empty for now
	decryptionHint := make([]byte, 32)

	return &trojanHeaders{
		span:           span,
		nonce:          nonce,
		decryptionHint: decryptionHint,
	}
}

// determines the nonce so that when the trojan message is hashed, it falls in the neighbourhood of the given address
func findMessageNonce(address chunk.Address, headers *trojanHeaders) error {
	// init BMT hash function
	hashFunc := storage.MakeHashFunc(storage.BMTHash)
	// start out with random nonce
	nonce := make([]byte, 32)
	_, err := rand.Read(nonce)
	if err != nil {
		return err
	}
	return nil
}

var emptyChunk = chunk.NewChunk([]byte{}, []byte{})

// creates a new addressed chunk structure with the given trojan message content serialized as its data
func newTrojanChunk(address chunk.Address, message trojanMessage) (chunk.Chunk, error) {
	chunkData, err := json.Marshal(message) // what is the correct way of serializing a trojan message?
	if err != nil {
		return emptyChunk, err
	}
	return chunk.NewChunk(address, chunkData), nil
}
