package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// 直接回傳api錯誤信息
func dealErr(msg string, err error, w http.ResponseWriter) bool {
	if err != nil {
		var res Response
		res.Data = map[string][]interface{}{}
		res.Status = wrong
		res.Data["errorMsg"] = []interface{}{
			msg,
			err,
		}

		log.Fatal(msg, err)
		json.NewEncoder(w).Encode(res)

		return true
	}

	return false
}

func checkErr(msg string, err error) bool {
	if err != nil {
		log.Fatal(msg, err)

		return true
	}

	return false
}
