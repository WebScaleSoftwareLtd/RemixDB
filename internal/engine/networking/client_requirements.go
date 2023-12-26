// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package networking

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
	"remixdb.io/internal/engine"
)

// ClientRequirements is used to define a client which is required for peer to peer networking.
type ClientRequirements interface {
	// Engine is used to get the local engine.
	Engine() engine.Engine

	// GetNetworkingData is used to get the networking data.
	GetNetworkingData() NetData

	// WriteNetworkingData is used to write the networking data to the database.
	WriteNetworkingData(NetData) error

	// GetHostname is used to get the hostname for this node.
	GetHostname() string
}

// NetData is used to get the networking data.
type NetData struct {
	// HostID is the host ID for this node.
	HostID string

	// JoinKey is the join key for this node.
	JoinKey string

	// KnownHosts is a list of known hosts for this node.
	KnownHosts []string
}

type clientRequirements struct {
	engine engine.Engine

	netdataLock sync.RWMutex
	netdata     NetData
	path        string
}

func (r *clientRequirements) Engine() engine.Engine {
	return r.engine
}

func (r *clientRequirements) GetNetworkingData() NetData {
	r.netdataLock.RLock()
	defer r.netdataLock.RUnlock()
	return r.netdata
}

func (r *clientRequirements) WriteNetworkingData(nd NetData) error {
	r.netdataLock.Lock()
	defer r.netdataLock.Unlock()

	// Marshal the netdata.
	b, err := msgpack.Marshal(nd)
	if err != nil {
		return err
	}

	// Make sure path is a folder.
	if err := os.MkdirAll(r.path, 0755); err != nil {
		return err
	}

	// Write the netdata.S file.
	netdataStagedPath := filepath.Join(r.path, "netdata.S")
	if err := os.WriteFile(netdataStagedPath, b, 0644); err != nil {
		return err
	}

	// Write the commit file.
	if err := os.WriteFile(filepath.Join(r.path, "netdata_commit"), []byte{}, 0644); err != nil {
		return err
	}

	// Delete the netdata file.
	netdataPath := filepath.Join(r.path, "netdata")
	if err := os.Remove(netdataPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	// Rename the netdata.S file to netdata.
	if err := os.Rename(netdataStagedPath, netdataPath); err != nil {
		return err
	}

	// Set netdata within the object.
	r.netdata = nd

	// Return no errors.
	return nil
}

func (r *clientRequirements) GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	return hostname
}

var _ ClientRequirements = &clientRequirements{}

// NewClientRequirements is used to create a new client requirements object.
// If path is unset, the default data path will be used.
func NewClientRequirements(path string, engine engine.Engine) (ClientRequirements, error) {
	// Handles the default path.
	if path == "" {
		path = os.Getenv("REMIXDB_PATH")
		if path == "" {
			homedir, err := os.UserHomeDir()
			if err != nil {
				panic(err)
			}
			path = filepath.Join(homedir, ".remixdb", "data")
		}
	}

	// Handle if the commit file exists.
	commitFile := filepath.Join(path, "netdata_commit")
	if _, err := os.Stat(commitFile); err == nil {
		// Delete netdata.
		if err := os.Remove(filepath.Join(path, "netdata")); err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
		}

		// Rename netdata.S to netdata.
		if err := os.Rename(filepath.Join(path, "netdata.S"), filepath.Join(path, "netdata")); err != nil {
			return nil, err
		}

		// Delete the commit file.
		if err := os.Remove(commitFile); err != nil {
			return nil, err
		}
	}

	// Read the netdata file.
	b, err := os.ReadFile(filepath.Join(path, "netdata"))
	var nd NetData
	doWrite := false
	if err == nil {
		// Unmarshal the netdata file.
		err = msgpack.Unmarshal(b, &nd)
		if err != nil {
			return nil, err
		}
	} else {
		if !os.IsNotExist(err) {
			// Handle error cases first.
			return nil, err
		}

		// Generate 32 random bytes that are a-z, A-Z, 0-9.
		joinKey := randString(32)

		// Generate a host ID.
		hostID := randString(32)

		// Create the netdata structure.
		nd = NetData{
			HostID:     hostID,
			JoinKey:    joinKey,
			KnownHosts: []string{},
		}

		// Do a write later.
		doWrite = true
	}

	// Make a new client requirements object.
	cr := &clientRequirements{engine: engine, netdata: nd}

	// If we need to write, do it now.
	if doWrite {
		if err := cr.WriteNetworkingData(nd); err != nil {
			return nil, err
		}
	}

	// Return the object.
	return cr, nil
}
