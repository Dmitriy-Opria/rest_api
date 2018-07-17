package recaptcha

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	postURL         = "https://www.google.com/recaptcha/api/siteverify"
	recaptchaSecret = "need to paste recaptchaSecret" //todo
)

type (
	Request struct {
		Response string
		RemoteIp string
	}

	googleResponse struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
)

func Verify(r Request) (err error) {

	if pos := strings.Index(r.RemoteIp, ":"); pos != -1 {
		r.RemoteIp = r.RemoteIp[:pos]
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	params := url.Values{
		"secret":   {recaptchaSecret},
		"remoteip": {r.RemoteIp},
		"response": {r.Response},
	}

	resp, err := client.PostForm(postURL, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	gr := googleResponse{}
	if err = json.Unmarshal(body, &gr); err != nil {
		return err
	}

	if !gr.Success {
		if len(gr.ErrorCodes) > 0 {
			return fmt.Errorf(gr.ErrorCodes[0])
		}
	}
	return nil
}
