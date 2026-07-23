package main

import (
	"errors"
	"time"
)

type MyTime time.Time

func GetMyTime() (MyTime, error) {
	err := errors.New("my error")
	/*error*/
}
