package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type AlertaNotifier struct {
	Url string
	Token string
}


func (alertaNotifier *AlertaNotifier) Notify(alerts Messages) bool {

	for _, alert := range alerts {
		alertSuccess := alertaNotifier.sendToAlerta(alert)
		if !alertSuccess{
			return false
		}
	}
	return true
}

func (alertaNotifier *AlertaNotifier) sendToAlerta(alert Message) bool {


	var Url *url.URL
	Url, err := url.Parse(alertaNotifier.Url + "/alert?api-key=" + alertaNotifier.Token)
	if err != nil {
		return false
	}

	var severity string
	switch alert.Status {
	case "passing":
		severity = "informational"
	case "warning":
		severity = "warning"
	case "critical":
		severity = "critical"
	default:
		severity = "indeterminate"
	}

	postData := make(map[string]interface{})
	postData["resource"] = alert.Node
	postData["event"] = alert.Check
	postData["service"] = []string{alert.CheckId}
	postData["environment"] = alert.Datacenter
	postData["severity"] = severity
	postData["group"] = "check"
	postData["value"] = alert.Output
	postData["text"] = alert.Notes
	postData["origin"] = "consul-" + alert.Datacenter


	var post bytes.Buffer
	enc := json.NewEncoder(&post)
	err = enc.Encode(postData)
	if err != nil {
		return false
	}

	resp, err := http.Post(Url.String(), "application/json", &post)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false
		}
		type response struct {
			Message string `json:"message"`
		}
		r := &response{Message: fmt.Sprintf("failed to understand Alerta response. code: %d content: %s", resp.StatusCode, string(body))}
		b := bytes.NewReader(body)
		dec := json.NewDecoder(b)
		dec.Decode(r)
		log.Println(r.Message)
		return false
	}
	return true
}
