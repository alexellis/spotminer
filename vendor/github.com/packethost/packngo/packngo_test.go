package packngo

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	packetTokenEnvVar = "PACKET_AUTH_TOKEN"
	packngoAccTestVar = "PACKNGO_TEST_ACTUAL_API"
	testProjectPrefix = "PACKNGO_TEST_DELME_2d768716_"
	testFacilityVar   = "PACKNGO_TEST_FACILITY"
)

func testFacility() string {
	envFac := os.Getenv(testFacilityVar)
	if envFac != "" {
		return envFac
	}
	return "ewr1"
}

func randString8() string {
	n := 8
	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("acdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func setupWithProject(t *testing.T) (*Client, string, func()) {
	c := setup(t)
	p, _, err := c.Projects.Create(&ProjectCreateRequest{Name: testProjectPrefix + randString8()})
	if err != nil {
		t.Fatal(err)
	}

	return c, p.ID, func() {
		_, err := c.Projects.Delete(p.ID)
		if err != nil {
			panic(fmt.Errorf("while deleting %s: %s", p, err))
		}
	}

}

func skipUnlessAcceptanceTestsAllowed(t *testing.T) {
	if os.Getenv(packngoAccTestVar) == "" {
		t.Skipf("%s is not set", packngoAccTestVar)
	}
}

func setup(t *testing.T) *Client {
	apiToken := os.Getenv(packetTokenEnvVar)
	if apiToken == "" {
		t.Fatalf("If you want to run packngo test, you must export %s.", packetTokenEnvVar)
	}
	c := NewClient("packngo test", apiToken, nil)
	return c
}

func projectTeardown(c *Client) {
	ps, _, err := c.Projects.List()
	if err != nil {
		panic(fmt.Errorf("while teardown: %s", err))
	}
	for _, p := range ps {
		if strings.HasPrefix(p.Name, testProjectPrefix) {
			_, err := c.Projects.Delete(p.ID)
			if err != nil {
				panic(fmt.Errorf("while deleting %s: %s", p, err))
			}
		}
	}
}

func TestAccInvalidCredentials(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	c := NewClient("packngo test", "wrongApiToken", nil)
	_, r, expectedErr := c.Projects.List()
	matched, err := regexp.MatchString(".*Invalid.*", expectedErr.Error())
	if err != nil {
		t.Fatalf("Err while matching err string from response err %s: %s", expectedErr, err)
	}
	if r.StatusCode != 401 {
		t.Fatalf("Expected 401 as response code, got: %d", r.StatusCode)
	}

	if !matched {
		t.Fatalf("Unexpected error string: %s", expectedErr)
	}

}
