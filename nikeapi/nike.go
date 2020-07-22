package nikeapi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	. "github.com/DanielEdeling/NikeGo/types"
)

//defining custom errors
var updateVersionError = errors.New("Error fetching newest app/exp version")
var getStyleCodeError = errors.New("Error fetching stylecode")
var getReleaseTimeError = errors.New("Error fetching release time")
var getProductInfoError = errors.New("Error fetching product info")
var getAvailableSizesError = errors.New("Error fetching available sizes")
var launchMethodError = errors.New("Error fetching launch method")

//functions
func LaunchMethod(productId string, proxy string) (string, error) {
	type Response struct {
		Objects []struct {
			ID             string      `json:"id"`
			ProductID      string      `json:"productId"`
			Method         string      `json:"method"`
			PaymentMethod  string      `json:"paymentMethod"`
			Audience       string      `json:"audience"`
			StartEntryDate time.Time   `json:"startEntryDate"`
			StopEntryDate  interface{} `json:"stopEntryDate"`
			Stores         interface{} `json:"stores"`
			ResourceType   string      `json:"resourceType"`
			Links          struct {
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
			} `json:"links"`
		} `json:"objects"`
		Pages struct {
			Next string `json:"next"`
			Prev string `json:"prev"`
		} `json:"pages"`
	}
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	url := "https://api.nike.com/launch/launch_views/v2?filter=productId(" + productId + ")"
	req, err := http.NewRequest("GET", url, nil)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", launchMethodError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", launchMethodError
	}
	j := []byte(body)
	info := Response{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return "", launchMethodError
	}
	le := len(info.Objects)
	if le == 0 {
		return "FIFO", nil
	} else {
		return "DRAW", nil
	}
}

func GetStylecode(url1 string, proxy string) (string, error) {
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	req, err := http.NewRequest("GET", url1, nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", getStyleCodeError

	} else {
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		re := regexp.MustCompile(`"style":"......"`)
		productStyle := re.FindString(string(content))
		a := strings.Replace(string(productStyle), `"style":"`, "", -1)
		style := strings.Replace(a, `"`, "", -1)
		re1 := regexp.MustCompile(`"colorCode":"..."`)
		cCode := re1.FindString(string(content))
		b := strings.Replace(cCode, `"colorCode":"`, "", -1)
		colorCode := strings.Replace(b, `"`, "", -1)
		styleCode := style + "-" + colorCode
		return styleCode, nil
	}
}

func GetVersion() (string, string, error) {
	resp, err := http.Get("https://s3.nikecdn.com/unite/scripts/unite.min.js")
	if err != nil {
		log.Println(err)
		return "", "", updateVersionError
	} else {
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)
		re := regexp.MustCompile(`"experience-version"]......`)
		exV := re.FindString(string(content))
		experienceVersion := strings.Replace(exV, `"experience-version"]||"`, "", -1)
		re1 := regexp.MustCompile(`"app-version"]......`)
		appV := re1.FindString(string(content))
		appVersion := strings.Replace(appV, `"app-version"]||"`, "", -1)
		return appVersion, experienceVersion, nil
	}
}

func GetProductInfo(styleCode string, config *Config, proxy string) (string, error) {
	type Info struct {
		Country                  string
		Locale                   string
		Channel                  string
		ID                       string
		ThreadID                 string
		InterestID               string
		Name                     string
		CreatedDate              string
		LastUpdatedDate          string
		EffectiveLastUpdatedDate string
		PublishedDate            string
		EffectivePublishedDate   string
		Product                  struct {
			ID          string
			InterestID  string
			Style       string
			ColorCode   string
			GlobalPid   string
			FullTitle   string
			Title       string
			Subtitle    string
			Description string
			ImageURL    string
			Genders     []string
			Price       struct {
				OnSale             bool
				Msrp               float64
				FullRetailPrice    float64
				CurrentRetailPrice float64
			}
			EstimatedLaunchDate           string
			PublishType                   string
			CommerceStartDate             string
			QuantityLimit                 int
			MerchStatus                   string
			ColorDescription              string
			ProductType                   string
			AccessCode                    bool
			StartSellDate                 string
			EffectiveInStockStartSellDate string
			EffectiveInStockStopSellDate  string
			TimeToStartSelectionSeconds   int
			TimeToStartSellSeconds        int
			WaitlineEnabled               bool
			Available                     bool
			SportTags                     []string
			Skus                          []struct {
				ID            string
				LocalizedSize string
				NikeSize      string
				Available     bool
			}
		}
		Restricted      bool
		Feed            string
		Title           string
		Subtitle        string
		ImageURL        string
		TabletImageURL  string
		DesktopImageURL string
		Tags            []string
		Cards           []struct {
			Country     string
			Locale      string
			Channel     string
			ID          string
			CardID      string
			SortOrder   int
			Type        string
			Title       string
			Subtitle    string
			Description string
			Images      []struct {
				Type            string
				ImageURL        string
				Alt             string
				SortOrder       int
				DesktopImageURL string
				TabletImageURL  string
			}
			CreatedDate     string
			LastUpdatedDate string
			ColorHint       struct {
				Text     string
				Active   string
				Inactive string
				Pressed  string
			}
			Cta struct {
				Text        string
				BuyingTools bool
			}
			IOSOnly bool
		}
		Relations []struct {
			Name    string
			Threads []string
		}
		Locations      []interface{}
		Active         bool
		SeoSlug        string
		SeoTitle       string
		SeoDescription string
		RelationalID   string
		SocialPattern  string
	}
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: transport,
	}
	url := "https://api.nike.com/commerce/productfeed/products/snkrs/" + styleCode + "/thread?country=" + config.Region + "&locale=" + config.Locale + "&withCards=true"
	req, err := http.NewRequest("GET", url, nil)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	//req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	//resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return "", getProductInfoError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", getProductInfoError
	}
	j := []byte(body)
	info := Info{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return "", getProductInfoError
	}
	return info.Product.ID, nil
}

func GetReleaseTime(styleCode string, config *Config, proxy string) (string, error) {
	type Info struct {
		Country                  string
		Locale                   string
		Channel                  string
		ID                       string
		ThreadID                 string
		InterestID               string
		Name                     string
		CreatedDate              string
		LastUpdatedDate          string
		EffectiveLastUpdatedDate string
		PublishedDate            string
		EffectivePublishedDate   string
		Product                  struct {
			ID          string
			InterestID  string
			Style       string
			ColorCode   string
			GlobalPid   string
			FullTitle   string
			Title       string
			Subtitle    string
			Description string
			ImageURL    string
			Genders     []string
			Price       struct {
				OnSale             bool
				Msrp               float64
				FullRetailPrice    float64
				CurrentRetailPrice float64
			}
			EstimatedLaunchDate           string
			PublishType                   string
			CommerceStartDate             string
			QuantityLimit                 int
			MerchStatus                   string
			ColorDescription              string
			ProductType                   string
			AccessCode                    bool
			StartSellDate                 string
			EffectiveInStockStartSellDate string
			EffectiveInStockStopSellDate  string
			TimeToStartSelectionSeconds   int
			TimeToStartSellSeconds        int
			WaitlineEnabled               bool
			Available                     bool
			SportTags                     []string
			Skus                          []struct {
				ID            string
				LocalizedSize string
				NikeSize      string
				Available     bool
			}
		}
		Restricted      bool
		Feed            string
		Title           string
		Subtitle        string
		ImageURL        string
		TabletImageURL  string
		DesktopImageURL string
		Tags            []string
		Cards           []struct {
			Country     string
			Locale      string
			Channel     string
			ID          string
			CardID      string
			SortOrder   int
			Type        string
			Title       string
			Subtitle    string
			Description string
			Images      []struct {
				Type            string
				ImageURL        string
				Alt             string
				SortOrder       int
				DesktopImageURL string
				TabletImageURL  string
			}
			CreatedDate     string
			LastUpdatedDate string
			ColorHint       struct {
				Text     string
				Active   string
				Inactive string
				Pressed  string
			}
			Cta struct {
				Text        string
				BuyingTools bool
			}
			IOSOnly bool
		}
		Relations []struct {
			Name    string
			Threads []string
		}
		Locations      []interface{}
		Active         bool
		SeoSlug        string
		SeoTitle       string
		SeoDescription string
		RelationalID   string
		SocialPattern  string
	}
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}

	client := &http.Client{
		Transport: transport,
	}
	url := "https://api.nike.com/commerce/productfeed/products/snkrs/" + styleCode + "/thread?country=" + config.Region + "&locale=" + config.Locale + "&withCards=true"
	req, err := http.NewRequest("GET", url, nil)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	//req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	//resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return "", getReleaseTimeError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", getReleaseTimeError
	}
	j := []byte(body)
	info := Info{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return "", getReleaseTimeError
	}
	return info.Product.StartSellDate, nil
}

func GetAvailableSizes(productId string, proxy string) ([]string, error) {
	type AvailableSizes struct {
		Pages struct {
		} `json:"pages"`
		Objects []struct {
			Links struct {
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
			} `json:"links"`
			ResourceType string `json:"resourceType"`
			ID           string `json:"id"`
			SkuID        string `json:"skuId"`
			ProductID    string `json:"productId"`
			Available    bool   `json:"available"`
			Level        string `json:"level"`
		} `json:"objects"`
	}
	skuList := make([]string, 0)
	proxyURL, _ := url.Parse(proxy)
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		TLSClientConfig:    &tls.Config{},
		DisableCompression: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	url := "https://api.nike.com/deliver/available_skus/v1/?filter=productIds(" + productId + ")"
	req, err := http.NewRequest("GET", url, nil)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return skuList, getAvailableSizesError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return skuList, getAvailableSizesError
	}
	j := []byte(body)
	info := AvailableSizes{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return skuList, getAvailableSizesError
	}
	for i := 0; i < len(info.Objects); i++ {
		if info.Objects[i].Available == true {
			skuList = append(skuList, info.Objects[i].SkuID)
		}
	}
	return skuList, nil
}
