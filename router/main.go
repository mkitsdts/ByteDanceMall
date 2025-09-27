package main

import (
	"bytedancemall/router/service"
)

func main() {
	s := service.NewRouterService()
	s.Run()
}
