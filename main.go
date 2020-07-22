package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"

	. "github.com/DanielEdeling/NikeGo/nikeapi"
	. "github.com/DanielEdeling/NikeGo/types"
)

var wg sync.WaitGroup
var mutex sync.Mutex
var savedAccounts []string

func init() {
	nf, err := os.Create("log.txt")
	if err != nil {
		fmt.Println(err)
	}
	log.SetOutput(nf)
}

func NikeGoTasks(appV string, expV string, config *Config, billing struct{ Billing }, proxy string, email string, password string, styleCode string) {
	defer wg.Done()
	var access_token *string
	var client *http.Client
	var err error
	access_token, client, err = LoginMethodTwo(proxy, email, password)

	if err != nil {
		log.Println(err)
		mutex.Lock()
		color.Magenta("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Login method one blocked, using fallback", email)
		mutex.Unlock()
		access_token, client, err = LoginMethodOne(appV, expV, config, proxy, email, password)
		if err != nil {
			return
		}
	}
	productId, err := GetProductInfo(styleCode, config, proxy)
	if err != nil {
		return
	}
	method, err := LaunchMethod(productId, proxy)
	if err != nil {
		return
	}
	creditcardId, err := SetCreditCard(*access_token, billing, *client, email)
	if err != nil {
		return
	}
	skuList, err := GetAvailableSizes(productId, proxy)
	if err != nil {
		mutex.Lock()
		color.Red("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Product is OOS", email)
		mutex.Unlock()
		return
	}
	n := rand.Intn(len(skuList))
	skuId := skuList[n]
	if config.Region == "US" {
		checkoutId, total, priceChecksum, err := CheckoutPreviewsUS(*access_token, email, productId, skuId, config, billing, *client)
		if err != nil {
			return
		}
		paymentToken, err := PaymentPreviewUS(*access_token, email, productId, checkoutId, total, config, billing, creditcardId, *client)
		if err != nil {
			return
		}
		if method == "FIFO" {
			err := StandardReleaseUS(*access_token, email, productId, skuId, checkoutId, paymentToken, config, billing, *client)
			if err != nil {
				return
			}
		} else {
			err := DrawReleaseUS(*access_token, email, productId, skuId, checkoutId, paymentToken, config, billing, *client, priceChecksum)
			if err != nil {
				return
			}
		}
	} else {
		checkoutId, total, priceChecksum, err := CheckoutPreviewsEU(*access_token, email, productId, skuId, config, billing, *client)
		if err != nil {
			return
		}
		paymentToken, err := PaymentPreviewEU(*access_token, email, productId, checkoutId, total, config, billing, creditcardId, *client)
		if err != nil {
			log.Println(err)
			return
		}
		if method == "FIFO" {
			err := StandardReleaseEU(*access_token, email, productId, skuId, checkoutId, paymentToken, config, billing, *client)
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			err := DrawReleaseEU(*access_token, email, productId, skuId, checkoutId, paymentToken, config, billing, *client, priceChecksum)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func CheckAccountsTasks(appV string, expV string, config *Config, proxy string, email string, password string) {
	defer wg.Done()
	var access_token *string
	var client *http.Client
	var err error
	access_token, client, err = LoginMethodTwo(proxy, email, password)

	if err != nil {
		log.Println(err)
		mutex.Lock()
		color.Magenta("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Login method one blocked, using fallback", email)
		mutex.Unlock()
		access_token, client, err = LoginMethodOne(appV, expV, config, proxy, email, password)
		if err != nil {
			log.Println(err)
			return
		}
	}
	savedAccounts, err = CheckStatus(email, password, *access_token, *client)
	if err != nil {
		log.Println(err)
		return
	}
}

func CheckOrdersTasks(appV string, expV string, config *Config, proxy string, email string, password string) {
	defer wg.Done()
	var access_token *string
	var client *http.Client
	var err error
	access_token, client, err = LoginMethodTwo(proxy, email, password)

	if err != nil {
		log.Println(err)
		mutex.Lock()
		color.Magenta("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [ERROR] %s - Login method one blocked, using fallback", email)
		mutex.Unlock()
		access_token, client, err = LoginMethodOne(appV, expV, config, proxy, email, password)
		if err != nil {
			log.Println(err)
			return
		}
	}
	GetOrders(*access_token, email, password, config, *client)
}

func main() {
	//LOADING ACCOUNTS
	aa := 0
	accountsHandle, err := os.Open("resources/accounts.txt")
	if err != nil {
		panic(err)
	}
	accounts := bufio.NewScanner(accountsHandle)
	for accounts.Scan() {
		aa++
	}
	accountsHandle.Close()
	accountsHandle, err = os.Open("resources/accounts.txt")
	defer accountsHandle.Close()
	if err != nil {
		panic(err)
	}
	accounts = bufio.NewScanner(accountsHandle)
	//LOADING PROXIES
	pa := 0
	proxyHandle, err := os.Open("resources/proxies.txt")
	defer proxyHandle.Close()
	if err != nil {
		panic(err)
	}
	proxies := bufio.NewScanner(proxyHandle)
	proxyList := make([]string, 0)
	for proxies.Scan() {
		proxyList = append(proxyList, proxies.Text())
		pa++
	}

	//LOADING CONFIG
	var config Config
	configFile, err := os.Open("resources/config.json")
	defer configFile.Close()
	if err != nil {
		panic(err)
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)

	//LOADING BILLING
	var billing BillingList
	billingFile, err := os.Open("resources/billing.json")
	defer billingFile.Close()
	if err != nil {
		panic(err)
	}
	billingParser := json.NewDecoder(billingFile)
	billingParser.Decode(&billing)
	lb := len(billing)

	//PRINTING WELCOME MESSAGE
	color.Cyan("###################################################################")
	fmt.Println("")
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [INFO] Welcome to NikeGo v1.5.0")

	//PRELOADING SOME SHIT
	color.Cyan("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] Accounts: %d - Proxies: %d", aa, pa)
	color.Cyan("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] Billing Profiles: %d", lb)
	color.Cyan("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] Use Timer: %t", config.UseTimer)

	fmt.Println("")
	color.Cyan("###################################################################")
	fmt.Println("")

	var appV string
	var expV string
	appV, expV, err = GetVersion()
	if err != nil {
		color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Could not fetch latest app/exp version")
		appV = "383"
		expV = "323"
		color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [INFO] Using fallback options: appV = 383 && expV = 323")
	}

	//START BOT
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [MODE] What do you want to do?")
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [MODE] [1] NikeGo")
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [MODE] [2] Check your accounts")
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [MODE] [3] Search for orders")
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [MODE] [4] Login Tests")
	color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [MODE] Enter anything else to stop the bot.")
	fmt.Println("")
	color.Cyan("###################################################################")
	fmt.Println("")
	var a int
	fmt.Print("--> ")
	_, err = fmt.Scanf("%d", &a)
	switch {
	case a == 1:
		//START BOT
		if aa > pa {
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Account:Proxy ratio NOT 1:1")
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Closing...")
			fmt.Println("")
			color.Cyan("###################################################################")
			fmt.Println("")
			return
		}
		if aa > lb {
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Account:Billing ratio NOT 1:1")
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Closing...")
			fmt.Println("")
			color.Cyan("###################################################################")
			fmt.Println("")
			return
		}
		styleCode, err := GetStylecode(config.Producturl, "https://"+proxyList[0])
		if err != nil {
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Could not fetch stylecode. Exiting...")
			return
		}
		if config.UseTimer == true {
			//TIMER
			startTime, err := GetReleaseTime(styleCode, &config, "https://"+proxyList[0])
			if err != nil {
				color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Could not fetch release time. Exiting...")
				return
			}
			color.Cyan("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] Waiting until: %s", startTime)
			loc, _ := time.LoadLocation("GMT")
			for {
				t := time.Now().In(loc)
				a := t.Format("2006-01-02T15:04:05.999")
				if a >= startTime {
					color.Cyan("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [INFO] Starting tasks now.")
					break
				}
				time.Sleep(time.Second)
			}
			//GOROUTINES
			i := 0
			for accounts.Scan() {
				input := accounts.Text()
				s := strings.Split(input, ":")
				email, password := s[0], s[1]
				proxy := "https://" + proxyList[i]
				wg.Add(1)
				time.Sleep(200 * time.Millisecond)
				go NikeGoTasks(appV, expV, &config, billing[i], proxy, email, password, styleCode)
				i++
			}
			wg.Wait()
		} else {
			//GOROUTINES
			i := 0
			for accounts.Scan() {
				input := accounts.Text()
				s := strings.Split(input, ":")
				email, password := s[0], s[1]
				proxy := "https://" + proxyList[i]
				wg.Add(1)
				time.Sleep(200 * time.Millisecond)
				go NikeGoTasks(appV, expV, &config, billing[i], proxy, email, password, styleCode)
				i++
			}
			wg.Wait()
		}
	case a == 2:
		accountsHandle.Close()
		billingFile.Close()
		//LOADING ACCOUNTS
		aa := 0
		accountsHandle, err := os.Open("resources/testaccounts.txt")
		if err != nil {
			panic(err)
		}
		accounts := bufio.NewScanner(accountsHandle)
		for accounts.Scan() {
			aa++
		}
		accountsHandle.Close()
		accountsHandle, err = os.Open("resources/testaccounts.txt")
		defer accountsHandle.Close()
		if err != nil {
			panic(err)
		}
		accounts = bufio.NewScanner(accountsHandle)
		//
		if aa > pa {
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Account:Proxy ratio NOT 1:1")
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Closing...")
			fmt.Println("")
			color.Cyan("###################################################################")
			fmt.Println("")
			return
		}
		fmt.Println("")
		color.Cyan("###################################################################")
		fmt.Println("")
		//
		i := 0
		if config.Type != "taobao" {
			for accounts.Scan() {
				input := accounts.Text()
				s := strings.Split(input, ":")
				email, password := s[0], s[1]
				proxy := "https://" + proxyList[i]
				wg.Add(1)
				time.Sleep(200 * time.Millisecond)
				go CheckAccountsTasks(appV, expV, &config, proxy, email, password)
				i++
			}
			wg.Wait()
		} else {
			for accounts.Scan() {
				input := accounts.Text()
				s := strings.Split(input, "----")
				email, password := s[0], s[1]
				proxy := "https://" + proxyList[i]
				wg.Add(1)
				time.Sleep(200 * time.Millisecond)
				go CheckAccountsTasks(appV, expV, &config, proxy, email, password)
				i++
			}
			wg.Wait()
		}
		fmt.Println("")
		color.Cyan("#####################################################################")
		fmt.Println("")
		color.Cyan("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] Found %d verified Accounts !", len(VerifiedAccounts))
		fmt.Println("")
		color.Cyan("#####################################################################")
		SaveAccounts(savedAccounts, "accounts.txt")

	case a == 3:
		//GOROUTINES
		if aa > pa {
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Account:Proxy ratio NOT 1:1")
			color.Red("[" + time.Now().Format("2006-01-02 15:04:05.000000") + "] [ERROR] Closing...")
			fmt.Println("")
			color.Cyan("###################################################################")
			fmt.Println("")
			return
		}
		var i int
		for accounts.Scan() {
			input := accounts.Text()
			s := strings.Split(input, ":")
			email, password := s[0], s[1]
			proxy := "https://" + proxyList[i]
			wg.Add(1)
			go CheckOrdersTasks(appV, expV, &config, proxy, email, password)
			i++
		}
		wg.Wait()

		// AFTER EVERYTHING HAS FINISHED
		SaveAccounts(VerifiedAccounts, "orders.txt")
		color.Cyan("["+time.Now().Format("2006-01-02 15:04:05.000000")+"] [INFO] Found %d account(s) with an order history !", len(VerifiedAccounts))
	case a == 4:
		proxy := "https://" + proxyList[0]
		email := ""
		password := ""
		locale := "de_DE"
		LoginMethodTest(appV, expV, proxy, email, password, locale)

	default:
		break
	}
	time.Sleep(time.Second)
	//STOP CLOSING WINDOW
	readerz := bufio.NewReader(os.Stdin)
	fmt.Print("Press any key to exit... ")
	readerz.ReadString('\n')
	fmt.Println("Bye.")
}
