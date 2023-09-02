package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	echo "github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

var (
	// Defines how many concurrent requests to the
	// 3rd party service can be made at a time
	MAX_CONCURRENT_REQUESTS int
	// Defines the request timeout of the initial request call,
	// not the timeout to the 3rd party service
	API_REQUEST_TIMEOUT_MS int
	// Defines the request timeout for the 3rd party service
	SERVICE_REQUEST_TIMEOUT_MS int
	// Defines how many idle connections we can have.
	// NOTE: Idle connections can be re-used by the HTTPClient
	MAX_IDLE_CONNECTIONS int
	// Defines the length of the request queue
	// NOTE: It is recommended to set the length of the
	//       request queue to be at least 3x the size of
	//       MAX_CONCURRENT_REQUESTS
	REQUEST_QUEUE_LENGTH int
	// Defines the grace perdiod which each currently active request
	// is given to complete before the API shuts down.
	API_SERVER_SHUTDOWN_GRACE_PREDOD_MS int
	// Defines the IP address which the webserver will be bound to
	BIND_ADDRESS string
	// Defines the PORT which the web server will be bound to
	BIND_PORT string

	E          = echo.New()
	HTTPClient *http.Client
	URL        = "http://router.project-osrm.org/route/v1/driving/"
	logger     = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	RequestSlice []*Request
	RequestQueue chan *Request

	RoutineMonior = make(chan int, 10)
)

func init() {

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("Unable to load .env file", err)
		os.Exit(1)
	}

	MAX_CONCURRENT_REQUESTS, err = strconv.Atoi(os.Getenv("MAX_CONCURRENT_REQUESTS"))
	API_REQUEST_TIMEOUT_MS, err = strconv.Atoi(os.Getenv("API_REQUEST_TIMEOUT_MS"))
	SERVICE_REQUEST_TIMEOUT_MS, err = strconv.Atoi(os.Getenv("SERVICE_REQUEST_TIMEOUT_MS"))
	MAX_IDLE_CONNECTIONS, err = strconv.Atoi(os.Getenv("MAX_IDLE_CONNECTIONS"))
	REQUEST_QUEUE_LENGTH, err = strconv.Atoi(os.Getenv("REQUEST_QUEUE_LENGTH"))
	API_SERVER_SHUTDOWN_GRACE_PREDOD_MS, err = strconv.Atoi(os.Getenv("API_SERVER_SHUTDOWN_GRACE_PREDOD_MS"))

	if err != nil {
		logger.Error("Invalid .env variables", err)
		os.Exit(1)
	}

	BIND_ADDRESS = os.Getenv("BIND_ADDRESS")
	BIND_PORT = os.Getenv("BIND_PORT")

	RequestSlice = make([]*Request, MAX_CONCURRENT_REQUESTS)
	RequestQueue = make(chan *Request, REQUEST_QUEUE_LENGTH)

	initAPI()

}

func main() {

	// The API Server is assigned RoutineMonitor ID 1
	RoutineMonior <- 1
	// The Queue Processor is assigned RoutineMonitor ID 2
	RoutineMonior <- 2

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGTERM)

	for {

		select {
		case _ = <-interrupt:

			ctx := context.Background()
			ctx, _ = context.WithDeadline(ctx, time.Now().Add(time.Millisecond*time.Duration(API_SERVER_SHUTDOWN_GRACE_PREDOD_MS)))

			if err := E.Shutdown(ctx); err != nil {
				logger.Error("Error shutting down API", "error", err)

			}

			os.Exit(1)
		default:
		}

		select {
		case ID := <-RoutineMonior:
			logger.Info("Routine Monitor", "Starting Routine with ID", ID)
			if ID == 1 {
				go StartAPI(ID)
			} else if ID == 2 {
				go ProcessRequestQueue(ID)
			}
		default:
		}

		time.Sleep(50 * time.Millisecond)
	}

}

func ProcessRequestQueue(ID int) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Queue processor panic", "error", r, "stack", string(debug.Stack()))
		}
		RoutineMonior <- ID
	}()

	for {

		for i := range RequestSlice {
			if RequestSlice[i] == nil {
				RequestSlice[i] = <-RequestQueue
				go RequestSlice[i].Process(i)
			}
		}

		time.Sleep(1 * time.Millisecond)
	}

}
