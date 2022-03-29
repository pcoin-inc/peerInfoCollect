package ethconfig

import (
	"math/big"
	"peerInfoCollect/common"
	"peerInfoCollect/core"
	"peerInfoCollect/params"
)

// Config contains configuration options for of the ETH and LES protocols.
type Config struct {
	// The genesis block, which is inserted if the database is empty.
	// If nil, the Ethereum main net block is used.
	Genesis *core.Genesis `toml:",omitempty"`

	// Protocol options
	NetworkId uint64 // Network ID to use for selecting peers to connect to
	//SyncMode  downloader.SyncMode

	// This can be set to list of enrtree:// URLs which will be queried for
	// for nodes to connect to.
	EthDiscoveryURLs  []string
	SnapDiscoveryURLs []string

	//NoPruning  bool // Whether to disable pruning and flush everything to disk
	//NoPrefetch bool // Whether to disable prefetching and only load state on demand

	//TxLookupLimit uint64 `toml:",omitempty"` // The maximum number of blocks from head whose tx indices are reserved.

	// PeerRequiredBlocks is a set of block number -> hash mappings which must be in the
	// canonical chain of all remote peers. Setting the option makes geth verify the
	// presence of these blocks for every new peer connection.
	PeerRequiredBlocks map[uint64]common.Hash `toml:"-"`

	//// Light client options
	//LightServ          int  `toml:",omitempty"` // Maximum percentage of time allowed for serving LES requests
	//LightIngress       int  `toml:",omitempty"` // Incoming bandwidth limit for light servers
	//LightEgress        int  `toml:",omitempty"` // Outgoing bandwidth limit for light servers
	//LightPeers         int  `toml:",omitempty"` // Maximum number of LES client peers
	//LightNoPrune       bool `toml:",omitempty"` // Whether to disable light chain pruning
	//LightNoSyncServe   bool `toml:",omitempty"` // Whether to serve light clients before syncing
	//SyncFromCheckpoint bool `toml:",omitempty"` // Whether to sync the header chain from the configured checkpoint

	//// Ultra Light client options
	//UltraLightServers      []string `toml:",omitempty"` // List of trusted ultra light servers
	//UltraLightFraction     int      `toml:",omitempty"` // Percentage of trusted servers to accept an announcement
	//UltraLightOnlyAnnounce bool     `toml:",omitempty"` // Whether to only announce headers, or also serve them

	//// Database options
	//SkipBcVersionCheck bool `toml:"-"`
	//DatabaseHandles    int  `toml:"-"`
	//DatabaseCache      int
	//DatabaseFreezer    string

	//TrieCleanCache          int
	//TrieCleanCacheJournal   string        `toml:",omitempty"` // Disk journal directory for trie cache to survive node restarts
	//TrieCleanCacheRejournal time.Duration `toml:",omitempty"` // Time interval to regenerate the journal for clean cache
	//TrieDirtyCache          int
	//TrieTimeout             time.Duration
	//SnapshotCache           int
	//Preimages               bool


	// Ethash options
	//Ethash ethash.Config

	//// Transaction pool options
	//TxPool core.TxPoolConfig
	//
	//// Gas Price Oracle options
	//GPO gasprice.Config

	// Enables tracking of SHA3 preimages in the VM
	EnablePreimageRecording bool

	// Miscellaneous options
	DocRoot string `toml:"-"`

	//// RPCGasCap is the global gas cap for eth-call variants.
	//RPCGasCap uint64

	//// RPCEVMTimeout is the global timeout for eth-call.
	//RPCEVMTimeout time.Duration
	//
	//// RPCTxFeeCap is the global transaction fee(price * gaslimit) cap for
	//// send-transction variants. The unit is ether.
	//RPCTxFeeCap float64

	// Checkpoint is a hardcoded checkpoint which can be nil.
	Checkpoint *params.TrustedCheckpoint `toml:",omitempty"`

	// CheckpointOracle is the configuration for checkpoint oracle.
	CheckpointOracle *params.CheckpointOracleConfig `toml:",omitempty"`

	// Arrow Glacier block override (TODO: remove after the fork)
	OverrideArrowGlacier *big.Int `toml:",omitempty"`

	// OverrideTerminalTotalDifficulty (TODO: remove after the fork)
	OverrideTerminalTotalDifficulty *big.Int `toml:",omitempty"`
}