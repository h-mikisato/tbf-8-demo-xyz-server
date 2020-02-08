package models

type Response struct {
	InteractionURL string    `json:",omitempty"`
	ServerNonce    string    `json:",omitempty"`
	Wait           int       `json:",omitempty"`
	UserCode       *Usercode `json:",omitempty"`
	Handle         *Token    `json:",omitempty"`
	AccessToken    *Token    `json:",omitempty"`
}

type Token struct {
	Value string
	Type  string
}

type Usercode struct {
	URL  string
	Code string
}
