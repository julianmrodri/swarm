// Copyright 2020 The Swarm Authors
// This file is part of the Swarm library.
//
// The Swarm library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Swarm library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Swarm library. If not, see <http://www.gnu.org/licenses/>.

package pss

import (
	"crypto/rand"
	"encoding/json"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
)

// TODO: can we re-use some types here?
type trojanHeaders struct {
	span  []byte
	nonce []byte
}

// TODO: can we re-use some types here?
type trojanMessage struct {
	length  []byte
	topic   []byte
	payload []byte
	padding []byte
}

type trojanData struct {
	trojanHeaders
	trojanMessage // TODO: this should be encrypted
}

type trojanChunk struct {
	address chunk.Address
	trojanData
}

// newTrojanChunk creates a new trojan chunk structure for the given address and message
func newTrojanChunk(address chunk.Address, message trojanMessage) (*trojanChunk, error) {
	chunk := &trojanChunk{
		address: address,
		trojanData: trojanData{
			trojanHeaders: newTrojanHeaders(),
			trojanMessage: message,
		},
	}
	// find nonce for chunk
	if err := chunk.setNonce(); err != nil {
		return nil, err
	}
	return chunk, nil
}

// newTrojanHeaders creates an empty trojan headers struct
func newTrojanHeaders() trojanHeaders {
	// TODO: what should be the value of this?
	span := make([]byte, 8)
	// create initial nonce
	nonce := make([]byte, 32)

	return trojanHeaders{
		span:  span,
		nonce: nonce,
	}
}

// setNonce determines the nonce so that when the trojan chunk fields are hashed, it falls in the neighbourhood of the trojan chunk address
func (tc *trojanChunk) setNonce() error {
	// init BMT hash function
	BMThashFunc := storage.MakeHashFunc(storage.BMTHash)()
	// iterate nonce
	nonce, err := iterateNonce(tc, BMThashFunc)
	if err != nil {
		return err
	}
	tc.nonce = nonce
	return nil
}

// iterateNonce iterates the BMT hash of the trojan chunk fields until the desired nonce is found
func iterateNonce(tc *trojanChunk, hashFunc storage.SwarmHash) ([]byte, error) {
	var emptyNonce []byte

	// start out with random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return emptyNonce, err
	}

	// hash nonce
	if _, err := hashFunc.Write(nonce); err != nil {
		return emptyNonce, err
	}
	nonce = hashFunc.Sum(nil)

	// TODO: iterate nonce
	return nonce, nil
}

// toContentAddressedChunk creates a new addressed chunk structure with the given trojan message content serialized as its data
func (tc *trojanChunk) toContentAddressedChunk() (chunk.Chunk, error) {
	var emptyChunk = chunk.NewChunk([]byte{}, []byte{})

	chunkData, err := json.Marshal(tc.trojanData)
	if err != nil {
		return emptyChunk, err
	}
	return chunk.NewChunk(tc.address, chunkData), nil
}

// func (tm *trojanData) MarshalJSON() ([]byte, error) {
// 	// append first 40 bytes
// 	s := append(tm.span, tm.nonce...)
// 	m, err := json.Marshal(tm.messageCyphertext)
// 	if err != nil {
// 		return []byte{}, err
// 	}
// 	// append with marshalled message
// 	return json.Marshal(append(s, m...))
// }

// func (tm *trojanChunk) UnmarshalJSON(data []byte) error {
// 	var b []byte
// 	if err := json.Unmarshal(data, &b); err != nil {
// 		return err
// 	}
// 	tm.span = b[0:8]   // first 8 bytes are span
// 	tm.nonce = b[8:40] // following 32 bytes are nonce

// 	// rest of the bytes are message
// 	var m message.Message
// 	if err := json.Unmarshal(b[40:], &m); err != nil {
// 		return err
// 	}
// 	tm.messageCyphertext = m
// 	return nil
// }
