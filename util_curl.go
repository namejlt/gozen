package gozen

//curl 封装转发

import (
	"bytes"
	"context"
	"fmt"
	ghttp "github.com/SkyAPM/go2sky/plugins/http"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"strings"
	"time"
)

const (
	CurlHeaderKeyContentType = "Content-Type"
)

type UploadFile struct {
	// 名称
	Name string
	// 文件全路径
	Filepath string
}

func resBodyClose(body io.ReadCloser) {
	_ = body.Close()
}

func CurlGet(ctx context.Context, url string, header []string, timeout time.Duration) (ret []byte, err error) {
	span, _ := ExitSpan(ctx, SpanDaoApi+UtilGetUrlHost(url), RunFuncNameUp(), v3.SpanLayer_Http)
	defer SpanEnd(span)
	var client *http.Client
	client, err = ghttp.NewClient(GTracer)
	if client == nil || err != nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	for _, v := range header {
		// 这儿使用的是strings.SplitN，而不是strings.Split
		// cookie值里可能会出现":"，导致分割出错
		t := strings.SplitN(v, ":", 2)
		length := len(t)
		if length == 2 {
			req.Header.Add(t[0], t[1])
		} else if length == 1 {
			req.Header.Add(t[0], "")
		}
	}

	resp, err := client.Do(req)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	defer resBodyClose(resp.Body)
	ret, err = ioutil.ReadAll(resp.Body)
	SpanLog(span, "url", url, "res", string(ret))
	return ret, err
}

func CurlPost(ctx context.Context, url string, header []string, data string, timeout time.Duration) (ret []byte, err error) {
	span, _ := ExitSpan(ctx, SpanDaoApi+UtilGetUrlHost(url), RunFuncNameUp(), v3.SpanLayer_Http)
	defer SpanEnd(span)
	var client *http.Client
	client, err = ghttp.NewClient(GTracer)
	if client == nil || err != nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data))
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	for _, v := range header {
		// 这儿使用的是strings.SplitN，而不是strings.Split
		// cookie值里可能会出现":"，导致分割出错
		t := strings.SplitN(v, ":", 2)
		length := len(t)
		if length == 2 {
			req.Header.Add(t[0], t[1])
		} else if length == 1 {
			req.Header.Add(t[0], "")
		}
	}
	if req.Header.Get(CurlHeaderKeyContentType) == "" {
		req.Header.Set(CurlHeaderKeyContentType, "application/x-www-form-urlencoded; charset=utf-8")
	}
	resp, err := client.Do(req)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	defer resBodyClose(resp.Body)
	ret, err = ioutil.ReadAll(resp.Body)
	SpanLog(span, "url", url, "param", data, "res", string(ret))
	return ret, err
}

func CurlPostFile(ctx context.Context, url string, reqParams map[string]string, files []UploadFile, timeout time.Duration) (ret []byte, err error) {
	span, _ := ExitSpan(ctx, SpanDaoApi+UtilGetUrlHost(url), RunFuncNameUp(), v3.SpanLayer_Http)
	defer SpanEnd(span)
	var client *http.Client
	client, err = ghttp.NewClient(GTracer)
	if client == nil || err != nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	if ctx == nil {
		ctx = context.Background()
	}

	//file update
	body := &bytes.Buffer{}
	// 文件写入 body
	writer := multipart.NewWriter(body)
	for _, uploadFile := range files {
		var file *os.File
		file, err = os.Open(uploadFile.Filepath)
		if err != nil {
			return
		}
		var part io.Writer
		part, err = writer.CreateFormFile(uploadFile.Name, filepath.Base(uploadFile.Filepath))
		if err != nil {
			return
		}
		_, err = io.Copy(part, file)
		file.Close()
	}
	// 其他参数列表写入 body
	for k, v := range reqParams {
		if err = writer.WriteField(k, v); err != nil {
			return
		}
	}
	if err = writer.Close(); err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	req.Header.Set(CurlHeaderKeyContentType, writer.FormDataContentType())

	resp, err := client.Do(req)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	defer resBodyClose(resp.Body)
	ret, err = ioutil.ReadAll(resp.Body)
	SpanLog(span, "url", url, "param", fmt.Sprint(reqParams), "res", string(ret))
	return ret, err
}

func CurlPut(ctx context.Context, url string, header []string, data string, timeout time.Duration) (ret []byte, err error) {
	span, _ := ExitSpan(ctx, SpanDaoApi+UtilGetUrlHost(url), RunFuncNameUp(), v3.SpanLayer_Http)
	defer SpanEnd(span)
	var client *http.Client
	client, err = ghttp.NewClient(GTracer)
	if client == nil || err != nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, strings.NewReader(data))
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	for _, v := range header {
		// 这儿使用的是strings.SplitN，而不是strings.Split
		// cookie值里可能会出现":"，导致分割出错
		t := strings.SplitN(v, ":", 2)
		length := len(t)
		if length == 2 {
			req.Header.Add(t[0], t[1])
		} else if length == 1 {
			req.Header.Add(t[0], "")
		}
	}
	if req.Header.Get(CurlHeaderKeyContentType) == "" {
		req.Header.Set(CurlHeaderKeyContentType, "application/x-www-form-urlencoded; charset=utf-8")
	}
	resp, err := client.Do(req)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	defer resBodyClose(resp.Body)
	ret, err = ioutil.ReadAll(resp.Body)
	SpanLog(span, "url", url, "param", data, "res", string(ret))
	return ret, err
}

func CurlDelete(ctx context.Context, url string, header []string, timeout time.Duration) (ret []byte, err error) {
	span, _ := ExitSpan(ctx, SpanDaoApi+UtilGetUrlHost(url), RunFuncNameUp(), v3.SpanLayer_Http)
	defer SpanEnd(span)
	var client *http.Client
	client, err = ghttp.NewClient(GTracer)
	if client == nil || err != nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	for _, v := range header {
		t := strings.Split(v, ":")
		length := len(t)
		if length == 2 {
			req.Header.Add(t[0], t[1])
		} else if length == 1 {
			req.Header.Add(t[0], "")
		}
	}
	resp, err := client.Do(req)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	defer resBodyClose(resp.Body)
	ret, err = ioutil.ReadAll(resp.Body)
	SpanLog(span, "url", url, "res", string(ret))
	return ret, err
}

func Curl(ctx context.Context, method string, url string, header map[string]string, body string, timeout time.Duration) (ret []byte, err error) {
	span, _ := ExitSpan(ctx, SpanDaoApi+UtilGetUrlHost(url), RunFuncNameUp(), v3.SpanLayer_Http)
	defer SpanEnd(span)
	var client *http.Client
	client, err = ghttp.NewClient(GTracer)
	if client == nil || err != nil {
		client = &http.Client{}
	}
	client.Timeout = timeout
	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	for key, value := range header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)
	SpanErrorFast(span, err)
	if err != nil {
		return ret, err
	}
	defer resBodyClose(resp.Body)
	ret, err = ioutil.ReadAll(resp.Body)
	SpanLog(span, "method", method, "url", url, "param", body, "res", string(ret))
	return ret, err
}
