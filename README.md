# raft-grpc-example

This is some example code for how to use [Hashicorp's Raft implementation](https://github.com/hashicorp/raft) with gRPC.

## Start your own cluster

```shell
$ mkdir /tmp/my-raft-cluster
$ mkdir /tmp/my-raft-cluster/node{A,B,C}
$ ./raft-grpc-example --raft_bootstrap --raft_id=nodeA --address=localhost:50051 --raft_data_dir /tmp/my-raft-cluster
$ ./raft-grpc-example --raft_id=nodeB --address=localhost:50052 --raft_data_dir /tmp/my-raft-cluster
$ ./raft-grpc-example --raft_id=nodeC --address=localhost:50053 --raft_data_dir /tmp/my-raft-cluster
$ go install github.com/Jille/raftadmin/cmd/raftadmin@latest
$ raftadmin localhost:50051 add_voter nodeB localhost:50052 0
$ raftadmin --leader multi:///localhost:50051,localhost:50052 add_voter nodeC localhost:50053 0
$ go run cmd/hammer/hammer.go &
$ raftadmin --leader multi:///localhost:50051,localhost:50052,localhost:50053 leadership_transfer
$ wait
```

You start up three nodes, and bootstrap one of them. Then you tell the bootstrapped node where to find peers. Those peers sync up to the state of the bootstrapped node and become members of the cluster. Once your cluster is running, you never need to pass `--raft_bootstrap` again.

[raftadmin](https://github.com/Jille/raftadmin) is used to communicate with the cluster and add the other nodes.

This example uses [Jille/raft-grpc-transport](https://github.com/Jille/raft-grpc-transport) to communicate between nodes using gRPC.

This example uses [Jille/raft-grpc-leader-rpc](https://github.com/Jille/raft-grpc-leader-rpc) to send RPCs to the leader.

Hammer is a client that connects to your raft cluster and sends a bunch of requests. Trigger some leadership failovers to show that it's unaffected.

## What's what

Raft uses logs to synchronize changes. Every change submitted to a Raft cluster is a log entry, which gets stored and replicated to the followers in the cluster. In this example, we use [raft-boltdb](https://github.com/hashicorp/raft-boltdb) to store these logs.
Once in a while Raft decides the logs have grown too large, and makes a snapshot. Your code is asked to write out its state. That state captures all previous logs. Now Raft can delete all the old logs and just use the snapshot. These snapshots are stored using the FileSnapshotStore, which means they'll just be files in your disk.
Raft also needs a way to talk to other nodes, that's called a Transport. This example uses [Jille/raft-grpc-transport](https://github.com/Jille/raft-grpc-transport) to communicate between nodes using gRPC.

You can see all this happening in `NewRaft()` in `main.go`.

## Your application

See `application.go`. You'll need to implement a `raft.FSM`, and you probably want a gRPC RPC interface.
