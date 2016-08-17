package xorm

import (
	"github.com/go-xorm/xorm"
)

type Table interface {
	TableName() string
}

// 返回设置好表名的数据库session实例
// 注：不传session参数时，返回的session只能使用一次就会被自动关闭。
func DBSession(t Table, session ...*xorm.Session) *xorm.Session {
	if len(session) > 0 {
		return session[0].Table(t.TableName())
	}
	return DefaultDB().Table(t.TableName())
}
func DBSession2(tableName string, session ...*xorm.Session) *xorm.Session {
	if len(session) > 0 {
		return session[0].Table(tableName)
	}
	return DefaultDB().Table(tableName)
}

// 无事务的数据库session回调函数
// session为空时，内部自动创建
func DBCallback(fn func(*xorm.Session) error, session ...*xorm.Session) error {
	if fn == nil {
		return nil
	}
	var sess *xorm.Session
	if len(session) > 0 {
		sess = session[0]
	}
	if sess == nil {
		sess = DefaultDB().NewSession()
		defer sess.Close()
	}
	return fn(sess)
}

// 启用事务的数据库session回调函数
// 不传入session参数时，内部自动创建事务处理
// 传入session参数时，内部不调用回滚操作，请在函数外部统一调用
func DBTransactCallback(fn func(*xorm.Session) error, session ...*xorm.Session) (err error) {
	if fn == nil {
		return
	}
	var sess *xorm.Session
	if len(session) > 0 {
		sess = session[0]
	}
	if sess == nil {
		sess = DefaultDB().NewSession()
		defer sess.Close()
		err = sess.Begin()
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				sess.Rollback()
				return
			}
			err = sess.Commit()
		}()
	}
	err = fn(sess)
	return
}
