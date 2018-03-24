package integration

import (
	"os"
	"os/exec"
	"strconv"
	"testing"
)

// Integration test skip test if WITH_INTEGRATION environment variable was not
// set to true.
func IntegrationTest(t *testing.T) {
	if !hasDocker() {
		t.Skip("Integration test. docker and/or docker-compose tools not available")
	}

	if ok, _ := strconv.ParseBool(os.Getenv("WITH_INTEGRATION")); !ok {
		t.Skip("Integration test. Set WITH_INTEGRATION=true to run.")
	}
}

func hasDocker() bool {
	if err := exec.Command("docker", "--version").Run(); err != nil {
		return false
	}
	if err := exec.Command("docker-compose", "--version").Run(); err != nil {
		return false
	}
	return true
}

func TestKafkaCluster(t *testing.T) {
	IntegrationTest(t)

	const clusterSize = 4

	cluster := NewKafkaCluster("kafka-docker/", clusterSize)
	if err := cluster.Start(); err != nil {
		t.Fatalf("cannot start kafka cluster: %s", err)
	}

	addrs, err := cluster.KafkaAddrs()
	if err != nil {
		t.Fatalf("cannot get kafka cluster addresses: %s", err)
	}
	if len(addrs) != clusterSize {
		t.Fatalf("expected %d addresses, got %d (%v)", clusterSize, len(addrs), addrs)
	}

	if err := cluster.Stop(); err != nil {
		t.Fatalf("cannot stop kafka cluster: %s", err)
	}
}

func TestContainerRestart(t *testing.T) {
	IntegrationTest(t)

	cluster := NewKafkaCluster("kafka-docker/", 4)
	if err := cluster.Start(); err != nil {
		t.Fatalf("cannot start kafka cluster: %s", err)
	}

	containers, err := cluster.Containers()
	if err != nil {
		t.Fatalf("cannot get containers info: %s", err)
	}
	// 4 kafka + zookeeper
	if len(containers) != 5 {
		t.Fatalf("expected 5 containers, got %d", len(containers))
	}

	// first stop all zookeeper containers
	for _, container := range containers {
		if container.RunningKafka() {
			continue
		}
		if err := container.Stop(); err != nil {
			t.Fatalf("cannot stop %q container: %s", container.ID, err)
		}
	}
	// then stop all kafka containers
	for _, container := range containers {
		if !container.RunningKafka() {
			continue
		}
		if err := container.Stop(); err != nil {
			t.Fatalf("cannot stop %q container: %s", container.ID, err)
		}
	}

	// first start all zookeeper containers
	for _, container := range containers[1:] {
		if container.RunningKafka() {
			continue
		}
		if err := container.Start(); err != nil {
			t.Fatalf("cannot start %q container: %s", container.ID, err)
		}
	}
	// then start all kafka containers
	for _, container := range containers[1:] {
		if !container.RunningKafka() {
			continue
		}
		if err := container.Start(); err != nil {
			t.Fatalf("cannot start %q container: %s", container.ID, err)
		}
	}

	if err := cluster.Stop(); err != nil {
		t.Fatalf("cannot stop kafka cluster: %s", err)
	}
}
