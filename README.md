# A Software gNB for free5GC

The gNB function was built on the model of the other free5GC CN functions using all the pattern and helper class defined by the free5GC team.

It ensures a seamless and immediate integration into free5GC without requiring any other dependencies.

The build and exection process is therefore the same as for the free5GC CN functions. 

## Installation

First you need to clone this forked version of free5GC using:

``` bash
git clone https://github.com/Srajdax/free5gc
```

Then

Build the function using 

``` bash
cd ~/free5gc
go build -o bin/gnb -x src/gnb/gnb.go
```

Execute the function with the following command

``` bash
cd bin/gnb
./gnb
```

/!\ MongoDB needs to be running before launching the gNB

## Configuration

The gNB `gnbcfg.cfg` configuration file is located in `free5gc/config` folder. A sample is also present into `gnb/config` folder.

``` yaml
info:
  version: 1.0.0
  description: "5G gNB initial local configuration"

configuration:
  ranName: RAN-1
  amfInterface:
    ipv4Addr: "127.0.0.1"
    port: 38412
  upfInterface:
    ipv4Addr: "10.200.200.102"
    port: 2152
  ngranInterface:
    ipv4Addr: "127.0.0.1"
    port: 9487
  gtpInterface:
    ipv4Addr: "10.200.200.1"
    port: 2152
  ueSubnet: "60.60.0.0/24"
  ue:

    - SUPI: imsi-2089300007487

      ipv4: 60.60.0.1

    - SUPI: imsi-2089300007486

      ipv4: 60.60.0.2
  sbi:
    scheme: http
    ipv4Addr: 127.0.0.1
    port: 32000
  networkName:
    full: free5GC
    short: free
```

The following Diagram gives represents configuration file above

![diagram_gNB](https://user-images.githubusercontent.com/41422704/88692144-07d6a700-d0fe-11ea-836d-56df98ffa93a.png)

## Service Exposed by REST Interface

The gNB exposes two command interfaces

| Service                   | Url                                    | Status |
|---------------------------|----------------------------------------|--------|
| Stream MPEG-DASH manifest | /run/stream_dash/:identifier/:manifest |On going|
| Ping a Device             | /run/ping_device/:identifier/:device   |Implemented|

## Usage

With simple tools such as curl, you can control the gNB using:

``` bash
curl -d {} http://localhost:32000/run/ping_device/0/60.60.0.101
```

## Limitation

For the moment, only one PDU session could be established per UE to match with the UE IP configuration

## TODO

* [ ] Clean the code
* [ ] Implement the gNB NF using Docker
* [ ] Implement the DASH function
* [ ] Proper implement of the tests
* [ ] Add HTTPS option for REST Interface
