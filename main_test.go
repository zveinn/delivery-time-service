package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestSorting(T *testing.T) {

	requestList := make([]*Request, 0)

	R1 := new(Request)
	R1.Dst = "13.428555,52.523219"
	R1.Src = "13.388860,52.517037"
	R1.Resp = new(DestinationServiceResponse)
	R1.Resp.Routes = append(R1.Resp.Routes, &DestinationRoute{
		Distance: 50,
		Duration: 100,
	})

	R2 := new(Request)
	R2.Dst = "13.428555,52.523219"
	R2.Src = "13.388860,52.517037"
	R2.Resp = new(DestinationServiceResponse)
	R2.Resp.Routes = append(R2.Resp.Routes, &DestinationRoute{
		Distance: 100,
		Duration: 100,
	})

	R3 := new(Request)
	R3.Dst = "13.428555,52.523219"
	R3.Src = "13.388860,52.517037"
	R3.Resp = new(DestinationServiceResponse)
	R3.Resp.Routes = append(R3.Resp.Routes, &DestinationRoute{
		Distance: 200,
		Duration: 200,
	})

	R4 := new(Request)
	R4.Dst = "13.428555,52.523219"
	R4.Src = "13.388860,52.517037"
	R4.Resp = new(DestinationServiceResponse)
	R4.Resp.Routes = append(R4.Resp.Routes, &DestinationRoute{
		Distance: 25,
		Duration: 100,
	})

	R5 := new(Request)
	R5.Dst = "13.428555,52.523219"
	R5.Src = "13.388860,52.517037"
	R5.Resp = new(DestinationServiceResponse)

	requestList = append(requestList, R1, R2, R3, R4, R5)

	SortRequestData(requestList)

	if requestList[0].Resp.Routes[0].Distance != 25 {
		T.Fatal("Shortest distance should have been 25 but is", requestList[0].Resp.Routes[0].Distance)
	}
	if requestList[1].Resp.Routes[0].Distance != 50 {
		T.Fatal("Shortest distance should have been 50 but is", requestList[1].Resp.Routes[0].Distance)
	}
	if requestList[2].Resp.Routes[0].Distance != 100 {
		T.Fatal("Shortest distance should have been 100 but is", requestList[2].Resp.Routes[0].Distance)
	}

	for i := range requestList {
		if requestList[i].Resp == nil || len(requestList[i].Resp.Routes) < 1 {
			log.Println("Dst:", requestList[i].Dst, " - NO ROUTE")
		} else {
			log.Println("Dst:", requestList[i].Dst, "Dur:", requestList[i].Resp.Routes[0].Duration, "Dist:", requestList[i].Resp.Routes[0].Distance)
		}
	}
}

func TestEndToEndConcurrent(T *testing.T) {
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
		log.Println(err)
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	APIResp := new(APIResponse)
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(string(out))

	err = json.Unmarshal(out, APIResp)
	if err != nil {
		log.Println(err)
		return
	}

	for i := range APIResp.Routes {
		if APIResp.Routes[i].Error != "" {
			log.Println(APIResp.Routes[i].Error)
		} else {
			log.Println("Dst:", APIResp.Routes[i].Destination, "Dist:", APIResp.Routes[i].Distance, "Dur:", APIResp.Routes[i].Duration)
		}
	}

}
