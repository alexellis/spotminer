spotminer
--------------------------

## What is this?

`spotminer` automates the [Packet.net](https://www.packet.net) [spot market](https://help.packet.net/technical/deployment-options/spot-market) and a cryptocurrency miner so that you can lower the costs of mining in the cloud and access bare metal performance.

> Packet's [Spot Market](https://help.packet.net/technical/deployment-options/spot-market) allows users to bid on spare server capacity at reduced rates. 

Features:

* Set a budget (max price) for each host
* Automatically sets the closest stratum server
* Places the minimum bid for the host
* Uses my [mine-with-docker project](https://github.com/alexellis/mine-with-docker) and `cpuminer`
* Can be run on a timer - i.e. every 5 minutes to ensure reclaimed hosts are replaced
* Easy configuration in YAML
* Docker Swarm used as an init process to keep the miner running if it crashes
* Configure one of the supported algorithms and Stratum of port i.e. hodl or cryptonight
* Atom hosts are supported through a separate Docker image

> Note: See disclaimer and check T&Cs with any cloud provider before embarking on mining.

## Q&A

* [Watch my video on why you shouldn't mine Bitcoin](https://www.youtube.com/watch?v=Apg8glATeto)

If you have additional questions or you want to try mining but don't want to use this example then consult my [mine-with-docker project](https://github.com/alexellis/mine-with-docker). You can also rebuild your own Docker image/binaries from source.

## Installation

* Install Go 1.9

* Run go install

```
go install github.com/alexellis/spotminer
```

This installs spotminer into your `$GOPATH/bin` directory, so update your `$PATH` variable if necessary. `$GOPATH` is normally set to `$HOME/go`.


The config file is read from `config.yml`, so copy `config.example.yml` as a template and fill in your [Packet API key and project ID](https://www.packet.net/developers/api/). Set the `CONFIG_FILE` environment variable for a different filename or path. You must also configure the bitcoin wallet address for your mining here.

```yaml
packet:
  project_id: ""
  api_key: ""
preferences:
  max_spot_instances: 6
  max_price: 0.15
  mine_algo: "cryptonight"
  port: 3355
  bitcoin_wallet: wallet_address
```

For mining Hodl use `mine_algo: hodl` and `port: 3352`

## Packages:

Dependencies are managed through the `dep` tool.

* github.com/packethost/packngo

Go package used to talk to the Packet API

* gopkg.in/yaml.v2 

Used to read YAML configuration files

## Disclaimer

This software is provided without any warranty or support. Use at your own risk and consult the T&Cs of your cloud or hosting provider before running this software.
