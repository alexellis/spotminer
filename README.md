spotminer
--------------------------

## What is this?

`spotminer` automates the Packet.net spot market and a cryptocurrency miner so that you can lower the costs of mining in the cloud and access bare metal performance.

Features:

* Set a budget (max price) for each host
* Automatically sets the closest stratum server
* Places the minimum bid for the host
* Can be run on a timer - i.e. every 5 minutes to ensure reclaimed hosts are replaced
* Easy configuration in YAML
* Docker Swarm used as an init process to keep the miner running if it crashes
* Configure the algorithm and port i.e. hodl or cryptonight
* Atom hosts are supported through a separate Docker image

The config file is read from `config.yml`, so copy `config.example.yml` as a template and fill in your packet API key and project ID. Set the `CONFIG_FILE` enviromental variable for a different filename or path.

## Q&A

For any questions please consult my [mine-with-docker project](https://github.com/alexellis/mine-with-docker)

## Packages:

Dependencies are managed through the `dep` tool.

* github.com/packethost/packngo

Go package used to talk to the Packet API

* gopkg.in/yaml.v2 

Used to read YAML configuration files

## Disclaimer

This software is provided without any warranty or support. Use at your own risk and consult the T&Cs of your cloud or hosting provider before running this software.
