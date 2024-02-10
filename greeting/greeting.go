package greeting

import "fmt"

func CalledMethod() {
	fmt.Println("Hello Incrementus")
}

type Result struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Success(m string, c int) *Result {
	return &Result{
		Success: true,
		Message: m,
		Code:    c,
	}
}

func Error(m string, c int) *Result {
	return &Result{
		Success: false,
		Message: m,
		Code:    c,
	}
}
