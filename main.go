package main

import (
	_ "StartApp/routers"
	// "StartApp/jobs"

	internal "StartApp/controllers"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"

	// "github.com/beego/beego/v2/task"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.RegisterDriver("mysql", orm.DRMySQL)
	logs.EnableFuncCallDepth(true) // show file name & line number
	logs.SetLogFuncCallDepth(3)
	logs.Async(1e3)
	mysqldriver, _ := beego.AppConfig.String("mysqldriver")
	mysqluser, _ := beego.AppConfig.String("mysqluser")
	mysqlpass, _ := beego.AppConfig.String("mysqlpass")
	mysqlurls, _ := beego.AppConfig.String("mysqlurls")
	mysqlport, _ := beego.AppConfig.String("mysqlport")
	mysqldb, _ := beego.AppConfig.String("mysqldb")

	orm.RegisterDataBase("default", mysqldriver, mysqluser+":"+
		mysqlpass)

	err := orm.RegisterDataBase("default", mysqldriver,
		mysqluser+":"+
			mysqlpass+"@tcp("+mysqlurls+":"+
			mysqlport+")/"+mysqldb+"?charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=Asia%2FBangkok&parseTime=true",
	)

	if err != nil {
		logs.Error("could not connect db: ", err)
		panic("could not connect db")
	}
	orm.Debug, _ = beego.AppConfig.Bool("mysqldebug")
}

func main() {

	debug, err := beego.AppConfig.Bool("debug")
	if err != nil {
		debug = false
	}

	info, err := beego.AppConfig.Bool("info")
	if err != nil {
		info = false
	}

	if beego.BConfig.RunMode == "dev" {
		logs.Debug("dev mode")

		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
		beego.BConfig.WebConfig.StaticDir["/lib"] = "lib"
		//beego.SetLogger(logs.AdapterMultiFile, `{"filename":"logs/app.log","separate":[ "error", "warning", "notice", "info", "debug"]}`)
		logs.SetLevel(logs.LevelDebug)
		if !debug {
			logs.SetLevel(logs.LevelError)
		}
		if info {
			logs.SetLevel(logs.LevelInformational)
		}
		// Database alias.
		name := "default"

		// Drop table and re-create.
		force := false

		// Print log.
		verbose := true

		// Error.
		err := orm.RunSyncdb(name, force, verbose)
		if err != nil {
			logs.Error(err)
		}
	} else if beego.BConfig.RunMode == "prd" {
		//beego.SetLogger("file", `{"filename":"logs/stdout.log"}`)
		logs.SetLogger(logs.AdapterMultiFile, `{"filename":"logs/app.log","separate":[ "error", "warning", "notice", "info", "debug"]}`)
		logs.SetLevel(logs.LevelError)
		if info {
			logs.SetLevel(logs.LevelInformational)
		}
		if debug {
			logs.SetLevel(logs.LevelDebug)
		}
	} else if beego.BConfig.RunMode == "uat" {
		//beego.SetLogger("file", `{"filename":"logs/stdout.log"}`)
		logs.SetLogger(logs.AdapterMultiFile, `{"filename":"logs/app.log","separate":[ "error", "warning", "notice", "info", "debug"]}`)
		logs.SetLevel(logs.LevelError)
		if info {
			logs.SetLevel(logs.LevelInformational)
		}
		if debug {
			logs.SetLevel(logs.LevelDebug)
		}
	} else if beego.BConfig.RunMode == "stg" {
		// Database alias.
		name := "default"

		// Drop table and re-create.
		force := false

		// Print log.
		verbose := true

		// Error.
		err := orm.RunSyncdb(name, force, verbose)
		if err != nil {
			logs.Error(err)
		}
		logs.SetLogger(logs.AdapterMultiFile, `{"filename":"logs/app.log","separate":[ "error", "warning", "notice", "info", "debug"]}`)
		logs.SetLevel(logs.LevelDebug)
		beego.BConfig.WebConfig.StaticDir["/cdn"] = "cdn"
	}
	beego.BConfig.Log.AccessLogs = true
	orm.RunCommand()

	filterFunc := cors.Allow(&cors.Options{
		//AllowOrigins: []string{"http://localhost:8080", "http://127.0.0.1:8080", "http://192.168.1.24:8080"},
		AllowAllOrigins:  true,
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Access-Credentials", "X-Auth-Token", "x-auth-token", "Token", "token", "Mobile-Id", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	})

	beego.InsertFilter("*", beego.BeforeRouter, filterFunc)
	internal.Init()
	enableTask, err := beego.AppConfig.Bool("EnableTask")
	if err != nil {
		enableTask = false
	}
	logs.Info("EnableTask: ", enableTask)
	if enableTask {
		// jobs.InitTask()
		// defer task.StopTask()
	}
	beego.Run()

}
