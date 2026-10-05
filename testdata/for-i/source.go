package main

import (
	"errors"
)

func main() {
	for i := 0; i < 100; i++ {
		err := errors.New("my error")
		/*error*/
	}
}
