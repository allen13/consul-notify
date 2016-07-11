package notifier

import (
  "testing"
  "reflect"
)

func TestNodeDiff(t *testing.T) {
	a := Nodes{Node{"a","dc1"}, Node{"b","dc1"}}
	b := Nodes{Node{"b","dc1"}}
	c := nodeDiff(a, b)
	if c[0].Name != "a"{
		t.Error("nodeDiff failed")
	}
}

func TestProcessNodes(t *testing.T) {
	serverNodes  := Nodes{Node{"s1","dc1"}, Node{"s2","dc1"}, Node{"s3","dc1"}}
	currentNodes := Nodes{Node{"s1","dc1"}, Node{"s3","dc1"}, Node{"s4","dc1"}}
	alerts,addToMongo := processNodes(serverNodes, currentNodes)
	if alerts[0].Node != "s2"{
		t.Error("Failed to add s2 to alerts.")
	}
	if addToMongo[0].Name != "s4"{
		t.Error("Failed to add s4 to Mongo.")
	}
}

func TestExtractNodesFromMessages(t *testing.T) {
	testInput := Messages{
    Message{
      Node: "s1",
      Datacenter: "dc1",
      Check: "random check",

    },
    Message{
      Node: "s1",
      Datacenter: "dc1",
      Check: "random check",

    },
    Message{
      Node: "s2",
      Datacenter: "dc1",
      Check: "random check",
    },
  }

  expectedOutput := Nodes{
    Node{
      Name: "s1",
      Dc: "dc1",
    },
    Node{
      Name: "s2",
      Dc: "dc1",
    },
  }

	actualOutput := extractNodesFromMessages(testInput)

  if !reflect.DeepEqual(expectedOutput, actualOutput) {
    t.Errorf("testInput:\n %v \n does not convert properly to node \n expectedOutput:\n %v", testInput, expectedOutput)
  }
}
