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
	UserService = userService{Session{DB: db}}
}

type User struct {
	number string
	name   string
	ege    int
	sex    int
}

var UserService userService

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

// TestDo 测试非事务方式
func TestDo(t *testing.T) {
	user, err := UserService.Get("1")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

	err = UserService.Insert(User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
	}
}

// TestDoTx 测试事务方式
func TestDoTx(t *testing.T) {
	err := UserService.Begin()
	if err != nil {
		fmt.Println(err)
	}

	user, err := UserService.Get("1")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(user)

	err = UserService.Insert(User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		fmt.Println(err)
	}

	UserService.Commit()
}

// TestDoNestingTx 测试嵌套事务
func TestDoNestingTx(t *testing.T) {
	err := UserService.Begin()
	if err != nil {
		fmt.Println(err)
	}

	err = UserService.Insert(User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		UserService.Rollback()
	}

	err = UserService.Add(User{number: "1", name: "1", ege: 1, sex: 1}, User{number: "1", name: "1", ege: 1, sex: 1})
	if err != nil {
		UserService.Rollback()
	}

	UserService.Commit()
}
