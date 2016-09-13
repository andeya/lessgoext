package sqlx

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

/**
 * Database transaction callback function
 */
func DBTransactCallback(fn func(*sqlx.Tx) error, tx ...*sqlx.Tx) (err error) {
	if fn == nil {
		return
	}
	var _tx *sqlx.Tx
	if len(tx) > 0 {
		_tx = tx[0]
	}
	if _tx == nil {
		_tx, err = DefaultDB().Beginx()
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				_tx.Rollback()
			} else {
				_tx.Commit()
			}
		}()
	}
	err = fn(_tx)
	return
}

/**
 * Compatible database callback function
 */

func DBCallback(fn func(DBTX) error, tx ...*sqlx.Tx) error {
	if fn == nil {
		return nil
	}
	if len(tx) > 0 && tx[0] != nil {
		return fn(tx[0])
	}
	return fn(DefaultDB())
}

type DBTX interface {
	BindNamed(query string, arg interface{}) (string, []interface{}, error)
	DriverName() string
	Get(dest interface{}, query string, args ...interface{}) error
	MustExec(query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	Preparex(query string) (*sqlx.Stmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
}
