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
	"encoding/binary"
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/ethersphere/swarm/chunk"
	"github.com/ethersphere/swarm/storage"
)

// TODO: can we re-use some existing types here?
type trojanHeaders struct {
	span  []byte
	nonce []byte
}

// TODO: can we re-use some existing types here?
type trojanMessage struct {
	length  []byte // this should be 2 bytes, not 8
	topic   []byte // redefine this as a type, alias for byte slice, init func which takes a string
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

// new struct: trojan message
// the iterate nonce func should take the trojan message payload serialization
// internally we use nonce and serialization
// new trojan chunk, takes topic and payload and take it directly from the api
// already in the constructor it will set the length and padding
// and calling on the iterator, take the byte slice, try a nonce, calculate the span
// gives back: it might even give you back the chunk.Chunk type

// setting the span: 4K written as an 8 byte binary little-endian (use the `put` from binary package)
// same one for all the iterations

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
	BMThashFunc := storage.MakeHashFunc(storage.BMTHash)() // init BMT hash function
	err := iterateNonce(tc, BMThashFunc)
	if err != nil {
		return err
	}
	return nil
}

// iterateNonce iterates the BMT hash of the trojan chunk fields until the desired nonce is found
func iterateNonce(tc *trojanChunk, hashFunc storage.SwarmHash) error {
	// start out with random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	tc.nonce = nonce

	// hash trojan chunk fields with different nonces until a desired one is found
	hashWithinNeighbourhood := false // TODO: this could be correct on the 1st try
	// TODO: add limit to tries
	for hashWithinNeighbourhood != true {
		serializedTrojanData, err := json.Marshal(tc.trojanData)
		if err != nil {
			return err
		}
		if _, err := hashFunc.Write(serializedTrojanData); err != nil {
			return err
		}
		hash := hashFunc.Sum(nil)

		// TODO: what is the correct way to check if hash is in the same neighbourhood as trojan chunk address?
		_ = chunk.Proximity(tc.address, hash)

		// we can give a list of 2 bytes prefixes
		// check the first 2 bytes of the hash result
		// from api call: take the length of the targets (complain if they are not all the same length)
		// every iteration: when we get an address, byte compare the first 2 bytes to all the targets, if it matches
		// one: yay. if not: re-roll
		// 16 bits

		// target bit vectors
		// address generated → bit sequence as well
		// check if it is a prefix: targets are 16 bits (arbitrary)
		// if it's 17 bits, then compare it with the target

		// so the target is a partial address

		// TODO: replace placeholder condition
		if true {
			// if nonce found, stop loop
			hashWithinNeighbourhood = true
		} else {
			// else, add 1 to nonce and try again
			// TODO: find non sinful way of adding 1 to byte slice
			// TODO: implement loop-around
			nonceInt := new(big.Int).SetBytes(tc.nonce)
			tc.nonce = nonceInt.Add(nonceInt, big.NewInt(1)).Bytes() // check what happens if this overflows the byte array
		}
	}

	return nil
}

// we call sum function with a byte

// toContentAddressedChunk creates a new addressed chunk structure with the given trojan message content serialized as its data
func (tc *trojanChunk) toContentAddressedChunk() (chunk.Chunk, error) {
	var emptyChunk = chunk.NewChunk([]byte{}, []byte{})

	chunkData, err := json.Marshal(tc.trojanData)
	if err != nil {
		return emptyChunk, err
	}
	return chunk.NewChunk(tc.address, chunkData), nil
}

// equals compares the underlying data of 2 trojanData variables and returns true if they match, false otherwise
// TODO: why doesn't a direct `reflect.DeepEqual` call of the whole variable work?
func (td *trojanData) equals(d *trojanData) bool {
	if !reflect.DeepEqual(td.trojanHeaders, d.trojanHeaders) {
		return false
	}
	if !reflect.DeepEqual(td.trojanMessage, d.trojanMessage) {
		return false
	}
	return true
}

// we can scrap json serialization
//

// UnmarshalJSON serializes a trojanData struct
// TODO: find a more elegant way of serializing trojan data
func (td *trojanData) MarshalJSON() ([]byte, error) {
	// append first 40 bytes, span & nonce
	s := append(td.span, td.nonce...)
	// marshal message
	m, err := json.Marshal(&td.trojanMessage)
	if err != nil {
		return []byte{}, err
	}
	// marshal appended result
	return json.Marshal(append(s, m...))
}

// UnmarshalJSON deserializes a trojanData struct
// TODO: find a more elegant way of de-serializing trojan data
func (td *trojanData) UnmarshalJSON(data []byte) error {
	var b []byte
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	td.span = b[0:8]   // first 8 bytes are span
	td.nonce = b[8:40] // following 32 bytes are nonce

	// rest of the bytes are message
	var m trojanMessage
	if err := json.Unmarshal(b[40:], &m); err != nil {
		return err
	}
	td.trojanMessage = m
	return nil
}

// let's use binary marshalling instead of json

// UnmarshalJSON serializes a trojanMessage struct
// TODO: find a more elegant way of serializing trojan messages
func (tm *trojanMessage) MarshalJSON() ([]byte, error) {
	s := append(tm.length, tm.topic...)
	s = append(s, tm.payload...)
	return json.Marshal(append(s, tm.padding...))
}

// UnmarshalJSON deserializes a trojanMesage struct
// TODO: find a more elegant way of de-serializing trojan messages
func (tm *trojanMessage) UnmarshalJSON(data []byte) error {
	var b []byte
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	tm.length = b[0:8] // first 8 bytes are length
	tm.topic = b[8:40] // following 32 bytes are topic

	// rest of the bytes are payload and padding
	length := binary.BigEndian.Uint64(tm.length)
	payloadEnd := 40 + length
	tm.payload = b[40:payloadEnd]
	tm.padding = b[payloadEnd:]
	return nil
}
