# DDFS - Distributed Deduplicating File System

PoC "filesystem" written for my master's thesis. This project is far from being able to be called complete. I have no drive to ever complete this but hopefully some parts of it may actually be useful for someone.

## Architecture

Server part consists of 3 components:
* Monitor server - basically frontend for etcd. Serves as configuration store and discovery server for nodes and volumes.
* Index server - stores volume indices. Each index is basically list of pointers to data blocks.
* Block server - stores actual volume data.

Client side of things is very rudimentary. There's a simple CLI client for accessing volumes.

Data blocks depending on configuration can either be fixed or variable sized. For variable sized blocks, Rabin Karp is used for chunking.

Pointers to data blocks are SHA256 hash of the data. This basically means there's a chance of collision, which is completely unsolved issue in this implementation - and probably the biggest flaw.

Data stored and served by index and block servers is sharded between all nodes. Dynamic node configuration is not implemented, but as far as architecture go - possible.

gRPC is used for communication between all parts of the system.

## Building

```
dep ensure
make
```

