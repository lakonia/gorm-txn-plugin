package gorm_txn

import (
	"context"
	"gorm.io/gorm"
)

type txnCtxObj struct {
	db *gorm.DB
}

type txnCtx struct{}

var txnCtxKey = txnCtx(struct{}{})

func beginTxn(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, txnCtxKey, &txnCtxObj{})
}

func isTxn(ctx context.Context) bool {
	if ctx == nil {
		return false
	}

	_, ok := ctx.Value(txnCtxKey).(*txnCtxObj)
	if !ok {
		return false
	}

	return true
}

func withDB(ctx context.Context, db *gorm.DB) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	v, ok := ctx.Value(txnCtxKey).(*txnCtxObj)
	if !ok {
		v = &txnCtxObj{}
	}

	v.db = db

	return context.WithValue(ctx, txnCtxKey, v)
}

func getDB(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}

	v, ok := ctx.Value(txnCtxKey).(*txnCtxObj)
	if !ok || v == nil {
		return nil
	}

	return v.db
}
