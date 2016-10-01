package main

import (
	"github.com/anticpp/gomqtt"
)

func main() {

	var err error
	ctx := gomqtt.New()
	err = ctx.Listen(":1883")
	if err != nil {
		panic(err)
	}

	ctx.Loop()
}
