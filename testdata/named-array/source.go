package main

import "errors"

type NamedArray [3]int

func GetNamedArray() (NamedArray, error) {
	err := errors.New("my error")
	/*error*/
}
