package main

import (
	"log"
)

// LOGIN_WORNG 登入失敗
const LOGIN_WORNG = "00001"

// DATA_EMPTY
const DATA_EMPTY = "00002"

// UNEXPECT_ERROR
const UNEXPECT_ERROR = "00003"

// EXIST_USER 已有存在user name不可重復
const EXIST_USER = "00004"

// EXIST_GAME_NOT_ALLOW_TO_CREATE_NEW_ONE 已加入遊戲不可再開新局
const EXIST_GAME_NOT_ALLOW_TO_CREATE_NEW_ONE = "00005"

// NOT_AUTHORIZATION
const NOT_AUTHORIZATION = "00006"

// GAME_NOT_FOUND
const GAME_NOT_FOUND = "00007"

// USER_ACTION_ERROR
const USER_ACTION_ERROR = "00008"

// REDIS_LOST_GAME_INFO Redis的資料有遺失
const REDIS_LOST_GAME_INFO = "00009"

// SESSION_NOT_FOUND 找不到session
const SESSION_NOT_FOUND = "00010"

// UNEXPECT_DB_ERROR DB錯誤
const UNEXPECT_DB_ERROR = "00011"

// UNEXPECT_REDIS_ERROR DB錯誤
const UNEXPECT_REDIS_ERROR = "00012"

func checkErr(msg string, err error) bool {
	if err != nil {
		log.Fatal(msg, err)

		return true
	}

	return false
}
