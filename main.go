package main

import (
	"bytedancemall/router/service"
)

func main() {
	s := service.InitRouterService()
	s.Router.Run(":8080")
}