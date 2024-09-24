# <img src="https://github.com/kqlite/kqlite/blob/main/kqlite-logo.png" width="65px">*kqlite*&nbsp;
[![CI](https://github.com/kqlite/kqlite/actions/workflows/go.yml/badge.svg)](https://github.com/kqlite/kqlite/actions/workflows/go.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/kqlite/kqlite)](https://goreportcard.com/report/github.com/kqlite/kqlite)

#### Database engine with high availability and automatic failover on top of SQLite.<br>

- Lightweight replicated SQLite database over the PostgreSQL wire protocol.
- Automatic failover to an active secondary instance and registering back as secondary a former primary.
- Quick and easy configuration and setup with only two nodes.

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
