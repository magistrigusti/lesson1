package main

import (
	"text/scanner"

	"ton-lessons2/internal/app"

)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	if err := app.InitApp(); err != nil {
		return err
	}

	scanner, err := scanner.NewScanner()
	if err != nil {
		return err
	}

	scanner.Listen()

	return nil
}
