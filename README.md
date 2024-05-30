# <img src="https://avatars.githubusercontent.com/u/166529745?s=400&u=41ea395203fc77b863f14d1079fb5f9bd9bdaadb&v=4" width="100px">*kqlite*&nbsp;
[![CI](https://github.com/kqlite/kqlite/actions/workflows/go.yml/badge.svg)](https://github.com/kqlite/kqlite/actions/workflows/go.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/kqlite/kqlite)](https://goreportcard.com/report/github.com/kqlite/kqlite)

#### High availability sqlite databases over PostgreSQL connection.<br>

- Lightweight replicated sqlite database for Edge and IoT devices over the PostgreSQL wire protocol.
- High availability K8s (Edge) clusters can consists of only two nodes when using *kqlite* as a backend storage.
- Automatic failover to an active secondary instance and registering back as secondary a former primary.

## Table of contents
* [Architecture]()
* [Installation]()
    * [Running in Docker]()
    * [Running in Kubernetes]()
    * [Running as a Systemd service]()
    * [Running as a Windows Service]()
* [Quick Start]()
* [Configuration]()
* [Development]()
   * [Setup test environment]()
   * [Building]()
   * [Running]()
