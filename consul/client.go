package consul

import (
	consulapi "github.com/hashicorp/consul/api"
	"log"
	"time"
)

type ConsulClient struct {
	api    *consulapi.Client
}

func NewClient(address, dc string) (*ConsulClient, error) {
	config := consulapi.DefaultConfig()
	config.Address = address
	config.Datacenter = dc
	api, _ := consulapi.NewClient(config)

	client := &ConsulClient{
		api:    api,
	}

	try := 1
	for {
		try += try
		_, err := client.api.Status().Leader()
		if err != nil {
			log.Println("Waiting for consul:", err)
			if try > 10 {
				return nil, err
			}
			time.Sleep(10000 * time.Millisecond)
		} else {
			break
		}
	}

	return client, nil
}

func (c *ConsulClient) GetAllChecks()(healthChecks []*consulapi.HealthCheck, err error){
	healthChecks, _, err = c.api.Health().State("any", &consulapi.QueryOptions{})
	return
}

func (c *ConsulClient) GetApiClient()(*consulapi.Client){
	return c.api
}