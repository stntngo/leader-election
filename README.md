# leader-election

This repo plays around with using the Raft Consensus Protocol for leader election.
It implements a very simple Raft cluster and HTTP server.

To install / build `leader-election` run the following after cloning the repo
```
$ go get github.com/hashicorp/raft
$ go build -o raft main.go
```

To start a cluster run the following
```
$ ./raft -id node0 -http 127.0.0.1:8080 -raft 127.0.0.1:9090 ~/raft0
```

This will spin up an HTTP server that listens on port 8080, a raft TCP server that opens on 9090 and will use `~/raft0` to store the raft logs.

To create new nodes and join them to the cluster use the following pattern:
```
$ ./raft -id {unique name} -http 127.0.0.1:{unique port} -raft 127.0.0.1:{unique port} -join {http address of running node} {unique path}
```

E.g., The following will create a three node cluster that joins the bootstrapepd node we started above running on `:8080`.
```
$ ./raft -id node1 -http 127.0.0.1:8081 -raft 127.0.0.1:9091 -join 127.0.0.1:8080 ~/raft1
$ ./raft -id node2 -http 127.0.0.1:8082 -raft 127.0.0.1:9092 -join 127.0.0.1:8080 ~/raft2
```

When making an HTTP request to any of the nodes they will reply in one of two ways. If the node you are making a request to is the leader it will reply `I am server {node id}.` If the node you are making a request to is not the leader then it will reply `I am not the leader.`.

```
niels ~  curl -XGET localhost:8080
I am server node0.
niels ~  curl -XGET localhost:8081
I am not the leader.
niels ~  curl -XGET localhost:8082
I am not the leader.

# node0's process is killed. Leaving only node1 running on localhost:8081 and node2 running on localhost:8082

niels ~  curl -XGET localhost:8082
I am not the leader.
niels ~  curl -XGET localhost:8081
I am server node1.
niels ~
```
