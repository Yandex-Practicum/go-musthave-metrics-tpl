package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusHandler(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "TestStatus#1",
			want: want{
				code:        200,
				response:    `{"status": "ok"}`,
				contentType: "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reguest := httptest.NewRequest(http.MethodGet, "/status", nil)
			w := httptest.NewRecorder()
			StatusHandler(w, reguest)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.JSONEq(t, tt.want.response, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestUserViewHandler(t *testing.T) {
	type want struct {
		contentType string
		stasusCode  int
		user        User
	}

	tests := []struct {
		name    string
		request string
		users   map[string]any
		want    want
	}{
		{
			name: "test #1 /200",
			users: map[string]any{
				"id1": User{
					ID:        "id1",
					FirstName: "Misha",
					LastName:  "Popov",
				},
			},
			want: want{
				contentType: "application/json",
				stasusCode:  200,
				user: User{
					ID:        "id1",
					FirstName: "Misha",
					LastName:  "Popov",
				},
			},
			request: "/users?user_id=id1",
		},
		{
			name: "test #2 /400",
			users: map[string]any{
				"id1": User{
					ID:        "id1",
					FirstName: "Misha",
					LastName:  "Popov",
				},
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				stasusCode:  400,
				user: User{
					ID:        "id1",
					FirstName: "Misha",
					LastName:  "Popov",
				},
			},
			request: "/users",
		},
		{
			name: "test #3 /404",
			users: map[string]any{
				"id1": User{
					ID:        "id1",
					FirstName: "Misha",
					LastName:  "Popov",
				},
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				stasusCode:  404,
				user:        User{},
			},
			request: "/users?user_id=unknown",
		},
		{
			name: "test #4 /500",
			users: map[string]any{
				"bad": struct{ Ch chan int }{Ch: make(chan int)},
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				stasusCode:  500,
			},
			request: "/users?user_id=bad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(UserViewHandler(tt.users))
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.stasusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-type"))

			userResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if tt.want.stasusCode == http.StatusOK {
				var user User
				err = json.Unmarshal(userResult, &user)
				require.NoError(t, err)

				assert.Equal(t, tt.want.user, user)
			}
		})
	}
}
