package czds

type authResponse struct {
	AccessToken string `json:"accessToken"`
	Message     string `json:"message"`
}

type TLD struct {
	Name          string `json:"tld"`
	Ulable        string `json:"ulable"`
	CurrentStatus string `json:"currentStatus"`
	SFTP          bool   `json:"sftp"`
}
