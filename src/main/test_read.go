package main

import (
	"fmt"
	"github.com/jotitan/music_server/logger"
)

func test() {
	fmt.Println(pa(1))
	fmt.Println(pa(2))
	fmt.Println(pa(0))
}

func pa(v int) *int {
	defer func() {
		if err := recover(); err != nil {
			logger.GetLogger().Error("Error when", err)
		}
	}()
	return throwPanic(v)
}

func throwPanic(v int) *int {
	if v == 0 {
		panic("Fuck of")
	}
	val := 1
	return &val
}
