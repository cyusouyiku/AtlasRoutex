package dto

type APIResponse struct {
Code    string      `json:"code"`
Message string      `json:"message"`
Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}) APIResponse {
return APIResponse{Code: "OK", Message: "success", Data: data}
}

func Fail(code, msg string) APIResponse {
return APIResponse{Code: code, Message: msg}
}
