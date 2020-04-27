package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"balango/internal/Configurations/RoundRobin"

	"balango/internal/ConfigParser"
	"balango/internal/Configurations/Adaptive"
	"balango/internal/Configurations/Config"
)

func buildConfig(poolConfig *Config.Config, Mode string) {

	switch Mode {
	case "RR":
		*poolConfig = RoundRobin.NewRRPool()
	case "A":
		*poolConfig = Adaptive.NewAdaptivePool()
	}
}

func main() {

	var config Config.Config

	var filePath string

	flag.StringVar(&filePath, "config", "config.xml", "Configuration file")

	serverList, mode, port, err := ConfigParser.ParseFile(filePath)

	if err != nil {
		fmt.Println(err)
	}

	buildConfig(&config, mode)
	config.BuildPool(serverList)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(config.LoadBalance),
	}

	go config.HealthCheck()

	log.Printf("Load balancer started at : %d\n", port)
	log.Printf("Configuration Mode : %s", mode)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
