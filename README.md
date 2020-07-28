# A Software gNB for free5GC

The gNB function was built on the model of the other free5GC CN functions using all the pattern and helper class defined by the free5GC team.

It ensures a seamless and immediate integration into free5GC without requiring any other dependencies.

The build and exection process is therefore the same as for the free5GC CN functions. 

## Installation

Build the function using 

``` bash
cd ~/free5gc
go build -o bin/gnb -x src/gnb/ran.go
```

Execute the function with the following command

``` bash
cd bin/gnb
./gnb
```

## Configuration

The gNB `rancfg.cfg` configuration file is in the traditionnal `free5gc/config` folder.

``` yaml
info:
  version: 1.0.0
  description: "5G-RAN initial local configuration"

configuration:
  ranName: RAN-1
  ueSubnet: "60.60.0.0/24"
  ue:

    - SUPI: imsi-2089300007487

      indentifier: 1
      ipv4: 60.60.0.10

    - SUPI: imsi-2089300007486

      indentifier: 2
      ipv4: 60.60.0.20
  sbi:
    scheme: http
    ipv4Addr: 127.0.0.1
    port: 32000
  networkName:
    full: free5GC
    short: free
```

**TODO ARRAY TO DESCRIBRE EACH LINE OF THE YAML FILE**

## Service Exposed

The goal of this gNB is to generate traffic for multiple UE, it exposes services

| Service                   | Url                                    |
|---------------------------|----------------------------------------|
| Stream MPEG-DASH manifest | /run/stream_dash/:identifier/:manifest |
| Ping a Device             | /run/ping_device/:identifier/:device   |

**TODO UPDATE ARRAY AND ADD EXAMPLE TOPOLOGY**
