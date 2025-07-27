package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	// Makes the json human readable, not sure why I would use this though if the data is going to a client side machine (i would assume it would make the response slower?)
	// I guess I'll stick to the below comment instead and remove the appended new line below
	// js, err := json.Marshal(data)
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func ReadIDParam(r *http.Request) (int64, error) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		return 0, errors.New("invalid ID parameter")
	}

	id, err := strconv.ParseInt(idParam, 10, 64) // Base 10, 64 bit int
	if err != nil {
		fmt.Println(err)
		return 0, errors.New("invalid ID parameter")
	}

	return id, nil
}