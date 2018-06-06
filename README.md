# golang的嵌套事务管理

golang的事务管理是一件很麻烦的事，，能不能像Java那样，通过Spring管理事务，可以支持事务扩散机制，最近琢磨了一下，写了一个demo,用来管理golang的事务，使其支持golang事务的嵌套调用。

其思想很简单，对于所有的写数据库操作，如果发现事务是开启的，就以事务方式执行数据库操作，如果事务没有开启，就以非事务的方式执行数据库操作。对于所有的读数据库的操作，都已非事务的方式进行（用以提高数据库的执行的性能）。

对于嵌套事务的执行，流程大体是这样的：

在第一次开启事务的时候，记录下来开启事务时函数，对于后面的开启事务操作，不进行任何操作，在提交事务的时候，判断提交事务所处的函数是否和打开事务的函数一致，如果函数不一致，就不提交事务，如果函数一致，就提交事务。

下面是一个演示示例：

```go
func Do(session *Session) {
	session.Begin()                // 开启事务
	func(session *Session) {
		session.Begin()            // 事务已经开启，不进行操作
		func(session *Session) {
			session.Begin()        // 事务已经开启，不进行操作
			session.Commit()       // 提交事务与开启事务不在一个函数中，不进行操作
		}(session)          
		session.Commit()           // 提交事务与开启事务不在一个函数中，不进行操作
	}(session)
	session.Commit()               // 提交事务与开启事务不在一个函数中，提交事务
}
```

我只是写了一个简单demo，这里贴出实现代码：

```go
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

```

