// Copyright 2019 The Swarm Authors
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
package outbox

import (
	"time"

	"github.com/ethersphere/swarm/pss/message"
)

type outboxMsg struct {
	msg       *message.Message
	startedAt time.Time
}

// NewOutboxMessage creates a new outbox message wrapping a pss message.
func NewOutboxMessage(msg *message.Message) *outboxMsg {
	return &outboxMsg{
		msg:       msg,
		startedAt: time.Now(),
	}
}