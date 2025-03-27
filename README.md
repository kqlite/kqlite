# <img src="https://github.com/kqlite/kqlite/blob/main/kqlite-logo.png" width="65px">*kqlite*&nbsp;
[![CI](https://github.com/kqlite/kqlite/actions/workflows/ci.yml/badge.svg)](https://github.com/kqlite/kqlite/actions/workflows/go.yml) 
[![Go Report Card](https://goreportcard.com/badge/github.com/kqlite/kqlite)](https://goreportcard.com/report/github.com/kqlite/kqlite)

#### Lightweight remote SQLite with high availability and auto failover.<br>

- Replicated SQLite database and remote access over the PostgreSQL wire protocol.
- Auto failover to an active secondary instance and registering back as secondary a former primary.
- Quick and easy setup and high availability configuration with only two DB Nodes.

Works by translating PostgreSQL frontend wire messages into SQLite transactions and converting results back into PostgreSQL response wire messages. 
Many PostgreSQL clients also inspect the pg_catalog to determine system information so ***kqlite*** mirrors this catalog by using an attached in-memory database with virtual tables. 
A rewrite on those system queries is performed to convert them to usable SQLite syntax.


This repo is very much under active development; as such there are no published artifacts at this time.
Interested developers can clone and run locally to try out things as they become available.

### How to Build and Run

This repo uses [Go 1.23 or higher](https://go.dev/dl/).

```sh
git clone https://github.com/kqlite/kqlite.git
```

Build is done via `make`

```sh
General
  help             Display this help.

Development
  docker-build     Build docker image.
  docker-push      Push kqlite image.
  kqlite           Build kqlite binary.
  example          Build example client program.
  fmt              Format source code.
  vet              Run go vet against code.
  vendor           Runs go mod vendor
  tidy             Runs go mod tidy
  test             Run unit tests.
  test-simple      Run unit tests without verbose/debug output.
  test-package     Run unit tests for specific package.
  test-coverage    Display test coverage as html output in the browser.
```

### Running `kqlite`

After running `make` without any arguments, you can run `bin/kqlite --help` to list available options.<br>
Usally `bin/kqlite -data-dir <dir>` is the common way for executing.

## What Works So Far?

This is still a work in progress and is not yet at full feature database engine. Bugs may exist. Please check this list carefully before logging a new issue or assuming an intentional change.

Status overview:

 * Access to a single database from multiple remote connections.
 * Remote access via <b>psql</b>, but some basic commands like `\dt` aren't yet supported.
 * Transaction support in terms of `sqlite`.
 * Currently ***pgx*** through `database/sql` is tested (https://github.com/jackc/pgx/wiki/Getting-started-with-pgx-through-database-sql).
 * A lightweight storage backend for K8s (https://docs.k3s.io/datastore) is proven to work.<br>
  > [!NOTE]
  > Encryption is not available yet,<br>
  > add `sslmode=disable` in the endpoint address ex. `postgres://127.0.0.1:5432/kine?sslmode=disable`.
    
