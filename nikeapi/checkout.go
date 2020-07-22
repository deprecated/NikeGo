package nikeapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
	"io/ioutil"
	"net/http"
	"log"

	"github.com/satori/go.uuid"
	"github.com/fatih/color"

	. "github.com/DanielEdeling/NikeGo/types"
	"errors"
)

//defining custom errors
var setCreditCardError = errors.New("Error sending credit card info")
var checkoutPreviewsError = errors.New("Error during CheckoutPreviews")
var paymentPreviewsError = errors.New("Error during PaymentPreviews")
var standardReleaseError = errors.New("Error during StandardRelease")
var drawReleaseError = errors.New("Error during DrawRelease")


func SetCreditCard(token string, billing struct { Billing }, clientz http.Client, email string) (string, error){
	type UserInfo struct {
		ExpirationMonth  string `json:"expirationMonth"`
		AccountNumber    string `json:"accountNumber"`
		CreditCardInfoID string `json:"creditCardInfoId"`
		CvNumber         string `json:"cvNumber"`
		CardType         string `json:"cardType"`
		ExpirationYear   string `json:"expirationYear"`
	}
	visitorId := uuid.Must(uuid.NewV4())
	url := "https://paymentcc.nike.com/creditcardsubmit/" + visitorId.String() + "/store"
	client := clientz

	u := UserInfo{billing.Cardmonth, billing.Cardnumber, visitorId.String(), billing.Cardcode, billing.Cardtype, billing.Cardyear}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("POST", url, b)
	//setting Headers
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com,https://www.nike.com")
	req.Header.Add("Referer", "https://paymentcc.nike.com/services?id=c48cdbda-8c0b-4992-9d69-bd4a27e91855")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	//setting Host
	req.Header.Add("Authorization", "Bearer "+token)
	req.Host = "paymentcc.nike.com"
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return "", setCreditCardError
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", setCreditCardError
	}
	fmt.Println(string(body))
	switch {
	case resp.StatusCode == 201:
		mutex.Lock()
		color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Added credit card", email)
		mutex.Unlock()
		return visitorId.String(), nil
	default:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Error, while setting credit card", email)
		mutex.Unlock()
		return "", setCreditCardError
	}
}

func CheckoutPreviewsUS(token string, email string, productId string, skuId string, config *Config, billing struct { Billing }, clientz http.Client) (string, json.Number, string, error){
	type contactinfo struct {
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type shippingaddress struct {
		Address1   string `json:"address1"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		Address2   string `json:"address2"`
	}
	type recipient struct {
		FirstName string `json:"firstName"`
		LastName string `json:"lastName"`
	}
	type itms []struct{
		ID string `json:"id"`
		SkuID string `json:"skuId"`
		Quantity int `json:"quantity"`
		Recipient recipient `json:"recipient"`
		ShippingAddress shippingaddress `json:"shippingAddress"`
		ContactInfo contactinfo `json:"contactInfo"`
		ShippingMethod string `json:"shippingMethod"`
	}
	type request struct {
		Email    string `json:"email"`
		Country  string `json:"country"`
		Currency string `json:"currency"`
		Locale   string `json:"locale"`
		Channel  string `json:"channel"`
		Items itms `json:"items"`
	}
	type Customer struct {
		Request request `json:"request"`
	}

	checkoutId := uuid.Must(uuid.NewV4())
	url := "https://api.nike.com/buy/checkout_previews/v2/" + checkoutId.String()
	client := clientz
	c := contactinfo{email, billing.Phonenumber}
	s := shippingaddress{billing.Address1, billing.City, billing.State,billing.Postalcode, config.Region, billing.Address2}
	r := recipient{billing.Firstname, billing.Lastname}
	i := itms{{productId, skuId, 1, r, s, c, "STANDARD"}}
	rq := request{email, config.Region, config.Currency, config.Locale, "SNKRS", i}
	u := Customer{rq}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("PUT", url, b)
	//setting Headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", "", "", checkoutPreviewsError
	}
	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202:
		type BillingResp struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Links  struct {
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
				Result struct {
					Ref string `json:"ref"`
				} `json:"result"`
			} `json:"links"`
			Response struct {
				ID             string `json:"id"`
				Country        string `json:"country"`
				Currency       string `json:"currency"`
				Locale         string `json:"locale"`
				ShippingGroups []struct {
					Items []struct {
						ID                 string        `json:"id"`
						SkuID              string        `json:"skuId"`
						Quantity           int           `json:"quantity"`
						ValueAddedServices []interface{} `json:"valueAddedServices"`
						PriceInfo          struct {
							Price              json.Number     `json:"price"`
							Discount           int     `json:"discount"`
							ValueAddedServices int     `json:"valueAddedServices"`
							Total              json.Number     `json:"total"`
							TaxTotal           float64 `json:"taxTotal"`
							PriceID            string  `json:"priceId"`
							PriceSnapshotID    string  `json:"priceSnapshotId"`
						} `json:"priceInfo"`
						Taxes []struct {
							Type  string  `json:"type"`
							Rate  float64 `json:"rate"`
							Total float64 `json:"total"`
						} `json:"taxes"`
						PromotionDiscounts       []interface{} `json:"promotionDiscounts"`
						EstimatedDelivery        time.Time     `json:"estimatedDelivery"`
						EstimatedDeliveryDetails struct {
							Date time.Time `json:"date"`
						} `json:"estimatedDeliveryDetails"`
					} `json:"items"`
					Recipient struct {
						FirstName string `json:"firstName"`
						LastName  string `json:"lastName"`
					} `json:"recipient"`
					ShippingAddress struct {
						Address1   string `json:"address1"`
						Address2   string `json:"address2"`
						City       string `json:"city"`
						PostalCode string `json:"postalCode"`
						Country    string `json:"country"`
					} `json:"shippingAddress"`
					ShippingMethod struct {
						ID                       string    `json:"id"`
						Cost                     int       `json:"cost"`
						DaysToArrive             int       `json:"daysToArrive"`
						EstimatedDelivery        time.Time `json:"estimatedDelivery"`
						EstimatedDeliveryDetails struct {
							Date time.Time `json:"date"`
						} `json:"estimatedDeliveryDetails"`
					} `json:"shippingMethod"`
					ShippingCosts struct {
						PriceInfo struct {
							Price    int     `json:"price"`
							Discount int     `json:"discount"`
							Total    int     `json:"total"`
							TaxTotal float64 `json:"taxTotal"`
						} `json:"priceInfo"`
						Taxes []struct {
							Type  string  `json:"type"`
							Rate  float64 `json:"rate"`
							Total float64 `json:"total"`
						} `json:"taxes"`
						PromotionDiscounts []interface{} `json:"promotionDiscounts"`
					} `json:"shippingCosts"`
					ContactInfo struct {
						PhoneNumber string `json:"phoneNumber"`
						Email       string `json:"email"`
					} `json:"contactInfo"`
				} `json:"shippingGroups"`
				Totals struct {
					ShippingTotal           int     `json:"shippingTotal"`
					Subtotal                json.Number     `json:"subtotal"`
					ValueAddedServicesTotal int     `json:"valueAddedServicesTotal"`
					TaxTotal                float64 `json:"taxTotal"`
					DiscountTotal           int     `json:"discountTotal"`
					Total                   json.Number     `json:"total"`
				} `json:"totals"`
				PriceChecksum  string        `json:"priceChecksum"`
				Email          string        `json:"email"`
				PromotionCodes []interface{} `json:"promotionCodes"`
				ResourceType   string        `json:"resourceType"`
				Links          struct {
					Self struct {
						Ref string `json:"ref"`
					} `json:"self"`
				} `json:"links"`
			} `json:"response"`
			ResourceType string `json:"resourceType"`
		}
		for {
			url := "https://api.nike.com/buy/checkout_previews/v2/jobs/"+checkoutId.String()
			//client := clientz
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
			req.Header.Add("Accept-Language", "en-US,en;q=0.8")
			req.Header.Add("Origin", "https://www.nike.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", "Bearer "+token)
			//setting Host
			req.Host = "api.nike.com"
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return "", "", "", checkoutPreviewsError
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return "", "", "", checkoutPreviewsError
			}
			j := []byte(body)
			info := BillingResp{}
			err = json.Unmarshal(j, &info)
			if err != nil {
				log.Println(err)
				return "", "", "", checkoutPreviewsError
			}
			switch {
			case info.Status == "COMPLETED":
				mutex.Lock()
				color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Added billing", email)
				mutex.Unlock()
				return checkoutId.String(), info.Response.Totals.Total, info.Response.PriceChecksum, nil
			case info.Status == "IN_PROGRESS":
				time.Sleep(2*time.Second)
				continue
			case info.Status == "PENDING":
				time.Sleep(2*time.Second)
				continue
			default:
				mutex.Lock()
				color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - There was an error creating billing job", email)
				mutex.Unlock()
				return "", "", "", checkoutPreviewsError
			}
			break
		}



	case resp.StatusCode > 300:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Failed setting billing", email)
		fmt.Println(resp.Body)
		log.Println(resp.Body)
		mutex.Unlock()
		return "", "", "", checkoutPreviewsError
	default:
		fmt.Println(resp.Body)
		log.Println(resp.Body)
		return "", "", "", checkoutPreviewsError
	}
	return "", "", "", checkoutPreviewsError
}

func CheckoutPreviewsEU(token string, email string, productId string, skuId string, config *Config, billing struct { Billing }, clientz http.Client) (string, json.Number, string, error){
	type contactinfo struct {
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type shippingaddress struct {
		Address1   string `json:"address1"`
		City       string `json:"city"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		Address2   string `json:"address2"`
	}
	type recipient struct {
		FirstName string `json:"firstName"`
		LastName string `json:"lastName"`
	}
	type itms []struct{
		ID string `json:"id"`
		SkuID string `json:"skuId"`
		Quantity int `json:"quantity"`
		Recipient recipient `json:"recipient"`
		ShippingAddress shippingaddress `json:"shippingAddress"`
		ContactInfo contactinfo `json:"contactInfo"`
		ShippingMethod string `json:"shippingMethod"`
	}
	type request struct {
		Email    string `json:"email"`
		Country  string `json:"country"`
		Currency string `json:"currency"`
		Locale   string `json:"locale"`
		Channel  string `json:"channel"`
		Items itms `json:"items"`
	}
	type Customer struct {
		Request request `json:"request"`
	}

	checkoutId := uuid.Must(uuid.NewV4())
	url := "https://api.nike.com/buy/checkout_previews/v2/" + checkoutId.String()
	client := clientz
	c := contactinfo{email, billing.Phonenumber}
	s := shippingaddress{billing.Address1, billing.City, billing.Postalcode, config.Region, billing.Address2}
	r := recipient{billing.Firstname, billing.Lastname}
	i := itms{{productId, skuId, 1, r, s, c, "GROUND_SERVICE"}}
	rq := request{email, config.Region, config.Currency, config.Locale, "SNKRS", i}
	u := Customer{rq}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("PUT", url, b)
	//setting Headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", "", "", checkoutPreviewsError
	}
	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202:
		type BillingResp struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Links  struct {
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
				Result struct {
					Ref string `json:"ref"`
				} `json:"result"`
			} `json:"links"`
			Response struct {
				ID             string `json:"id"`
				Country        string `json:"country"`
				Currency       string `json:"currency"`
				Locale         string `json:"locale"`
				ShippingGroups []struct {
					Items []struct {
						ID                 string        `json:"id"`
						SkuID              string        `json:"skuId"`
						Quantity           int           `json:"quantity"`
						ValueAddedServices []interface{} `json:"valueAddedServices"`
						PriceInfo          struct {
							Price              json.Number     `json:"price"`
							Discount           int     `json:"discount"`
							ValueAddedServices int     `json:"valueAddedServices"`
							Total              json.Number     `json:"total"`
							TaxTotal           float64 `json:"taxTotal"`
							PriceID            string  `json:"priceId"`
							PriceSnapshotID    string  `json:"priceSnapshotId"`
						} `json:"priceInfo"`
						Taxes []struct {
							Type  string  `json:"type"`
							Rate  float64 `json:"rate"`
							Total float64 `json:"total"`
						} `json:"taxes"`
						PromotionDiscounts       []interface{} `json:"promotionDiscounts"`
						EstimatedDelivery        time.Time     `json:"estimatedDelivery"`
						EstimatedDeliveryDetails struct {
							Date time.Time `json:"date"`
						} `json:"estimatedDeliveryDetails"`
					} `json:"items"`
					Recipient struct {
						FirstName string `json:"firstName"`
						LastName  string `json:"lastName"`
					} `json:"recipient"`
					ShippingAddress struct {
						Address1   string `json:"address1"`
						Address2   string `json:"address2"`
						City       string `json:"city"`
						PostalCode string `json:"postalCode"`
						Country    string `json:"country"`
					} `json:"shippingAddress"`
					ShippingMethod struct {
						ID                       string    `json:"id"`
						Cost                     int       `json:"cost"`
						DaysToArrive             int       `json:"daysToArrive"`
						EstimatedDelivery        time.Time `json:"estimatedDelivery"`
						EstimatedDeliveryDetails struct {
							Date time.Time `json:"date"`
						} `json:"estimatedDeliveryDetails"`
					} `json:"shippingMethod"`
					ShippingCosts struct {
						PriceInfo struct {
							Price    int     `json:"price"`
							Discount int     `json:"discount"`
							Total    int     `json:"total"`
							TaxTotal float64 `json:"taxTotal"`
						} `json:"priceInfo"`
						Taxes []struct {
							Type  string  `json:"type"`
							Rate  float64 `json:"rate"`
							Total float64 `json:"total"`
						} `json:"taxes"`
						PromotionDiscounts []interface{} `json:"promotionDiscounts"`
					} `json:"shippingCosts"`
					ContactInfo struct {
						PhoneNumber string `json:"phoneNumber"`
						Email       string `json:"email"`
					} `json:"contactInfo"`
				} `json:"shippingGroups"`
				Totals struct {
					ShippingTotal           int     `json:"shippingTotal"`
					Subtotal                json.Number     `json:"subtotal"`
					ValueAddedServicesTotal int     `json:"valueAddedServicesTotal"`
					TaxTotal                float64 `json:"taxTotal"`
					DiscountTotal           int     `json:"discountTotal"`
					Total                   json.Number     `json:"total"`
				} `json:"totals"`
				PriceChecksum  string        `json:"priceChecksum"`
				Email          string        `json:"email"`
				PromotionCodes []interface{} `json:"promotionCodes"`
				ResourceType   string        `json:"resourceType"`
				Links          struct {
					Self struct {
						Ref string `json:"ref"`
					} `json:"self"`
				} `json:"links"`
			} `json:"response"`
			ResourceType string `json:"resourceType"`
		}
		for {
			url := "https://api.nike.com/buy/checkout_previews/v2/jobs/"+checkoutId.String()
			//client := clientz
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
			req.Header.Add("Accept-Language", "en-US,en;q=0.8")
			req.Header.Add("Origin", "https://www.nike.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", "Bearer "+token)
			//setting Host
			req.Host = "api.nike.com"
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return "", "", "", checkoutPreviewsError
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return "", "", "", checkoutPreviewsError
			}
			j := []byte(body)
			info := BillingResp{}
			err = json.Unmarshal(j, &info)
			if err != nil {
				log.Println(err)
				return "", "", "", checkoutPreviewsError
			}
			switch {
			case info.Status == "COMPLETED":
				mutex.Lock()
				color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Added billing", email)
				mutex.Unlock()
				return checkoutId.String(), info.Response.Totals.Total, info.Response.PriceChecksum, nil
			case info.Status == "IN_PROGRESS":
				time.Sleep(2*time.Second)
				continue
			case info.Status == "PENDING":
				time.Sleep(2*time.Second)
				continue
			default:
				mutex.Lock()
				color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - There was an error creating billing job", email)
				mutex.Unlock()
				return "", "", "", checkoutPreviewsError
			}
			break
		}
	case resp.StatusCode > 300:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Failed setting billing", email)
		fmt.Println(resp.Body)
		mutex.Unlock()
		return "", "", "", checkoutPreviewsError
	default:
		fmt.Println(resp.StatusCode)
		log.Println(resp.Body)
		return "", "", "", checkoutPreviewsError
	}
	return "", "", "", checkoutPreviewsError
}

func PaymentPreviewUS(token string, email string, productId string, checkoutId string, total json.Number, config *Config, billing struct { Billing }, creditcardId string, clientz http.Client) (string, error){
	type contactinfo struct {
		PhoneNumber string `json:"phoneNumber"`
		Email       string `json:"email"`
	}
	type address struct {
		Address1   string `json:"address1"`
		City       string `json:"city"`
		State		string `json:"state"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		Address2   string `json:"address2"`
	}
	type name struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	type billingz struct {
		Name name `json:"name"`
		Address address `json:"address"`
		ContactInfo contactinfo `json:"contactInfo"`
	}
	type paymentinfo []struct{
		ID string `json:"id"`
		BillingInfo billingz `json:"billingInfo"`
		Type string `json:"type"`
		CreditCardInfoID string `json:"creditCardInfoId"`
	}
	type items []struct{
		ProductID       string `json:"productId"`
		ShippingAddress address `json:"shippingAddress"`
	}
	type Preview struct {
		CheckoutID string  `json:"checkoutId"`
		Total      json.Number `json:"total"`
		Currency   string  `json:"currency"`
		Country    string  `json:"country"`
		Items      items `json:"items"`
		PaymentInfo paymentinfo `json:"paymentInfo"`
	}
	type self struct {
		Ref string `json:"ref"`
	}
	type links1 struct {
		Self self `json:"self"`
	}
	type payments []struct {
		ID     string  `json:"id"`
		Type   string  `json:"type"`
		Amount json.Number `json:"amount"`
	}
	type response struct {
		ID       string  `json:"id"`
		Total    json.Number `json:"total"`
		Currency string  `json:"currency"`
		Payments payments `json:"payments"`
		Links links1 `json:"links"`
		ResourceType string `json:"resourceType"`
	}
	type result struct {
		Ref string `json:"ref"`
	}
	type self2 struct {
		Ref string `json:"ref"`
	}
	type links2 struct {
		Result result`json:"result"`
		Self self2 `json:"self"`
	}
	type jsonResp struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		ResourceType string `json:"resourceType"`
		Links links2 `json:"links"`
		Response response `json:"response"`
	}


	url := "https://api.nike.com/payment/preview/v2/"
	client := clientz
	paymentinfoId := uuid.Must(uuid.NewV4())
	ci := contactinfo{billing.Phonenumber, email}
	ad := address{billing.Address1, billing.City, billing.State, billing.Postalcode, billing.Country, billing.Address2}
	na := name{billing.Firstname, billing.Lastname}
	bInfo := billingz{na, ad, ci}
	pInfo := paymentinfo{{paymentinfoId.String(), bInfo, "CreditCard", creditcardId}}
	it := items{{productId, ad}}
	u := Preview{checkoutId, total, config.Currency, billing.Country, it, pInfo}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("POST", url, b)
	//setting Headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", paymentPreviewsError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", paymentPreviewsError
	}

	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202:
		type PaymentPreviewResp struct {
			ID           string `json:"id"`
			Status       string `json:"status"`
			ResourceType string `json:"resourceType"`
			Links        struct {
				Result struct {
					Ref string `json:"ref"`
				} `json:"result"`
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
			} `json:"links"`
			Response struct {
				ID       string  `json:"id"`
				Total    float64 `json:"total"`
				Currency string  `json:"currency"`
				Payments []struct {
					ID     string  `json:"id"`
					Type   string  `json:"type"`
					Amount float64 `json:"amount"`
				} `json:"payments"`
				Links struct {
					Self struct {
						Ref string `json:"ref"`
					} `json:"self"`
				} `json:"links"`
				ResourceType string `json:"resourceType"`
			} `json:"response"`
		}
		j := []byte(body)
		info := jsonResp{}
		err = json.Unmarshal(j, &info)
		if err != nil {
			log.Println(err)
			return "", paymentPreviewsError
		}
		paymentToken := info.ID
		for {
			url := "https://api.nike.com/payment/preview/v2/jobs/"+paymentToken
			//client := clientz
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
			req.Header.Add("Accept-Language", "en-US,en;q=0.8")
			req.Header.Add("Origin", "https://www.nike.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", "Bearer "+token)
			//setting Host
			req.Host = "api.nike.com"
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return "", paymentPreviewsError
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return "", paymentPreviewsError
			}
			j := []byte(body)
			info := PaymentPreviewResp{}
			err = json.Unmarshal(j, &info)
			if err != nil {
				log.Println(err)
				return "", paymentPreviewsError
			}
			switch {
			case info.Status == "COMPLETED":
				mutex.Lock()
				color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Payment preview", email)
				mutex.Unlock()
				return paymentToken, nil
			case info.Status == "IN_PROGRESS":
				continue
			case info.Status == "PENDING":
				continue
			default:
				mutex.Lock()
				color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - There was an error while payment preview", email)
				mutex.Unlock()
				return "", paymentPreviewsError
			}
		}
	default:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - While payment/preview/v2...", email)
		mutex.Unlock()
		return "", paymentPreviewsError
	}
	return "", paymentPreviewsError
}

func PaymentPreviewEU(token string, email string, productId string, checkoutId string, total json.Number, config *Config, billing struct { Billing }, creditcardId string, clientz http.Client) (string, error){
	type contactinfo struct {
		PhoneNumber string `json:"phoneNumber"`
		Email       string `json:"email"`
	}
	type address struct {
		Address1   string `json:"address1"`
		City       string `json:"city"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		Address2   string `json:"address2"`
	}
	type name struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	type billingz struct {
		Name name `json:"name"`
		Address address `json:"address"`
		ContactInfo contactinfo `json:"contactInfo"`
	}
	type paymentinfo []struct{
		ID string `json:"id"`
		BillingInfo billingz `json:"billingInfo"`
		Type string `json:"type"`
		CreditCardInfoID string `json:"creditCardInfoId"`
	}
	type items []struct{
		ProductID       string `json:"productId"`
		ShippingAddress address `json:"shippingAddress"`
	}
	type Preview struct {
		CheckoutID string  `json:"checkoutId"`
		Total      json.Number `json:"total"`
		Currency   string  `json:"currency"`
		Country    string  `json:"country"`
		Items      items `json:"items"`
		PaymentInfo paymentinfo `json:"paymentInfo"`
	}
	type self struct {
		Ref string `json:"ref"`
	}
	type links1 struct {
		Self self `json:"self"`
	}
	type payments []struct {
		ID     string  `json:"id"`
		Type   string  `json:"type"`
		Amount json.Number `json:"amount"`
	}
	type response struct {
		ID       string  `json:"id"`
		Total    json.Number `json:"total"`
		Currency string  `json:"currency"`
		Payments payments `json:"payments"`
		Links links1 `json:"links"`
		ResourceType string `json:"resourceType"`
	}
	type result struct {
		Ref string `json:"ref"`
	}
	type self2 struct {
		Ref string `json:"ref"`
	}
	type links2 struct {
		Result result`json:"result"`
		Self self2 `json:"self"`
	}
	type jsonResp struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		ResourceType string `json:"resourceType"`
		Links links2 `json:"links"`
		Response response `json:"response"`
	}


	url := "https://api.nike.com/payment/preview/v2/"
	client := clientz
	paymentinfoId := uuid.Must(uuid.NewV4())
	ci := contactinfo{billing.Phonenumber, email}
	ad := address{billing.Address1, billing.City, billing.Postalcode, billing.Country, billing.Address2}
	na := name{billing.Firstname, billing.Lastname}
	bInfo := billingz{na, ad, ci}
	pInfo := paymentinfo{{paymentinfoId.String(), bInfo, "CreditCard", creditcardId}}
	it := items{{productId, ad}}
	u := Preview{checkoutId, total, config.Currency, billing.Country, it, pInfo}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("POST", url, b)
	//setting Headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", paymentPreviewsError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", paymentPreviewsError
	}

	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202:
		type PaymentPreviewResp struct {
			ID           string `json:"id"`
			Status       string `json:"status"`
			ResourceType string `json:"resourceType"`
			Links        struct {
				Result struct {
					Ref string `json:"ref"`
				} `json:"result"`
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
			} `json:"links"`
			Response struct {
				ID       string  `json:"id"`
				Total    float64 `json:"total"`
				Currency string  `json:"currency"`
				Payments []struct {
					ID     string  `json:"id"`
					Type   string  `json:"type"`
					Amount float64 `json:"amount"`
				} `json:"payments"`
				Links struct {
					Self struct {
						Ref string `json:"ref"`
					} `json:"self"`
				} `json:"links"`
				ResourceType string `json:"resourceType"`
			} `json:"response"`
		}
		j := []byte(body)
		info := jsonResp{}
		err = json.Unmarshal(j, &info)
		if err != nil {
			log.Println(err)
			return "", paymentPreviewsError
		}
		paymentToken := info.ID
		for {
			url := "https://api.nike.com/payment/preview/v2/jobs/"+paymentToken
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
			req.Header.Add("Accept-Language", "en-US,en;q=0.8")
			req.Header.Add("Origin", "https://www.nike.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", "Bearer "+token)
			//setting Host
			req.Host = "api.nike.com"
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return "", paymentPreviewsError
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return "", paymentPreviewsError
			}
			j := []byte(body)
			info := PaymentPreviewResp{}
			err = json.Unmarshal(j, &info)
			if err != nil {
				log.Println(err)
				return "", paymentPreviewsError
			}
			switch {
			case info.Status == "COMPLETED":
				mutex.Lock()
				color.White("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [LOG] %s - Payment preview", email)
				mutex.Unlock()
				return paymentToken, nil
			case info.Status == "IN_PROGRESS":
				time.Sleep(3*time.Second)
				continue
			case info.Status == "PENDING":
				time.Sleep(3*time.Second)
				continue
			default:
				mutex.Lock()
				color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - There was an error while payment preview", email)
				mutex.Unlock()
				return "", paymentPreviewsError
			}
		}
	default:
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - While payment/preview/v2...", email)
		mutex.Unlock()
		return "", paymentPreviewsError
	}
	return "", paymentPreviewsError
}

func StandardReleaseUS(token string, email string, productId string, skuId string, checkoutId string, paymentToken string, config *Config, billing struct { Billing }, clientz http.Client) error{
	type contactinfo struct {
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type shippingaddress struct {
		Address1   string `json:"address1"`
		City       string `json:"city"`
		State	   string `json:"state"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		Address2   string `json:"address2"`
	}
	type recipient struct {
		FirstName string `json:"firstName"`
		LastName string `json:"lastName"`
	}
	type itms []struct{
		ID string `json:"id"`
		SkuID string `json:"skuId"`
		Quantity int `json:"quantity"`
		Recipient recipient `json:"recipient"`
		ShippingAddress shippingaddress `json:"shippingAddress"`
		ContactInfo contactinfo `json:"contactInfo"`
		ShippingMethod string `json:"shippingMethod"`
	}
	type request struct {
		Email    string `json:"email"`
		Country  string `json:"country"`
		Currency string `json:"currency"`
		Locale   string `json:"locale"`
		Channel  string `json:"channel"`
		Items itms `json:"items"`
		PaymentToken string `json:"paymentToken"`
	}
	type Customer struct {
		Request request `json:"request"`
	}

	url := "https://api.nike.com/buy/checkouts/v2/"+checkoutId
	client := clientz
	c := contactinfo{email, billing.Phonenumber}
	s := shippingaddress{billing.Address1, billing.City, billing.State, billing.Postalcode, config.Region, billing.Address2}
	r := recipient{billing.Firstname, billing.Lastname}
	i := itms{{productId, skuId, 1, r, s, c, "GROUND_SERVICE"}}
	rq := request{email, config.Region, config.Currency, config.Locale, "SNKRS", i, paymentToken}
	u := Customer{rq}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("PUT", url, b)
	//setting Headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return standardReleaseError
	}
	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202:
		type jsonResponse struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Error  struct {
				Message    string `json:"message"`
				HTTPStatus int    `json:"httpStatus"`
				Code       string `json:"code"`
				Errors     []struct {
					Message string `json:"message"`
					Code    string `json:"code"`
				} `json:"errors"`
			} `json:"error"`
			Links struct {
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
			} `json:"links"`
			ResourceType string `json:"resourceType"`
		}
		for {
			url := "https://api.nike.com/buy/checkouts/v2/jobs/"+checkoutId
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
			req.Header.Add("Accept-Language", "en-US,en;q=0.8")
			req.Header.Add("Origin", "https://www.nike.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", "Bearer "+token)
			//setting Host
			req.Host = "api.nike.com"
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return standardReleaseError
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return standardReleaseError
			}
			j := []byte(body)
			info := jsonResponse{}
			err = json.Unmarshal(j, &info)
			if err != nil {
				log.Println(err)
				return standardReleaseError
			}
			switch {
			case info.Status == "COMPLETED":
				mutex.Lock()
				color.Green("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [SUCCESS] %s - ENTERED SUCCESSFULLY", email)
				mutex.Unlock()
				return nil
			case info.Status == "IN_PROGRESS":
				time.Sleep(2*time.Second)
				continue
			case info.Status == "PENDING":
				time.Sleep(2*time.Second)
				continue
			default:
				mutex.Lock()
				color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - There was an error during checkout process", email)
				mutex.Unlock()
				return standardReleaseError
			}
		}
	}
	return standardReleaseError
}

func StandardReleaseEU(token string, email string, productId string, skuId string, checkoutId string, paymentToken string, config *Config, billing struct { Billing }, clientz http.Client) error{
	type contactinfo struct {
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type shippingaddress struct {
		Address1   string `json:"address1"`
		City       string `json:"city"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
		Address2   string `json:"address2"`
	}
	type recipient struct {
		FirstName string `json:"firstName"`
		LastName string `json:"lastName"`
	}
	type itms []struct{
		ID string `json:"id"`
		SkuID string `json:"skuId"`
		Quantity int `json:"quantity"`
		Recipient recipient `json:"recipient"`
		ShippingAddress shippingaddress `json:"shippingAddress"`
		ContactInfo contactinfo `json:"contactInfo"`
		ShippingMethod string `json:"shippingMethod"`
	}
	type request struct {
		Email    string `json:"email"`
		Country  string `json:"country"`
		Currency string `json:"currency"`
		Locale   string `json:"locale"`
		Channel  string `json:"channel"`
		Items itms `json:"items"`
		PaymentToken string `json:"paymentToken"`
	}
	type Customer struct {
		Request request `json:"request"`
	}

	url := "https://api.nike.com/buy/checkouts/v2/"+checkoutId
	client := clientz
	c := contactinfo{email, billing.Phonenumber}
	s := shippingaddress{billing.Address1, billing.City, billing.Postalcode, config.Region, billing.Address2}
	r := recipient{billing.Firstname, billing.Lastname}
	i := itms{{productId, skuId, 1, r, s, c, "GROUND_SERVICE"}}
	rq := request{email, config.Region, config.Currency, config.Locale, "SNKRS", i, paymentToken}
	u := Customer{rq}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(u)

	req, err := http.NewRequest("PUT", url, b)
	//setting Headers
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8")
	req.Header.Add("Origin", "https://www.nike.com")
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Authorization", "Bearer "+token)
	//setting Host
	req.Host = "api.nike.com"
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return standardReleaseError
	}
	switch {
	case resp.StatusCode == 200 || resp.StatusCode == 201 || resp.StatusCode == 202:
		type jsonResponse struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Error  struct {
				Message    string `json:"message"`
				HTTPStatus int    `json:"httpStatus"`
				Code       string `json:"code"`
				Errors     []struct {
					Message string `json:"message"`
					Code    string `json:"code"`
				} `json:"errors"`
			} `json:"error"`
			Links struct {
				Self struct {
					Ref string `json:"ref"`
				} `json:"self"`
			} `json:"links"`
			ResourceType string `json:"resourceType"`
		}
		for {
			url := "https://api.nike.com/buy/checkouts/v2/jobs/"+checkoutId
			req, err := http.NewRequest("GET", url, nil)
			req.Header.Add("Accept", "application/json")
			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
			req.Header.Add("Accept-Language", "en-US,en;q=0.8")
			req.Header.Add("Origin", "https://www.nike.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Authorization", "Bearer "+token)
			//setting Host
			req.Host = "api.nike.com"
			resp, err := client.Do(req)
			if err != nil {
				log.Println(err)
				return standardReleaseError
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				return standardReleaseError
			}
			j := []byte(body)
			info := jsonResponse{}
			err = json.Unmarshal(j, &info)
			if err != nil {
				log.Println(err)
				return standardReleaseError
			}
			switch {
			case info.Status == "COMPLETED":
				mutex.Lock()
				color.Green("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [SUCCESS] %s - ENTERED SUCCESSFULLY", email)
				mutex.Unlock()
				return nil
			case info.Status == "IN_PROGRESS":
				time.Sleep(2*time.Second)
				continue
			case info.Status == "PENDING":
				time.Sleep(2*time.Second)
				continue
			default:
				mutex.Lock()
				color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - There was an error during checkout process", email)
				mutex.Unlock()
				return standardReleaseError
			}
		}
	}
	return standardReleaseError
}

func DrawReleaseUS(token string, email string, productId string, skuId string, checkoutId string, paymentToken string, config *Config, billing struct { Billing }, clientz http.Client, priceChecksum string) error{
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
	type address struct {
		Address1   string `json:"address1"`
		Address2   string `json:"address2"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	}
	type recipient struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type shipping struct {
		Recipient recipient `json:"recipient"`
		Address address `json:"address"`
		Method string `json:"method"`
	}
	type Launch struct {
		LaunchID   string `json:"launchId"`
		SkuID      string `json:"skuId"`
		Locale     string `json:"locale"`
		Currency   string `json:"currency"`
		CheckoutID string `json:"checkoutId"`
		Shipping shipping `json:"shipping"`
		PriceChecksum string `json:"priceChecksum"`
		PaymentToken  string `json:"paymentToken"`
		Channel       string `json:"channel"`
	}
	type Resp struct {
		Message                      string    `json:"message"`
		Code                        string    `json:"code"`
		ID                          string    `json:"id"`
		LaunchID                    string    `json:"launchId"`
		SkuID                       string    `json:"skuId"`
		EstimatedResultAvailability time.Time `json:"estimatedResultAvailability"`
		CreationDate                time.Time `json:"creationDate"`
		ResourceType                string    `json:"resourceType"`
		Links                       struct {
			Self struct {
				Ref string `json:"ref"`
			} `json:"self"`
		} `json:"links"`
	}
	var url string
	url = "https://api.nike.com/launch/launch_views/v2/?filter=productId("+productId+")"
	client := clientz
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
		return drawReleaseError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return drawReleaseError
	}
	j := []byte(body)
	info := Response{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return drawReleaseError
	}
	launchId := info.Objects[0].ID
	url = "https://api.nike.com/launch/entries/v2/"
	for {

		a := address{billing.Address1, billing.Address2, billing.City, billing.State, billing.Postalcode, billing.Country}
		r := recipient{billing.Firstname, billing.Lastname, email, billing.Phonenumber}
		s := shipping{r, a, "GROUND_SERVICE"}
		u := Launch{launchId, skuId, config.Locale, config.Currency, checkoutId, s, priceChecksum, paymentToken, "SNKRS"}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(u)

		req, err = http.NewRequest("POST", url, b)
		//setting Headers
		req.Header.Add("Accept", "application/json")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
		req.Header.Add("Accept-Language", "en-US,en;q=0.8")
		req.Header.Add("Origin", "https://www.nike.com")
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Authorization", "Bearer "+token)
		//setting Host
		req.Host = "api.nike.com"
		resp, err = client.Do(req)
		if err != nil {
			log.Println(err)
			return drawReleaseError
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return drawReleaseError
		}
		j = []byte(body)
		infoz := Resp{}
		err = json.Unmarshal(j, &infoz)
		if err != nil {
			log.Println(err)
			return drawReleaseError
		}
		if infoz.Message != "Launch is not active" {
			mutex.Lock()
			color.Green("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [SUCCESS] %s - Successfully entered the draw", email)
			mutex.Unlock()

			return nil

		} else {
			time.Sleep(2*time.Second)
			continue
		}
	}
}

func DrawReleaseEU(token string, email string, productId string, skuId string, checkoutId string, paymentToken string, config *Config, billing struct { Billing }, clientz http.Client, priceChecksum string) error{
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
	type address struct {
		Address1   string `json:"address1"`
		Address2   string `json:"address2"`
		City       string `json:"city"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	}
	type recipient struct {
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}
	type shipping struct {
		Recipient recipient `json:"recipient"`
		Address address `json:"address"`
		Method string `json:"method"`
	}
	type Launch struct {
		LaunchID   string `json:"launchId"`
		SkuID      string `json:"skuId"`
		Locale     string `json:"locale"`
		Currency   string `json:"currency"`
		CheckoutID string `json:"checkoutId"`
		Shipping shipping `json:"shipping"`
		PriceChecksum string `json:"priceChecksum"`
		PaymentToken  string `json:"paymentToken"`
		Channel       string `json:"channel"`
	}
	type Resp struct {
		Message                      string    `json:"message"`
		Code                        string    `json:"code"`
		ID                          string    `json:"id"`
		LaunchID                    string    `json:"launchId"`
		SkuID                       string    `json:"skuId"`
		EstimatedResultAvailability time.Time `json:"estimatedResultAvailability"`
		CreationDate                time.Time `json:"creationDate"`
		ResourceType                string    `json:"resourceType"`
		Links                       struct {
			Self struct {
				Ref string `json:"ref"`
			} `json:"self"`
		} `json:"links"`
	}
	url := "https://api.nike.com/launch/launch_views/v2/?filter=productId("+productId+")"
	client := clientz
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
		return drawReleaseError
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return drawReleaseError
	}
	j := []byte(body)
	info := Response{}
	err = json.Unmarshal(j, &info)
	if err != nil {
		log.Println(err)
		return drawReleaseError
	}
	launchId := info.Objects[0].ID
	url = "https://api.nike.com/launch/entries/v2/"
	for {
		a := address{billing.Address1, billing.Address2, billing.City, billing.Postalcode, billing.Country}
		r := recipient{billing.Firstname, billing.Lastname, email, billing.Phonenumber}
		s := shipping{r, a, "GROUND_SERVICE"}
		u := Launch{launchId, skuId, config.Locale, config.Currency, checkoutId, s, priceChecksum, paymentToken, "SNKRS"}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(u)

		req, err = http.NewRequest("POST", url, b)
		//setting Headers
		req.Header.Add("Accept", "application/json")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")
		req.Header.Add("Accept-Language", "en-US,en;q=0.8")
		req.Header.Add("Origin", "https://www.nike.com")
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Authorization", "Bearer "+token)
		//setting Host
		req.Host = "api.nike.com"
		resp, err = client.Do(req)
		if err != nil {
			log.Println(err)
			return drawReleaseError
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return drawReleaseError
		}
		j = []byte(body)
		infoz := Resp{}
		err = json.Unmarshal(j, &infoz)
		if err != nil {
			log.Println(err)
			return drawReleaseError
		}
		if infoz.Message != "Launch is not active" {
			mutex.Lock()
			color.Green("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [SUCCESS] %s - Successfully entered the draw", email)
			mutex.Unlock()

			return nil

		} else {
			time.Sleep(2*time.Second)
			continue
		}
	}
}
