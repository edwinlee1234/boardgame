package error

import (
	"encoding/json"
	"log"
	"net/http"
)

// ErrorRespone 直接回傳api錯誤信息
func ErrorRespone(err error, errorCode string, w http.ResponseWriter, statusCode int) bool {
	if err != nil {
		var res Response
		res.Status = "error"
		res.Data = map[string][]interface{}{}
		res.Error = map[string]interface{}{}

		res.Error["message"] = err.Error()
		res.Error["error_code"] = errorCode

		w.WriteHeader(statusCode)
		log.Println(err)

		json.NewEncoder(w).Encode(res)

		return true
	}

	return false
}

// LogsMessage
func LogsMessage(err error, message string) {
	log.Println(err.Error())
	log.Println(message)
}
