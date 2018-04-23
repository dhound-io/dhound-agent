# dhound-agent

A cross-platform tool for collecting security events information on Servers (Windows, Free-BSD, Ubuntu, Raspberry (Debian) and other Linux versions) for further processing and aggregating in [dhound.io](https://dhound.io) (Lightweight Intrusion Detection System for Internet facing systems).

Some additional information about dhound-agent configuration can be found [here](https://knowledge.dhound.io/how-to-use-dhound).

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

What things you need to install the software and how to install them.

1. install go v1.9.4 and higher - https://golang.org/doc/install
```
wget https://dl.google.com/go/go1.9.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.9.4.linux-amd64.tar.gz
```

2. Set into ~/.profile
```
export PATH=$PATH:/usr/local/go/bin
export GOROOT=/usr/local/go
```

3. Install dependencies on Linux
```
sudo apt-get install libpcap0.8-dev
```

4. IDE: [LiteIDE](http://sourceforge.net/projects/liteide/files/)

Build options for LIteIDE (Build Confiration -> TARGETARGS):
```
-config-dir config -log-file dhound.log -verbose
```
For profiling use additional parameter (profile available by address http://localhost:5061/debug/pprof/)
```
-pprof :5061
```

5. Configure iptables to track output traffic ip addresses
Linux 
```
iptables -I OUTPUT -m state -p tcp --state NEW  --syn -j LOG --log-prefix "OUT TCP: " --log-level 4 -m hashlimit --hashlimit-upto 1/hour --hashlimit-burst 1 --hashlimit-mode dstip --hashlimit-name dhoundtcpout --hashlimit-htable-expire 3600000 --hashlimit-htable-size 1000 -m comment --comment "dhound: Log OUT Tcp Connections to syslog"
```

Windows
```
Windows Defender Firewall with Advanced Security -> Properties -> Public Profile -> Customize... -> Log successfull connections = true
```
6. Download go packages
```
go get gopkg.in/yaml.v2
		go get gopkg.in/natefinch/lumberjack.v2
		go get github.com/judwhite/go-svc/svc
		go get github.com/google/gopacket
		go get github.com/google/gopacket/layers
		go get github.com/google/gopacket/pcap
		go get golang.org/x/text/encoding
```

### Build
Build dhound-agent
```
cd <project>
go build
```

### Run

Linux
```
dhound-agent -config-dir config -verbose
```

Windows
```
dhound-agent -config-dir config -verbose
```

## Versioning

Version specified in 2 files:
* Makefile
* version.go

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE.md](LICENSE.md) file for details
