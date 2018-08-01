package session

import (
	"database/sql"
	"runtime"
)

// Session 会话工厂
type SessionFactory struct {
	*sql.DB
}

// NewSession 创建一个Session
func NewSessionFactory(driverName, dataSourseName string) (*SessionFactory, error) {
	db, err := sql.Open(driverName, dataSourseName)
	if err != nil {
		panic(err)
	}
	factory := new(SessionFactory)
	factory.DB = db
	return factory, nil
}

// NewSession 创建一个Session
func (sf *SessionFactory) GetSession() *Session {
	session := new(Session)
	session.db = sf.DB
	return session
}

// Session 会话
type Session struct {
	db *sql.DB
	tx *sql.Tx
	f  *runtime.Func
}

// Begin 开启事务：如果已经开启，则对事务不进行任何操作
func (s *Session) Begin() error {
	if s.tx == nil {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		s.tx = tx

		// 记录下首次开启事务的函数
		pc, _, _, _ := runtime.Caller(1)
		s.f = runtime.FuncForPC(pc)
	}
	return nil
}

// Rollback 回滚事务
func (s *Session) Rollback() error {
	if s.tx != nil {
		return s.tx.Rollback()
		s.tx = nil
	}
	return nil
}

// Commit 提交事务：如果提交事务的函数和开启事务的函数在一个函数栈内，则提交事务，否则，不提交
func (s *Session) Commit() error {
	if s.tx != nil {
		pc, _, _, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		if s.f == f {
			err := s.tx.Commit()
			if err != nil {
				return err
			}
			s.tx = nil
		}
	}
	return nil
}

// Exec 执行sql语句，如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.tx != nil {
		return s.tx.Exec(query, args...)
	}
	return s.db.Exec(query, args...)
}

// QueryRow 查询单条数据，始终以非事务方式执行（查询都以非事务方式执行）
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}

// Query 查询数据，始终以非事务方式执行
func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

// Prepare 预执行
func (s *Session) Prepare(query string) (*sql.Stmt, error) {
	if s.tx != nil {
		return s.tx.Prepare(query)
	}
	return s.db.Prepare(query)
}
