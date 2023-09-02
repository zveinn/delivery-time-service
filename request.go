package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"runtime/debug"
)

// The Request object is the most crucial object in this service.
// It gets created during the initial API request and its pointer
// is loaded into the RequestList by the ProcessRequestQueue function.
// Once loaded its Process function is called and on its defer the
// RequestList index is cleared, making room for new Request objects
// in the RequestList.
type Request struct {
	Src      string
	Dst      string
	Resp     *DestinationServiceResponse
	Err      error
	HTTPCode int
	Done     chan byte
	CTX      context.Context
}

func (R *Request) Finished() {
	select {
	case R.Done <- 1:
	default:
	}
}

func (R *Request) Process(index int) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Panic while processing request", "error", r, "stack", string(debug.Stack()))
		}

		RequestSlice[index] = nil
		R.Finished()
	}()

	select {
	case <-R.CTX.Done():
		return
	default:
	}

	req, err := http.NewRequest("GET", URL+R.Src+";"+R.Dst+"?overview=false&steps=false&alternatives=false&annotations=false", nil)
	if err != nil {
		R.Err = err
		return
	}

	req.Header.Add("Content-Type", "application/json")
	// R.CTX is linked with the original requests context.
	// if the original request is canceled this request
	// should be canceled as well.
	req.WithContext(R.CTX)

	resp, err := HTTPClient.Do(req)
	if err != nil {
		if resp != nil {
			R.HTTPCode = resp.StatusCode
			R.Err = err
			return
		} else {
			R.Err = err
			return
		}
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		R.HTTPCode = resp.StatusCode
		R.Err = err
		return
	}

	R.Err = nil
	R.HTTPCode = resp.StatusCode
	R.Resp = new(DestinationServiceResponse)
	R.Resp.Routes = make([]*DestinationRoute, 0)

	// Requests returning code 429 will not include a body
	if resp.StatusCode == 429 {
		R.Err = errors.New("Service ratelimit reached")
		return
	}

	if len(out) < 1 {
		return
	}

	err = json.Unmarshal(out, R.Resp)
	if err != nil {
		R.Err = err
		return
	}

	return

}
