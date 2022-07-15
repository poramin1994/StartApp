package routers

import (
	v1app "StartApp/api/v1/app"
	v1cms "StartApp/api/v1/cms"
	v1manage "StartApp/api/v1/manage"

	"StartApp/controllers"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func init() {
	beego.Options("/*", func(ctx *context.Context) {
		ctx.Output.SetStatus(200)
		ctx.Output.Body([]byte("OK"))
		return
	})
	beego.Router("/", &controllers.API{})
	// Api App
	appv1 := beego.NewNamespace("/v1/app/api",
		beego.NSBefore(FilterDebug),
		beego.NSNamespace("/user",
			beego.NSRouter("/login", &v1app.User{}, "post:Login"),
			beego.NSRouter("/logout", &v1app.User{}, "post:LogOut"),
			beego.NSRouter("/changePassword", &v1app.User{}, "post:ChangePassword"),
		),
	)

	cmsv1 := beego.NewNamespace("/v1/cms/api",
		beego.NSBefore(FilterDebug),
		beego.NSNamespace("/user",
			beego.NSRouter("/login", &v1cms.User{}, "post:Login"),
			beego.NSRouter("/logout", &v1cms.User{}, "post:LogOut"),
			beego.NSRouter("/changePassword", &v1cms.User{}, "post:ChangePassword"),
		),
	)
	managev1 := beego.NewNamespace("/v1/manage/api",
		beego.NSBefore(FilterDebug),
		beego.NSNamespace("/upload",
			beego.NSRouter("/image", &v1manage.Upload{}, "post:UploadImage"),
		),
		beego.NSNamespace("/export",
			beego.NSNamespace("/pdf",
				beego.NSRouter("/spareList", &v1manage.Pdf{}, "get:GeneratePDFSpareList"),
			),
		),
	)

	beego.AddNamespace(appv1)
	beego.AddNamespace(cmsv1)
	beego.AddNamespace(managev1)
}

var FilterDebug = func(ctx *context.Context) {
	logs.Debug("--------------------------------------------")
	logs.Debug("RequestURI", ctx.Request.RequestURI)
	logs.Debug("method", ctx.Request.Method)
	logs.Debug("content-type:", ctx.Request.Header.Values("Content-Type"))
	logs.Debug("params:", ctx.Input.Params())
	logs.Debug("body:", string(ctx.Input.RequestBody))
	logs.Debug("form:", ctx.Request.Form)
	logs.Debug("postform:", ctx.Request.PostForm)
	logs.Debug("=== Request Headers ===")
	for name, value := range ctx.Request.Header {
		logs.Debug(name, ":", value)
	}
	logs.Debug("=== END Request Headers ===")
	return
}
