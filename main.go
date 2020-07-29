package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/agalue/gominion/api"
	"github.com/agalue/gominion/broker"
	_ "github.com/agalue/gominion/rpc"
	_ "github.com/agalue/gominion/sink"
)

func main() {
	hostname, _ := os.Hostname()
	config := &api.MinionConfig{}
	flag.StringVar(&config.ID, "id", hostname, "Minion ID")
	flag.StringVar(&config.Location, "location", "", "Minion Location")
	flag.StringVar(&config.OnmsURL, "onms-url", "http://localhost:8980/opennms", "OpenNMS URL")
	flag.StringVar(&config.BrokerURL, "broker-url", "localhost:8990", "OpenNMS gRPC server connection string")
	flag.IntVar(&config.TrapPort, "trap-port", 1162, "Trap Listener Port")
	flag.IntVar(&config.SyslogPort, "syslog-port", 1514, "Syslog Listener Port")
	flag.Parse()

	log.Printf("Starting OpenNMS Minion, using %s", config.String())
	if err := config.IsValid(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	client := &broker.GrpcClient{}
	if err := client.Start(config); err != nil {
		log.Fatalf("Cannot connect to OpenNMS: %v", err)
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	client.Stop()
}
