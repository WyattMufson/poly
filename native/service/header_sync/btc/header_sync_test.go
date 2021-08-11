/*
 * Copyright (C) 2021 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package btc

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/polynetwork/poly/account"
	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/core/store/leveldbstore"
	"github.com/polynetwork/poly/core/store/overlaydb"
	"github.com/polynetwork/poly/core/types"
	"github.com/polynetwork/poly/native"
	scom "github.com/polynetwork/poly/native/service/header_sync/common"
	"github.com/polynetwork/poly/native/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	acct *account.Account = account.NewAccount("")

	getNativeFunc = func(args []byte, db *storage.CacheDB) *native.NativeService {
		if db == nil {
			store, _ := leveldbstore.NewMemLevelDBStore()
			db = storage.NewCacheDB(overlaydb.NewOverlayDB(store))
		}
		ns, _ := native.NewNativeService(db, new(types.Transaction), 0, 0, common.Uint256{0}, 0, args, false)
		return ns
	}

	getHeaders = func() []*wire.BlockHeader {
		res := make([]*wire.BlockHeader, 0)
		for _, v := range chain {
			b, _ := hex.DecodeString(v)
			h := &wire.BlockHeader{}
			_ = h.BtcDecode(bytes.NewBuffer(b), wire.ProtocolVersion, wire.LatestEncoding)
			res = append(res, h)
		}

		return res
	}

	getHdrsInBytes = func() [][]byte {
		res := make([][]byte, len(chain))
		for i, v := range chain {
			b, _ := hex.DecodeString(v)
			res[i] = b
		}
		return res
	}

	getFork = func() []*wire.BlockHeader {
		res := make([]*wire.BlockHeader, 0)
		for _, v := range fork {
			b, _ := hex.DecodeString(v)
			h := &wire.BlockHeader{}
			_ = h.BtcDecode(bytes.NewBuffer(b), wire.ProtocolVersion, wire.LatestEncoding)
			res = append(res, h)
		}

		return res
	}

	syncGHeader = func() (*native.NativeService, *BTCHandler) {
		var buf bytes.Buffer
		_ = netParam.GenesisBlock.Header.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)

		h := make([]byte, 4)
		binary.BigEndian.PutUint32(h, 0)

		params := new(scom.SyncGenesisHeaderParam)
		params.ChainID = 0
		params.GenesisHeader = append(buf.Bytes(), h...)

		sink := common.NewZeroCopySink(nil)
		params.Serialization(sink)

		ns := getNativeFunc(sink.Bytes(), nil)
		handler := NewBTCHandler()
		_ = handler.SyncGenesisHeader(ns)

		return ns, handler
	}

	syncNormalHeaders = func(db *storage.CacheDB, handler *BTCHandler) {
		hdrs := getHdrsInBytes()
		param := new(scom.SyncBlockHeaderParam)
		param.ChainID = 0
		param.Address = acct.Address
		param.Headers = hdrs

		sink := common.NewZeroCopySink(nil)
		param.Serialization(sink)
		ns := getNativeFunc(sink.Bytes(), db)
		_ = handler.SyncBlockHeader(ns)
	}

	getForkInBytes = func() [][]byte {
		res := make([][]byte, len(fork))
		for i, v := range fork {
			b, _ := hex.DecodeString(v)
			res[i] = b
		}
		return res
	}
)

func TestBTCHandler_SyncGenesisHeader(t *testing.T) {
	var buf bytes.Buffer
	_ = netParam.GenesisBlock.Header.BtcEncode(&buf, wire.ProtocolVersion, wire.LatestEncoding)

	hb := make([]byte, 4)
	binary.BigEndian.PutUint32(hb, 0)

	params := new(scom.SyncGenesisHeaderParam)
	params.ChainID = 0
	params.GenesisHeader = append(buf.Bytes(), hb...)

	sink := common.NewZeroCopySink(nil)
	params.Serialization(sink)

	ns := getNativeFunc(sink.Bytes(), nil)
	handler := NewBTCHandler()
	err := handler.SyncGenesisHeader(ns)
	assert.NoError(t, err)
}

func TestBTCHandler_SyncBlockHeader(t *testing.T) {
	netParam = &chaincfg.RegressionNetParams
	ns, handler := syncGHeader()

	// normal case
	hdrs := getHdrsInBytes()
	param := new(scom.SyncBlockHeaderParam)
	param.ChainID = 0
	param.Address = acct.Address
	param.Headers = hdrs

	sink := common.NewZeroCopySink(nil)
	param.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err := handler.SyncBlockHeader(ns)
	assert.NoError(t, err)

	normal := getHeaders()
	for i, v := range normal {
		sh, _ := GetHeaderByHeight(ns, 0, uint32(i+1))
		assert.Equal(t, v.BlockHash().String(), sh.Header.BlockHash().String(), fmt.Sprintf("wrong header %d", i+1))
	}

	// add 5 forks and best is not changed
	fHdrs := getForkInBytes()
	param.Headers = fHdrs[:5]

	sink.Reset()
	param.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = handler.SyncBlockHeader(ns)
	assert.NoError(t, err)
	best, _ := GetBestBlockHeader(ns, 0)
	assert.Equal(t, normal[len(normal)-1].BlockHash().String(), best.Header.BlockHash().String(), "wrong best")

	forks := getFork()
	for _, v := range forks[:5] {
		sh, _ := GetHeaderByHash(ns, 0, v.BlockHash())
		assert.Equal(t, v.BlockHash().String(), sh.Header.BlockHash().String())
	}

	// add one more fork, best should be changed
	param.Headers = fHdrs[5:6]

	sink.Reset()
	param.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = handler.SyncBlockHeader(ns)
	assert.NoError(t, err)
	best, _ = GetBestBlockHeader(ns, 0)
	assert.Equal(t, forks[5].BlockHash().String(), best.Header.BlockHash().String(), "wrong best")
	for i, v := range forks[:6] {
		sh, _ := GetHeaderByHeight(ns, 0, uint32(i+6))
		assert.Equal(t, v.BlockHash().String(), sh.Header.BlockHash().String(), fmt.Sprintf("wrong header %d", i+6))
	}

	// add replicated header
	err = handler.SyncBlockHeader(ns)
	assert.NoError(t, err)

	// orphan
	orphan, _ := hex.DecodeString("00000020cc29dfd714165a2a9fd3bfcc05d81baaecb72dda02262a538128b01651d49241c62a812dd83899eef572dae19ca22267bfa90a36a427ffccc469a2f99b2b412c45ff235effff7f2000000000")
	param.Headers = [][]byte{orphan}
	sink.Reset()
	param.Serialization(sink)
	ns = getNativeFunc(sink.Bytes(), ns.GetCacheDB())
	err = handler.SyncBlockHeader(ns)
	assert.Error(t, err, "should be error")
}
