package http

import (
	"bytes"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// FileFiled FormData中文件字段信息
type FileFiled struct {
	FileName string
	Data     []byte
}

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // ignore SSL certificates
		},
		Timeout: time.Second * 6,
	}
}

// FormData 构建FormData, 返回带Boundary的内容类型和数据
func FormData(fileFiled map[string]FileFiled, textField map[string]string) (string, *bytes.Buffer, error) {
	// 构建multipart/form-data
	bytesBuffer := &bytes.Buffer{}
	mw := multipart.NewWriter(bytesBuffer)
	for k, v := range fileFiled {
		w, e := mw.CreateFormFile(k, v.FileName)
		if e != nil {
			return "", nil, e
		}
		_, _ = w.Write(v.Data)
	}
	for k, v := range textField {
		w, e := mw.CreateFormField(k)
		if e != nil {
			return "", nil, e
		}
		_, _ = w.Write([]byte(v))
	}
	_ = mw.Close() //必须在数据构建后立即关闭, 否则数据缺少结束符号不可用!

	return mw.FormDataContentType(), bytesBuffer, nil
}

// RequestFormData 发送FormData
func RequestFormData(uri string, method string, formDataContentType string, formDataBytesBuffer *bytes.Buffer) (int, []byte, error) {
	defer formDataBytesBuffer.Reset()

	req, e := http.NewRequest(method, uri, formDataBytesBuffer)
	if e != nil {
		return 0, nil, e
	}
	req.Header.Set("Content-Type", formDataContentType) //带boundary,例如 multipart/form-data; boundary=----WebKitFormBoundaryNH6384gjCcRFQGlr
	resp, e := httpClient.Do(req)
	if e != nil {
		return 0, nil, e
	}
	defer resp.Body.Close()
	bodyBytes, e := io.ReadAll(resp.Body)
	if e != nil {
		return 0, nil, e
	}
	return resp.StatusCode, bodyBytes, nil
}

// HttpGet 发出http的get请求
func HttpGet(uri string) (resp *http.Response, err error) {
	return httpClient.Get(uri)
}

// HttpDelete 发出http的delete请求
func HttpDelete(uri string) (resp *http.Response, err error) {
	hr, e := http.NewRequest(http.MethodDelete, uri, nil)
	if e != nil {
		return nil, e
	}
	return httpClient.Do(hr)
}
