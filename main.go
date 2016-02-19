package main

import (
	"github.com/airware/vili/app"
)

func main() {
	s := vili.New()
	if s == nil {
		return
	}
	s.Start()
}
