package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	ErrorManner "./error"

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
	allowOrigin(w, r)
	if r.Method == "OPTIONS" {
		return
	}

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
	_, userHashPassword, err := getUserInfoByUserName(userName)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	if !checkPasswordHash(password, userHashPassword) {
		ErrorManner.ErrorRespone(errors.New("User Name or Password Error"), LOGIN_WORNG, w, 400)
	}

	// Save session
	session, _ := store.Get(r, "userInfo")
	session.Values["login"] = true
	session.Values["userName"] = userName
	session.Save(r, w)

	var res Response
	res.Status = success

	json.NewEncoder(w).Encode(res)
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	allowOrigin(w, r)
	if r.Method == "OPTIONS" {
		return
	}

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
	exist, _, _ := getUserInfoByUserName(userName)
	if exist != "" {
		ErrorManner.ErrorRespone(errors.New("UserName is Exist"), EXIST_USER, w, 400)
		return
	}

	hash, err := hashPassword(password)
	if ErrorManner.ErrorRespone(err, UNEXPECT_ERROR, w, 500) {
		return
	}

	_, err = regsiterUser(userName, hash)
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
