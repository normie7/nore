package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/normie7/nore/internal/api"
	"github.com/normie7/nore/internal/noiseremover"
	"github.com/normie7/nore/internal/repository/mysql"
	"github.com/normie7/nore/internal/storage"
)

const (
	defaultPort   = "8080"
	FSThresholdMb = 2048 // 2GB
)

type config struct {
	Mysql struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Dbname   string `yaml:"dbname"`
	} `yaml:"mysql"`
	WebDir              string `yaml:"webdir"`
	GoogleGlobalSiteTag string `yaml:"gst"`
	Port                string `yaml:"port"`
}

func main() {
	log.Println("Starting Server")

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	timeToStop := make(chan bool)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	cfg := parseConfig()

	repo := mysql.NewMysqlRepo(cfg.Mysql.User, cfg.Mysql.Password, cfg.Mysql.Host, cfg.Mysql.Port, cfg.Mysql.Dbname)
	defer repo.Close()

	st := storage.NewFileStorage(FSThresholdMb, cfg.WebDir)

	var s noiseremover.Service
	s = noiseremover.NewNoiseRemoverService(st, repo)
	s = noiseremover.NewNoiseRemoverLoggingService(s)

	var b noiseremover.BackgroundService
	b = noiseremover.NewBackGroundService(st, repo)

	r := api.NewRouter(s, cfg.WebDir, cfg.GoogleGlobalSiteTag)

	var wg sync.WaitGroup
	wg.Add(2)
	go startHttpServer(cfg.Port, r, timeToStop, &wg)
	go b.ProcessTicker(250*time.Millisecond, timeToStop, &wg)
	<-termChan // Blocks here until interrupted
	close(timeToStop)
	wg.Wait()
	return
}

func parseConfig() config {

	// webDir must be set to the absolute path to assets
	configPath := flag.String("c", "", "config path")
	flag.Parse()

	f, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var cfg config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.WebDir == "" {
		log.Fatal("webDir must be set to the absolute path to the web folder with static assets and templates")
	}

	if cfg.Port == "" {
		log.Println("port was not provided. using default port:", defaultPort)
		cfg.Port = defaultPort
	}

	return cfg
}

func startHttpServer(port string, r http.Handler, timeToStop chan bool, wg *sync.WaitGroup) {
	server := &http.Server{Addr: ":" + port, Handler: r}
	go gracefullServerShutdown(server, timeToStop, wg)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// todo send os.Signal instead of Fatal
		log.Fatalf("Could not listen on %s: %v\n", port, err)
	}
}

func gracefullServerShutdown(server *http.Server, timeToStop chan bool, wg *sync.WaitGroup) {
	<-timeToStop
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Could not gracefully shutdown the server: %v\n", err)
	}
	log.Println("server done")
	wg.Done()
	return
}
