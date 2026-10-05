package main

import (
	"errors"
)

func main() {
	s := [_]int{0xDE, 0xAD, 0xBE, 0xEF}
	for i, x := range s {
		err := errors.New("my error")
		/*error*/
	}
}
