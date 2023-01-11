package app

import (
	v1 "StartApp/api"
	"StartApp/models"
	"math/rand"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	v1.API
}

func (this *User) Login() {

	result := map[string]interface{}{}
	email := this.GetString("email", "")
	password := this.GetString("password", "")

	user, err := models.GetUserByUsernameAndDelete(email, false)
	if err != nil || user == nil {
		// 404
		this.ResponseJSONWithCode(result, 404, 40400, "บัญชีอีเมลหรือรหัสผ่านไม่ถูกต้อง", false)
		return
	}

	if !user.Activate {
		this.ResponseJSONWithCode(result, 403, 40301, "ผู้ใช้ไม่ได้รับอนุญาตให้ใช้งาน", false)
		return
	}
	if user.Delete {
		this.ResponseJSONWithCode(result, 403, 40302, "ผู้ใช้ไม่ได้รับอนุญาตให้ใช้งาน", false)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		this.ResponseJSONWithCode(nil, 404, 40401, "ไม่พบรหัสผ่านหรือบัญชีไม่ถูกต้อง", false)
		return
	}
	// generate token
	token, err := GanNewToken(user)
	if err != nil {
		this.ResponseJSONWithCode(nil, 500, 50000, "ไม่สามารถสร้างรหัสผ่านได้", true)
		return
	}
	result = map[string]interface{}{
		"token": token,
		"user":  user.Username,
	}
	this.ResponseJSONWithCode(result, 200, 20000, v1.Success, false)
	return
}

func (this *User) LogOut() {
	result := map[string]interface{}{}
	user := this.GetUser()
	if user == nil {
		this.ResponseJSONWithCode(result, 401, 401, v1.Unauthorized, false)
		return
	}

	userToken, _ := models.GetUserTokenByUser(user)

	err := models.DeleteUserToken(userToken.Id)
	if err != nil {
		this.ResponseJSONWithCode(result, 500, 500, "DeleteUserToken fail", true)
		return
	}

	this.ResponseJSONWithCode(result, 200, 20000, v1.Success, false)
	return
}

func (this *User) ChangePassword() {
	now := time.Now()
	result := map[string]interface{}{}
	user := this.GetUser()
	if user == nil {
		this.ResponseJSONWithCode(map[string]interface{}{}, 401, 401, v1.Unauthorized, false)
		return
	}
	password := this.GetString("password")
	newPassword := this.GetString("newPassword")

	if trimString(password) == "" || trimString(newPassword) == "" {
		logs.Debug("passwordTrim:", trimString(password))
		logs.Debug("newPasswordTrim:", trimString(newPassword))
		this.ResponseJSONWithCode(result, 400, 40001, "รูปแบบรหัสผ่านไม่ถูกต้อง", false)
		return
	}

	if password != newPassword {
		logs.Debug("passwordTrim:", password)
		logs.Debug("newPasswordTrim:", newPassword)
		this.ResponseJSONWithCode(result, 400, 40002, "รูปแบบรหัสผ่านไม่ถูกต้อง", false)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logs.Error("err", err)
		this.ResponseJSONWithCode(result, 401, 40100, "รหัสผ่านปัจจุบันตรงกับรหัสผ่านเดิม", false)
		return
	}

	err = this.ValidatePassword(newPassword)
	if err != nil {
		this.ResponseJSONWithCode(result, 400, 40003, err.Error(), false)
		return
	}
	enc, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logs.Error("update reset password: err gen pass:", err)
		this.ResponseJSONWithCode(result, 500, 50001, v1.SomethingWentWrong, true)
		return
	}
	user.Password = string(enc)
	user.Updated = now
	err = models.UpdateUserById(nil, user)
	if err != nil {
		logs.Error("update reset password: err update u:", err)
		this.ResponseJSONWithCode(result, 500, 50002, v1.SomethingWentWrong, true)
		return
	}

	this.ResponseJSONWithCode(result, 200, 20000, v1.Success, false)
	return
}

func GanNewToken(user *models.User) (token string, err error) {
	removeOldToken(user)

gen:
	token = randString(64)
	exist, _ := models.GetUserTokenByToken(token)
	if exist != nil {
		goto gen
	}
	tokenData := &models.UserToken{
		User:  user,
		Token: token,
	}
	_, err = models.AddUserToken(tokenData)
	return token, err
}

func removeOldToken(u *models.User) {
	oldToken, _ := models.GetUserTokenByUser(u)
	if oldToken != nil {
		models.DeleteUserToken(oldToken.Id)
	}
}

func randString(digit int64) (res string) {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, digit)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	res = string(b)
	return
}

func trimString(s string) string {
	return strings.TrimSpace(s)
}
