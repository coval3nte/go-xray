# go-xray

Easy [XTLS] [Xray Core](https://github.com/XTLS/Xray-core) **MacOS** Client

* CLI first
* bundled [tun2socks](https://github.com/xjasonlyu/tun2socks)
    * improved network stack
    * `tun` support, tunnels all device connections into the Xray Instance
* tiny size
* CGO-less


## Installation

### Prerequisites

* make
* go 1.25

### Building

`make go-xray`

## Usage

### Prerequisites

1. Xray Server (not covered in this guide, follow their docs)
2. Config File (JSON notation)

### Running

```bash
go-xray \
  -remote-address ${REMOTE_ADDRESS} \ # IP Address of the remote server
  -config-file ${CONFIG_FILE_PATH} # ./config.json
```
