package main

import (
	"log"
	"github.com/docopt/docopt-go"
	"github.com/pelletier/go-toml"
	"github.com/allen13/consul-notify/notifier"
	"os"
	"syscall"
	"time"
	"os/exec"
	"io/ioutil"
	"encoding/json"
)

const version = "Consul Notify 0.0.1"
const usage = `Consul Notify.

Usage:
  consul-notify start [--config=<config>]
  consul-notify watch [--config=<config>]
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
		log.Fatalf("config file error: ", err.Error())
	}

	consulAddr := config.GetDefault("consul.addr", "localhost:8500").(string)
	consulDc := config.GetDefault("consul.dc", "dc1").(string)

	switch {
	case args["start"].(bool):
		runWatcher(consulAddr, consulDc, "checks")
	case args["watch"].(bool):
		handleWatch(consulDc, config)
	}
}


func handleWatch(consulDc string, config *toml.TomlTree){
	var checks []Check
	readConsulStdinToWatchObject(&checks)

	messages := ProcessChecks(checks, consulDc)
	notifiers := notifier.GetNotifiers(config)

	for _,notifier := range notifiers{
		notifier.Notify(messages)
	}
}

func runWatcher(consulAddr, datacenter, watchType string) {
	consulNotify := os.Args[0]
	cmd := exec.Command(
		"consul", "watch",
		"-http-addr", consulAddr,
		"-datacenter", datacenter,
		"-type", watchType,
		consulNotify, "watch")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("Starting watcher...")
	if err := cmd.Run(); err != nil {
		var exitCode int
		switch err.(type) {
		case *exec.ExitError:
			exitError, _ := err.(*exec.ExitError)
			status, _ := exitError.Sys().(syscall.WaitStatus)
			exitCode = status.ExitStatus()
			log.Println("Shutting down watcher --> Exit Code: ", exitCode)
		case *exec.Error:
			exitCode = 1
			log.Println("Shutting down watcher --> Something went wrong running consul watch: ", err.Error())
		default:
			exitCode = 127
			log.Println("Shutting down watcher --> Unknown error: ", err.Error())
		}
		os.Exit(exitCode)
	} else {
		log.Printf("Execution complete.")
	}
}

func readConsulStdinToWatchObject(watchObject interface{}) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Println("stdin read error: ", err)
	}
	err = json.Unmarshal(data, watchObject)
	if err != nil {
		log.Println("json unmarshall error: ", err)
	}
}

type Check struct {
	Node        string
	CheckID     string
	Name        string
	Status      string
	Notes       string
	Output      string
	ServiceID   string
	ServiceName string
}

func ProcessChecks(checks []Check, datacenter string) (messages notifier.Messages){
	messages = make(notifier.Messages, len(checks))
	for i, check := range checks {
		messages[i] = notifier.Message{
			Node:      check.Node,
			ServiceId: check.ServiceID,
			Service:   check.ServiceName,
			CheckId:   check.CheckID,
			Check:     check.Name,
			Status:    check.Status,
			Output:    check.Output,
			Notes:     check.Notes,
			Datacenter: datacenter,
			Timestamp: time.Now(),
		}
	}
	return
}


