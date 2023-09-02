package main

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"sort"
	"time"

	echo "github.com/labstack/echo/v4"
	m "github.com/labstack/echo/v4/middleware"
)

type DestinationServiceResponse struct {
	Routes []*DestinationRoute `json:"routes"`
	Code   string              `json:"code"`
}

type DestinationRoute struct {
	Duration float64 `json:"duration"`
	Distance float64 `json:"distance"`
}

type APIResponse struct {
	Source string   `json:"source"`
	Routes []*Route `json:"routes"`
}

type Route struct {
	Destination string  `json:"destination"`
	Duration    float64 `json:"duration"`
	Distance    float64 `json:"distance"`

	Error       string `json:"error,omitempty"`
	HTTPCode    int    `json:"statusCode,omitempty"`
	ServiceCode string `json:"serviceCode"`
}

func initAPI() {

	HTTPClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns: MAX_IDLE_CONNECTIONS,
			// TCP-KeepAlive is disable due to the fact that this service
			// does not use long lived TCP connections
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(SERVICE_REQUEST_TIMEOUT_MS) * time.Millisecond,
	}

	E.Use(m.RecoverWithConfig(m.RecoverConfig{
		StackSize:         4 << 10, // 4 KB
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogLevel:          1,
	}))

	E.Use(m.SecureWithConfig(m.DefaultSecureConfig))

	corsConfig := m.CORSConfig{
		Skipper:      m.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}

	E.Use(m.CORSWithConfig(corsConfig))

	E.GET("/routes", Routes)
}

func StartAPI(ID int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("API server panic", "error", r, "stack", string(debug.Stack()))
		}
		RoutineMonior <- ID
	}()

	if err := E.Start(BIND_ADDRESS + ":" + BIND_PORT); err != nil && err != http.ErrServerClosed {
		logger.Error("Unable to start API", err)
	}
}

func Routes(c echo.Context) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("/Routes Panic", "error", r, "stack", string(debug.Stack()))
		}

	}()

	qp := c.QueryParams()

	src, ok := qp["src"]
	if !ok {
		return c.JSON(400, "Source location is missing")
	}

	if len(src) > 1 {
		return c.JSON(400, "Only one source location can be provided")
	}

	dsts, ok := qp["dst"]
	if !ok {
		return c.JSON(400, "Destination location is missing")
	}

	ctx := c.Request().Context()
	ctx, cancel := context.WithCancel(ctx)
	resp, err := GetDurationForDestination(ctx, cancel, src[0], dsts)
	if err != nil {
		return c.JSON(500, err)
	}

	cancel()
	return c.JSON(200, resp)
}

func GetDurationForDestination(ctx context.Context, cancel context.CancelFunc, source string, destinations []string) (resp *APIResponse, err error) {

	requestList := make([]*Request, len(destinations))
	requestIndex := 0
	Timeout := time.After(time.Millisecond * time.Duration(API_REQUEST_TIMEOUT_MS))

	for _, v := range destinations {
		requestList[requestIndex] = new(Request)
		requestList[requestIndex].CTX = ctx
		requestList[requestIndex].Done = make(chan byte, 1)
		requestList[requestIndex].Src = source
		requestList[requestIndex].Dst = v

		select {
		case RequestQueue <- requestList[requestIndex]:
		case <-Timeout:
			cancel()
			return nil, errors.New("Request timeout")
		}
		requestIndex++

	}

	for i := range requestList {
		select {
		case <-requestList[i].Done:
		case <-Timeout:
			cancel()
			return nil, errors.New("Request timeout")
		}
	}

	resp = new(APIResponse)
	resp.Source = source
	resp.Routes = GenerateResponseRoutes(requestList)

	SortRequestData(resp.Routes)

	return resp, nil
}

func SortRequestData(routeList []*Route) {

	// Since go 1.18 sort.Slice has been replaced with a more
	// efficient algorithm which is used by sort.SliceStable
	sort.SliceStable(routeList, func(a, b int) bool {

		// First we compare durations
		if routeList[a].Duration < routeList[b].Duration {
			return true

		} else if routeList[a].Duration == routeList[b].Duration {
			// If durations are equal we compare distances
			if routeList[a].Distance < routeList[b].Distance {
				return true
			}

			return false
		}

		return false
	})

	return
}

func GenerateResponseRoutes(requestList []*Request) (routeList []*Route) {

	for i := range requestList {

		var duration float64 = 0
		var distance float64 = 0
		var code string = ""
		var errString = ""

		if requestList[i].Err != nil {
			errString = requestList[i].Err.Error()
			logger.Error("Error from 3rd party service", "error", errString)
		}

		if requestList[i].Resp != nil && len(requestList[i].Resp.Routes) > 0 {
			duration = requestList[i].Resp.Routes[0].Duration
			distance = requestList[i].Resp.Routes[0].Distance
			code = requestList[i].Resp.Code
		}

		routeList = append(routeList, &Route{
			Error:       errString,
			Destination: requestList[i].Dst,
			HTTPCode:    requestList[i].HTTPCode,
			ServiceCode: code,
			Duration:    duration,
			Distance:    distance,
		})

	}

	return
}
