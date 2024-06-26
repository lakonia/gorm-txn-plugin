package gorm_txn

import (
	"context"
	"gorm.io/gorm"
)

const pluginName = "gorm:txn-plugin"

type GormTxnPlugin struct {
	*gorm.DB

	disableNestedTransaction bool
	debug bool
}

func (txn *GormTxnPlugin) Name() string {
	return pluginName
}

func (txn *GormTxnPlugin) Initialize(db *gorm.DB) error {
	txn.DB = db
	txn.registerCallbacks(db)
	return nil
}


func (txn *GormTxnPlugin) WithSkipDefaultTransaction(flag bool) *GormTxnPlugin {
	txn.disableNestedTransaction = flag
	return txn
}

func (txn *GormTxnPlugin) Debug() *GormTxnPlugin {
	txn.debug = true
	return txn
}

func (txn *GormTxnPlugin) registerCallbacks(db *gorm.DB) {
	txn.Callback().Create().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Update().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Delete().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Raw().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Query().Before("*").Register(pluginName, txn.beginTxnIfRequired)
	txn.Callback().Row().Before("*").Register(pluginName, txn.beginTxnIfRequired)
}

func (txn *GormTxnPlugin) beginTxnIfRequired(db *gorm.DB) {
	ctx := db.Statement.Context
	if isTxn(ctx) {
		if getDB(ctx) == nil {
			db.Statement.SkipDefaultTransaction = true
			db.Statement.DisableNestedTransaction = txn.disableNestedTransaction

			db = db.Begin()
			if txn.debug {
				db = db.Debug()
			}

			ctx = withDB(ctx, db)
		}

		db.Statement.ConnPool = getDB(ctx).Statement.ConnPool
	}
}

func RunInTxn(ctx context.Context, fn func(ctx context.Context) error) error {
	if !isTxn(ctx) {
		ctx = beginTxn(ctx)
	}
	
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
