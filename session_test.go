package session

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"fmt"
	"database/sql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:Liu123456@tcp(localhost:3306)/test?charset=utf8")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

type User struct {
	number string
	name   string
	ege    int
	sex    int
}

func UserService() *userService{
	return &userService{Session{DB: db}}
}

type userService struct {
	Session
}

func (s *userService) Insert(user User) error {
	_, err := s.Exec("insert into user(number,name,ege,sex) values(?,?,?,?)",
		user.number, user.name, user.ege, user.sex)
	return err
}

func (s *userService) Get(number string) (*User, error) {
	row := db.QueryRow("select number,name,ege,sex from user where number = ?", number)
	user := new(User)
	err := row.Scan(&user.number, &user.name, &user.ege, &user.sex)
	return user, err
}

func (s *userService) Add(user1, user2 User) error {
	s.Begin()

	err := s.Insert(user1)
	if err != nil {
		s.Rollback()
		return err
	}

	err = s.Insert(user2)
	if err != nil {
		s.Rollback()
		return err
	}

	s.Commit()
	return nil
}

// Do 事务
func (s *userService)Do() {
	user, err := s.Get("1")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

	err = s.Insert(User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
	}
}

// DoTx 事务
func (s *userService)DoTx() {
	err := s.Begin()
	if err != nil {
		fmt.Println(err)
	}

	user, err := s.Get("1")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

	err = s.Insert(User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
	}

	s.Commit()
}

// DoNestingTx 嵌套事务
func (s *userService)DoNestingTx() {

	err := s.Begin()
	if err != nil {
		fmt.Println(err)
	}

	err = s.Insert(User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		s.Rollback()
	}

	err = s.Add(User{number: "1", name: "1", ege: 1, sex: 1}, User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		s.Rollback()
	}

	s.Commit()
}

// DoNestingTx 嵌套事务
func DoNoNestingTx() {
	tx,err := db.Begin()
	if err != nil {
		fmt.Println(err)
	}

	user:=User{number: "1", name: "1", ege: 1, sex: 1}
	_, err = tx.Exec("insert into user(number,name,ege,sex) values(?,?,?,?)",
		user.number, user.name, user.ege, user.sex)
	if err != nil {
		tx.Rollback()
	}

	_, err = tx.Exec("insert into user(number,name,ege,sex) values(?,?,?,?)",
		user.number, user.name, user.ege, user.sex)
	if err != nil {
		tx.Rollback()
	}

	_, err = tx.Exec("insert into user(number,name,ege,sex) values(?,?,?,?)",
		user.number, user.name, user.ege, user.sex)
	if err != nil {
		tx.Rollback()
	}

	tx.Commit()
}

// TestDo 测试非事务方式
func TestDo(t *testing.T) {
	userService :=UserService()
	userService.Do()
}

// TestDoTx 测试事务方式
func TestDoTx(t *testing.T) {
	userService :=UserService()
	userService.DoTx()
}

// TestDoNestingTx 测试嵌套事务
func TestDoNestingTx(t *testing.T) {
	userService:=UserService()
	userService.DoNestingTx()
}

func BenchmarkDoNestingTx(b *testing.B) {
	for i:=0;i<b.N;i++{
		userService:=UserService()
		userService.DoNestingTx()
	}
}

func BenchmarkNoDoNestingTx(b *testing.B) {
	for i:=0;i<b.N;i++{
		DoNoNestingTx()
	}
}
// Do 解释
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
	session.Commit()               // 提交事务与开启事务在一个函数中，提交事务
}
