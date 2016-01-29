package main

import (
	"bankcode"
)

func main() {
	err := bankcode.Update()
	if err != nil {
		panic(err)
	}
}
