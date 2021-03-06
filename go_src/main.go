package main

import (
	"bytes"
	"encoding/json"
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
		Jar:     jar,
		Timeout: 10 * time.Second,
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
	req.Header.Set("Origin", rfc.rootUrl)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	req.Header.Set("Referer", "https://portal.fedsfm.ru/account/login")

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
}

func (rfc RosfinClient) downloadDbfFile(fileName string) string {
	var url = fmt.Sprintf("%s/SkedDownload/GetActiveSked?type=dbf", rfc.rootUrl)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	req.Header.Set("Referer", "https://portal.fedsfm.ru/account/login")
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

type Notification struct {
	ID string `json:"idNotification"`
}

type NotificationContainer struct {
	Notifications []Notification `json:"notifications"`
}

type NotificationPayload struct {
	Data NotificationContainer `json:"data"`
}

func (rfc RosfinClient) getUnreadNotifications() []string {
	var url = fmt.Sprintf("%s/EventNotifications/GetNotifications", rfc.rootUrl)
	var payload = []byte(`{"pageIndex": 1, "pageSize": 10, "isRead": false}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	if err != nil {
		panic(err)
	}
	resp, err := rfc.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		panic("get unread notifications fail")
	}
	data := new(NotificationPayload)
	json.NewDecoder(resp.Body).Decode(data)
	ids := make([]string, 0)
	container := data.Data
	for _, elm := range container.Notifications {
		ids = append(ids, elm.ID)
	}
	return ids
}

func (rfc RosfinClient) postCheckedNotifications(ids []string) {
	var url = fmt.Sprintf("%s/EventNotifications/GetCheckedNotifications", rfc.rootUrl)
	var payload, err = json.Marshal(ids)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	if err != nil {
		panic(err)
	}
	resp, err := rfc.client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic("get unread notifications fail")
	}

}

func main() {
	var login, password string
	var loginFlag = flag.String("login", "", "login from rosfin cabinet")
	var passwordFlag = flag.String("password", "", "password from rosfin cabinet")
	flag.Parse()
	if len(*loginFlag) == 0 || len(*passwordFlag) == 0 {
		login = os.Getenv("ROSFIN_LOGIN")
		password = os.Getenv("ROSFIN_PASS")
		if len(login) == 0 || len(password) == 0 {
			panic("login or password may not be empty!")
		}
	} else {
		login = *loginFlag
		password = *passwordFlag
	}
	var client = newRosfinClient(login, password)
	client.makeLoginRequest()
	var ids = client.getUnreadNotifications()
	client.postCheckedNotifications(ids)
	var result = client.downloadDbfFile(getFileName())
	fmt.Println(result)
}
