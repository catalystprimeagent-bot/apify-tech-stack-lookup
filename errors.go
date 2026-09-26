package main

import "errors"

var (
	errEmptyInput = errors.New("empty input")
	errNoHost     = errors.New("no hostname in URL")
)
