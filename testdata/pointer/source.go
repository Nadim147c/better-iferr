package main

import (
	"bytes"
	"errors"
)

func GetPointer() (*bytes.Buffer, error) {
	err := errors.New("my error")
	/*error*/
}
