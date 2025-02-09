/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package rpc

import (
	"encoding/hex"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/polynetwork/poly/common/config"
	"github.com/polynetwork/poly/common/log"
	"github.com/polynetwork/poly/consensus/vbft"
	scom "github.com/polynetwork/poly/core/store/common"
	"github.com/polynetwork/poly/core/types"
	ontErrors "github.com/polynetwork/poly/errors"
	bactor "github.com/polynetwork/poly/http/base/actor"
	bcomn "github.com/polynetwork/poly/http/base/common"
	berr "github.com/polynetwork/poly/http/base/error"
)

//get best block hash
func GetBestBlockHash(params []interface{}) map[string]interface{} {
	hash := bactor.CurrentBlockHash()
	return responseSuccess(hash.ToHexString())
}

// get block by height or hash
// Input JSON string examples for getblock method as following:
//   {"jsonrpc": "2.0", "method": "getblock", "params": [1], "id": 0}
//   {"jsonrpc": "2.0", "method": "getblock", "params": ["aabbcc.."], "id": 0}
func GetBlock(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var err error
	var hash common.Uint256
	switch (params[0]).(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		hash = bactor.GetBlockHashFromStore(index)
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		// block hash
	case string:
		str := params[0].(string)
		hash, err = common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	block, err := bactor.GetBlockFromStore(hash)
	if err != nil {
		return responsePack(berr.UNKNOWN_BLOCK, "unknown block")
	}
	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				return responseSuccess(bcomn.GetBlockInfo(block))
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	return responseSuccess(common.ToHexString(block.ToArray()))
}

//get block height
func GetBlockCount(params []interface{}) map[string]interface{} {
	height := bactor.GetCurrentBlockHeight()
	return responseSuccess(height + 1)
}

// get latest block msg snapshot
func GetLatestBlockMsgsSnap(params []interface{}) map[string]interface{} {
	result := vbft.GetLatestBlockMsgsSnap()
	return responseSuccess(result)
}

// get cross state root
func GetCrossStateRoot(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch (params[0]).(type) {
	// block height
	case float64:
		height := uint32(params[0].(float64))
		result, err := bactor.GetCrossStateRoot(height)
		if err != nil {
			return responsePack(berr.UNKNOWN_BLOCK, err.Error())
		}
		return responseSuccess(result)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

}

//get block hash
// A JSON example for getblockhash method as following:
//   {"jsonrpc": "2.0", "method": "getblockhash", "params": [1], "id": 0}
func GetBlockHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash := bactor.GetBlockHashFromStore(height)
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.UNKNOWN_BLOCK, "")
		}
		return responseSuccess(hash.ToHexString())
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

//get node connection count
func GetConnectionCount(params []interface{}) map[string]interface{} {
	count, err := bactor.GetConnectionCnt()
	if err != nil {
		log.Errorf("GetConnectionCount error:%s", err)
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responseSuccess(count)
}

func GetRawMemPool(params []interface{}) map[string]interface{} {
	txs := []*bcomn.Transactions{}
	txpool := bactor.GetTxsFromPool(false)
	for _, t := range txpool {
		txs = append(txs, bcomn.TransArryByteToHexString(t))
	}
	if len(txs) == 0 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	return responseSuccess(txs)
}

//get memory pool transaction count
func GetMemPoolTxCount(params []interface{}) map[string]interface{} {
	count, err := bactor.GetTxnCount()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, nil)
	}
	return responseSuccess(count)
}

//get memory pool transaction state
func GetMemPoolTxState(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		txEntry, err := bactor.GetTxFromPool(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, "unknown transaction")
		}
		attrs := []bcomn.TXNAttrInfo{}
		for _, t := range txEntry.Attrs {
			attrs = append(attrs, bcomn.TXNAttrInfo{t.Height, int(t.Type), int(t.ErrCode)})
		}
		info := bcomn.TXNEntryInfo{attrs}
		return responseSuccess(info)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

// get raw transaction in raw or json
// A JSON example for getrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "getrawtransaction", "params": ["transactioin hash in hex"], "id": 0}
func GetRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var tx *types.Transaction
	var height uint32
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		h, t, err := bactor.GetTxnWithHeightByTxHash(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_TRANSACTION, fmt.Sprintf("unknown transaction:%s", err))
		}
		height = h
		tx = t
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	if len(params) >= 2 {
		switch (params[1]).(type) {
		case float64:
			json := uint32(params[1].(float64))
			if json == 1 {
				txinfo := bcomn.TransArryByteToHexString(tx)
				txinfo.Height = height
				return responseSuccess(txinfo)
			}
		default:
			return responsePack(berr.INVALID_PARAMS, "")
		}
	}
	return responseSuccess(common.ToHexString(tx.Raw))
}

//get storage from contract
//   {"jsonrpc": "2.0", "method": "getstorage", "params": ["code hash", "key"], "id": 0}
func GetStorage(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	var address common.Address
	var key []byte
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		var err error
		address, err = bcomn.GetAddress(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

	switch params[1].(type) {
	case string:
		str := params[1].(string)
		hex, err := hex.DecodeString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		key = hex
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	value, err := bactor.GetStorageItem(address, key)
	if err != nil {
		if err == scom.ErrNotFound {
			return responseSuccess(nil)
		}
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(common.ToHexString(value))
}

//send raw transaction
// A JSON example for sendrawtransaction method as following:
//   {"jsonrpc": "2.0", "method": "sendrawtransaction", "params": ["raw transactioin in hex"], "id": 0}
func SendRawTransaction(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	var hash common.Uint256
	switch params[0].(type) {
	case string:
		str := params[0].(string)
		raw, err := common.HexToBytes(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		txn, err := types.TransactionFromRawBytes(raw)
		if err != nil {
			return responsePack(berr.INVALID_TRANSACTION, "")
		}
		hash = txn.Hash()
		log.Debugf("SendRawTransaction recv %s", hash.ToHexString())
		if txn.TxType == types.Invoke {
			if len(params) > 1 {
				preExec, ok := params[1].(float64)
				if ok && preExec == 1 {
					result, err := bactor.PreExecuteContract(txn)
					if err != nil {
						log.Infof("PreExec: ", err)
						return responsePack(berr.SMARTCODE_ERROR, err.Error())
					}
					return responseSuccess(bcomn.ConvertPreExecuteResult(result))
				}
			}
		}

		log.Debugf("SendRawTransaction send to txpool %s", hash.ToHexString())
		if errCode, desc := bcomn.SendTxToPool(txn); errCode != ontErrors.ErrNoError {
			log.Warnf("SendRawTransaction verified %s error: %s", hash.ToHexString(), desc)
			return responsePack(int64(errCode), desc)
		}
		log.Debugf("SendRawTransaction verified %s", hash.ToHexString())
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responseSuccess(hash.ToHexString())
}

//get node version
func GetNodeVersion(params []interface{}) map[string]interface{} {
	return responseSuccess(config.Version)
}

// get networkid
func GetNetworkId(params []interface{}) map[string]interface{} {
	return responseSuccess(config.DefConfig.P2PNode.NetworkId)
}

//get smartconstract event
func GetSmartCodeEvent(params []interface{}) map[string]interface{} {
	if !config.DefConfig.Common.EnableEventLog {
		return responsePack(berr.INVALID_METHOD, "")
	}
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	switch (params[0]).(type) {
	// block height
	case float64:
		height := uint32(params[0].(float64))
		eventInfos, err := bactor.GetEventNotifyByHeight(height)
		if err != nil {
			if err == scom.ErrNotFound {
				return responseSuccess(nil)
			}
			return responsePack(berr.INTERNAL_ERROR, "")
		}
		eInfos := make([]*bcomn.ExecuteNotify, 0, len(eventInfos))
		for _, eventInfo := range eventInfos {
			_, notify := bcomn.GetExecuteNotify(eventInfo)
			eInfos = append(eInfos, &notify)
		}
		return responseSuccess(eInfos)
		//txhash
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		eventInfo, err := bactor.GetEventNotifyByTxHash(hash)
		if err != nil {
			if scom.ErrNotFound == err {
				return responseSuccess(nil)
			}
			return responsePack(berr.INTERNAL_ERROR, "")
		}
		_, notify := bcomn.GetExecuteNotify(eventInfo)
		return responseSuccess(notify)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

//get block height by transaction hash
func GetBlockHeightByTxHash(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}

	switch (params[0]).(type) {
	// tx hash
	case string:
		str := params[0].(string)
		hash, err := common.Uint256FromHexString(str)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		height, _, err := bactor.GetTxnWithHeightByTxHash(hash)
		if err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		return responseSuccess(height)
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.INVALID_PARAMS, "")
}

//get merkle proof by the block height to be verified
func GetMerkleProof(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	height, ok := params[0].(float64)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	rootHeight, ok := params[1].(float64)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	if height >= rootHeight || height == 0 {
		responsePack(berr.INVALID_PARAMS, fmt.Sprintf("Cannot get proof of block hash at height: %d when the block root is at height: %d", uint32(height), uint32(rootHeight)))
	}
	proof, err := bactor.GetMerkleProof(uint32(height), uint32(rootHeight))
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, "")
	}
	return responseSuccess(bcomn.MerkleProof{"MerkleProof", hex.EncodeToString(proof)})
}

//get cross chain state proof
func GetCrossStatesProof(params []interface{}) map[string]interface{} {
	if len(params) < 2 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	height, ok := params[0].(float64)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	str, ok := params[1].(string)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	key, err := hex.DecodeString(str)
	if err != nil {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	proof, err := bactor.GetCrossStatesProof(uint32(height), key)
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}
	return responseSuccess(bcomn.MerkleProof{"CrossStatesProof", hex.EncodeToString(proof)})
}

func GetHeaderByHeight(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	height, ok := params[0].(float64)
	if !ok {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	header, err := bactor.GetHeaderByHeight(uint32(height))
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, err.Error())
	}

	return responseSuccess(hex.EncodeToString(header.ToArray()))
}

//get block transactions by height
func GetBlockTxsByHeight(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	case float64:
		height := uint32(params[0].(float64))
		hash := bactor.GetBlockHashFromStore(height)
		if hash == common.UINT256_EMPTY {
			return responsePack(berr.INVALID_PARAMS, "")
		}
		block, err := bactor.GetBlockFromStore(hash)
		if err != nil {
			return responsePack(berr.UNKNOWN_BLOCK, "")
		}
		return responseSuccess(bcomn.GetBlockTransactions(block))
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
}

// get state merkle root by block height
func GetStateMerkleRoot(params []interface{}) map[string]interface{} {

	if len(params) != 1 {
		return responsePack(berr.INVALID_PARAMS, nil)
	}
	switch params[0].(type) {
	// block height
	case float64:
		index := uint32(params[0].(float64))
		root, err := bactor.GetStateMerkleRoot(index)
		if err != nil {
			return responsePack(berr.INTERNAL_ERROR, err.Error())
		}

		return responseSuccess(root.ToHexString())
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}

}
