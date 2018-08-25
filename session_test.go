package session

import (
	_ "github.com/go-sql-driver/mysql"
	"testing"
	"fmt"
	)

var sf *SessionFactory

func init() {
	var err error
	sf, err = NewSessionFactory("mysql", "root:Liu123456@tcp(localhost:3306)/test?charset=utf8")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

type User struct {
	mobile string
	name   string
	age    int
	sex    int
}

type UserService struct {
	session *Session
}

func NewUserService() *UserService {
	return &UserService{sf.GetSession()}
}

func (s *UserService) Insert(user User) error {
	_, err := s.session.Exec("insert into user(mobile,name,age,sex) values(?,?,?,?)",
		user.mobile, user.name, user.age, user.sex)
	return err
}

func (s *UserService) Get(mobile string) (*User, error) {
	row := s.session.QueryRow("select mobile,name,age,sex from user where mobile = ?", mobile)
	user := new(User)
	err := row.Scan(&user.mobile, &user.name, &user.age, &user.sex)
	return user, err
}

func (s *UserService) AddInTx(user1, user2 User) error {
	err:=s.session.Begin()
	if err != nil {
		return err
	}
	defer s.session.Rollback()

	err = s.Insert(user1)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = s.Insert(user2)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// s.session.Commit()
	return nil
}

// Do 事务
func (s *UserService)Do() {
	user, err := s.Get("1")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

	err = s.Insert(User{mobile: "1", name: "1", age: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
	}
}

// DoTx 事务
func (s *UserService)DoTx() {
	err := s.session.Begin()
	if err != nil {
		fmt.Println(err)
	}
	defer s.session.Rollback()

	user, err := s.Get("1")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

	err = s.Insert(User{mobile: "1", name: "1", age: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
	}

	s.session.Commit()
}

// DoNestingTx 嵌套事务
func (s *UserService)DoNestingTx() {
	err := s.session.Begin()
	if err != nil {
		fmt.Println(err)
	}
	defer s.session.Rollback()

	err = s.Insert(User{mobile: "1", name: "1", age: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.AddInTx(User{mobile: "1", name: "1", age: 1, sex: 1}, User{mobile: "1", name: "1", age: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
		return
	}

	err=s.session.Commit()
	if err!=nil{
		fmt.Println(err)
		return
	}
}



// TestDo 测试非事务方式
func TestDo(t *testing.T) {
	userService :=NewUserService()
	userService.Do()
}

// TestDoTx 测试事务方式
func TestDoTx(t *testing.T) {
	userService :=NewUserService()
	userService.DoTx()
}

// TestDoNestingTx 测试嵌套事务
func TestDoNestingTx(t *testing.T) {
	userService:=NewUserService()
	userService.DoNestingTx()
}


func BenchmarkDoNestingTx(b *testing.B) {
	for i:=0;i<b.N;i++{
		userService:=NewUserService()
		userService.DoNestingTx()
	}
}

// DoNestingTx 原生事务方法
func DoNoNestingTx() {
	tx,err := sf.DB.Begin()
	if err != nil {
		fmt.Println(err)
	}

	user:=User{mobile: "1", name: "1", age: 1, sex: 1}
	_, err = tx.Exec("insert into user(mobile,name,age,sex) values(?,?,?,?)",
		user.mobile, user.name, user.age, user.sex)
	if err != nil {
		tx.Rollback()
	}

	_, err = tx.Exec("insert into user(mobile,name,age,sex) values(?,?,?,?)",
		user.mobile, user.name, user.age, user.sex)
	if err != nil {
		tx.Rollback()
	}

	_, err = tx.Exec("insert into user(mobile,name,age,sex) values(?,?,?,?)",
		user.mobile, user.name, user.age, user.sex)
	if err != nil {
		tx.Rollback()
	}

	tx.Commit()
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
