package httpd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

// Cluster is the interface Raft-backed leadership election.
type Cluster interface {
	// Join joins the node, identitifed by nodeID and reachable at addr, to the cluster.
	Join(nodeID string, addr string) error

	// Returns whether or not the current node is the leader
	Leader() bool

	// NodeID identifies the Raft nodeID of the currently running process.
	NodeID() string
}

// Service provides HTTP service.
type Service struct {
	addr string
	ln   net.Listener

	cluster Cluster
}

// New returns an uninitialized HTTP service.
func New(addr string, cluster Cluster) *Service {
	return &Service{
		addr:    addr,
		cluster: cluster,
	}
}

// Start starts the service.
func (s *Service) Start() error {
	server := http.Server{
		Handler: s,
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.ln = ln

	http.Handle("/", s)

	go func() {
		err := server.Serve(s.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	return nil
}

// Close closes the service.
func (s *Service) Close() {
	s.ln.Close()
	return
}

// ServeHTTP allows Service to serve HTTP requests.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/join" {
		s.handleJoin(w, r)
		return
	}

	if s.cluster.Leader() {
		id := s.cluster.NodeID()
		io.WriteString(w, fmt.Sprintf("I am server %s.\n", id))
	} else {
		io.WriteString(w, "I am not the leader.\n")
	}
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(m) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr, ok := m["addr"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeID, ok := m["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := s.cluster.Join(nodeID, remoteAddr); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// Addr returns the address on which the Service is listening
func (s *Service) Addr() net.Addr {
	return s.ln.Addr()
}
