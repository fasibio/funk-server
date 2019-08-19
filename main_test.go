package main

import (
	"net/http"
	"testing"
)

func TestHandler_ConnectionAllowed(t *testing.T) {

	type test struct {
		name       string
		key        string
		requestKey string
		want       bool
	}

	data := []test{
		test{
			name:       "keys are the same ",
			key:        "test12359okö_?==98§$%^°*'",
			requestKey: "test12359okö_?==98§$%^°*'",
			want:       true,
		},
		test{
			name:       "keys are different ",
			key:        "test12359okö_?==98§$%^°*'",
			requestKey: "annljsdflk",
			want:       false,
		},
	}
	for _, one := range data {
		handler := Handler{
			connectionkey: one.key,
		}
		r := http.Request{
			Header: make(http.Header),
		}
		r.Header.Add("funk.connection", one.requestKey)

		result := handler.ConnectionAllowed(&r)

		if one.want != result {
			t.Errorf("ConnectionAllowed error by %v want allowed: %v but got %v", one.name, one.want, result)
		}
	}

}
