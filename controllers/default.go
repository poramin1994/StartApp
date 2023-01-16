package controllers

import (
	"StartApp/models"
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

// API for Client
type API struct {
	beego.Controller
}

type ResponseObject struct {
	Code           int         `json:"code"`
	Message        string      `json:"message"`
	ResponseObject interface{} `json:"responseObject"`
}

type ResponseObjectWithCode struct {
	Code           int64       `json:"code"`
	Message        string      `json:"message"`
	ResponseObject interface{} `json:"responseObject"`
}

var (
	defAdminPassword = "password"
)

func Init() {
	logs.Debug("init db")
	initPrefix()
	initAdminUser()
}

func (api *API) BaseURL() string {
	var baseUrl string = api.Ctx.Input.Site() + fmt.Sprintf(":%d", api.Ctx.Input.Port())
	if api.Ctx.Input.Header("X-Forwarded-Host") != "" {
		baseUrl = api.Ctx.Input.Scheme() + "://" + api.Ctx.Input.Header("X-Forwarded-Host")
	}
	return baseUrl
}

func (api *API) ResponseJSON(results interface{}, code int, msg string) {
	if results == nil {
		results = struct{}{}
	}
	response := &ResponseObject{
		Code:           code,
		Message:        msg,
		ResponseObject: results,
	}
	api.Data["json"] = response
	api.Ctx.Output.SetStatus(code)
	api.ServeJSON()
	return
}

func (api *API) ResponseJSONWithCode(results interface{}, statusCode int, code int64, msg string) {
	if results == nil {
		results = struct{}{}
	}
	response := &ResponseObjectWithCode{
		Code:           code,
		Message:        msg,
		ResponseObject: results,
	}
	api.Data["json"] = response
	api.Ctx.Output.SetStatus(statusCode)
	api.ServeJSON()
	return
}

func (c *API) Get() {
	c.Data["Website"] = "beego.me"
	c.Data["Email"] = "astaxie@gmail.com"
	c.TplName = "index.tpl"
}

func initPrefix() {
	prefixList := []string{"นาย", "นาง", "นางสาว"}
	prefixListEn := []string{"Mr.", "Ms.", "Mrs."}
	now := time.Now()

	for i, data := range prefixList {
		exist, _ := models.GetPrefixByPrefixTh(data)
		if exist == nil {
			_, err := models.AddPrefix(&models.Prefix{
				PrefixTh: data,
				PrefixEn: prefixListEn[i],
				Created:  now,
				Updated:  now,
			})
			if err != nil {
				logs.Error("err init user :", data, "err", err)
			}
		}
	}
}

func initAdminUser() {
	usernameList := []string{"wisdomvastTester1@gmail.com", "wisdomvastTester2@gmail.com", "wisdomvastTester3@gmail.com", "wisdomvastDev@wisdomvast.com", "wisdomvastDev2@wisdomvast.com"}
	now := time.Now()
	encPass, _ := bcrypt.GenerateFromPassword([]byte(defAdminPassword), bcrypt.DefaultCost)
	pass := string(encPass)

	for _, username := range usernameList {
		exist, _ := models.GetUserByUsername(username)
		if exist == nil {
			profileObj := models.GetDefaultProfile()
			profileId, err := models.AddProfile(&profileObj)
			if err != nil {
				logs.Error("err add profile", err)
			}
			profile, _ := models.GetProfileById(profileId)
			_, err = models.AddUser(&models.User{
				Username:  username,
				Password:  pass,
				Profile:   profile,
				Activate:  true,
				Delete:    false,
				DeletedBy: 0,
				Created:   now,
				Updated:   now,
			})
			if err != nil {
				logs.Error("err init user :", username, "err", err)
			}
			//Generate user_token
			user, _ := models.GetUserByUsername(username)
			_, err = models.AddUserToken(&models.UserToken{
				Token:   models.GenerateAccessToken(user),
				User:    user,
				Created: now,
				Updated: now,
			})
			if err != nil {
				logs.Error("err init user token :", username, "err", err)
			}
		}
	}
}
