package types


type Config struct {
	Locale  string `json:"locale"`
	Region string `json:"region"`
	Currency string `json:"currency"`
	Producturl string `json:"producturl"`
	UseTimer bool `json:"usetimer"`
	Type string `json:"type"`
}

type Billing struct {
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	Address1    string `json:"address1"`
	Address2    string `json:"address2"`
	City        string `json:"city"`
	State		string `json:"state"`
	Postalcode  string `json:"postalcode"`
	Country     string `json:"country"`
	Phonenumber string `json:"phonenumber"`
	Cardnumber string `json:"cardnumber"`
	Cardmonth string `json:"cardmonth"`
	Cardyear string `json:"cardyear"`
	Cardcode string `json:"cvv"`
	Cardtype string `json:"cardtype"`
}

type BillingList []struct {
	Billing
}