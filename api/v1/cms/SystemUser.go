package app

import (
	v1 "StartApp/api"
	"StartApp/models"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/crypto/bcrypt"
)

type SystemUser struct {
	v1.API
}

func (this *User) CreateNewUser() {

	result := map[string]interface{}{}
	user := this.GetUser()
	if user == nil {
		this.ResponseJSONWithCode(result, 401, 401, v1.Unauthorized, false)
		return
	}
	username := this.GetString("email", "")
	password := this.GetString("password", "")
	prefixId, _ := this.GetInt64("prefixId", 0)
	fname := this.GetString("fname", "")
	lname := this.GetString("lname", "")
	activate, _ := this.GetBool("activate", false)

	if TrimString(password) == "" {
		this.ResponseJSONWithCode(result, 400, 40001, "รูปแบบรหัสผ่านไม่ถูกต้อง", false)
		return
	}

	encryptPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logs.Error("Update reset password: err gen pass:", err)
		this.ResponseJSONWithCode(result, 500, 50000, err.Error(), true)
		return
	}

	prefixObj, err := models.GetPrefixById(prefixId)
	if err != nil {
		this.ResponseJSONWithCode(result, 404, 40400, "Prefix NotFound", true)
		return
	}

	profileObj := models.GetDefaultProfile()
	profileObj.FnameEn = fname
	profileObj.LnameEn = lname
	profileObj.Prefix = prefixObj
	profileObj.Activate = activate

	profileId, err := models.AddProfile(&profileObj)
	if err != nil {
		logs.Error("err add profile", err)
		this.ResponseJSONWithCode(result, 500, 50001, err.Error(), true)
		return
	}
	profile, _ := models.GetProfileById(profileId)

	userData := &models.User{
		Username:  username,
		Password:  string(encryptPassword),
		Profile:   profile,
		Activate:  activate,
		Delete:    false,
		Deleted:   time.Time{},
		DeletedBy: 0,
	}
	userId, err := models.AddUser(userData)
	if err != nil {
		this.ResponseJSONWithCode(result, 500, 50002, err.Error(), true)
		return
	}

	result = map[string]interface{}{
		"userId": userId,
	}

	this.ResponseJSONWithCode(result, 200, 20000, v1.Success, false)
	return
}
