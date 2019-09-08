package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/stntngo/leader-election/cluster"
	"github.com/stntngo/leader-election/http"
)

const (
	DefaultHTTPAddr = "127.0.0.1:8080"
	DefaultRaftAddr = "127.0.0.1:9090"
)

// Command line parameters
var httpAddr string
var raftAddr string
var joinAddr string
var nodeID string

func init() {
	flag.StringVar(&httpAddr, "http", DefaultHTTPAddr, "Set the HTTP Listening Address and Port")
	flag.StringVar(&raftAddr, "raft", DefaultRaftAddr, "Set the Raft Listening Address and Port")
	flag.StringVar(&joinAddr, "join", "", "Set the HTTP Address of an existing cluster member to contact to join the cluster")
	flag.StringVar(&nodeID, "id", "", "Node ID")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	// Ensure Raft storage exists.
	raftDir := flag.Arg(0)
	if raftDir == "" {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}
	os.MkdirAll(raftDir, 0700)

	enableSingle := joinAddr == ""

	c := cluster.New()
	c.RaftDir = raftDir
	c.RaftBind = raftAddr
	if err := c.Open(enableSingle, nodeID); err != nil {
		log.Fatalf("Failed to open cluster [%s]", err.Error())
	}

	h := httpd.New(httpAddr, c)
	if err := h.Start(); err != nil {
		log.Fatalf("Failed to start HTTP service [%s]", err.Error())
	}

	if !enableSingle {
		if err := joinCluster(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("Failed to join cluster at %s [%s]", joinAddr, err.Error())
		}
	}

	log.Printf("Node %s started successfully\n", nodeID)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Printf("Node %s exiting...", nodeID)
}

func joinCluster(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
