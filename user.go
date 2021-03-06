package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	ErrorManner "boardgame_server/error"
	model "boardgame_server/model"
	redisClient "boardgame_server/redis"

	//Hash
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	UserName string
	Password string
}

type loginRequest struct {
	UserName string
	Password string
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	var loginData loginRequest
	err = json.Unmarshal(body, &loginData)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	userName := loginData.UserName
	if userName == "" {
		ErrorManner.ErrorRespone(errors.New("User Name or Password is required"), DATA_EMPTY, w, 400)
		return
	}

	password := loginData.Password
	if password == "" {
		ErrorManner.ErrorRespone(errors.New("User Name or Password is required"), DATA_EMPTY, w, 400)
		return
	}

	// 驗證帳密
	ID, _, userHashPassword, err := model.GetUserInfoByUserName(userName)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	if !checkPasswordHash(password, userHashPassword) {
		ErrorManner.ErrorRespone(errors.New("User Name or Password Error"), LOGIN_WORNG, w, 400)
		return
	}

	// Save session
	session, _ := redisClient.Store.Get(r, "userInfo")
	session.Values["login"] = true
	session.Values["userName"] = userName
	session.Values["userID"] = ID
	session.Save(r, w)

	var res Response
	res.Status = success

	json.NewEncoder(w).Encode(res)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	var registerData registerRequest
	err = json.Unmarshal(body, &registerData)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	userName := registerData.UserName
	if userName == "" {
		ErrorManner.ErrorRespone(errors.New("UserName is required"), DATA_EMPTY, w, 400)
		return
	}

	password := registerData.Password
	if password == "" {
		ErrorManner.ErrorRespone(errors.New("password is required"), DATA_EMPTY, w, 400)
		return
	}

	// 檢查使用者存在
	_, existUserName, _, _ := model.GetUserInfoByUserName(userName)
	if existUserName != "" {
		ErrorManner.ErrorRespone(errors.New("UserName is Exist"), EXIST_USER, w, 400)
		return
	}

	hash, err := hashPassword(password)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	_, err = model.RegsiterUser(userName, hash)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	var res Response
	res.Status = success

	json.NewEncoder(w).Encode(res)
}

// hash
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)

	return string(bytes), err
}

// check hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}

// 取得session的會員資料
func getSessionUserInfo(r *http.Request) (authorization bool, userID int32, userName string, gameID int32, err error) {
	session, err := redisClient.Store.Get(r, "userInfo")

	if err != nil {
		return false, 0, "", 0, err
	}

	authorization, _ = session.Values["login"].(bool)
	userName, _ = session.Values["userName"].(string)
	userID, _ = session.Values["userID"].(int32)
	gameID, _ = session.Values["gameID"].(int32)

	return authorization, userID, userName, gameID, nil
}

// 會員登出
func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := redisClient.Store.Get(r, "userInfo")
	session.Values["login"] = false
	session.Values["userName"] = ""
	session.Values["userID"] = 0
	session.Save(r, w)

	var res Response
	res.Status = success

	json.NewEncoder(w).Encode(res)
}

func clearGame(w http.ResponseWriter, r *http.Request) {
	session, _ := redisClient.Store.Get(r, "userInfo")
	session.Values["gameID"] = ""
	session.Save(r, w)
}
