package api

type ErrorMsg struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
