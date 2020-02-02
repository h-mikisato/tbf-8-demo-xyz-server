package models

type Response struct {
	InteractionURL string    `json:",omitempty"`
	ServerNonce    string    `json:",omitempty"`
	Wait           int       `json:",omitempty"`
	UserCode       *Usercode `json:",omitempty"`
	Handle         *Handle   `json:",omitempty"`
}

type Handle struct {
	Value string
	Type  string
}

type Usercode struct {
	URL  string
	Code string
}
