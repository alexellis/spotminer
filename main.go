// Copyright (c) Alex Ellis 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/packethost/packngo"
	yaml "gopkg.in/yaml.v2"
)

// Config for spotd
type Config struct {
	Packet struct {
		ProjectID string `yaml:"project_id"`
		APIKey    string `yaml:"api_key"`
	} `yaml:"packet"`
	Preferences struct {
		MaxSpotInstances int     `yaml:"max_spot_instances"`
		MaxPrice         float64 `yaml:"max_price"`
		Algorithm        string  `yaml:"mine_algo"`
		Port             int
		BitcoinWallet    string `yaml:"bitcoin_wallet"`
	} `yaml:"preferences"`
}

func main() {
	configFile := "./config.yml"
	if val, exists := os.LookupEnv("CONFIG_FILE"); exists {
		configFile = val
	}

	config := Config{}
	ymlBytes, readErr := ioutil.ReadFile(configFile)
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", readErr.Error())
		os.Exit(1)
	}

	err := yaml.Unmarshal(ymlBytes, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(2)
	}

	if len(config.Packet.ProjectID) == 0 {
		fmt.Fprintf(os.Stderr, "Provide a value for ProjectID\n")
		os.Exit(1)
	}
	if len(config.Packet.APIKey) == 0 {
		fmt.Fprintf(os.Stderr, "Provide a value for apiKey\n")
		os.Exit(1)
	}

	httpClient := http.Client{}
	api := packngo.NewClient("", config.Packet.APIKey, &httpClient)

	devices, _, err := api.Devices.List(config.Packet.ProjectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	var printIP bool
	flag.BoolVar(&printIP, "print-ip", false, "print IPs only")
	flag.Parse()

	spots := 0
	for _, device := range devices {
		if device.SpotInstance {
			if printIP {
				var ip4 string
				for _, ip := range device.Network {
					if ip.Public && ip.AddressFamily == 4 {
						ip4 = ip.Address
					}
				}
				fmt.Printf("%s\t%s\t%s\n", ip4, since(device.Created), device.TerminationTime)
			}
			spots++
		}
	}

	fmt.Printf("You have %d/%d spot instances.\n", spots, config.Preferences.MaxSpotInstances)

	if printIP {
		os.Exit(0)
	}

	if spots > config.Preferences.MaxSpotInstances {
		fmt.Printf("Too many hosts allocated got: %d\nRemediating...\n", spots)
		deleteHosts(spots, config.Preferences.MaxSpotInstances, devices, api)
		return
	} else if spots >= config.Preferences.MaxSpotInstances {
		fmt.Printf("Cannot allocate more hosts, got: %d\n", spots)
		return
	}

	// fmt.Println(api.Projects.List())
	priceMap, _, err := api.SpotMarket.Prices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		return
	}

	desiredPlans := []string{
		"baremetal_3",
		"baremetal_2",
		"baremetal_1",
		"baremetal_1e",
		"baremetal_0",
	}

	fmt.Println(desiredPlans)
	// fmt.Println(priceMap)
	matches := getMatches(priceMap, desiredPlans)

	for _, match := range matches {
		fmt.Printf("Name: %s Location: %s Price: %f\n", match.Plan, match.Installation, match.Price)
	}

	fmt.Println()

	sort.Sort(BySpotMatch(matches))

	fmt.Println("Sorted matches")
	for _, match := range matches {
		fmt.Printf("Name: %s Location: %s Price: %f Weight: %f\n", match.Plan, match.Installation, match.Price, getPowerWeights()[match.Plan])
	}
	for _, match := range matches {
		if match.Price <= config.Preferences.MaxPrice {

			fmt.Printf("[*] Name: %s Location: %s Price: %f\n", match.Plan, match.Installation, match.Price)
			imageTag := "2018-1-2"

			if match.Plan == "baremetal_0" {
				imageTag = "atom"
			}

			hostname := fmt.Sprintf("spot%s%s", match.Installation, strings.Replace(match.Plan, "_", "", -1))

			// userData aka cloud-init is used to 1) install Docker, setup the miner and
			// configure the stratum server / bitcoin wallet address
			userData := `#!/bin/bash
curl -sL get.docker.com | sh
docker swarm init --advertise-addr=$(hostname -i)
docker service rm wolf ; docker service create --mode=global --name wolf alexellis2/cpu-opt:` + imageTag + ` ./cpuminer -a ` + config.Preferences.Algorithm + ` -o ` + getStratumServer(match.Installation, config.Preferences.Algorithm, config.Preferences.Port) + ` -u ` + config.Preferences.BitcoinWallet + `.` + hostname + `
docker service logs wolf -f`

			createReq := &packngo.DeviceCreateRequest{
				Plan:         match.Plan,
				Facility:     match.Installation,
				Hostname:     hostname,
				ProjectID:    config.Packet.ProjectID,
				SpotInstance: true,
				SpotPriceMax: match.Price,
				OS:           "ubuntu_16_04",
				BillingCycle: "hourly",
				UserData:     userData,
			}

			fmt.Printf("Creating: %s\n", hostname)
			device, resp, err := api.Devices.Create(createReq)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s", err)
				os.Exit(-1)
			}

			fmt.Println(resp.StatusCode)
			fmt.Println(device)

			spots++
			fmt.Printf("Spots - %d/%d\n", spots, config.Preferences.MaxSpotInstances)
			if spots >= config.Preferences.MaxSpotInstances {
				fmt.Println("Cannot allocate more hosts, maximum reached.")
				return
			}
		}
	}
}

func getMatches(priceMap packngo.PriceMap, desiredPlans []string) []SpotMatch {
	matches := []SpotMatch{}

	for k, v := range priceMap {

		for planName, planPrice := range v {
			for _, desiredPlan := range desiredPlans {
				if planName == desiredPlan {
					matches = append(matches, SpotMatch{
						Installation: k,
						Plan:         planName,
						Price:        planPrice,
					})
					// fmt.Printf("Name: %s Price: %f\n", planName, planPrice)
				}
			}
		}
	}
	return matches
}

type SpotMatch struct {
	Installation string
	Plan         string
	Price        float64
}

// BySpotMatch sorts by plan weighting (i.e. our preference for one machine other another)
type BySpotMatch []SpotMatch

func (a BySpotMatch) Len() int      { return len(a) }
func (a BySpotMatch) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySpotMatch) Less(i, j int) bool {
	weights := getPowerWeights()
	return weights[a[i].Plan] > weights[a[j].Plan]
}

func since(val string) string {
	t, err := time.Parse(time.RFC3339Nano, val)
	x := time.Since(t).String()

	if err != nil {
		panic(err)
	}
	if strings.Contains(x, ".") {
		x = x[:strings.Index(x, ".")] + "s"
	}
	return x
}

func deleteHosts(spots int, MaxSpotInstances int, devices []packngo.Device, api *packngo.Client) error {
	deviceIDs := []string{}
	for i := len(devices) - 1; i >= 0; i-- {

		if spots-len(deviceIDs) == MaxSpotInstances {
			break
		}

		if devices[i].SpotInstance {
			deviceIDs = append(deviceIDs, devices[i].ID)
		}
	}

	if len(deviceIDs) > 0 {
		log.Printf("Devices to delete: %d", len(deviceIDs))

		for _, device := range deviceIDs {
			resp, err := api.Devices.Delete(device)
			log.Printf("Deleting device: %s, response: %s", device, resp.Status)
			if err != nil {
				log.Printf("Error deleting device %s %s\n", device, err.Error())
			}
		}
	}
	return nil
}

func getPowerWeights() map[string]float32 {

	powerWeights := make(map[string]float32)
	powerWeights["baremetal_3"] = 2
	powerWeights["baremetal_2"] = 3
	powerWeights["baremetal_1"] = 1
	powerWeights["baremetal_1e"] = 1.5
	powerWeights["baremetal_0"] = 0.5
	return powerWeights
}

func getStratumServer(val string, mineAlgo string, port int) string {
	algo := mineAlgo

	var region string
	switch val {
	case "fra1",
		"ams1":
		region = "eu"

		break
	case "syd1",
		"hkg1":
		region = "hk"
		break
	case "nrt1":
		region = "jp"
		break
	default:
		region = "usa"
	}

	server := fmt.Sprintf("stratum+tcp://%s.%s.nicehash.com:%d", algo, region, port)

	return server
}
