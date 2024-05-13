# <img src="https://github.com/kqlite/kqlite/blob/main/kqlite-logo.png" width="65px">*kqlite*&nbsp; [![CI](https://github.com/kqlite/kqlite/actions/workflows/go.yml/badge.svg)](https://github.com/kqlite/kqlite/actions/workflows/go.yml)
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
