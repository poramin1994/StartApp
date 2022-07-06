package main

import (
	_ "StartApp/routers"
	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	beego.Run()
}

