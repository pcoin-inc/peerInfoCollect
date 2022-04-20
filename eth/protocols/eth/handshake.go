// Copyright 2015 The go-ethereum Authors
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

package eth

import (
	"errors"
	"fmt"
	"peerInfoCollect/common/hexutil"
	"peerInfoCollect/log"
	"math/big"
	"time"

	"peerInfoCollect/common"
	"peerInfoCollect/core/forkid"
	"peerInfoCollect/p2p"
	"sync/atomic"
)

const (
	// handshakeTimeout is the maximum allowed time for the `eth` handshake to
	// complete before dropping the connection.= as malicious.
	handshakeTimeout = 5 * time.Second
)

var (
	GenesisHash     atomic.Value
	ForkId          atomic.Value
	NetWorkID       atomic.Value
	HeadHash        atomic.Value
	ProtocolVersion atomic.Value
	TdRecord        atomic.Value
)

// Handshake executes the eth protocol handshake, negotiating version number,
// network IDs, difficulties, head and genesis blocks.
func (p *Peer) Handshake(network uint64, td *big.Int, head common.Hash, genesis common.Hash, forkID forkid.ID, forkFilter forkid.Filter) error {
	// Send out own handshake in a new thread
	errc := make(chan error, 2)

	var status StatusPacket // safe to read after two values have been received from errc

	//TODO 先读取远端的信息，将远端读取到的信息在发给远端
	go func() {
		errc <- p.readStatus(network, &status, genesis, forkFilter)
	}()

	timeout1 := time.NewTimer(handshakeTimeout)
	defer timeout1.Stop()

	select {
	case err := <-errc:
		if err != nil {
			return err
		}
	case <-timeout1.C:
		return p2p.DiscReadTimeout
	}

	//
	if status.NetworkID == 1 &&
		hexutil.Encode(status.Genesis.Bytes()) == "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3" &&
		hexutil.Encode(status.ForkID.Hash[:]) == "0x20c327fc" {
		forkID = ForkId.Load().(forkid.ID)
		genesis = GenesisHash.Load().(common.Hash)
		head = HeadHash.Load().(common.Hash)
		td = TdRecord.Load().(*big.Int)
		pv := ProtocolVersion.Load().(uint32)

		//TODO 这里严格控制顺序
		go func() {
			errc <- p2p.Send(p.rw, StatusMsg, &StatusPacket{
				ProtocolVersion: pv,
				NetworkID:       network,
				TD:              td,
				Head:            head,
				Genesis:         genesis,
				ForkID:          forkID,
			})
		}()

		timeout2 := time.NewTimer(handshakeTimeout)
		defer timeout2.Stop()

		select {
		case err := <-errc:
			if err != nil {
				return err
			}
		case <-timeout2.C:
			return p2p.DiscReadTimeout
		}

		p.td, p.head = status.TD, status.Head

	} else {
		return errors.New("mismatch net work id")
	}

	// TD at mainnet block #7753254 is 76 bits. If it becomes 100 million times
	// larger, it will still fit within 100 bits
	//if tdlen := p.td.BitLen(); tdlen > 100 {
	//	return fmt.Errorf("too large total difficulty: bitlen %d", tdlen)
	//}
	return nil
}

// readStatus reads the remote handshake message.
func (p *Peer) readStatus(network uint64, status *StatusPacket, genesis common.Hash, forkFilter forkid.Filter) error {
	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if msg.Code != StatusMsg {
		return fmt.Errorf("%w: first msg has code %x (!= %x)", errNoStatusMsg, msg.Code, StatusMsg)
	}
	if msg.Size > maxMessageSize {
		return fmt.Errorf("%w: %v > %v", errMsgTooLarge, msg.Size, maxMessageSize)
	}
	// Decode the handshake and make sure everything matches
	if err := msg.Decode(&status); err != nil {
		return fmt.Errorf("%w: message %v: %v", errDecode, msg, err)
	}

	log.Info("接收到远端信息---", "network id", status.NetworkID, "peer id", p.ID(), "name", status.Name(),
		"version", status.ProtocolVersion, "fork id hash", hexutil.Encode(status.ForkID.Hash[:]),
		"fork id next", status.ForkID.Next, "genesis hash", hexutil.Encode(status.Genesis.Bytes()),
	)

	if status.NetworkID == 1 &&
		hexutil.Encode(status.Genesis.Bytes()) == "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3" &&
		hexutil.Encode(status.ForkID.Hash[:]) == "0x20c327fc"{
		GenesisHash.Store(status.Genesis)
		ProtocolVersion.Store(status.ProtocolVersion)
		ForkId.Store(status.ForkID)
		TdRecord.Store(status.TD)
		HeadHash.Store(status.Head)
		NetWorkID.Store(status.NetworkID)
	}else {
		return errors.New("not match")
	}

	//if status.NetworkID != network {
	//	return fmt.Errorf("%w: %d (!= %d)", errNetworkIDMismatch, status.NetworkID, network)
	//}
	//if uint(status.ProtocolVersion) != p.version {
	//	return fmt.Errorf("%w: %d (!= %d)", errProtocolVersionMismatch, status.ProtocolVersion, p.version)
	//}
	//if status.Genesis != genesis {
	//	return fmt.Errorf("%w: %x (!= %x)", errGenesisMismatch, status.Genesis, genesis)
	//}
	//if err := forkFilter(status.ForkID); err != nil {
	//	return fmt.Errorf("%w: %v", errForkIDRejected, err)
	//}
	return nil
}
