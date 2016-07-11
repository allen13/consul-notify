package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"gopkg.in/mgo.v2"
)


type AlertaNotifier struct {
	Url string
	Token string
	MongoHosts string
	MongoDB string
	VerifyActiveNodes bool
}

//Represents the active nodes which will be recorded to a datastore (mongodb)
type Node struct {
	Name string
	Dc string
}

type Nodes []Node

func nodeDiff(a Nodes, b Nodes)(c Nodes) {
	for _,aNode := range a{
		aInB := false
		for _,bNode := range b{
			if aNode.Name == bNode.Name && aNode.Dc == bNode.Dc{
				aInB = true
			}
		}
		if !aInB{
			c = append(c, aNode)
		}
	}

	return
}


func processNodes(serverNodes Nodes, currentNodes Nodes) (alerts Messages, addToMongo Nodes){
	addToMongo = nodeDiff(currentNodes, serverNodes)
	alertNodes := nodeDiff(serverNodes, currentNodes)

  for _,node := range alertNodes {
    alerts = append(alerts, createAlertMessage(node))
  }

	return
}


func (alertaNotifier *AlertaNotifier) CheckActiveNodes(messages Messages) {
	err,serverNodes := alertaNotifier.retrieveNodesFromMongo()
  if err != nil{
    return
  }

	currentNodes := extractNodesFromMessages(messages)
	alerts,addToMongo := processNodes(serverNodes, currentNodes)
  alertaNotifier.alertOnMessages(alerts)
  alertaNotifier.addNodesToMongo(addToMongo)
}


func createAlertMessage(node Node)(Message){
  return Message{
		Node: node.Name,
		Datacenter: node.Dc,
		Check: "Active Node Check",
		Status: "critical",
		CheckId: "active-node-check",
		Output: "Node is absent from current checks.",
		Notes: "Node is absent from current checks. This may not be a problem if the node was manually removed. Remove it from the node cache if this is the case.",
	}
}

func extractNodesFromMessages(messages Messages)(nodes Nodes){
	for _,message := range messages{
    messageNodeExists := false
    for _,node := range nodes{
      if node.Name == message.Node && node.Dc == message.Datacenter{
        messageNodeExists = true
      }
    }
    if !messageNodeExists{
      nodes = append(nodes, Node{message.Node, message.Datacenter})
    }
	}
  return
}

func (alertaNotifier *AlertaNotifier) alertOnMessages(messages Messages){
	for _, message := range messages {
		alertaNotifier.sendToAlerta(message)
	}
}

func (AlertaNotifier *AlertaNotifier) retrieveNodesFromMongo()(err error, nodes Nodes){
	sess, err := mgo.Dial(AlertaNotifier.MongoHosts)
	if err != nil {
		log.Println(err)
    return
	}
	defer sess.Close()

	collection := sess.DB(AlertaNotifier.MongoDB).C("nodes")
	err = collection.Find(nil).Iter().All(&nodes)
	if err != nil {
		log.Println(err)
    return
	}
	return
}

func (alertaNotifier *AlertaNotifier) addNodesToMongo(nodes Nodes) {
	session, err := mgo.Dial(alertaNotifier.MongoHosts)
	if err != nil{
    log.Println(err)
    return
	}
	defer session.Close()

	conn := session.DB(alertaNotifier.MongoDB).C("nodes")
	err = conn.Insert(nodes)

	if err != nil {
    log.Println(err)
    return
	}
}


func (alertaNotifier *AlertaNotifier) Notify(alerts Messages) bool {
	if alertaNotifier.VerifyActiveNodes{
		alertaNotifier.CheckActiveNodes(alerts)
	}

  alertaNotifier.alertOnMessages(alerts)
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
