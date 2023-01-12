package api

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	beegoAPI "StartApp/controllers"

	"StartApp/models"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"

	"github.com/beego/beego/v2/client/orm"
	"github.com/golang-jwt/jwt"
)

type API struct {
	beegoAPI.API
}

var (
	ImagePath, _    = beego.AppConfig.String("imagePath")
	PathCallData, _ = beego.AppConfig.String("pathCall")

	BackEndDataPath, _ = beego.AppConfig.String("backEndDataPath")
	GomoAccessKey, _   = beego.AppConfig.String("gomoAccessKey")
)

const (
	// header
	BearerPrefix              = "Bearer "
	AccessKey                 = "VDeoa0934lkfaZ30ds"
	Success                   = "Success"
	BadRequest                = "Bad Request!"
	SomethingWentWrong        = "Something went wrong!"
	DuplicatedRequest         = "Duplicated Request!"
	RateLimits                = "Too many Request!"
	RequestTimedOut           = "Request timed out"
	InvalidArgument           = "Invalid argument"
	InvalidEmail              = "Invalid E-Mail Address"
	NotFound                  = "NOT FOUND"
	AccountOrPasswordNotFound = "Invalid email account or password."
	UserIsNotAdminLevel       = "User is not admin level."
	MaxFileSize               = 500000000

	// header
	HeaderAuthToken = "X-Auth-Token"
	HeaderMobileId  = "Mobile-Id"
	HeaderToken     = "Token"

	// user api error messages
	Unauthorized     = "Unauthorized"
	PermissionDenied = "ขออภัย คุณไม่มีสิทธิ์เข้าถึงข้อมูลนี้"
	FileSizeLarge    = "ขออภัยขนาดไฟล์ใหญ่เกิน 500 MB"

	MissionNotFound           = "Mission Not Found"
	MissionTaskNotFound       = "Mission Task Not Found"
	MissionTaskResultNotFound = "Mission Task Result Not Found"

	//Role
	Admin = "admin"
	User  = "user"
)

func (api *API) GetAccessCredentials() string {
	return api.Ctx.Input.Header(HeaderAuthToken)
}

func (api *API) getHeaderAuthToken() string {
	token := api.Ctx.Input.Header(HeaderAuthToken)

	return GetBearer(token)
}

func (api *API) GetHeaderMobileId() string {
	return strings.TrimSpace(api.Ctx.Input.Header(HeaderMobileId))
}
func (api *API) GetHeaderToken() string {
	return api.Ctx.Input.Header(HeaderToken)
}

func NewAccessToken() *string {
	//token := jwt.New(jwt.SigningMethodHS512)
	//Set some claims
	//Create the Claims
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour * 72).UnixNano(),
		Issuer:    "ind-platform.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	//token.Claims["time"] = time.Now().Unix()
	//token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(AccessKey))
	if err != nil {
		logs.Error("Error cannot gen new token | ", err.Error())
		return nil
	}
	return &tokenString
}

type CustomClaim struct {
	Channel       string `json:"channel"`
	TransactionId string `json:"transactionId"`
	MobileId      string `json:"mobileId"`
	//Action        string `json:"action"`
	//Schema        Schema `json:"schema"`
	//Iat           int64  `json:"iat"`
	jwt.StandardClaims
}
type Schema struct {
	Home     string `json:"home"`
	Toggle   string `json:"toggle"`
	Register string `json:"register"`
}

func (api *API) ResponseJSONWithCode(results interface{}, statusCode int, code int64, msg string, pushNoti bool) {
	if results == nil {
		results = struct{}{}
	}

	response := &beegoAPI.ResponseObjectWithCode{
		Code:           code,
		Message:        msg,
		ResponseObject: results,
	}
	if pushNoti {
		appname, _ := beego.AppConfig.String("appname")
		runmode, _ := beego.AppConfig.String("runmode")
		apiPath := api.Ctx.Request.RequestURI
		headerString := ""
		reqBodyString := string(api.Ctx.Input.RequestBody)
		paramString := fmt.Sprintf("%v", api.Ctx.Input.Params())
		formString := fmt.Sprintf("%v", api.Ctx.Request.Form)

		for name, value := range api.Ctx.Request.Header {
			justString := strings.Join(value, " ")
			headerString += name + ":" + justString + ","
		}
		for name, value := range api.Ctx.Input.Params() {
			paramString += name + ":" + value + ","
		}

		errorCaseData := &models.ErrorCase{
			StatusCode: statusCode,
			Code:       code,
			ApiPath:    apiPath,
			ReqHeader:  headerString,
			ReqForm:    formString,
			ReqBody:    reqBodyString,
			Params:     paramString,
			Message:    msg,
		}
		errId, err := models.AddErrorCase(errorCaseData)
		if err != nil {
			logs.Error("err AddErrorCase :", err)
		}
		detailNoti := map[string]string{
			"Appname":    appname,
			"Runmode":    runmode,
			"ApiPath":    apiPath,
			"StatusCode": strconv.Itoa(statusCode),
			"Code":       strconv.FormatInt(code, 10),
			"Message":    msg,
			"Info":       "ErrCaseID :" + strconv.FormatInt(errId, 10),
		}
		LineNotifyErrCase(detailNoti)
	}

	api.Data["json"] = response
	api.Ctx.ResponseWriter.Header().Set("access-control-allow-headers", "Origin,Accept,Content-Length,Content-Type,X-Atmosphere-tracking-id,X-Atmosphere-Framework,X-Cache-Dat,Cache-Control,X-Requested-With,X-Auth-Token,Authorization,Access-Control-Allow-Origin")
	api.Ctx.ResponseWriter.Header().Set("access-control-expose-headers", "access-control-allow-origin;Content-Type")
	api.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	api.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "PUT,PATCH,GET,POST,DELETE,OPTIONS")
	api.Ctx.ResponseWriter.WriteHeader(statusCode)

	api.Ctx.Output.SetStatus(statusCode)
	api.ServeJSON()
	return
}

func (api *API) ResponseJSON(results interface{}, code int, msg string) {
	if results == nil {
		results = struct{}{}
	}
	response := &beegoAPI.ResponseObject{
		Code:           code,
		Message:        msg,
		ResponseObject: results,
	}

	api.Data["json"] = response
	api.Ctx.ResponseWriter.Header().Set("access-control-allow-headers", "Origin,Accept,Content-Length,Content-Type,X-Atmosphere-tracking-id,X-Atmosphere-Framework,X-Cache-Dat,Cache-Control,X-Requested-With,X-Auth-Token,Authorization,Access-Control-Allow-Origin")
	api.Ctx.ResponseWriter.Header().Set("access-control-expose-headers", "access-control-allow-origin;Content-Type")
	api.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	api.Ctx.ResponseWriter.Header().Set("Access-Control-Allow-Methods", "PUT,PATCH,GET,POST,DELETE,OPTIONS")
	api.Ctx.ResponseWriter.WriteHeader(code)

	api.Ctx.Output.SetStatus(code)
	api.ServeJSON()
	return
}

func (api *API) GetUser() (user *models.User) {
	token := api.getHeaderAuthToken()
	user = models.GetUserByToken(token)
	signingKey := models.MySigningKey

	claims := jwt.MapClaims{}
	rawData, _ := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(signingKey), nil
	})
	logs.Debug("rawData: ", rawData)
	logs.Debug("claims: ", claims)
	username := claims["user"].(string)
	userVeri, err := models.GetUserByUsernameAndDelete(username, false)
	if userVeri.Id != user.Id || err != nil {
		api.Data["user"] = nil
	}

	api.Data["user"] = user
	return
}

func GetBearer(authorization string) string {
	n := len(BearerPrefix)
	if len(authorization) < n || authorization[:n] != BearerPrefix {
		return ""
	}
	return authorization[n:]
}

func (api *API) ValidatePassword(s string) error {
	if len(s) < 8 || len(s) > 64 {
		return errors.New("Password's length have to be between 8 - 64 characters.")
	}
	regex := regexp.MustCompile("^[a-zA-Z0-9]+$")
	if regex.MatchString(s) == false {
		return errors.New("Password can contains only english characters and numbers.")
	}
	return nil
}

func (api *API) CheckTransaction(err error, to orm.TxOrmer) error {
	if err != nil {
		logs.Error("execute transaction's sql fail, rollback.", err)
		err = to.Rollback()
		if err != nil {
			logs.Error("roll back transaction failed", err)
		}
	} else {
		err = to.Commit()
		if err != nil {
			logs.Error("commit transaction failed.", err)
		}
	}
	return err
}

func (api *API) CheckAndCreatesDirectory(filePath string, creates bool) (err error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logs.Debug("no dir")
		if creates {
			err = CreatesDirectory(filePath)
		}
	}
	return
}

func CreatesDirectory(filePath string) (err error) {
	if err = os.Mkdir(filePath, os.ModePerm); err != nil {
		logs.Debug(err)
	}
	return
}
func (api *API) TrimString(message string) string {
	return strings.TrimSpace(message)
}

// ToDateTime
// support dd/mm/yyyy , dd/mm/yyyy hh:mm dd/mm/yyyy hh:mm:ss
// support yyyy-mm-dd hh:mm:ss
func (api *API) ToDateTime(s string) (t time.Time) {
	if s == "" {
		return time.Time{}
	}
	ss := strings.Split(s, " ")
	sections := strings.Split(s, ":")
	// hh:mm
	if len(ss) == 2 && len(ss[1]) == 5 {
		logs.Debug("case 0")
		s += ":00"
	} else if sections == nil || len(sections) == 1 {
		logs.Debug("case 1")
		s += " 00:00:00"
	}
	isDash := (strings.Replace(s, "-", "x", -1)) != s
	var err error
	if isDash {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", s, time.Now().Location())
	} else {
		t, err = time.ParseInLocation("02/01/2006 15:04:05", s, time.Now().Location())
	}
	if err != nil {
		logs.Error("err parse date", err)
		return time.Time{}
	}
	return t
}

func (api *API) FormatDateUnix(t time.Time) string {
	if (t == time.Time{}) {
		return ""
	}
	return strconv.Itoa(int(t.UnixNano()))
}

func (api *API) FormatDateNoTime(t time.Time) string {
	if (t == time.Time{}) {
		return ""
	}
	return t.Format("02/01/2006")
}

func (api *API) DateToBuddhistEra(time time.Time) time.Time {
	date := time.AddDate(543, 0, 0)
	return date
}

func (api *API) ZipFolder(zippath string, path []string) (string, string, error) {
	logs.Debug("ZipPediaFolder :")
	rand := randString(16, "")
	zipFileName := rand + "-" + time.Now().Format("02012006150405")
	tmpDir := ImagePath + "/tmp/" + zipFileName + "/"
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		err = os.Mkdir(tmpDir, os.ModePerm)
		if err != nil {
			return "", "", err
		}
	}
	logs.Debug("zipFileName :", zipFileName)
	contentDir := moveFileToParentFolder(tmpDir, zipFileName, ImagePath, path)
	logs.Debug("> : contentDir :", contentDir)
	src, err := os.Open(contentDir)
	if err != nil {
		logs.Error("ERROR 'ZipPediaFolder' 1 :", err)
	}
	dst, err := os.Open(contentDir)
	if err != nil {
		logs.Error("ERROR 'ZipPediaFolder' 2 :", err)
	}
	io.Copy(dst, src)
	err = zipFilePath(contentDir, tmpDir, zipFileName+".zip")
	if err != nil {
		logs.Error("ERROR 'ZipPediaFolder' 3 :", err)
	}
	logs.Debug("return dir :", tmpDir+zipFileName+".zip")
	logs.Debug("tmpDir :", tmpDir)
	logs.Debug("zipFileName :", zipFileName)
	return tmpDir + zipFileName + ".zip", contentDir, nil
}

func randString(n int64, prefix string) (name string) {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	name = prefix + string(b)
	return
}

func moveFileToParentFolder(tmpDir, name string, base string, paths []string) string {
	rand := randString(16, "")
	dir := base + rand + "-" + time.Now().Format("02012006150405")
	if name != "" {
		dir = base + "/tmp/" + name + "/"
	}
	logs.Debug("dir:", dir)
	dir = tmpDir
	for _, path := range paths {
		ss := strings.Split(path, "/")
		fname := ss[len(ss)-1]
		src, err := os.Open(path)
		if err != nil {
			logs.Error("ERROR 'moveFileToParentFolder' 1 :", err)
			return ""
		}
		logs.Debug("dst path:", dir+fname)
		dst, err := os.Create(dir + fname)
		if err != nil {
			logs.Error("ERROR 'moveFileToParentFolder' 2 :", err)
			return ""
		}
		wb, err := io.Copy(dst, src)
		logs.Error("wb 'copy'  :", wb)
		if err != nil {
			logs.Error("ERROR 'copy'  :", err)

		}
	}
	return dir
}

func zipFilePath(path string, zipToPath string, zipName string) error {
	logs.Debug("ZipFilePath :", path, " -- ", zipName)
	zipFile, err := os.Create(zipToPath + zipName)
	if err != nil {
		logs.Debug("Create err:", err)
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fileToZip, err := os.Open(path)

	if err != nil {
		logs.Debug("Open err:", err)
		return err
	}

	defer fileToZip.Close()
	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		logs.Debug("fileToZip err:", err)
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		logs.Debug("FileInfoHeader err:", err)
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	//TODO : Change file name to original file name here
	header.Name = zipName

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate
	//writer, err := zipWriter.CreateHeader(header)
	//if err != nil {
	//	logs.Debug("CreateHeader err:", err)
	//	return err
	//}

	addFileToZip(zipWriter, path, "", zipName)

	//_, err = io.Copy(writer, fileToZip)
	//logs.Debug("Copy err:", err)
	return err
}

func addFileToZip(w *zip.Writer, path string, baseInZip string, zipName string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err)
	}
	for _, file := range files {
		if file.Name() != zipName {
			fmt.Println(path + file.Name())
			if !file.IsDir() {
				dat, err := ioutil.ReadFile(path + file.Name())
				if err != nil {
					fmt.Println(err)
				}
				// Add some files to the archive.
				f, err := w.Create(baseInZip + file.Name())
				if err != nil {
					fmt.Println(err)
				}
				_, err = f.Write(dat)
				if err != nil {
					fmt.Println(err)
				}
			} else if file.IsDir() {
				// Recurse
				newBase := path + file.Name() + "/"
				fmt.Println("Recursing and Adding SubDir: " + file.Name())
				fmt.Println("Recursing and Adding SubDir: " + newBase)
				addFileToZip(w, newBase, baseInZip+file.Name()+"/", zipName)
			}
		}
	}
}

func LineNotifyErrCase(data map[string]string) (res bool) {
	// Create a new HTTP client
	client := &http.Client{}

	lineNotifyAPI, _ := beego.AppConfig.String("lineNotifyAPI")
	accessToken, _ := beego.AppConfig.String("lineNotifyAccessToken")

	// Create a new POST request
	var messageNoti string
	messageNoti += " \n"

	for name, value := range data {
		messageNoti += name + " : " + value + " \n"
	}

	messageNoti += "ฝากแก้ด้วยจ้าาาาาา~~~~"
	req, err := http.NewRequest("POST", lineNotifyAPI, strings.NewReader("message="+messageNoti))
	if err != nil {
		fmt.Println("Err LineNotifyAPI: ", err)
		return
	}

	// Set the "Authorization" and "Content-Type" headers
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Err LineNotifyAPI: ", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != 200 {
		fmt.Println("Err LineNotifyAPI: ", err)
		return
	}
	return true
}
