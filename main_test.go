package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
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

func Test_genUID(t *testing.T) {
	res, err := genUID()
	if err != nil {
		t.Fatalf("Error: %v by genUID", err.Error())
	}
	if len(res) != 18 {
		t.Errorf("genUID have to generate a %v sing long uid but got %v", 18, len(res))
	}
}

type IlmPolicyDBMock struct {
	t                             *testing.T
	setIlmPolicyReturnsError      bool
	ilmPlan                       DataRolloverPattern
	setPolicyTemplateReturnsError bool
	isExecute                     []bool
}

func (db *IlmPolicyDBMock) AddLog(data LogData, index string) {
}

func (db *IlmPolicyDBMock) AddStats(data StatsData, index string) {
}

func (db *IlmPolicyDBMock) SetIlmPolicy(indextype DataRolloverPattern) error {
	db.isExecute[0] = true
	if db.setIlmPolicyReturnsError {
		return errors.New("Mock Error but error")
	}
	if db.ilmPlan != indextype {
		db.t.Errorf("SetIlmPolicy Want deleteage %v but got %v", db.ilmPlan, indextype)
	}
	return nil
}

func (db *IlmPolicyDBMock) SetPolicyTemplate() error {
	if db.setPolicyTemplateReturnsError {
		return errors.New("Mock Error but error")
	}
	return nil
}
func Test_setIlmPolicy(t *testing.T) {
	tests := []struct {
		name                          string
		setIlmPolicyReturnsError      bool
		setPolicyTemplateReturnsError bool
		rolloverplan                  DataRolloverPattern
	}{
		{
			name:                          "Call and no error will be fallen minAge = 20d",
			setIlmPolicyReturnsError:      false,
			setPolicyTemplateReturnsError: false,
			rolloverplan:                  DataRolloverPattern("monthly"),
		},
		{
			name:                          "Call and error will be fallen minAge = 20d",
			setIlmPolicyReturnsError:      true,
			setPolicyTemplateReturnsError: true,
			rolloverplan:                  DataRolloverPattern("monthly"),
		},
		{
			name:                          "Call and error will be fallen from setPolicyTemplateReturnsError minAge = 20d",
			setIlmPolicyReturnsError:      false,
			setPolicyTemplateReturnsError: true,
			rolloverplan:                  DataRolloverPattern("monthly"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := IlmPolicyDBMock{
				t:                             t,
				setIlmPolicyReturnsError:      tt.setIlmPolicyReturnsError,
				setPolicyTemplateReturnsError: tt.setPolicyTemplateReturnsError,
				isExecute:                     make([]bool, 2),
				ilmPlan:                       tt.rolloverplan,
			}
			err := setIlmPolicy(&db, tt.rolloverplan)
			if err != nil && !tt.setIlmPolicyReturnsError && !tt.setPolicyTemplateReturnsError {
				t.Error("Got error but all error mocks are set to false")
			}
			if !db.isExecute[0] && !db.isExecute[1] {
				t.Errorf("Mockfunctions was not been called... so db will not be execute")
			}
		})
	}
}

type ResolverMock struct{}

func (rolv *ResolverMock) Root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
func (rolv *ResolverMock) Subscribe(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Test_registerHandler(t *testing.T) {
	resolver := ResolverMock{}
	handler := registerHandler(&resolver)
	server := httptest.NewServer(handler)
	defer server.Close()
	res, err := http.Get(server.URL)
	if err != nil {
		t.Errorf("Get Error by call server %v", err)
	}
	if res.StatusCode != http.StatusAccepted {
		t.Errorf("Want to call Rootmock but statuscode is incorrect want %v got %v", http.StatusAccepted, res.StatusCode)
	}

	res, err = http.Get(server.URL + "/data/subscribe")
	if err != nil {
		t.Errorf("Get Error by call server %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Want to call Subscribe but statuscode is incorrect want %v got %v", http.StatusAccepted, res.StatusCode)
	}

}
