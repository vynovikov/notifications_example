package receiver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/vynovikov/study/notifications_example/internal/pkg/model"
)

type receiverSuite struct {
	suite.Suite
}

func TestTpSuite(t *testing.T) {
	suite.Run(t, new(receiverSuite))
}

var nilError error

type mockApp struct {
	mock.Mock
}

func (m *mockApp) Save(model.WrappedReq) error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockApp) Extract(model.WrappedReq) ([][]byte, error) {
	args := m.Called()
	return args.Get(0).([][]byte), args.Error(1)
}
func (m *mockApp) Count(model.WrappedReq) (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}
func (m *mockApp) AuthInternal(model.WrappedReq) error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockApp) AuthExternal(model.WrappedReq) error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockApp) DeleteLast(model.WrappedReq) error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockApp) Start()                        {}
func (m *mockApp) Stop()                         {}
func (m *mockApp) Log(model.UUIDWrapper, string) {}

func (s *receiverSuite) TestHandleGet() {
	tt := []struct {
		name        string
		number      int
		method      string
		url         string
		on          []string
		ret         [][]interface{}
		reqError    error
		doError     error
		readError   error
		wantResBody []byte
	}{

		{
			name:        "Data absent",
			number:      0,
			method:      "GET",
			url:         "http://localhost:8080/api/v1/notifications?page=1&per_page=10&user_uuid=2593ede0-2301-4480-a452-752f03dcfab0&filter=%7B%7D",
			on:          []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:         [][]interface{}{{nilError}, {[][]byte{}, errors.New("no rows")}, {0, nilError}, {nilError}, {nilError}, {}, {}, {}},
			reqError:    nil,
			doError:     nil,
			readError:   nil,
			wantResBody: []byte(`{"success":true,"meta":{"per_page":10,"current_page":1,"from":1,"to":10,"last_page":1,"total":251},"data":[]}`),
		},

		{
			name:        "Data present",
			number:      0,
			method:      "GET",
			url:         "http://localhost:8080/api/v1/notifications?page=1&per_page=10&user_uuid=2593ede0-2301-4480-a452-752f03dcfab0&filter=%7B%7D",
			on:          []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:         [][]interface{}{{nilError}, {[][]byte{[]byte(`{"category":"cat1","name":"alice","uuid":"azaza"}`)}, nilError}, {0, nilError}, {nilError}, {nilError}, {}, {}, {}},
			reqError:    nil,
			doError:     nil,
			readError:   nil,
			wantResBody: []byte(`{"success":true,"meta":{"per_page":10,"current_page":1,"from":1,"to":10,"last_page":1,"total":251},"data":[{"category":"cat1","id":0,"name":"alice","uuid":"azaza"}]}`),
			//,
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {

			// setting Application
			ma := &mockApp{}
			for j, w := range v.on {
				if len(v.ret[j]) == 0 {
					ma.On(w).Once()
				} else {
					ma.On(w).Return(v.ret[j]...)
				}
			}
			rcvr := NewReceiver(ma, &sync.WaitGroup{}, &sync.WaitGroup{})

			// setting mux
			mux := http.NewServeMux()

			mux.HandleFunc("/api/v1/notifications/batch", rcvr.HandlePut())
			mux.HandleFunc("/api/v1/notifications", rcvr.HandleGet())
			mux.HandleFunc("/api/v1/notifications/count", rcvr.HandleCount())

			// setting server
			srv := http.Server{
				Addr:    ":8080",
				Handler: mux,
			}

			// Starting server
			go func(s *http.Server) {
				s.ListenAndServe()
			}(&srv)

			// Checking server API
			time.Sleep(time.Millisecond * 100)
			req, err := http.NewRequest(v.method, v.url, nil)
			if v.reqError != nil {
				s.Equal(v.reqError, err)
			}

			res, err := http.DefaultClient.Do(req)

			if v.doError != nil {
				s.Equal(v.doError, err)
			}

			resBody, err := io.ReadAll(res.Body)
			if v.readError != nil {
				s.Equal(v.readError, err)
			}
			s.Equal(v.wantResBody, resBody)

			// Stopping server
			srv.Shutdown(context.Background())
			time.Sleep(time.Millisecond * 100)

		})
	}
}

func (s *receiverSuite) TestHandleCount() {
	tt := []struct {
		name        string
		number      int
		method      string
		url         string
		on          []string
		ret         [][]interface{}
		reqError    error
		doError     error
		readError   error
		wantResBody []byte
	}{
		{
			name:        "Data absent",
			number:      0,
			method:      "GET",
			url:         "http://localhost:8080/api/v1/notifications/count?user_uuid=2593ede0-2301-4480-a452-752f03dcfab0",
			on:          []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:         [][]interface{}{{nilError}, {[][]byte{}, nilError}, {0, errors.New("no rows")}, {nilError}, {nilError}, {nilError}, {}, {}, {}},
			reqError:    nil,
			doError:     nil,
			readError:   nil,
			wantResBody: []byte(`{"success":true,"data":{"count":0}}`),
		},

		{
			name:        "Bad request",
			number:      1,
			method:      "GET",
			url:         "http://localhost:8080/api/v1/notifications/count",
			on:          []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:         [][]interface{}{{nilError}, {[][]byte{}, nilError}, {-1, errors.New("in application.Count request has empty user_uuid parameter")}, {nilError}, {nilError}, {nilError}, {}, {}, {}},
			reqError:    nil,
			doError:     nil,
			readError:   nil,
			wantResBody: []byte(`{"success":false,"error":[{"code":50002300,"msg":"Wrong request"}]}`),
		},

		{
			name:        "Data present",
			number:      2,
			method:      "GET",
			url:         "http://localhost:8080/api/v1/notifications/count?user_uuid=2593ede0-2301-4480-a452-752f03dcfab0",
			on:          []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:         [][]interface{}{{nilError}, {[][]byte{}, nilError}, {10, nilError}, {nilError}, {nilError}, {nilError}, {}, {}, {}},
			reqError:    nil,
			doError:     nil,
			readError:   nil,
			wantResBody: []byte(`{"success":true,"data":{"count":10}}`),
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {

			// setting Application
			ma := &mockApp{}
			for j, w := range v.on {
				if len(v.ret[j]) == 0 {
					ma.On(w).Once()
				} else {
					ma.On(w).Return(v.ret[j]...)
				}
			}
			rcvr := NewReceiver(ma, &sync.WaitGroup{}, &sync.WaitGroup{})

			// setting mux
			mux := http.NewServeMux()

			mux.HandleFunc("/api/v1/notifications/batch", rcvr.HandlePut())
			mux.HandleFunc("/api/v1/notifications", rcvr.HandleGet())
			mux.HandleFunc("/api/v1/notifications/count", rcvr.HandleCount())

			// setting server
			srv := http.Server{
				Addr:    ":8080",
				Handler: mux,
			}

			// Starting server
			go func(s *http.Server) {
				s.ListenAndServe()
			}(&srv)

			// Starting server
			go func(s *http.Server) {
				s.ListenAndServe()
			}(&srv)

			// Checking server API
			time.Sleep(time.Millisecond * 100)
			req, err := http.NewRequest(v.method, v.url, nil)
			if v.reqError != nil {
				s.Equal(v.reqError, err)
			}

			res, err := http.DefaultClient.Do(req)

			if v.doError != nil {
				s.Equal(v.doError, err)
			}

			resBody, err := io.ReadAll(res.Body)
			if v.readError != nil {
				s.Equal(v.readError, err)
			}
			s.Equal(v.wantResBody, resBody)

			// Stopping server
			srv.Shutdown(context.Background())
			time.Sleep(time.Millisecond * 100)
		})
	}
}

func (s *receiverSuite) TestHandlePut() {
	tt := []struct {
		name         string
		number       int
		method       string
		url          string
		on           []string
		ret          [][]interface{}
		body         []byte
		appId        string
		appSignature string
		writeError   error
		reqError     error
		doError      error
		readError    error
		wantResBody  []byte
	}{
		{
			name:         "Auth fail",
			number:       0,
			method:       "PUT",
			url:          "http://localhost:8080/api/v1/notifications/batch",
			on:           []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:          [][]interface{}{{nilError}, {[][]byte{}}, {0, errors.New("no rows")}, {nilError}, {errors.New("failed")}, {nilError}, {}, {}, {}},
			body:         []byte(`[{"user_uuid":"2593ede0-2301-4480-a452-752f03dcfab0","category":"new_rank","uuid":"75359b90-a0de-4e50-bbcf-ba400d17033f","task_uuid":null,"object_uuid":"fd7f3b4e-008d-4629-af8e-05fadfe4bd29","name":"25.09 \u041b\u041a\u041b D","description":"\u0412\u044b \u0431\u044b\u043b\u0438 \u043f\u0440\u0438\u0433\u043b\u0430\u0448\u0435\u043d\u044b \u043d\u0430 \u0440\u0430\u0431\u043e\u0442\u0443: 25.09 \u041b\u041a\u041b D.","created_at":"2022-10-02T12:43:46.000000Z"}]`),
			appId:        "",
			appSignature: "",
			writeError:   nil,
			reqError:     nil,
			doError:      nil,
			readError:    nil,
			wantResBody:  []byte(`{"success":false,"data":null,"error":[{"code":50002100,"msg":"Unauthorized"}]}`),
		},

		{
			name:         "Auth success, data added",
			number:       1,
			method:       "PUT",
			url:          "http://localhost:8080/api/v1/notifications/batch",
			on:           []string{"Save", "Extract", "Count", "AuthInternal", "AuthExternal", "Start", "Stop", "Log"},
			ret:          [][]interface{}{{nilError}, {[][]byte{}}, {0, errors.New("no rows")}, {nilError}, {nilError}, {nilError}, {}, {}, {}},
			body:         []byte(`[{"user_uuid":"2593ede0-2301-4480-a452-752f03dcfab0","category":"new_rank","uuid":"75359b90-a0de-4e50-bbcf-ba400d17033f","task_uuid":null,"object_uuid":"fd7f3b4e-008d-4629-af8e-05fadfe4bd29","name":"25.09 \u041b\u041a\u041b D","description":"\u0412\u044b \u0431\u044b\u043b\u0438 \u043f\u0440\u0438\u0433\u043b\u0430\u0448\u0435\u043d\u044b \u043d\u0430 \u0440\u0430\u0431\u043e\u0442\u0443: 25.09 \u041b\u041a\u041b D.","created_at":"2022-10-02T12:43:46.000000Z"}]`),
			appId:        "",
			appSignature: "",
			writeError:   nil,
			reqError:     nil,
			doError:      nil,
			readError:    nil,
			wantResBody:  []byte(`{"success":true,"data":null}`), // null interpreted as []
		},
	}
	for _, v := range tt {
		s.Run(v.name, func() {

			// setting Application
			ma := &mockApp{}
			for j, w := range v.on {
				if len(v.ret[j]) == 0 {
					ma.On(w).Once()
				} else {
					ma.On(w).Return(v.ret[j]...)
				}
			}
			rcvr := NewReceiver(ma, &sync.WaitGroup{}, &sync.WaitGroup{})

			// setting mux
			mux := http.NewServeMux()

			mux.HandleFunc("/api/v1/notifications/batch", rcvr.HandlePut())
			mux.HandleFunc("/api/v1/notifications", rcvr.HandleGet())
			mux.HandleFunc("/api/v1/notifications/count", rcvr.HandleCount())

			// setting server
			srv := http.Server{
				Addr:    ":8080",
				Handler: mux,
			}

			// Starting server
			go func(s *http.Server) {
				s.ListenAndServe()
			}(&srv)

			// Starting server
			go func(s *http.Server) {
				s.ListenAndServe()
			}(&srv)

			// Checking server API
			time.Sleep(time.Millisecond * 100)

			buf := bytes.NewBuffer(v.body)

			req, err := http.NewRequest(v.method, v.url, buf)
			if v.reqError != nil {
				s.Equal(v.reqError, err)
			}
			req.Header.Add("APPID", v.appId)
			req.Header.Add("APPSIGNATURE", v.appSignature)

			res, err := http.DefaultClient.Do(req)

			if v.doError != nil {
				s.Equal(v.doError, err)
			}

			resBody, err := io.ReadAll(res.Body)
			if v.readError != nil {
				s.Equal(v.readError, err)
			}
			s.Equal(v.wantResBody, resBody)

			// Stopping server
			srv.Shutdown(context.Background())
			time.Sleep(time.Millisecond * 100)
		})
	}
}
