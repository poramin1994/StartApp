package manager

import (
	v1 "StartApp/api"
	"StartApp/models"
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/beego/beego/v2/core/logs"
)

type Pdf struct {
	v1.API
}

func (this *Pdf) GeneratePDFv1() {
	now := time.Now()
	expiredDate := now.Add(7 * 24 * time.Hour)
	expiredDateTimeStamp := strconv.FormatInt(expiredDate.Unix(), 10)
	result := map[string]interface{}{}
	spareLists := make([]map[string]interface{}, 0) //this 's mockup
	var (
		templ *template.Template
		body  bytes.Buffer
		err   error
	)
	user := this.GetUser()
	if user == nil {
		this.ResponseJSONWithCode(result, 401, 40100, v1.Unauthorized, false)
		return
	}

	data := map[string]interface{}{
		"spareLists": spareLists,
	}

	if templ, err = template.ParseFiles("views/pdf-template/tmp1.html"); err != nil {
		this.ResponseJSONWithCode(result, 500, 50000, err.Error(), true)
		return
	}

	if err = templ.Execute(&body, data); err != nil {
		this.ResponseJSONWithCode(result, 500, 50001, err.Error(), true)
		return
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		this.ResponseJSONWithCode(result, 500, 50002, err.Error(), true)
		return
	}

	// read the HTML page as a PDF page
	page := wkhtmltopdf.NewPageReader(bytes.NewReader(body.Bytes()))
	page.HeaderFontName.Set("Sarabun-Regular")
	page.EnableLocalFileAccess.Set(true)

	// add the page to your generator
	pdfg.AddPage(page)

	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.MarginBottom.Set(0)
	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationPortrait)

	err = pdfg.Create()
	if err != nil {
		this.ResponseJSONWithCode(result, 500, 50003, err.Error(), true)
		return
	}

	//directory
	filePath := v1.ImagePath + "TempFile"
	if err := this.CheckAndCreatesDirectory(filePath, true); err != nil {
		this.ResponseJSONWithCode(result, 500, 50004, err.Error(), true)
		return
	}

	timeString := now.Format("02-01-2006")
	fileName := filePath + "/" + timeString
	err = os.WriteFile(fileName+".pdf", pdfg.Bytes(), os.ModePerm)
	if err != nil {
		this.ResponseJSONWithCode(result, 500, 50005, err.Error(), true)
		return
	}

	tmpfile := &models.TmpFileList{
		User:        user,
		Path:        fileName + ".pdf",
		ExpiredDate: expiredDateTimeStamp,
		Extension:   "",
		Created:     now,
		Updated:     now,
	}
	if _, err = models.AddTmpFileList(tmpfile); err != nil {
		logs.Error("Error UploadImage : add AddTmpFileList")
		this.ResponseJSONWithCode(result, 500, 50006, err.Error(), true)
		return
	}

	// export PDF

	transcodeDstPath := v1.ImagePath + "/TempFile"
	filename := timeString
	output := transcodeDstPath + "/" + filename + ".pdf"
	downloadName := filename + ".pdf"
	logs.Debug("output :", output)

	w := this.Ctx.ResponseWriter
	r := this.Ctx.Request
	v := url.Values{}
	v.Add("filename", downloadName)
	encoded := v.Encode()
	attachment := "attachment; " + encoded
	downloadName = strings.Replace(encoded, "filename=", "", 1)

	w.Header().Set("Content-Disposition", attachment)
	w.Header().Set("File-Name", downloadName)
	w.Header().Set("Access-Control-Expose-Headers", "File-Name")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	http.ServeFile(w, r, output)

	return
}
