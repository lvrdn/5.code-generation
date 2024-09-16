package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"strconv"
	)

type RR map[string]interface{}
func (srv *MyApi) handlerProfile(w http.ResponseWriter, r *http.Request) (*User, error) {

ctx := r.Context()

genStruct := &ProfileParams {}

fieldLogin := r.FormValue("login")

	if fieldLogin == "" {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("login must be not empty")}
	}
	

genStruct.Login = fieldLogin

return srv.Profile(ctx, *genStruct)
}

func (srv *MyApi) handlerCreate(w http.ResponseWriter, r *http.Request) (*NewUser, error) {


					if r.Header.Get("X-Auth") != "100500" {
					return nil, ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
					}
					

					if r.Method != "POST" {
					return nil, ApiError{http.StatusNotAcceptable, fmt.Errorf("bad method")}
					}
					
ctx := r.Context()

genStruct := &CreateParams {}

fieldLogin := r.FormValue("login")

	if fieldLogin == "" {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("login must be not empty")}
	}
	

	if len(fieldLogin)  < 10 {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("login len must be >= 10")}
	}
	

genStruct.Login = fieldLogin

fieldName := r.FormValue("full_name")

genStruct.Name = fieldName

fieldStatus := r.FormValue("status")
if fieldStatus == "" {
 fieldStatus = "user" }
if fieldStatus != "user" && fieldStatus != "moderator" && fieldStatus != "admin" {
return nil, ApiError{http.StatusBadRequest, fmt.Errorf("status must be one of [user, moderator, admin]")}
}

genStruct.Status = fieldStatus

fieldAge := r.FormValue("age")
fieldAgeInt, err := strconv.Atoi(fieldAge)
	if err != nil {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("age must be int")}
	}
	

	if fieldAgeInt < 0 {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("age must be >= 0")}
	}
	

	if fieldAgeInt > 128 {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("age must be <= 128")}
	}
	

genStruct.Age = fieldAgeInt
return srv.Create(ctx, *genStruct)
}

func (srv *OtherApi) handlerCreate(w http.ResponseWriter, r *http.Request) (*OtherUser, error) {


					if r.Header.Get("X-Auth") != "100500" {
					return nil, ApiError{http.StatusForbidden, fmt.Errorf("unauthorized")}
					}
					

					if r.Method != "POST" {
					return nil, ApiError{http.StatusNotAcceptable, fmt.Errorf("bad method")}
					}
					
ctx := r.Context()

genStruct := &OtherCreateParams {}

fieldUsername := r.FormValue("username")

	if fieldUsername == "" {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("username must be not empty")}
	}
	

	if len(fieldUsername)  < 3 {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("username len must be >= 3")}
	}
	

genStruct.Username = fieldUsername

fieldName := r.FormValue("account_name")

genStruct.Name = fieldName

fieldClass := r.FormValue("class")
if fieldClass == "" {
 fieldClass = "warrior" }
if fieldClass != "warrior" && fieldClass != "sorcerer" && fieldClass != "rouge" {
return nil, ApiError{http.StatusBadRequest, fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")}
}

genStruct.Class = fieldClass

fieldLevel := r.FormValue("level")
fieldLevelInt, err := strconv.Atoi(fieldLevel)
	if err != nil {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("level must be int")}
	}
	

	if fieldLevelInt < 1 {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("level must be >= 1")}
	}
	

	if fieldLevelInt > 50 {
	return nil, ApiError{http.StatusBadRequest, fmt.Errorf("level must be <= 50")}
	}
	

genStruct.Level = fieldLevelInt
return srv.Create(ctx, *genStruct)
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
case "/user/profile" :
result, err := h.handlerProfile(w, r)

	
	if err != nil {

		if errorApi, ok := err.(ApiError); ok {

			w.WriteHeader(errorApi.HTTPStatus)

			dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

			w.Write(dataResutl)
			return
		}

			if err.Error() != "" {

			w.WriteHeader(http.StatusInternalServerError)

			dataResutl, _ := json.Marshal(RR{"error": err.Error()})

			w.Write(dataResutl)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	dataResult, _ := json.Marshal(
		RR{
			"error":    "",
			"response": result,
		},
	)
	w.Write(dataResult)
	
case "/user/create" :
result, err := h.handlerCreate(w, r)

	
	if err != nil {

		if errorApi, ok := err.(ApiError); ok {

			w.WriteHeader(errorApi.HTTPStatus)

			dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

			w.Write(dataResutl)
			return
		}

			if err.Error() != "" {

			w.WriteHeader(http.StatusInternalServerError)

			dataResutl, _ := json.Marshal(RR{"error": err.Error()})

			w.Write(dataResutl)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	dataResult, _ := json.Marshal(
		RR{
			"error":    "",
			"response": result,
		},
	)
	w.Write(dataResult)
	
default:

	errorApi := ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}

	w.WriteHeader(errorApi.HTTPStatus)

	dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

	w.Write(dataResutl)
	
}

}

func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
switch r.URL.Path {
case "/user/create" :
result, err := h.handlerCreate(w, r)

	
	if err != nil {

		if errorApi, ok := err.(ApiError); ok {

			w.WriteHeader(errorApi.HTTPStatus)

			dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

			w.Write(dataResutl)
			return
		}

			if err.Error() != "" {

			w.WriteHeader(http.StatusInternalServerError)

			dataResutl, _ := json.Marshal(RR{"error": err.Error()})

			w.Write(dataResutl)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	dataResult, _ := json.Marshal(
		RR{
			"error":    "",
			"response": result,
		},
	)
	w.Write(dataResult)
	
default:

	errorApi := ApiError{http.StatusNotFound, fmt.Errorf("unknown method")}

	w.WriteHeader(errorApi.HTTPStatus)

	dataResutl, _ := json.Marshal(RR{"error": errorApi.Err.Error()})

	w.Write(dataResutl)
	
}

}

