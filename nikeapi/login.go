package nikeapi

import (
	"net/url"
	"crypto/tls"
	"net/http/cookiejar"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
	"net/http"
	"log"
	"sync"
	"errors"

	"github.com/fatih/color"
	"github.com/satori/go.uuid"

	. "github.com/DanielEdeling/NikeGo/types"
	"strings"
)

//defining custom errors
var loginMethodOneError = errors.New("Error while logging in with LoginMethodOne")
var loginMethodTwoError = errors.New("Error while logging in with LoginMethodTwo")
//var loginMethodThreeError = errors.New("Error while logging in with LoginMethodThree")


//defining some variables
var mutex sync.Mutex

func LoginMethodOne(appVersion string, expVersion string, config *Config, proxy string, email string, password string) (*string,*http.Client, error){
	type JSONRes struct {
		User_id string
		Access_token string
	}
	type Login struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		ClientID       string `json:"client_id"`
		KeepMeLoggedIn string `json:"keepMeLoggedIn"`
		UxID           string `json:"ux_id"`
		GrantType      string `json:"grant_type"`
	}


	// proxy
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{},
	}
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: transport,
		Jar: cookieJar,
	}
	visitorId := uuid.Must(uuid.NewV4())
	apiUrl := "https://unite.nikedev.com"
	resource := "/login?appVersion="+appVersion+"&experienceVersion="+expVersion+"&uxid=com.nike.commerce.snkrs.web&locale="+config.Locale+"&backendEnvironment=identity&browser=Google%20Inc.&os=undefined&mobile=false&native=false&visit=1&visitor="
	nilString := ""

	u := Login{email, password, "0Zrbh9wN0CwjeczJAoKvc8US44Ogf49X", "true", "com.nike.commerce.snkrs.ios", "password"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	req, err := http.NewRequest("POST", apiUrl+resource+visitorId.String(), b)
	//setting Headers

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Origin", "https://unite.nikedev.com")
	req.Header.Add("Referer", "https://awr.svs.nike.com/activity/login")
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("X-Requested-With", "com.nike.omega")
	//setting Host
	req.Host = "unite.nikedev.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return &nilString, client, loginMethodOneError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return &nilString, client, loginMethodOneError
	}
	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201:
		mutex.Lock()
		color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Successfully logged in", email)
		mutex.Unlock()
		j := []byte(body)
		token := JSONRes{}
		err = json.Unmarshal(j, &token)
		if err != nil {
			log.Println(err)
		}
		atoken := token.Access_token

		return &atoken, client, nil

	case resp.StatusCode > 400:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Failed logging in", email)
		mutex.Unlock()
		return &resp.Status, client, loginMethodOneError
	default:
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Body)
		return &resp.Status, client, loginMethodOneError
	}
}

func LoginMethodTwo(proxy string, email string, password string) (*string,*http.Client, error){
	type JSONRes struct {
		User_id string
		Access_token string
	}
	type Login struct {
		Username       string `json:"username"`
		Password       string `json:"password"`
		ClientID       string `json:"client_id"`
		KeepMeLoggedIn bool `json:"keepMeLoggedIn"`
		UxID           string `json:"ux_id"`
		GrantType      string `json:"grant_type"`
	}
	// proxy
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{},
	}
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: transport,
		Jar: cookieJar,
	}
	apiUrl := "https://api.nike.com"
	resource := "/idn/shim/oauth/2.0/token"
	nilString := ""

	u := Login{email, password, "PbCREuPr3iaFANEDjtiEzXooFl7mXGQ7", true, "com.nike.commerce.snkrs.droid", "password"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	req, err := http.NewRequest("POST", apiUrl+resource, b)
	//setting Headers

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.186 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,ms;q=0.8")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Host = "api.nike.com"

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return &nilString, client, loginMethodTwoError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return &nilString, client, loginMethodTwoError
	}
	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201:
		mutex.Lock()
		color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Successfully logged in", email)
		mutex.Unlock()
		j := []byte(body)
		token := JSONRes{}
		err = json.Unmarshal(j, &token)
		if err != nil {
			log.Println(err)
		}
		atoken := token.Access_token
		return &atoken, client, nil

	case resp.StatusCode > 400:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Failed logging in", email)
		mutex.Unlock()
		return &resp.Status, client, loginMethodTwoError
	default:
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Body)
		return &resp.Status, client, loginMethodTwoError
	}
}


//testing new login methods
func LoginMethodTest (appV string, expV string, proxy string, email string, password string, locale string) {
	// proxy
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{},
	}
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Transport: transport,
		Jar: cookieJar,
	}

	preloadUrl := "https://secure-nikeplus.nike.com/login/login.do?app=fuelband&client_id=57ee5f99fd3d87ff05cdea6283784060&style=esp&locale="+locale+"&client_secret=46148c0f3e7133f0&format=json"
	loginUrl := "https://secure-nikeplus.nike.com/login/loginViaNike.do?mode=login"


	//preload
	req, err := http.NewRequest("GET", preloadUrl, nil)
	//setting Headers
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Origin", "https://secure-nikeplus.nike.com")
	req.Header.Add("Referer", "https://secure-nikeplus.nike.com/login/loginViaNike.do?uihint=login&mode=login")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	//setting Host
	req.Host = "secure-nikeplus.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	//login
	//Post Body
	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	form.Add("external_access_secret", "")
	form.Add("external_access_token", "")
	form.Add("external_email", "")
	form.Add("external_id", "")
	form.Add("network", "")
	req.PostForm = form
	req, err = http.NewRequest("POST", loginUrl, strings.NewReader(form.Encode()))
	//setting Headers
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Origin", "https://secure-nikeplus.nike.com")
	req.Header.Add("Referer", "https://secure-nikeplus.nike.com/login/loginViaNike.do?uihint=login&mode=login")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	//setting Host
	req.Host = "secure-nikeplus.nike.com"
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	/*body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}*/
	var access_token string
	var refresh_token string
	for _, element := range resp.Cookies() {
		if element.Name == "access_token" {
			access_token = element.Value
		}
		if element.Name == "refresh_token" {
			refresh_token = element.Value
		}
	}

	fmt.Println(access_token)
	fmt.Println(refresh_token)

	//refresh token
	//test1
	/*type Refresh struct {
		GrantType    string `json:"grant_type"`
		RefreshToken string `json:"refresh_token"`
		UxID         string `json:"ux_id"`
		ClientID     string `json:"client_id"`
	}
	refreshUrl := "https://api.nike.com/idn/shim/oauth/2.0/token"
	u := Refresh{"refresh_token", refresh_token, "com.nike.commerce.snkrs.droid", "0Zrbh9wN0CwjeczJAoKvc8US44Ogf49X"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	req, err = http.NewRequest("POST", refreshUrl, b)
	//setting Headers

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Encoding", "gzip;q=1.0, compress;q=0.5")
	req.Header.Add("User-Agent", "NikeRunClub/5.13.1 (com.nike.nikeplus-gps; build:1802261854; iOS 11.2.1) Alamofire/4.5.1")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Nike-TokenAuthStatic", "d682c79eae01b45b55e357f95168b658")
	req.Header.Add("Authorization", "Bearer "+access_token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(body))*/



	//test2
	/*type RefreshToken struct {
		RefreshToken string `json:"refresh_token"`
		ClientID     string `json:"client_id"`
		GrantType    string `json:"grant_type"`
	}
	visitorId := uuid.Must(uuid.NewV4())
	tokenRefreshUrl := "https://unite.nike.com/tokenRefresh?appVersion="+appV+"&experienceVersion="+expV+"&uxid=com.nike.commerce.snkrs.ios&locale="+locale+"&backendEnvironment=identity&browser=Apple%20Computer%2C%20Inc.&os=undefined&mobile=true&native=true&visit=1&visitor="
	u := RefreshToken{refresh_token, "0Zrbh9wN0CwjeczJAoKvc8US44Ogf49X", "refresh_token"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	req, err = http.NewRequest("POST", tokenRefreshUrl+visitorId.String(), b)
	//setting Headers

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://unite.nike.com")
	req.Header.Add("Referer", "https://s3.nikecdn.com/unite/mobile.html?iOSSDKVersion=2.8.5&clientId=g0RB5OJf6SdOASOGyMShmXCfsdS31Al8&uxId=com.nike.brand.ntc.ios.5.6&view=none&locale=de_DE&backendEnvironment=identity&facebookAppId=1428363014144760&wechatAppId=wxfb4e9b7ba7fb5d93")
	req.Header.Add("Content-Type", "text/plain")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("X-NewRelic-ID", "VQYGVF5SCBADUlRbDgcCXg==")
	//setting Host
	req.Host = "unite.nike.com"
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(body))*/

	//test3
	secureUrl := "https://unite.nike.com/loginWithSetCookie?appVersion="+appV+"&experienceVersion="+expV+"&uxid=com.nike.commerce.snkrs.web&locale=en_US&backendEnvironment=identity&browser=Google%20Inc.&os=undefined&mobile=false&native=false&visit=1&visitor="
	//Post Body
	type Request struct {
		ClientID       string `json:"client_id"`
		GrantType      string `json:"grant_type"`
		KeepMeLoggedIn bool   `json:"keepMeLoggedIn"`
		Password       string `json:"password"`
		Username       string `json:"username"`
		UxID           string `json:"ux_id"`
	}


	u := Request{"PbCREuPr3iaFANEDjtiEzXooFl7mXGQ7", "password",true, password, email, "com.nike.commerce.snkrs.web"}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)
	req, err = http.NewRequest("POST", secureUrl, b)
	//setting Headers
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US")
	req.Header.Add("Origin", "https://secure-nikeplus.nike.com")
	req.Header.Add("Referer", "https://secure-nikeplus.nike.com/login/loginViaNike.do?uihint=login&mode=login")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	//setting Host
	req.Host = "unite.nike.com"
	resp, err = client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println(string(body))
}