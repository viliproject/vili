package main

import (
	"github.com/viliproject/vili/app"
)

func main() {
	s := vili.New()
	if s == nil {
		return
	}
	s.Start()
}
