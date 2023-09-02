package main

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestSorting(T *testing.T) {

	routeList := make([]*Route, 0)

	for i := 10; i <= 100; i++ {
		routeList = append(routeList, &Route{
			Destination: "13.428555,52.523219",
			Duration:    100,
			Distance:    float64(i),
		})
	}
	routeList = append(routeList, &Route{
		Destination: "13.428555,52.523219",
		Duration:    101,
		Distance:    20,
	})

	SortRequestData(routeList)

	for i := range routeList {
		logger.Info("SORT", "Dur", routeList[i].Duration, "Dist", routeList[i].Distance)
	}

	if routeList[0].Distance != 10 {
		T.Fatal("Shortest distance was not 10")
	}
	if routeList[1].Distance != 11 {
		T.Fatal("Second shortest distance was not 11")
	}
	if routeList[len(routeList)-1].Distance != 20 {
		T.Fatal("Longest distance was not 101")
	}

}

func TestEtoEConcurrent(T *testing.T) {
	for i := 0; i < 10; i++ {
		go TestEndToEnd(T)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}

func TestEndToEnd(T *testing.T) {

	httpClient := new(http.Client)

	req, err := http.NewRequest("GET", "http://127.0.0.1/routes?src=13.388860,52.517037&dst=13.397634,52.529407&dst=13.428555,52.523219", nil)
	if err != nil {
		logger.Info("Err", err)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Info("Err", err)
		return
	}

	APIResp := new(APIResponse)
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("Err", err)
		return
	}

	// logger.Info(string(out))

	err = json.Unmarshal(out, APIResp)
	if err != nil {
		logger.Info("Err", err)
		return
	}

	for i := range APIResp.Routes {
		if APIResp.Routes[i].Error != "" {
			logger.Error(APIResp.Routes[i].Error)
		} else {
			logger.Info("Resp", "Dst:", APIResp.Routes[i].Destination, "Dist:", APIResp.Routes[i].Distance, "Dur:", APIResp.Routes[i].Duration)
		}
	}

}
