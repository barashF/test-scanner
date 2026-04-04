package transaction

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/internships-backend/test-backend-barashF/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type manager struct {
	pool *pgxpool.Pool
}

type (
	txKey         struct{}
	txRequiredKey struct{}
)

func NewManager(pool *pgxpool.Pool) *manager {
	return &manager{pool: pool}
}

func (tm *manager) InTransaction(ctx context.Context, options *model.TransactionOptions, fn func(context.Context) error) (err error) {
	var pgxOptions pgx.TxOptions

	if options != nil {
		if options.IsolationLevel != nil {
			pgxOptions.IsoLevel = pgx.TxIsoLevel(*options.IsolationLevel)
		}
		if options.AccessMode != nil {
			pgxOptions.AccessMode = pgx.TxAccessMode(*options.AccessMode)
		}
		if options.DeferrableMode != nil {
			pgxOptions.DeferrableMode = pgx.TxDeferrableMode(*options.DeferrableMode)
		}
	}

	tx, e := tm.pool.BeginTx(ctx, pgxOptions)
	if e != nil {
		return e
	}

	defer func() {
		if p := recover(); p != nil {
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				log.Println(fmt.Errorf("tx rollback err: %w", rollbackErr))
			}
			panic(p)
		} else if err != nil {
			log.Println("tx rollback, err: ", err)

			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				log.Println(fmt.Errorf("tx rollback err: %w", err))
			}
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				log.Println(fmt.Errorf("tx commit err: %w", err))
			}
		}
	}()

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)
	ctxWithTx = context.WithValue(ctxWithTx, txRequiredKey{}, true)
	err = fn(ctxWithTx)

	return
}

func (tm *manager) GetConn(ctx context.Context) (conn Connection, err error) {
	required, err := tm.isTransactionRequired(ctx)
	if err != nil {
		return nil, err
	}

	if required {
		tx := tm.getTx(ctx)
		return tx, nil
	}

	return tm.pool, nil
}

func (tm *manager) isTransactionRequired(ctx context.Context) (bool, error) {
	required, ok := ctx.Value(txRequiredKey{}).(bool)
	if ok {
		if required == true {
			tx := tm.getTx(ctx)
			if tx == nil {
				return false, errors.New("transaction required, but not found")
			}
		}
		return required, nil
	} else {
		return false, nil
	}
}

func (tm *manager) getTx(ctx context.Context) *pgxpool.Tx {
	if tx, ok := ctx.Value(txKey{}).(*pgxpool.Tx); ok {
		return tx
	}
	return nil
}
