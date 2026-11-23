package client

import (
	"encoding/json"
	"net/http"
)

func asJSON(obj any, w http.ResponseWriter) error {
	enc := json.NewEncoder(w)
	return enc.Encode(obj)
}
