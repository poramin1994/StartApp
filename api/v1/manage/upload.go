package manager

import (
	v1 "StartApp/api"
	"StartApp/models"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

type Upload struct {
	v1.API
}

func (this *Upload) UploadImage() {
	result := map[string]interface{}{}
	now := time.Now()
	expiredDate := now.Add(7 * 24 * time.Hour)
	expiredDateTimeStamp := strconv.FormatInt(expiredDate.Unix(), 10)
	filePath := v1.ImagePath

	user := this.GetUser()
	if user == nil {
		this.ResponseJSONWithCode(result, 401, 401, v1.Unauthorized, false)
		return
	}

	file, header, err := this.GetFile("import")
	if err != nil {
		this.ResponseJSONWithCode(result, 400, 40000, "Error Retrieving the File", false)
		return
	}

	if file == nil {
		this.ResponseJSONWithCode(result, 400, 40001, v1.BadRequest, false)
		return
	}

	defer file.Close()

	if err = this.CheckAndCreatesDirectory(filePath, true); err != nil {
		this.ResponseJSONWithCode(result, 404, 40401, err.Error(), false)
		return
	}

	//Create File for get
	filePath = v1.ImagePath + "/TmpFileImage"
	if err = this.CheckAndCreatesDirectory(filePath, true); err != nil {
		this.ResponseJSONWithCode(result, 404, 40402, err.Error(), false)
		return
	}
	filePath = filePath + "/" + header.Filename
	out, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating new file:", err)
		this.ResponseJSONWithCode(result, 500, 50000, err.Error(), true)
		return
	}
	defer out.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		this.ResponseJSONWithCode(result, 500, 50001, err.Error(), true)

		return
	}
	// Encode the image with a lower quality level
	var opt jpeg.Options
	opt.Quality = 80
	jpeg.Encode(out, img, &opt)

	// Copy the uploaded image to the new file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Println("Error copying image file:", err)
		this.ResponseJSONWithCode(result, 500, 50002, err.Error(), true)
		return
	}

	tmpfile := &models.TmpFileList{
		User:        user,
		Path:        filePath,
		ExpiredDate: expiredDateTimeStamp,
		Extension:   "",
	}
	tmpfileId, err := models.AddTmpFileList(tmpfile)
	if err != nil {
		logs.Error("Error UploadImage : add AddTmpFileList")
		this.ResponseJSONWithCode(result, 500, 50003, err.Error(), true)
		return
	}

	result = map[string]interface{}{
		"tmpfileId": tmpfileId,
	}

	this.ResponseJSONWithCode(result, 200, 200, "Successfully Uploaded", false)
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
