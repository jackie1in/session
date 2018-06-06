package session

import (
	"runtime"
	"database/sql"
)

// Session 会话
type Session struct {
	DB *sql.DB
	Tx *sql.Tx
	F  *runtime.Func
}

// Begin 开启事务：如果已经开启，则对事务不进行任何操作
func (s *Session) Begin() error {
	if s.Tx == nil {
		tx, err := s.DB.Begin()
		if err != nil {
			return err
		}
		s.Tx = tx

		// 记录下首次开启事务的函数
		pc, _, _, _ := runtime.Caller(1)
		s.F = runtime.FuncForPC(pc)
	}
	return nil
}

// Rollback 回滚事务
func (s *Session) Rollback() error {
	if s.Tx != nil {
		return s.Tx.Rollback()
	}
	return nil
}

// Commit 提交事务：如果提交事务的函数和开启事务的函数在一个函数栈内，则提交事务，否则，不提交
func (s *Session) Commit() error {
	if s.Tx != nil {
		pc, _, _, _ := runtime.Caller(1)
		f:=runtime.FuncForPC(pc)
		if s.F == f {
			err := s.Tx.Commit()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Exec 执行sql语句，如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *Session) Exec(query string, args ...interface{}) (sql.Result, error) {
	if s.Tx != nil {
		return s.Tx.Exec(query, args...)
	}
	return s.DB.Exec(query, args...)
}

// QueryRow 查询单条数据，始终以非事务方式执行（查询都已非事务方式执行）
func (s *Session) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.Tx.QueryRow(query, args...)
}

// Query 查询数据，始终以非事务方式执行
func (s *Session) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.DB.Query(query, args...)
}
