// Copyright (c) 2021
// Author: Jeff Weisberg <jaw @ tcp4me.com>
// Created: 2021-Jan-13 15:07 (EST)
// Function: http utils

package alertscript

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dop251/goja"
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Body    string `json:"body"`
}

func (as *AS) webRequest(c goja.FunctionCall) (*Result, error) {

	// (url string, method string, hdrs map[string]string, content string)

	url := ""
	method := ""
	hdrs := map[string]string{}
	content := ""

	if len(c.Arguments) != 4 {
		return nil, fmt.Errorf("webRequest: incorrect number of parameters (url, method, headers, body)")
	}

	as.vm.ExportTo(c.Arguments[0], &url)
	as.vm.ExportTo(c.Arguments[1], &method)
	as.vm.ExportTo(c.Arguments[2], &hdrs)
	as.vm.ExportTo(c.Arguments[3], &content)

	if as.WebReqs >= as.cf.WebMax {
		return nil, fmt.Errorf("Maximum number of web requests exceeded!")
	}

	as.WebReqs++

	// for debugging
	as.Diagf("web: %s %s\nheaders: %+v\nbody: %s\n", method, url, hdrs, content)
	if as.cf.WebMock {
		return &Result{200, "not tried", ""}, nil
	}

	// build request
	client := &http.Client{Timeout: as.cf.WebTimeout}
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(content)))
	if err != nil {
		return nil, fmt.Errorf("webRequest: error %v", err)
	}

	for k, v := range hdrs {
		req.Header.Set(k, v)
	}

	// send request
	resp, err := client.Do(req)

	if err != nil {
		as.WebErrs++
		as.Logf("Request Failed: %v", err)
		return &Result{500, "Request Failed", err.Error()}, nil
	}

	if resp.Status[0] != '2' {
		as.Logf("Request Failed: %s", resp.Status)
		as.WebErrs++
	} else {
		as.Diag(resp.Status)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	return &Result{resp.StatusCode, resp.Status[4:], string(body)}, nil
}
