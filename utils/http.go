package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

/*
  Follows the Google JSON style guide
*/

type httpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type httpResponse struct {
	NextLink string      `json:"nextLink,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Error    *httpError  `json:"error,omitempty"`
}

func SendSuccess(w http.ResponseWriter, data interface{}, status int) {
	resp := httpResponse{Data: data}

	respBody, err := json.Marshal(resp)
	if err == nil {
		w.WriteHeader(status)
		w.Write(respBody)
	} else {
		log.Printf("error marshalling response body\n %v\n error\n %v", resp, err)
		SendError(w, "Could not convert response to JSON", http.StatusInternalServerError)
	}
}

func SendPage(w http.ResponseWriter, r *http.Request, data interface{}, offset int, count int, more bool) {
	nextLink := ""
	if more {
		query := r.URL.RawQuery
		if query != "" {
			query += "&"
		} else {
			query += "?"
		}
		query += fmt.Sprintf("offset=%v&count=%v", offset, count)
		nextLink = fmt.Sprintf("https://%v%v%v", r.Host, r.RequestURI, query)
	}

	resp := httpResponse{NextLink: nextLink, Data: data}

	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(resp)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write(respBody)
	} else {
		log.Printf("error marshalling response body\n %v\n error\n %v", resp, err)
		SendError(w, "Could not convert response to JSON", http.StatusInternalServerError)
	}
}

func SendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")

	resp := httpResponse{Error: &httpError{Code: status, Message: message}}

	respBody, err := json.Marshal(resp)
	if err == nil {
		w.WriteHeader(status)
		w.Write(respBody)
	} else {
		log.Printf("error marshalling response body\n %v\n error\n %v", resp, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// takes a pointer to the target structure, structure may be modified on error
func ParseArgs(r *http.Request, target interface{}) error {
	q := r.URL.Query()
	targTypes := reflect.TypeOf(target).Elem()
	targVals := reflect.ValueOf(target).Elem()
	for i := 0; i < targVals.NumField(); i++ {
		targType := targTypes.Field(i)
		targVal := targVals.Field(i)

		key := targType.Tag.Get("query")
		val := q.Get(key)
		if val == "" {
			continue
		}

		switch targType.Type.Kind() {
		case reflect.String:
			targVal.SetString(val)
		case reflect.Int:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return fmt.Errorf("parameter '%v' must be be an integer", key)
			}
			targVal.SetInt(int64(intVal))
		case reflect.TypeOf(time.Time{}).Kind():
			timeVal, err := time.Parse(time.RFC3339, val)
			if err != nil {
				return fmt.Errorf("parameter '%v' must be be a time value", key)
			}
			targVal.Set(reflect.ValueOf(timeVal))
		}
	}

	return nil
}
