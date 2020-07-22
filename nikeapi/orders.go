package nikeapi

import (
	"fmt"
	"io/ioutil"
	"encoding/xml"
	"net/http"

	. "github.com/DanielEdeling/NikeGo/types"
	"log"
)

type Listitem struct {
	Id string `xml:"id"`
}

type Root struct {
	Listitems []Listitem `xml:"o"` 	// == <o>
	Us string `xml:"us"`			// == <us>
}

func GetOrders(access_token string, email string, password string, config *Config, clientz http.Client) {

	urlz := "https://api.nike.com/commerce/eu/orderhistory?action=getOrderHistoryList&country="+config.Region
	client := clientz

	req, err := http.NewRequest("GET", urlz, nil)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,de;q=0.8,es;q=0.7")
	req.Header.Add("User-Agent", "SNEAKRS/1.2.0 (iPhone; iOS 11.2.1; Scale/3.00)")
	req.Header.Add("X-NewRelic-ID", "VQYGVF5SCBAEVVBUBgMDVg==")
	req.Header.Add("Authorization", "Bearer "+access_token)
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
		return
	}
	if resp.StatusCode != 200 {

		return
	}

	var r Root
	xml.Unmarshal(body, &r)
	var myslice []string
	for _, element := range r.Listitems {
		if element.Id != "" {
			myslice = append(myslice, element.Id)
		}
	}
	if myslice != nil {
		mutex.Lock()
		VerifiedAccounts = append(VerifiedAccounts, email+":"+password)
		mutex.Unlock()
	}

}
