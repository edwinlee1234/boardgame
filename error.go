package main

import "log"

func checkErr(msg string, err error) bool {
	if err != nil {
		log.Fatal(msg, err)

		return true
	}

	return false
}
