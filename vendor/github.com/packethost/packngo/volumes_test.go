package packngo

import (
	"fmt"
	"testing"
	"time"
)

func waitVolumeActive(id string, c *Client) (*Volume, error) {
	// 15 minutes = 180 * 5sec-retry
	for i := 0; i < 180; i++ {
		<-time.After(5 * time.Second)
		c, _, err := c.Volumes.Get(id)
		if err != nil {
			return nil, err
		}
		if c.State == "active" {
			return c, nil
		}
	}
	return nil, fmt.Errorf("volume %s is still not active after timeout", id)
}

func TestAccVolume(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)

	c, projectID, teardown := setupWithProject(t)
	defer teardown()

	sp := SnapshotPolicy{
		SnapshotFrequency: "1day",
		SnapshotCount:     3,
	}

	vcr := VolumeCreateRequest{
		Size:             10,
		BillingCycle:     "hourly",
		PlanID:           "storage_1",
		FacilityID:       testFacility(),
		SnapshotPolicies: []*SnapshotPolicy{&sp},
	}

	v, _, err := c.Volumes.Create(&vcr, projectID)
	if err != nil {
		t.Fatal(err)
	}

	v, err = waitVolumeActive(v.ID, c)
	defer c.Volumes.Delete(v.ID)

	if len(v.SnapshotPolicies) != 1 {
		t.Fatal("Test volume should have one snapshot policy")
	}

	if v.SnapshotPolicies[0].SnapshotFrequency != sp.SnapshotFrequency {
		t.Fatal("Test volume has wrong snapshot frequency")
	}

	if v.SnapshotPolicies[0].SnapshotCount != sp.SnapshotCount {
		t.Fatal("Test volume has wrong snapshot count")
	}

	if v.Facility.Code != testFacility() {
		t.Fatal("Test volume has wrong facility")
	}
}
