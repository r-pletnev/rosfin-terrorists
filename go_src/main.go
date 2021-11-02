package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"
)

const rootUrl = "https://portal.fedsfm.ru"

func getCurrentDate() string {
	var now = time.Now()
	var result = now.Format("02.01.2006")
	return result
}

func getFileName() string {
	return fmt.Sprintf("%s.dbf.zip", getCurrentDate())
}

func getHttpClient() http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := http.Client{
		Jar: jar,
	}
	return client
}

type RosfinClient struct {
	client   http.Client
	rootUrl  string
	login    string
	password string
}

func newRosfinClient(login string, password string) RosfinClient {
	return RosfinClient{
		client:   getHttpClient(),
		rootUrl:  rootUrl,
		login:    login,
		password: password,
	}
}

func (rfc RosfinClient) makeLoginRequest() {
	var url = fmt.Sprintf("%s/account/login", rfc.rootUrl)
	var payload = []byte(fmt.Sprintf(`{"Login": "%s", "Password": "%s"}`, rfc.login, rfc.password))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		panic(err)
	}
	resp, err := rfc.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("login request fail")
	}
	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("resp body:", string(body))
}

func (rfc RosfinClient) downloadDbfFile(fileName string) string {
	var url = fmt.Sprintf("%s/SkedDownload/GetActiveSked?type=dbf", rfc.rootUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	resp, err := rfc.client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("download file request fail")
	}
	defer resp.Body.Close()
	out, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic("file copy error")
	}
	result, err := filepath.Abs(fileName)
	if err != nil {
		panic(err)
	}
	return result
}

func main() {
	var login = flag.String("login", "", "login from rosfin cabinet")
	var password = flag.String("password", "", "password from rosfin cabinet")
	flag.Parse()
	if len(*login) == 0 || len(*password) == 0 {
		panic("login or password may not be empty!")
	}
	var client = newRosfinClient(*login, *password)
	client.makeLoginRequest()
	var result = client.downloadDbfFile(getFileName())
	fmt.Println(result)
}
