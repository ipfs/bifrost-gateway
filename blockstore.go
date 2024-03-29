package main

import (
	"context"
	"errors"
	"time"

	"github.com/ipfs/bifrost-gateway/lib"
	blockstore "github.com/ipfs/boxo/blockstore"
	exchange "github.com/ipfs/boxo/exchange"
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"go.uber.org/zap/zapcore"
)

var errNotImplemented = errors.New("not implemented")

const GetBlockTimeout = time.Second * 60

func newExchange(bs blockstore.Blockstore) (exchange.Interface, error) {
	return &exchangeBsWrapper{bstore: bs}, nil
}

type exchangeBsWrapper struct {
	bstore blockstore.Blockstore
}

func (e *exchangeBsWrapper) GetBlock(ctx context.Context, c cid.Cid) (blocks.Block, error) {
	ctx, cancel := context.WithTimeout(ctx, GetBlockTimeout)
	defer cancel()

	if goLog.Level().Enabled(zapcore.DebugLevel) {
		goLog.Debugw("block requested from remote blockstore", "cid", c.String())
	}

	blk, err := e.bstore.Get(ctx, c)
	if err != nil {
		return nil, lib.GatewayError(err)
	}
	return blk, nil
}

func (e *exchangeBsWrapper) GetBlocks(ctx context.Context, cids []cid.Cid) (<-chan blocks.Block, error) {
	out := make(chan blocks.Block)

	go func() {
		defer close(out)
		for _, c := range cids {
			blk, err := e.GetBlock(ctx, c)
			if err != nil {
				return
			}
			out <- blk
		}
	}()
	return out, nil
}

func (e *exchangeBsWrapper) NotifyNewBlocks(ctx context.Context, blks ...blocks.Block) error {
	return nil
}

func (e *exchangeBsWrapper) Close() error {
	return nil
}

var _ exchange.Interface = (*exchangeBsWrapper)(nil)
