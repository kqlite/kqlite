# <img src="https://github.com/kqlite/kqlite/blob/main/kqlite-logo.png" width="65px">*kqlite*&nbsp;
[![CI](https://github.com/kqlite/kqlite/actions/workflows/go.yml/badge.svg)](https://github.com/kqlite/kqlite/actions/workflows/go.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/kqlite/kqlite)](https://goreportcard.com/report/github.com/kqlite/kqlite)

#### Lightweight remote SQLite with high availability and automatic failover.<br>

- Replicated SQLite database and remote access over the PostgreSQL wire protocol.
- Automatic failover to an active secondary instance and registering back as secondary a former primary.
- Quick and easy configuration and setup with only two DB Nodes.

Works by translating PostgreSQL frontend wire messages into SQLite transactions and converting results back into PostgreSQL response wire messages. 
Many PostgreSQL clients also inspect the pg_catalog to determine system information so ***kqlite*** mirrors this catalog by using an attached in-memory database with virtual tables. 
A rewrite on those system queries is performed to convert them to usable SQLite syntax.


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
