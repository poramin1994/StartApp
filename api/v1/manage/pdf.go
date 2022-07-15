package manager

import (
	v1 "StartApp/api"
	"StartApp/models"
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/beego/beego/v2/core/logs"
)

type Pdf struct {
	v1.API
}

func (this *Pdf) GeneratePDFSpareList() {
	now := time.Now()
	expiredDate := now.Add(7 * 24 * time.Hour)

	result := map[string]interface{}{}
	user := this.GetUser()
	spareLists := make([]map[string]interface{}, 0)
	var (
		templ *template.Template
		body  bytes.Buffer
		err   error
	)
	if user == nil {
		this.ResponseJSON(result, 401, v1.Unauthorized)
		return
	}

	data := map[string]interface{}{
		"spareLists": spareLists,
	}

	if templ, err = template.ParseFiles("views/pdf-template/tmp1.html"); err != nil {
		this.ResponseJSON(err.Error(), 500, "error")
		return
	}

	if err = templ.Execute(&body, data); err != nil {
		this.ResponseJSON(err.Error(), 500, "error")
		return
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		this.ResponseJSON(err.Error(), 500, "error")
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
		this.ResponseJSON(err.Error(), 500, "error")
		return
	}

	//directory
	filePath := v1.PathCallData + "/TempFile"
	if err := this.CheckAndCreatesDirectory(filePath, true); err != nil {
		this.ResponseJSON(err.Error(), 500, "error")
		return
	}

	timeString := now.Format("02-01-2006")
	fileName := v1.ImagePath + "/TempFile/" + "approval" + timeString
	err = os.WriteFile(fileName+".pdf", pdfg.Bytes(), os.ModePerm)
	if err != nil {
		this.ResponseJSON(err.Error(), 500, "error")
		return
	}

	tmpfile := &models.TmpFileList{
		User:        user,
		Path:        "/TempFile/approval" + timeString,
		ExpiredDate: expiredDate,
		Extension:   "",
		Created:     now,
		Updated:     now,
	}
	if _, err = models.AddTmpFileList(tmpfile); err != nil {
		logs.Error("Error UploadImage : add AddTmpFileList")
	}

	// export PDF

	transcodeDstPath := v1.ImagePath + "/TempFile"
	filename := "approval" + timeString
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

	//this.ResponseJSON(result, 200, v1.Success)
	return
}
