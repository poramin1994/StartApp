package manager

import (
	v1 "StartApp/api"
	"StartApp/models"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"


)

type Upload struct {
	v1.API
}

func (this *Upload) UploadImage() {
	user := this.GetUser()
	now := time.Now()
	expiredDate := now.Add(7 * 24 * time.Hour)

	// if user == nil {
	// 	this.ResponseJSONWithCode(map[string]interface{}{}, 401, 401, v1.Unauthorized)
	// 	return
	// }

	tempFilePath := v1.ImagePath + "/TempFile"
	result := map[string]interface{}{}

	file, handler, err := this.GetFile("import")
	if err != nil {
		this.ResponseJSONWithCode(map[string]interface{}{}, 400, 40000, "Error Retrieving the File")
		return
	}
	if file == nil {
		this.ResponseJSONWithCode(map[string]interface{}{}, 400, 40001, v1.BadRequest)
		return
	}

	defer file.Close()
	logs.Debug("Uploaded File: %+v\n", handler.Filename)
	logs.Debug("File Size: %+v\n", handler.Size)
	logs.Debug("MIME Header: %+v\n", handler.Header)

	if err = this.CheckAndCreatesDirectory(tempFilePath, true); err != nil {
		this.ResponseJSONWithCode(result, 404, 40401, err.Error())
		return
	}

	// tempFile, err := ioutil.TempFile(tempFilePath, "upload-*.png")
	tempFile, err := ioutil.TempFile(tempFilePath, "upload-*.jpeg")
	if err != nil {
		this.ResponseJSONWithCode(result, 404, 40402, err.Error())
		return
	}

	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		this.ResponseJSONWithCode(result, 404, 40403, err.Error())
		return
	}
	tempFile.Write(fileBytes)
	logs.Debug("Successfully UploadedTmp")

	//Create File for get

	mp := v1.ImagePath + "/Image"
	if err = this.CheckAndCreatesDirectory(mp, true); err != nil {
		this.ResponseJSONWithCode(result, 404, 40404, err.Error())
		return
	}
	tempFileName := strings.Replace(tempFile.Name(), v1.ImagePath+"/TempFile/", "", 3)
	createPath := mp + "/" + tempFileName

	err = CreateImage(tempFile, createPath)
	if err != nil {
		this.ResponseJSONWithCode(result, 404, 40405, err.Error())
		return
	}

	//Delete TmpFile
	err = os.RemoveAll(tempFile.Name())
	if err != nil {
		logs.Debug("Error delete dir image")
	}

	tmpfile := &models.TmpFileList{
		User:        user,
		Path:        "/Image/" + tempFileName,
		ExpiredDate: expiredDate,
		Extension:   "",
		Created:     now,
		Updated:     now,
	}
	if _, err = models.AddTmpFileList(tmpfile); err != nil {
		logs.Error("Error UploadImage : add AddTmpFileList")
	}

	result = map[string]interface{}{
		"fileName": v1.PathCallData + "/Image/" + tempFileName,
	}

	this.ResponseJSONWithCode(result, 200, 200, "Successfully Uploaded")
	return
}

func CreateImage(tempFile *os.File, createPath string) (err error) {
	imgFile, _ := getImageFromFilePath(tempFile.Name())
	f, err := os.Create(createPath)
	if err != nil {
		logs.Error("Create imageCard error")
		return
	}
	defer f.Close()
	// if err = png.Encode(f, imgFile); err != nil {
	// 	return
	// }

	if err = jpeg.Encode(f, imgFile, nil); err != nil {
		return
	}
	return
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	typeFlie := checkTypeImage(filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if typeFlie == "image/jpeg" || typeFlie == "image/jpg" {
		image, err := jpeg.Decode(f)
		return image, err
	} else { //typeFlie == "image/png"
		image, err := png.Decode(f)
		return image, err
	}

}

func checkTypeImage(f string) (res string) {
	// runtime.GOMAXPROCS(runtime.NumCPU())

	file, err := os.Open(f)
	if err != nil {
		fmt.Println(err)
		// os.Exit(1)
	}

	buff := make([]byte, 512)
	if _, err = file.Read(buff); err != nil {
		fmt.Println(err)
		// os.Exit(1)
	}

	filetype := http.DetectContentType(buff)
	switch filetype {
	case "image/jpeg", "image/jpg":
		res = filetype
	case "image/png":
		res = filetype

	case "application/pdf":
		res = filetype

	default:
		logs.Debug("unknown file type uploaded")
	}
	return
}
