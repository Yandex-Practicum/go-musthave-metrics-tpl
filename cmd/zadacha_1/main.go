package main

import (
	"errors"
	"fmt"
	"math"
)

// User — пользователь в системе.
type User struct {
	FirstName string
	LastName  string
}

type Relationship string

// Возможные роли в семье.
const (
	Father      = Relationship("father")
	Mother      = Relationship("mother")
	Child       = Relationship("child")
	GrandMother = Relationship("grandMother")
	GrandFather = Relationship("grandFather")
)

// Family описывает семью.
type Family struct {
	Members map[Relationship]Person
}

type Person struct {
	FirstName string
	LastName  string
	Age       int
}

var (
	// ErrRelationshipAlreadyExists возвращает ошибку, если роль уже занята.
	ErrRelationshipAlreadyExists = errors.New("relationship already exists")
)

func main() {
	//задание 1
	v := Abs(3)
	fmt.Println(v)

	//задание 2
	u := User{
		FirstName: "Misha",
		LastName:  "Popov",
	}

	fmt.Println(u.FullName())

	//задание 3
	f := Family{}
	err := f.AddNew(Father, Person{
		FirstName: "Misha",
		LastName:  "Popov",
		Age:       56,
	})
	fmt.Println(f, err)

	err = f.AddNew(Father, Person{
		FirstName: "Drug",
		LastName:  "Mishi",
		Age:       57,
	})
	fmt.Println(f, err)
}

// Abs возвращает абсолютное значение.
// Например: 3.1 => 3.1, -3.14 => 3.14, -0 => 0.
func Abs(value float64) float64 {
	return math.Abs(value)
}

// FullName возвращает имя и фамилию пользователя.
func (u User) FullName() string {
	return u.FirstName + " " + u.LastName
}

func (f *Family) AddNew(r Relationship, p Person) error {
	if f.Members == nil {
		f.Members = map[Relationship]Person{}
	}
	if _, ok := f.Members[r]; ok {
		return ErrRelationshipAlreadyExists
	}
	f.Members[r] = p
	return nil
}
