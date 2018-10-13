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

func (s *UserService) AddInTx(user1, user2 User) error {
	err := s.session.Begin()
	if err != nil {
		return err
	}
	defer s.session.Rollback()

	err = s.Insert(user1)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// return errors.New("err") 回滚测试

	err = s.Insert(user2)
	if err != nil {
		fmt.Println(err)
		return err
	}

	s.session.Commit()
	return nil
}

// DoNestingTx 嵌套事务
func (s *UserService) DoNestingTx() {
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

	err = s.session.Commit()
	if err != nil {
		fmt.Println(err)
		return
	}
}

// TestDoNestingTx 测试嵌套事务
func TestDoNestingTx(t *testing.T) {
	userService := NewUserService()
	userService.DoNestingTx()
}
