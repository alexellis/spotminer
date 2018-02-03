package packngo

import "testing"

func TestAccSpotMarket(t *testing.T) {
	skipUnlessAcceptanceTestsAllowed(t)
	t.Parallel()

	c := setup(t)
	prices, _, err := c.SpotMarket.Prices()
	if err != nil {
		t.Fatal(err)
	}

	dcs := []string{"dfw1", "ewr1", "nrt1", "ord1", "sea1", "sjc1", "ams1", "atl1", "iad1", "lax1"}
	for _, dc := range dcs {
		if val, ok := prices[dc]; ok {
			if len(val) == 0 {
				t.Errorf("spot market listing for facility %s doesn't contain any plan prices: %v", dc, val)
			}
		} else {
			t.Errorf("facility %s not in spot prices market map: %v", dc, prices)
		}
	}

}
