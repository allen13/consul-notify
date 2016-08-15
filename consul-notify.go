package main

import (
	"log"
	"os"
	"time"

	"github.com/allen13/consul-notify/consul"
	"github.com/allen13/consul-notify/notifier"
	"github.com/docopt/docopt-go"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/pelletier/go-toml"
	"os/signal"
	"github.com/allen13/consul-notify/election"
)

const version = "Consul Notify 1.0.0"
const usage = `Consul Notify.

Usage:
  consul-notify [--config=<config>]
  consul-notify --help
  consul-notify --version

Options:
  --config=<config>            The consul-notify config [default: /etc/consul-notify/consul-notify.conf].
  --help                       Show this screen.
  --version                    Show version.
`

func main() {
	log.SetPrefix("[consul-notifier] ")
	args, _ := docopt.Parse(usage, nil, true, version, false)
	configFile := args["--config"].(string)
	config, err := toml.LoadFile(configFile)
	if err != nil {
		log.Fatalf("config file error: %s", err.Error())
	}

	consulAddr := config.GetDefault("consul.addr", "localhost:8500").(string)
	consulDc := config.GetDefault("consul.dc", "dc1").(string)
	gatherInterval, err := time.ParseDuration(config.GetDefault("consul.gather_interval", "5s").(string))
	if err != nil {
		log.Fatalln(err)
	}

	gatherTimeout, err := time.ParseDuration(config.GetDefault("consul.gather_timeout", "20s").(string))
	if err != nil {
		log.Fatalln(err)
	}

	client, err := consul.NewClient(consulAddr, consulDc)
	if err != nil {
		log.Fatalln(err)
	}

	notifiers := notifier.GetNotifiers(config)
	shutdown := createShutdownChannel()

	leaderElection := election.StartLeaderElection(client.GetApiClient(), shutdown)
	defer leaderElection.Stop()

	gatherTicker := time.NewTicker(gatherInterval)
	defer gatherTicker.Stop()

	for {
		select {
		case <-shutdown:
			log.Println("shutting down consul-notify")
			return
		case <-gatherTicker.C:
			if leaderElection.IsLeader() {
				gatherChecks(client, notifiers, consulDc, gatherTimeout)
			}
		}
	}
}

func gatherChecks(client *consul.ConsulClient, notifiers []notifier.Notifier, dc string, gatherTimeout time.Duration) {

	checks, err := client.GetAllChecks()
	if err != nil {
		log.Println(err)
		return
	}

	messages := processChecks(checks, dc, gatherTimeout)

	for _, notifier := range notifiers {
		notifier.Notify(messages)
	}
}

func processChecks(checks []*consulapi.HealthCheck, datacenter string, gatherTimeout time.Duration) (messages notifier.Messages) {
	messages = make(notifier.Messages, len(checks))

	for i, check := range checks {
		messages[i] = notifier.Message{
			Node:       check.Node,
			ServiceId:  check.ServiceID,
			Service:    check.ServiceName,
			CheckId:    check.CheckID,
			Check:      check.Name,
			Status:     check.Status,
			Output:     check.Output,
			Notes:      check.Notes,
			Datacenter: datacenter,
			Timestamp:  time.Now(),
			Timeout: int(gatherTimeout.Seconds()),
		}
	}
	return
}

func createShutdownChannel() chan struct{} {
	shutdown := make(chan struct{})
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)

	go func() {
		sig := <-signals
		if sig == os.Interrupt {
			close(shutdown)
		}
	}()

	return shutdown
}
