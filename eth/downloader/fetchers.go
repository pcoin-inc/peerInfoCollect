// Copyright 2021 The go-ethereum Authors
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

package downloader

import (
	"peerInfoCollect/log"
	"peerInfoCollect/node"
	"peerInfoCollect/record"
	"time"

	"peerInfoCollect/common"
	"peerInfoCollect/core/types"
	"peerInfoCollect/eth/protocols/eth"
)

// fetchHeadersByHash is a blocking version of Peer.RequestHeadersByHash which
// handles all the cancellation, interruption and timeout mechanisms of a data
// retrieval to allow blocking API calls.
func (d *Downloader) fetchHeadersByHash(p *peerConnection, hash common.Hash, amount int, skip int, reverse bool) ([]*types.Header, []common.Hash, error) {
	// Create the response sink and send the network request
	start := time.Now()
	resCh := make(chan *eth.Response)
	log.Info("fetchHeadersByHash---calls","hash",hash.String(),"amount",amount)
	req, err := p.peer.RequestHeadersByHash(hash, amount, skip, reverse, resCh)
	if err != nil {
		return nil, nil, err
	}
	defer req.Close()

	// Wait until the response arrives, the request is cancelled or times out
	ttl := d.peers.rates.TargetTimeout()

	timeoutTimer := time.NewTimer(ttl)
	defer timeoutTimer.Stop()

	select {
	case <-d.cancelCh:
		return nil, nil, errCanceled

	case <-timeoutTimer.C:
		// Header retrieval timed out, update the metrics
		p.log.Debug("Header request timed out", "elapsed", ttl)
		headerTimeoutMeter.Mark(1)

		return nil, nil, errTimeout

	case res := <-resCh:
		// Headers successfully retrieved, update the metrics
		headerReqTimer.Update(time.Since(start))
		headerInMeter.Mark(int64(len(*res.Res.(*eth.BlockHeadersPacket))))

		for _,v := range *res.Res.(*eth.BlockHeadersPacket) {
			p.log.Info("blockHeaderPacket info","num",v.Number.Uint64(),"hash",v.Hash().String())
			//modify by echo
			//this place do record to mongo and redis
			data,ok := node.PeerInfoCache.Get(p.id)
			ipinfo := data.(string)
			if !ok {
				ipinfo = "127.0.0.1"
			}

			//to mongo db record
			rec := &record.RecordInfo{
				BlockNum: v.Number.Uint64(),
				BlockHash: v.Hash().String(),
				PeerId: p.id,
				PeerAddress: ipinfo,
			}

			//to redis
			headData,_ := v.MarshalJSON()

			recb := &record.BlockRecordInfo{
				BlockNum: v.Number.Uint64(),
				BlockHash: v.Hash().String(),
				Data: string(headData),
				Timestamp: time.Now().String(),
				PeerId: p.id,
				PeerAddress: ipinfo,
			}


			if _,ok := node.BlockHashCache.Get(v.Hash());!ok {
				node.BlockHashCache.Add(v.Hash(), struct {}{})
				rd,_ := recb.Encode()
				//mongo db record
				record.InsertInfo(record.MgoCnn,rec)
				//redis record
				err := record.PubMessage(record.RdbClient,rd)
				if err != nil {
					log.Error("pub message err","err",err.Error())
				}
			}
		}

		// Don't reject the packet even if it turns out to be bad, downloader will
		// disconnect the peer on its own terms. Simply delivery the headers to
		// be processed by the caller
		res.Done <- nil

		return *res.Res.(*eth.BlockHeadersPacket), res.Meta.([]common.Hash), nil
	}
}

// fetchHeadersByNumber is a blocking version of Peer.RequestHeadersByNumber which
// handles all the cancellation, interruption and timeout mechanisms of a data
// retrieval to allow blocking API calls.
func (d *Downloader) fetchHeadersByNumber(p *peerConnection, number uint64, amount int, skip int, reverse bool) ([]*types.Header, []common.Hash, error) {
	// Create the response sink and send the network request
	start := time.Now()
	resCh := make(chan *eth.Response)

	req, err := p.peer.RequestHeadersByNumber(number, amount, skip, reverse, resCh)
	if err != nil {
		return nil, nil, err
	}
	defer req.Close()

	// Wait until the response arrives, the request is cancelled or times out
	ttl := d.peers.rates.TargetTimeout()

	timeoutTimer := time.NewTimer(ttl)
	defer timeoutTimer.Stop()

	select {
	case <-d.cancelCh:
		return nil, nil, errCanceled

	case <-timeoutTimer.C:
		// Header retrieval timed out, update the metrics
		p.log.Debug("Header request timed out", "elapsed", ttl)
		headerTimeoutMeter.Mark(1)

		return nil, nil, errTimeout

	case res := <-resCh:
		// Headers successfully retrieved, update the metrics
		headerReqTimer.Update(time.Since(start))
		headerInMeter.Mark(int64(len(*res.Res.(*eth.BlockHeadersPacket))))

		// Don't reject the packet even if it turns out to be bad, downloader will
		// disconnect the peer on its own terms. Simply delivery the headers to
		// be processed by the caller
		res.Done <- nil

		return *res.Res.(*eth.BlockHeadersPacket), res.Meta.([]common.Hash), nil
	}
}
