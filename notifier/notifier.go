package notifier

import (
	"time"
	"github.com/pelletier/go-toml"
	"log"
)

const (
	SYSTEM_HEALTHY  string = "HEALTHY"
	SYSTEM_UNSTABLE string = "UNSTABLE"
	SYSTEM_CRITICAL string = "CRITICAL"
)

type Message struct {
	Node         string
	ServiceId    string
	Service      string
	CheckId      string
	Check        string
	Status       string
	Output       string
	Notes        string
	Datacenter   string
	//timeout in seconds before alert is considered stale
	Timeout      int
	Interval  int
	Timestamp time.Time
}

type Messages []Message

type Notifier interface {
	Notify(alerts Messages) bool
}

func (m Message) IsCritical() bool {
	return m.Status == "critical"
}

func (m Message) IsWarning() bool {
	return m.Status == "warning"
}

func (m Message) IsPassing() bool {
	return m.Status == "passing"
}

func (m Messages) Summary() (overallStatus string, pass, warn, fail int) {
	hasCritical := false
	hasWarnings := false
	for _, message := range m {
		switch {
		case message.IsCritical():
			hasCritical = true
			fail++
		case message.IsWarning():
			hasWarnings = true
			warn++
		case message.IsPassing():
			pass++
		}
	}
	if hasCritical {
		overallStatus = SYSTEM_CRITICAL
	} else if hasWarnings {
		overallStatus = SYSTEM_UNSTABLE
	} else {
		overallStatus = SYSTEM_HEALTHY
	}
	return
}

func GetNotifiers(config *toml.TomlTree) (notifiers []Notifier) {
	notifiers = []Notifier{}
	if config.GetDefault("log.enabled", false).(bool) {
		logNotifier := &LogNotifier{
			LogFile:   config.GetDefault("log.file", "/var/log/consul-notify/consul-notify.log").(string),
		}
		notifiers = append(notifiers, logNotifier)
	}
	if config.GetDefault("alerta.enabled", false).(bool) {
		alertaNotifier := &AlertaNotifier{
			Url: config.GetDefault("alerta.url", "http://localhost:8000").(string),
			Token: config.GetDefault("alerta.token", "").(string),
			TLSSkipVerify: config.GetDefault("alerta.tls_skip_verify", true).(bool),
		}
		notifiers = append(notifiers, alertaNotifier)
	}
	if config.GetDefault("pagerduty.enabled", false).(bool) {
		log.Println("Pagerduty enabled?")
		pagerdutyNotifier := &PagerDutyNotifier{
			ServiceKey: config.Get("pagerduty.servicekey").(string),
			ClientName: config.Get("pagerduty.clientname").(string),
			ClientUrl:  config.Get("pagerduty.clienturl").(string),
		}
		notifiers = append(notifiers, pagerdutyNotifier)
	}
	return
}