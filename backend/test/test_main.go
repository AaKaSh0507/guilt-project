package test

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// start docker compose test
	cmd := exec.Command("docker", "compose", "-f", "test/docker/docker-compose.test.yml", "up", "-d")
	if err := cmd.Run(); err != nil {
		fmt.Println("failed to start docker compose test:", err)
		os.Exit(1)
	}

	// wait for postgres + redis
	time.Sleep(3 * time.Second)

	// create test db
	create := exec.Command("bash", "test/scripts/create_test_db.sh")
	out, err := create.Output()
	if err != nil {
		fmt.Println("failed to create test db:", err)
		teardown()
		os.Exit(1)
	}
	testDBURL := string(out)
	os.Setenv("TEST_DB_URL", testDBURL)
	os.Setenv("TEST_REDIS_URL", "redis://localhost:6380")

	// run tests
	code := m.Run()

	// drop db
	_ = exec.Command("bash", "test/scripts/drop_test_db.sh").Run()

	teardown()
	os.Exit(code)
}

func teardown() {
	_ = exec.Command("docker", "compose", "-f", "test/docker/docker-compose.test.yml", "down", "-v").Run()
}
