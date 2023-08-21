package Models

type Fail struct {
	Error string `json:"error"`
}

type Response_APIKeyList struct {
	APIKeys []APIKeyJSON `json:"api_keys"`
}

type Response_APIKey struct {
	APIKey APIKeyJSON `json:"api_key"`
}

type APIKeyJSON struct {
	Key        string `json:"key"`
	Expiration string `json:"expiration"`
}
