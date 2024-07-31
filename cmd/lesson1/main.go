package main

import (
	"github.com/sirupsen/logrus"

	"ton-lessons2/internal/app"

)

type TestStruct struct {
	Abc string
}

func main() {
	smthg, err := run()
	if err != nil {
		panic(err)
	}

	logrus.Info(smthg)
}

func run() (TestStruct, error) {
	if err := app.InitApp(); err != nil {
		return TestStruct{}, err
	}

	return TestStruct{
		Abc: "123",
	}, nil
}
