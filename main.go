package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	httproutermiddleware "github.com/slok/go-http-metrics/middleware/httprouter"
	"go.uber.org/zap"
)

type server struct {
	redis  redis.UniversalClient
	logger *zap.Logger
}

func main() {

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("unable to initialize logger")
	}

	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{os.Getenv("REDIS_ADDR")},
		Password: "",
		DB:       0,
	})

	srv := &server{
		redis:  rdb,
		logger: logger,
	}

	router := httprouter.New()

	router.GET("/", httproutermiddleware.Handler("/", srv.indexHandler, mdlw))
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})

	router.GET("/ready", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		logger.Info("metrics listening at 8081")
		if err := http.ListenAndServe(":8081", promhttp.Handler()); err != nil {
			log.Panicf("error while serving metrics: %s", err)
		}
	}()

	go func() {
		logger.Info("server started on port 8080")
		log.Fatal(http.ListenAndServe(":8080", router))
	}()

	<-ctx.Done()
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var v string
	var err error
	if v, err = s.redis.Get(context.Background(), "updated_time").Result(); err != nil {
		s.logger.Info("updated_time not found, setting it")
		v = time.Now().Format("2006-01-02 03:04:05")
		s.redis.Set(context.Background(), "updated_time", v, 5*time.Second)
	} else {
		s.logger.Info("got updated_time")
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "hello world: updated_time=%s\n", v)
}
