// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

import (
	"context"
	"errors"
	"math/big"

	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/ethclient"
	"github.com/ava-labs/subnet-evm/interfaces"
	"github.com/ava-labs/subnet-evm/precompile/contracts/warp"
	"github.com/ethereum/go-ethereum/common"
)

var WarpPrecompileLogFilter = warp.WarpABI.Events["SendWarpMessage"].ID
var ErrInvalidLog = errors.New("invalid warp message log")

// WarpBlockInfo describes the block height and logs needed to process Warp messages.
// WarpBlockInfo instances are populated by the subscriber, and forwared to the
// listener to process
type WarpBlockInfo struct {
	BlockNumber uint64
	WarpLogs    []types.Log
}

// WarpLogInfo describes the transaction information for the Warp message
// sent on the source chain, and includes the Warp Message payload bytes
// WarpLogInfo instances are either derived from the logs of a block or
// from the manual Warp message information provided via configuration
type WarpLogInfo struct {
	SourceAddress    common.Address
	UnsignedMsgBytes []byte
}

// Extract Warp logs from the block, if they exist
func NewWarpBlockInfo(header *types.Header, ethClient ethclient.Client) (*WarpBlockInfo, error) {
	var (
		logs []types.Log
		err  error
	)
	// Check if the block contains warp logs, and fetch them from the client if it does
	if header.Bloom.Test(WarpPrecompileLogFilter[:]) {
		logs, err = ethClient.FilterLogs(context.Background(), interfaces.FilterQuery{
			Topics:    [][]common.Hash{{WarpPrecompileLogFilter}},
			Addresses: []common.Address{warp.ContractAddress},
			FromBlock: big.NewInt(int64(header.Number.Uint64())),
			ToBlock:   big.NewInt(int64(header.Number.Uint64())),
		})
		if err != nil {
			return nil, err
		}
	}
	return &WarpBlockInfo{
		BlockNumber: header.Number.Uint64(),
		WarpLogs:    logs,
	}, nil
}

// Extract the Warp message information from the raw log
func NewWarpLogInfo(log types.Log) (*WarpLogInfo, error) {
	if len(log.Topics) != 3 {
		return nil, ErrInvalidLog
	}
	if log.Topics[0] != WarpPrecompileLogFilter {
		return nil, ErrInvalidLog
	}

	return &WarpLogInfo{
		// BytesToAddress takes the last 20 bytes of the byte array if it is longer than 20 bytes
		SourceAddress:    common.BytesToAddress(log.Topics[1][:]),
		UnsignedMsgBytes: log.Data,
	}, nil
}