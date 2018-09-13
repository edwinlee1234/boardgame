package error

// Response 回應資訊格式
type Response struct {
	Status string                   `json:"status"`
	Data   map[string][]interface{} `json:"data"`
	Error  map[string]interface{}   `json:"error"`
}
