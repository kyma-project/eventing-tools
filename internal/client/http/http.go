package http

import (
	"net/http"
)

func NewClient(transport *http.Transport) *http.Client {
	return &http.Client{Transport: transport}
}
