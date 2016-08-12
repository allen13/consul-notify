package election

import (
	"log"
	"time"
	consulapi "github.com/hashicorp/consul/api"
)

const LockKey = "consul-notify/leader"

type LeaderElection struct {
	lock     *consulapi.Lock
	shutdown chan struct{}
	leader   bool
}

func (l *LeaderElection) start() {
	for {
		log.Println("Running for leader election...")
		lockC, _ := l.lock.Lock(l.shutdown)
		if lockC != nil {
			log.Println("Now acting as leader.")
			l.leader = true
			<-lockC
			l.leader = false
			log.Println("Lost leadership.")
			l.lock.Unlock()
			l.lock.Destroy()
		} else {
			time.Sleep(10000 * time.Millisecond)
		}
	}
}

func (l *LeaderElection) Stop() {
	log.Println("cleaning up")
	l.lock.Unlock()
	l.lock.Destroy()
	l.leader = false
	log.Println("cleanup done")
}

func StartLeaderElection(client *consulapi.Client, shutdown chan struct{})(leader *LeaderElection) {
	lock, _ := client.LockKey(LockKey)

	leader = &LeaderElection{
		lock:     lock,
		shutdown: shutdown,
	}

	go leader.start()

	return
}

func (l *LeaderElection) IsLeader()(bool) {
	return l.leader
}
