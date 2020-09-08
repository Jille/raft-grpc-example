# raft-grpc-example

This is some example code for how to use [Hashicorp's Raft implementation](https://github.com/hashicorp/raft) with gRPC.

## Start your own cluster

```shell
$ mkdir /tmp/my-raft-cluster
$ mkdir /tmp/my-raft-cluster/node{A,B,C}
$ ./raft-grpc-example --raft_id=nodeA --address=localhost:50051 --raft_data_dir /tmp/my-raft-cluster --raft_bootstrap
$ ./raft-grpc-example --raft_id=nodeB --address=localhost:50052 --raft_data_dir /tmp/my-raft-cluster
$ ./raft-grpc-example --raft_id=nodeC --address=localhost:50053 --raft_data_dir /tmp/my-raft-cluster
$ go get github.com/Jille/raftadmin
$ raftadmin localhost:50051 add_voter nodeB localhost:50052 0
$ raftadmin localhost:50051 add_voter nodeC localhost:50053 0
```
