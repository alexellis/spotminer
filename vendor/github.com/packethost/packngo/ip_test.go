package packngo

import (
	"path"
	"testing"
)

func TestAccIPReservation(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)

	c, projectID, teardown := setupWithProject(t)
	defer teardown()
	quantityToMask := map[int]int{
		1: 32, 2: 31, 4: 30, 8: 29, 16: 28,
	}

	testFac := "ewr1"
	quantity := 2

	ipList, _, err := c.ProjectIPs.List(projectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(ipList) != 0 {
		t.Fatalf("There should be no reservations a new project, existing list: %s", ipList)
	}

	req := IPReservationRequest{
		Type:     "public_ipv4",
		Quantity: quantity,
		Comments: "packngo test",
		Facility: testFac,
	}

	res, _, err := c.ProjectIPs.Request(projectID, &req)
	if err != nil {
		t.Fatal(err)
	}

	if res.CIDR != quantityToMask[quantity] {
		t.Fatalf(
			"CIDR prefix length for requested reservation should be %d, was %d",
			quantityToMask[quantity], res.CIDR)
	}

	if path.Base(res.Project.Href) != projectID {
		t.Fatalf("Wrong project linked in reserved block: %s", res.Project.Href)
	}

	if res.Management {
		t.Fatal("Management flag of new reservation block must be False")
	}
	if res.Facility.Code != testFac {
		t.Fatalf(
			"Facility of new reservation should be %s, was %s", testFac,
			res.Facility.Code)
	}

	ipList, _, err = c.ProjectIPs.List(projectID)
	if len(ipList) != 1 {
		t.Fatalf("There should be only one reservation, was: %s", ipList)
	}
	if err != nil {
		t.Fatal(err)
	}

	sameRes, _, err := c.ProjectIPs.Get(res.ID)
	if err != nil {
		t.Fatal(err)
	}
	if sameRes.ID != res.ID {
		t.Fatalf("re-requested test reservation should be %s, is %s",
			res, sameRes)
	}

	availableAddresses, _, err := c.ProjectIPs.AvailableAddresses(
		res.ID, &AvailableRequest{CIDR: 32})
	if err != nil {
		t.Fatal(err)
	}
	if len(availableAddresses) != quantity {
		t.Fatalf("New block should have %d available addresses, got %s",
			quantity, availableAddresses)
	}

	_, err = c.ProjectIPs.Remove(res.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = c.ProjectIPs.Get(res.ID)
	if err == nil {
		t.Fatalf("Reservation %s should be deleted at this point", res)
	}
}
