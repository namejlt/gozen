package gozen

import (
	"errors"
	"fmt"
	"testing"
)

func TestGoFuncAll(t *testing.T) {
	name := "golang"
	age := 18
	tag := map[string]string{
		"a": "aaa",
		"b": "bbb",
		"c": "ccc",
	}
	num := 999

	var retANum int
	var retBName string
	var retCTag map[string]string

	err := GoFunc(func() error {
		var err error
		retANum, err = goFuncA(name, age, tag)
		return err
	}, func() error {
		var err error
		retBName, err = goFuncB(num, age, tag)
		return err
	}, func() error {
		var err error
		retCTag, err = goFuncC(name, age, tag)
		return err
	},
	)

	t.Log(retANum)
	t.Log(retBName)
	t.Log(retCTag)
	t.Log(err)
}

func TestGoFuncOneC(t *testing.T) {
	name := "golang"
	age := 18
	tag := map[string]string{
		"a": "aaa",
		"b": "bbb",
		"c": "ccc",
	}

	var retCTag map[string]string

	GoFuncOne(func() error {
		var err error
		retCTag, err = goFuncC(name, age, tag)
		return err
	})

	t.Log(retCTag)
}

func TestGoFuncOneA(t *testing.T) {
	name := "golang"
	age := 18
	tag := map[string]string{
		"a": "aaa",
		"b": "bbb",
		"c": "ccc",
	}

	var retANum int

	GoFuncOne(func() error {
		var err error
		retANum, err = goFuncA(name, age, tag)
		return err
	})

	t.Log(retANum)
}

func goFuncA(name string, age int, tag map[string]string) (num int, err error) {
	fmt.Println("goFuncA", name)
	fmt.Println("goFuncA", age)
	fmt.Println("goFuncA", tag)
	num = 18
	panic("goFuncA")
}

func goFuncB(num int, age int, tag map[string]string) (name string, err error) {
	fmt.Println("goFuncB", num)
	fmt.Println("goFuncB", age)
	fmt.Println("goFuncB", tag)
	name = "18🈲"
	err = errors.New("goFuncB")
	return
}

func goFuncC(name string, age int, tag map[string]string) (data map[string]string, err error) {
	fmt.Println("goFuncC", name)
	fmt.Println("goFuncC", age)
	fmt.Println("goFuncC", tag)
	data = tag
	panic("goFuncC")
	return
}
