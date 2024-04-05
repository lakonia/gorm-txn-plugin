package gorm_txn_plugin

import (
	"context"
	"gorm.io/gorm"
)

const pluginName = "gorm:txn-plugin"

type Txn struct {
	*gorm.DB
}

func (txn *Txn) Name() string {
	return pluginName
}

func (txn *Txn) Initialize(db *gorm.DB) error {
	txn.DB = db
	txn.registerCallbacks(db)
	return nil
}

func (txn *Txn) registerCallbacks(db *gorm.DB) {
	txn.Callback().Create().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Update().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Delete().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Raw().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Query().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Row().Before("*").Register(pluginName, txn.beginTxnIfRequired)
}

func (txn *Txn) beginTxnIfRequired(db *gorm.DB) {
	ctx := db.Statement.Context
	if isTxn(ctx) {
		if txn := getDB(ctx); txn == nil {
			db.Statement.SkipDefaultTransaction = true
			db.Statement.DisableNestedTransaction = false
			ctx = withDB(ctx, db.Begin())
		}

		db.Statement.ConnPool = getDB(ctx).Statement.ConnPool
	}
}

func RunInTxn(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx = beginTxn(ctx)
	err := fn(ctx)

	if db := getDB(ctx); db != nil {
		if err != nil {
			if err := db.Rollback().Error; err != nil {
				return err
			}

			return err
		}

		return db.Commit().Error
	}

	return err
}
