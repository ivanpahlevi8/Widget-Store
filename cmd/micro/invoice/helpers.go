package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// create json response
type JsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Content string `json:"content"`
	Id      string `json:"id"`
}

// create function to read json from request body
func (app *application) ReadJsonBodyRequest(w http.ResponseWriter, r *http.Request, payload interface{}) error {
	// set maximum data that can be hold by body
	maximumData := 1104857

	// set to body as maximum data
	r.Body = http.MaxBytesReader(w, r.Body, int64(maximumData))

	// get decoder
	decoder := json.NewDecoder(r.Body)

	// decode payload from http request body
	err := decoder.Decode(&payload)

	// check for an error
	if err != nil {
		log.Println("error when decode json payload from request body to payload object")
		return err
	}

	// check another decoder
	err = decoder.Decode(&struct{}{})

	// check if there is an error
	// error will happen if there are two json object in body request
	if err != io.EOF {
		return err
	}

	// if success
	return nil
}

// create function to response an error
func (app *application) ErrorJsonResponse(w http.ResponseWriter, httpStatus int, err error) {
	// set status
	w.WriteHeader(httpStatus)

	// creaet error payload
	jsonResp := JsonResponse{
		OK:      true,
		Message: fmt.Sprintf("error happen with status code : %d", httpStatus),
		Content: fmt.Sprintf("error message : %s", err.Error()),
		Id:      uuid.New().String(),
	}

	// marshalling response
	objJson, err := json.MarshalIndent(jsonResp, "", "\t")

	// check for an error
	if err != nil {
		log.Println("error when marshalling obejct to json")
		return
	}

	// send json
	w.Write(objJson)
}

// create function to write obejct
func (app *application) WriteJsonObject(w http.ResponseWriter, item interface{}, status int, header ...http.Header) error {
	// set header as json response
	w.Header().Set("Content-Type", "application/json")

	// check if there is header or not
	if len(header) > 0 {
		for k, v := range header[0] {
			w.Header()[k] = v
		}
	}

	// set header status
	w.WriteHeader(status)

	// create json object
	jsonObject, err := json.MarshalIndent(item, "", "\t")

	// check for an error
	if err != nil {
		log.Println("error when converting object to json")
		return err
	}

	// write to output
	_, err = w.Write(jsonObject)

	// check for an error
	if err != nil {
		log.Println("error when write to http output")
		return err
	}

	return nil
}

// create function to make directory in root level to hold invoices
func (app *application) CreateDir(dirName string) error {
	const mode = 0755
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.Mkdir(dirName, mode)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}
	return nil
}
