package node

import (
	"os"
	"peerInfoCollect/rpc"
	"sync"
	"peerInfoCollect/event"
	"peerInfoCollect/accounts"
	"peerInfoCollect/log"
	"github.com/prometheus/tsdb/fileutil"
	"peerInfoCollect/p2p"
	"path/filepath"
	"strings"
	"errors"
)

// Node is a container on which services can be registered.
type Node struct {
	eventmux      *event.TypeMux
	config        *Config
	accman        *accounts.Manager
	log           log.Logger
	keyDir        string            // key store directory
	keyDirTemp    bool              // If true, key directory will be removed by Stop
	dirLock       fileutil.Releaser // prevents concurrent use of instance directory
	stop          chan struct{}     // Channel to wait for termination notifications
	server        *p2p.Server       // Currently running P2P networking layer
	startStopLock sync.Mutex        // Start/Stop are protected by an additional lock
	state         int               // Tracks state of node lifecycle

	lock          sync.Mutex
	lifecycles    []Lifecycle // All registered backends, services, and auxiliary services that have a lifecycle
	rpcAPIs       []rpc.API   // List of APIs currently provided by the node

	//http          *httpServer //
	//httpAuth      *httpServer //
	inprocHandler *rpc.Server // In-process RPC request handler to process the API requests

	//databases map[*closeTrackingDB]struct{} // All open databases
}

const (
	initializingState = iota
	runningState
	closedState
)

// New creates a new P2P node, ready for protocol registration.
func New(conf *Config) (*Node, error) {
	// Copy config and resolve the datadir so future changes to the current
	// working directory don't affect the node.
	confCopy := *conf
	conf = &confCopy
	if conf.DataDir != "" {
		absdatadir, err := filepath.Abs(conf.DataDir)
		if err != nil {
			return nil, err
		}
		conf.DataDir = absdatadir
	}
	if conf.Logger == nil {
		conf.Logger = log.New()
	}

	// Ensure that the instance name doesn't cause weird conflicts with
	// other files in the data directory.
	if strings.ContainsAny(conf.Name, `/\`) {
		return nil, errors.New(`Config.Name must not contain '/' or '\'`)
	}
	if conf.Name == datadirDefaultKeyStore {
		return nil, errors.New(`Config.Name cannot be "` + datadirDefaultKeyStore + `"`)
	}
	if strings.HasSuffix(conf.Name, ".ipc") {
		return nil, errors.New(`Config.Name cannot end in ".ipc"`)
	}

	node := &Node{
		config:        conf,
		inprocHandler: rpc.NewServer(),
		eventmux:      new(event.TypeMux),
		log:           conf.Logger,
		stop:          make(chan struct{}),
		server:        &p2p.Server{Config: conf.P2P},
		//databases:     make(map[*closeTrackingDB]struct{}),
	}

	// Register built-in APIs.
	node.rpcAPIs = append(node.rpcAPIs, node.apis()...)

	// Acquire the instance directory lock.
	if err := node.openDataDir(); err != nil {
		return nil, err
	}
	keyDir, isEphem, err := getKeyStoreDir(conf)
	if err != nil {
		return nil, err
	}
	node.keyDir = keyDir
	node.keyDirTemp = isEphem
	// Creates an empty AccountManager with no backends. Callers (e.g. cmd/geth)
	// are required to add the backends later on.
	node.accman = accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: conf.InsecureUnlockAllowed})

	// Initialize the p2p server. This creates the node key and discovery databases.
	node.server.Config.PrivateKey = node.config.NodeKey()
	//node.server.Config.Name = node.config.NodeName()
	//node.server.Config.Logger = node.log
	//if node.server.Config.StaticNodes == nil {
	//	node.server.Config.StaticNodes = node.config.StaticNodes()
	//}
	//if node.server.Config.TrustedNodes == nil {
	//	node.server.Config.TrustedNodes = node.config.TrustedNodes()
	//}
	//if node.server.Config.NodeDatabase == "" {
	//	node.server.Config.NodeDatabase = node.config.NodeDB()
	//}
	//
	//// Check HTTP prefixes are valid.
	//if err := validatePrefix("HTTP", conf.HTTPPathPrefix); err != nil {
	//	return nil, err
	//}
	//
	//// Configure RPC servers.
	//node.http = newHTTPServer(node.log, conf.HTTPTimeouts)
	//node.httpAuth = newHTTPServer(node.log, conf.HTTPTimeouts)

	return node, nil
}

// AccountManager retrieves the account manager used by the protocol stack.
func (n *Node) AccountManager() *accounts.Manager {
	return n.accman
}

// KeyStoreDir retrieves the key directory
func (n *Node) KeyStoreDir() string {
	return n.keyDir
}

// Config returns the configuration of node.
func (n *Node) Config() *Config {
	return n.config
}

func (n *Node) openDataDir() error {
	if n.config.DataDir == "" {
		return nil // ephemeral
	}

	instdir := filepath.Join(n.config.DataDir, n.config.name())
	if err := os.MkdirAll(instdir, 0700); err != nil {
		return err
	}
	// Lock the instance directory to prevent concurrent use by another instance as well as
	// accidental use of the instance directory as a database.
	release, _, err := fileutil.Flock(filepath.Join(instdir, "LOCK"))
	if err != nil {
		return convertFileLockError(err)
	}
	n.dirLock = release
	return nil
}