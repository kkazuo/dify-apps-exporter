/*
Copyright (c) 2025 Koga Kazuo <kkazuo@kkazuo.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/go-retryablehttp"
)

func main() {
	con := newDifyConsole()
	cred, err := con.Login()
	if err != nil {
		slog.Error("Login", slog.Any("err", err))
		return
	}

	file, err := os.Create("apps.zip")
	if err != nil {
		slog.Error("Create Zip", slog.Any("err", err))
		return
	}
	defer file.Close()

	z := zip.NewWriter(file)
	defer z.Close()

	err = con.Apps(cred, func(id string) error {
		app, err := con.ExportApp(cred, id, true)
		if err != nil {
			return err
		}
		err = zipFile(z, id+".yml", app)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		slog.Error("Apps", slog.Any("err", err))
	}
}

func zipFile(z *zip.Writer, name string, body string) error {
	w, err := z.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(body))
	return err
}

type DifyConsole struct {
	APIBase string
	email   string
	passwd  string
	client  *http.Client
}

type LoginResponse struct {
	Result string `json:"result"`
	Data   struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	} `json:"data"`
}

func newDifyConsole() *DifyConsole {
	url := os.Getenv("DIFY_CONSOLE_API")
	email := os.Getenv("DIFY_EMAIL")
	passwd := os.Getenv("DIFY_PASSWORD")

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 2

	client := retryClient.StandardClient()

	return &DifyConsole{APIBase: url, client: client, email: email, passwd: passwd}
}

func (dify *DifyConsole) Request(method string, uri string, query *url.Values, login *LoginResponse, body io.Reader) (*http.Request, error) {
	endpoint, err := url.Parse(dify.APIBase + uri)
	if err != nil {
		return nil, err
	}
	if query != nil {
		endpoint.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}

	if login != nil {
		req.Header.Set("Authorization", login.Auth())
	}

	return req, nil
}

func (dify *DifyConsole) Login() (*LoginResponse, error) {
	bodySend, err := json.Marshal(map[string]any{
		"email":    dify.email,
		"password": dify.passwd,
	})
	if err != nil {
		return nil, err
	}

	req, err := dify.Request("POST", "/login", nil, nil, bytes.NewBuffer(bodySend))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := dify.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var parsed LoginResponse
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (login *LoginResponse) Auth() string {
	return "Bearer " + login.Data.AccessToken
}

func (dify *DifyConsole) Apps(login *LoginResponse, f func(string) error) error {
	page := 1
	for {
		qp := url.Values{}
		qp.Set("page", fmt.Sprint(page))
		qp.Set("limit", fmt.Sprint(100))
		req, err := dify.Request("GET", "/apps", &qp, login, nil)
		if err != nil {
			return err
		}

		res, err := dify.client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var parsed struct {
			HasMore bool `json:"has_more"`
			Data    []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		err = json.Unmarshal(body, &parsed)
		if err != nil {
			return err
		}

		for _, obj := range parsed.Data {
			err = f(obj.ID)
			if err != nil {
				return err
			}
		}

		if !parsed.HasMore {
			break
		}
		page += 1
	}

	return nil
}

func (dify *DifyConsole) ExportApp(login *LoginResponse, id string, include_secret bool) (string, error) {
	qp := url.Values{}
	qp.Set("include_secret", fmt.Sprint(include_secret))
	req, err := dify.Request("GET", "/apps/"+id+"/export", &qp, login, nil)
	if err != nil {
		return "", err
	}

	res, err := dify.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var parsed struct {
		Data string `json:"data"`
	}
	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return "", err
	}

	return parsed.Data, nil
}
