# OS-EDGEC

## Description

Open Steering edge controller is written in Go Lang, and can be run or built for many operating systems.

A Publish/Subscribe system, where sensors publish data on unique topics to the subscribing edge controller, which in turn publishes control commands to relays, motors, actuators and more.

## System Setup

Edge controller relies on MQTT for telemetry. You may use your MQTT server of choice. We suggest [EMQX](https://www.emqx.io/) with username/password, jailed users, and ssl/tls. [EMQX](https://www.emqx.io/) is open source and free for on premise installation. Edge controller will also work with many MQTT free to use services online. We do however suggest you run your services on premise.

Edge controller also relies on InfluxDB for time based data store.

Edit `config.go` to match your mqtt and influxdb server details.

Setup and deploy sensors or control. You can find SDI-12, RS485, and power control examples in our repository. You should configure each sensor mcu with a unique mqtt/device id, and a unique topic or "zone name" for organizational purposes.

Create a ca.pem file in the same directory as your edgec_server.go (or binary)

run or build the edge controller

```bash
go run .
```

```bash
go build
```

## Using Docker

Prerequisites

- Docker [installed](https://docs.docker.com/engine/install/)

### Building

```bash

docker build -t os-edgec:latest .

```

### Running

Please make sure that MQTT and Influxdb variables are set properly in ./config-docker if you running both servers locally.

[Build](#building) service in docker, then run it.

```bash

docker run -p 8081:8081 os-edgec:latest

```

### Running using docker compose

`docker-compose.yml` contains InfluxDB, EMQX and edge controller.

```bash
docker compose build
```

```bash
docker compose up -d
```

### Check logs

```bash

docker compose logs -f controller

```

### Cleaning up

```bash
docker compose down
```

To free up docker rubbish (optional).

```bash
docker system prune
```

```bash
docker image prune -a -f
```

Example of `ca.pem`:

```text
-----BEGIN CERTIFICATE-----
MIIM9DCCC9wCAQAwXDEVMBMGA1UEAwwMKi5nb29nbGUuY29tMQswCQYDVQQGEwJD
QTEiMCAGA1UECAwZTmV3Zm91bmRsYW5kIGFuZCBMYWJyYWRvcjESMBAGA1UEBwwJ
U3QuIEpvaG5zMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvKILbKB+
9qUzK3Vr98hnG4T4UVZGKh1Q9jkrBOU9M2vDeWTMwQt+Eu9HMUe4E2vj+iBxSDes
6pi6LwlUDWqHBckj/SOzMQblUD3r+cyaP30hONfGxdynJwp25PGTuYUgk9fHm5GA
I9BHhM6f+Uh+tTsQESJgAmGCUJj9FEToM2/CRTRG6SDb8SR3bdOWzlun3S9y35CW
ndZ7R3wY0jgOzYkMnsRp0q+p+FaVmMpPSCUKirZ0GgUi53HtX+dhFNFMSwTJ0OEE
AzvHzRh7hQLf9to7v+fCa5KeNq4QP5GtGH8vTx5RHsoIA82kwDFSVOY01L8xJ99H
b03A/qlcG6diWQIDAQABoIIKUTCCCk0GCSqGSIb3DQEJDjGCCj4wggo6MA4GA1Ud
DwEB/wQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMIIK
AwYDVR0RBIIJ+jCCCfaCDCouZ29vZ2xlLmNvbYIWKi5hcHBlbmdpbmUuZ29vZ2xl
LmNvbYIJKi5iZG4uZGV2ghUqLm9yaWdpbi10ZXN0LmJkbi5kZXaCEiouY2xvdWQu
Z29vZ2xlLmNvbYIYKi5jcm93ZHNvdXJjZS5nb29nbGUuY29tghgqLmRhdGFjb21w
zAlloJswHfU=
-----END CERTIFICATE-----
```
