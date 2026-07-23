package main

import (
	"errors"
)

func GetChan() (<-chan int, error) {
	err := errors.New("my error")
	/*error*/
}
