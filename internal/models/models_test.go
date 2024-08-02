package models

import (
    "testing"
)

func TestNewHTTPError(t *testing.T) {
    tests := []struct {
        name     string
        status   int
        message  string
        want     *HTTPError
    }{
        {
            name:    "Status 404",
            status:  404,
            message: "Not Found",
            want:    &HTTPError{Status: 404, Message: "Not Found"},
        },
        {
            name:    "Status 500",
            status:  500,
            message: "Internal Server Error",
            want:    &HTTPError{Status: 500, Message: "Internal Server Error"},
        },
        {
            name:    "Status 400",
            status:  400,
            message: "Bad Request",
            want:    &HTTPError{Status: 400, Message: "Bad Request"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := NewHTTPError(tt.status, tt.message)
            if got.Status != tt.want.Status || got.Message != tt.want.Message {
                t.Errorf("NewHTTPError() = %v, want %v", got, tt.want)
            }
        })
    }
}