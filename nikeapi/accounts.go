package nikeapi

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"net/http"
	"log"
	"os"
	"bufio"
	"errors"

	"github.com/fatih/color"
)

var VerifiedAccounts []string
var checkStatusError = errors.New("Error while CheckStatus")

func CheckStatus(email string, password string, token string, clientz http.Client) ([]string, error) {
	type Response struct {
		Measurements struct {
			Height float64 `json:"height"`
			Weight float64 `json:"weight"`
		} `json:"measurements"`
		Verifiedphone string `json:"verifiedphone"`
		Location      struct {
			Country string `json:"country"`
		} `json:"location"`
		UpmID                    string `json:"upmId"`
		LeaderboardAccessFriends bool   `json:"leaderboardAccess_friends"`
		Preferences              struct {
			HeightUnit string `json:"heightUnit"`
			WeightUnit string `json:"weightUnit"`
		} `json:"preferences"`
		Healthdata struct {
			AnonymousAcceptance bool   `json:"anonymousAcceptance"`
			EnhancedAcceptance  string `json:"enhancedAcceptance"`
		} `json:"healthdata"`
		Optin struct {
			Lb struct {
				Friends string `json:"friends"`
			} `json:"lb"`
		} `json:"optin"`
	}
	url := "https://idn.nike.com/user/accountsettings"
	client := clientz
	req, err := http.NewRequest("GET", url, nil)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,de;q=0.8,es;q=0.7")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Referer", "https://www.nike.com/us/en_us/p/settings")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "idn.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return VerifiedAccounts, checkStatusError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return VerifiedAccounts, checkStatusError
	}
	j := []byte(body)
	info := Response{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return VerifiedAccounts, checkStatusError
	}
	if info.Verifiedphone != "" {
		mutex.Lock()
		color.Green("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] %s-> Verified", email)
		VerifiedAccounts = append(VerifiedAccounts, email+":"+password)
		mutex.Unlock()
	} else {
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] %s-> NOT Verified", email)
		mutex.Unlock()
	}
	return VerifiedAccounts, nil
}

func SaveAccounts(verAccs []string, fileName string) {
	// open output file
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("File does not exists or cannot be created")
		os.Exit(1)
	}
	defer file.Close()
	//write every element to file
	w := bufio.NewWriter(file)
	for _, acc := range verAccs {
		fmt.Fprintln(w, acc)
	}
	w.Flush()
}