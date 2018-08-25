package session

import (
	"database/sql"
	)

const beginStatus = 1

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
	db       *sql.DB
	tx       *sql.Tx
	sign     int8
	snapshot int8
}

// Begin 开启事务，如果事务没有开启，开启事务；如果事务已经开启，对事务不做任何操作
func (s *Session) Begin() error {
	if s.tx == nil {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		s.tx = tx
		s.sign = beginStatus
		return nil
	}
	s.sign++
	s.snapshot = s.sign
	return nil
}

// Rollback 回滚事务
func (s *Session) Rollback() error {
	if s.tx != nil {
		if s.sign == s.snapshot {
			err:= s.tx.Rollback()
			if err!=nil{
				return err
			}
			s.tx = nil
			return nil
		}
	}
	return nil
}

// Commit 提交事务：如果提交事务的函数和开启事务的函数在一个函数栈内，则提交事务，否则，不提交
func (s *Session) Commit() error {
	if s.tx != nil {
		if s.sign == beginStatus {
			err := s.tx.Commit()
			if err != nil {
				return err
			}
			s.tx = nil
			return nil
		} else {
			s.sign--
		}
		return nil
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

// Prepare 预执行，如果已经开启事务，就以事务方式执行，如果没有开启事务，就以非事务方式执行
func (s *Session) Prepare(query string) (*sql.Stmt, error) {
	if s.tx != nil {
		return s.tx.Prepare(query)
	}
	return s.db.Prepare(query)
}
