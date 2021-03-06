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
	"fmt"
	"peerInfoCollect/common"
	"peerInfoCollect/core"
	"peerInfoCollect/core/types"
	"peerInfoCollect/eth/protocols/eth"
	"peerInfoCollect/log"
	"peerInfoCollect/node"
	"peerInfoCollect/p2p/enode"
	"peerInfoCollect/record"
	"math/big"
	"sync/atomic"
	"time"
)

// ethHandler implements the eth.Backend interface to handle the various network
// packets that are sent as replies or broadcasts.

type ethHandler handler

func (h *ethHandler) Chain() *core.BlockChain { return h.chain }
func (h *ethHandler) TxPool() eth.TxPool      { return h.txpool }

// RunPeer is invoked when a peer joins on the `eth` protocol.
func (h *ethHandler) RunPeer(peer *eth.Peer, hand eth.Handler) error {
	return (*handler)(h).runEthPeer(peer, hand)
}

// PeerInfo retrieves all known `eth` information about a peer.
func (h *ethHandler) PeerInfo(id enode.ID) interface{} {
	if p := h.peers.peer(id.String()); p != nil {
		return p.info()
	}
	return nil
}

// AcceptTxs retrieves whether transaction processing is enabled on the node
// or if inbound transactions should simply be dropped.
func (h *ethHandler) AcceptTxs() bool {
	return atomic.LoadUint32(&h.acceptTxs) == 1
}

// Handle is invoked from a peer's message handler when it receives a new remote
// message that the handler couldn't consume and serve itself.
func (h *ethHandler) Handle(peer *eth.Peer, packet eth.Packet) error {
	// Consume any broadcasts and announces, forwarding the rest to the downloader
	switch packet := packet.(type) {
	case *eth.NewBlockHashesPacket:
		hashes, numbers := packet.Unpack()

		_,ok := node.PeerInfoCache.Get(peer.ID())
		if !ok {
			node.PeerInfoCache.Add(peer.ID(),peer.RemoteAddr().String())
		}
		return h.handleBlockAnnounces(peer, hashes, numbers)

	case *eth.NewBlockPacket:
		log.Info("????????????????????????---","??????num",packet.Block.NumberU64(),"??????hash",packet.Block.Hash().String(),
			"peer id",peer.ID(),"peer ip",peer.RemoteAddr().String(),
		)
		node.BlockHashCache.Add(packet.Block.Hash(), struct {}{})

		//to redis
		headData,_ := packet.Block.Header().MarshalJSON()

		recb := &record.BlockRecordInfo{
			BlockNum: packet.Block.NumberU64(),
			BlockHash: packet.Block.Hash().String(),
			Data: string(headData),
			Timestamp: time.Now().String(),
			PeerId: peer.ID(),
			PeerAddress: peer.RemoteAddr().String(),
		}

		rd,_ := recb.Encode()
		err := record.PubMessage(record.RdbClient,record.ChanBlockID,string(rd))
		if err != nil {
			log.Error("pub message","err",err.Error())
		}

		return h.handleBlockBroadcast(peer, packet.Block, packet.TD)

	case *eth.NewPooledTransactionHashesPacket:
		return h.txFetcher.Notify(peer.ID(), *packet)

	case *eth.TransactionsPacket:
		for _,v := range *packet {
			log.Info("??????????????????---","tx hash",v.Hash().String())
			txData,_ := v.MarshalJSON()
			td := record.TxRecordInfo{
				TxHash: v.Hash().String(),
				Payload: string(txData),
				PeerId: peer.ID(),
				PeerAddr: peer.RemoteAddr().String(),
			}

			data,_  := td.Encode()
			record.PubMessage(record.RdbClient,record.ChanTxID,string(data))
		}
		return h.txFetcher.Enqueue(peer.ID(), *packet, false)

	case *eth.PooledTransactionsPacket:
		for _,v := range *packet{
			log.Info("??????????????????????????????????????????--","tx hash",v.Hash())
			txData,_ := v.MarshalJSON()
			td := record.TxRecordInfo{
				TxHash: v.Hash().String(),
				Payload: string(txData),
				PeerId: peer.ID(),
				PeerAddr: peer.RemoteAddr().String(),
			}
			data,_  := td.Encode()
			record.PubMessage(record.RdbClient,record.ChanTxID,string(data))
		}
		return h.txFetcher.Enqueue(peer.ID(), *packet, true)

	default:
		return fmt.Errorf("unexpected eth packet type: %T", packet)
	}
}

// handleBlockAnnounces is invoked from a peer's message handler when it transmits a
// batch of block announcements for the local node to process.
func (h *ethHandler) handleBlockAnnounces(peer *eth.Peer, hashes []common.Hash, numbers []uint64) error {
	// Drop all incoming block announces from the p2p network if
	// the chain already entered the pos stage and disconnect the
	// remote peer.
	if h.merger.PoSFinalized() {
		// TODO (MariusVanDerWijden) drop non-updated peers after the merge
		return nil
		// return errors.New("unexpected block announces")
	}
	// Schedule all the unknown hashes for retrieval
	var (
		unknownHashes  = make([]common.Hash, 0, len(hashes))
		unknownNumbers = make([]uint64, 0, len(numbers))
	)
	for i := 0; i < len(hashes); i++ {
		if !h.chain.HasBlock(hashes[i], numbers[i]) {
			unknownHashes = append(unknownHashes, hashes[i])
			unknownNumbers = append(unknownNumbers, numbers[i])
		}
	}
	for i := 0; i < len(unknownHashes); i++ {
		flag :=  peer.KnownBlock(unknownHashes[i])
		if !flag {
			log.Info("handle block announce--","peer id",peer.ID(),"unknownNumbers",unknownNumbers[i],"unknownHashes",unknownHashes[i])
			h.blockFetcher.Notify(peer.ID(), unknownHashes[i], unknownNumbers[i], time.Now(), peer.RequestOneHeader, peer.RequestBodies)
		}
	}
	return nil
}

// handleBlockBroadcast is invoked from a peer's message handler when it transmits a
// block broadcast for the local node to process.
func (h *ethHandler) handleBlockBroadcast(peer *eth.Peer, block *types.Block, td *big.Int) error {
	// Drop all incoming block announces from the p2p network if
	// the chain already entered the pos stage and disconnect the
	// remote peer.
	if h.merger.PoSFinalized() {
		// TODO (MariusVanDerWijden) drop non-updated peers after the merge
		return nil
		// return errors.New("unexpected block announces")
	}

	//
	//// Schedule the block for import
	//h.blockFetcher.Enqueue(peer.ID(), block)

	// Assuming the block is importable by the peer, but possibly not yet done so,
	// calculate the head hash and TD that the peer truly must have.
	var (
		trueHead = block.ParentHash()
		trueTD   = new(big.Int).Sub(td, block.Difficulty())
	)
	// Update the peer's total difficulty if better than the previous
	if _, td := peer.Head(); trueTD.Cmp(td) > 0 {
		peer.SetHead(trueHead, trueTD)
		h.chainSync.handlePeerEvent(peer)
	}
	return nil
}
