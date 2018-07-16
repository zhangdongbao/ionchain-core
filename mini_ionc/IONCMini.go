package mini_ionc

import (
	"runtime"

	"github.com/ionchain/ionchain-core/common/hexutil"
	"github.com/ionchain/ionchain-core/accounts_ionc"
	core "github.com/ionchain/ionchain-core/core_ionc"
	consensus "github.com/ionchain/ionchain-core/consensus_ionc"
	"github.com/ionchain/ionchain-core/consensus_ionc/ipos"
	miner "github.com/ionchain/ionchain-core/miner_ionc"
	"github.com/ionchain/ionchain-core/ethdb"
	"github.com/ionchain/ionchain-core/common"
	"github.com/ionchain/ionchain-core/log"
	"github.com/ionchain/ionchain-core/core_ionc/vm"
	"github.com/ionchain/ionchain-core/params"
	"github.com/ionchain/ionchain-core/rlp"
	"github.com/ionchain/ionchain-core/event"
	"github.com/ionchain/ionchain-core/node_ionc"
	"github.com/ionchain/ionchain-core/p2p"
	"github.com/ionchain/ionchain-core/rpc"

	"fmt"
	"math/big"
	"sync"
	"github.com/ionchain/ionchain-core/mini_ionc/gasprice"
	ethapi "github.com/ionchain/ionchain-core/internal/ioncapi"
	//"github.com/ionchain/ionchain-core/eth/downloader"
	"github.com/ionchain/ionchain-core/mini_ionc/filters"
	"github.com/ionchain/ionchain-core/core_ionc/bloombits"
)

type IONCMini struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the ethereum
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool     *core.TxPool     //交易池
	blockchain *core.BlockChain //区块链

	// DB interfaces
	// leveldb数据库
	chainDb ethdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine  //共识引擎
	accountManager *accounts.Manager //账户管理

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *EthApiBackend

	miner     *miner.Miner //挖矿
	gasPrice  *big.Int
	etherbase common.Address

	networkId     uint64
	netRPCService *ethapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and etherbase)
}

// New creates a new Ethereum object (including the
// initialisation of the common Ethereum object)
func New(ctx *node.ServiceContext, config *Config) (*IONCMini, error) {

	chainDb, err := CreateDB(ctx, config, "chaindata") // 创建leveldb数据库
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb) // 数据库格式升级
	// 设置创世区块。 如果数据库里面已经有创世区块那么从数据库里面取出(私链)。或者是从代码里面获取默认值。
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	//构建以太坊对象
	eth := &IONCMini{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, config, chainConfig, chainDb), //共识引擎
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		etherbase:      config.Etherbase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks), //创建布隆过滤器
	}

	// 检查数据库里面存储的BlockChainVersion和客户端的BlockChainVersion的版本是否一致
	if !config.SkipBcVersionCheck {
		//数据库中的BlockChainVersion
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run geth upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	// vm虚拟机配置
	vmConfig := vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
	//创建区块链 主链
	eth.blockchain, err = core.NewBlockChain(chainDb, eth.chainConfig, eth.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		eth.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	eth.bloomIndexer.Start(eth.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	//交易池
	eth.txPool = core.NewTxPool(config.TxPool, eth.chainConfig, eth.blockchain)

	// 创建协议管理器
	/*if eth.protocolManager, err = NewProtocolManager(eth.chainConfig, config.SyncMode, config.NetworkId, eth.eventMux, eth.txPool, eth.engine, eth.blockchain, chainDb); err != nil {
		return nil, err
	}*/

	eth.miner = miner.New(eth, eth.chainConfig, eth.EventMux(), eth.engine)
	eth.miner.SetExtra(makeExtraData(config.ExtraData))

	// 对外api 接口
	eth.ApiBackend = &EthApiBackend{eth, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	eth.ApiBackend.gpo = gasprice.NewOracle(eth.ApiBackend, gpoParams)

	return eth, nil
}


func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata 创建默认extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch), //版本号
			"geth",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}


// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (ethdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*ethdb.LDBDatabase); ok {
		db.Meter("eth/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Ethereum service
func CreateConsensusEngine(ctx *node.ServiceContext, config *Config, chainConfig *params.ChainConfig, db ethdb.Database) consensus.Engine {

	return ipos.New(db)

}

/*, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}*/
func (s *IONCMini) APIs() []rpc.API {
	apis := ethapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}


func (s *IONCMini) Etherbase() (eb common.Address, err error) {
	s.lock.RLock()
	etherbase := s.etherbase
	s.lock.RUnlock()

	if etherbase != (common.Address{}) {
		return etherbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			return accounts[0].Address, nil
		}
	}
	return common.Address{}, fmt.Errorf("etherbase address must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *IONCMini) SetEtherbase(etherbase common.Address) {
	self.lock.Lock()
	self.etherbase = etherbase
	self.lock.Unlock()

	self.miner.SetEtherbase(etherbase)
}

//启动挖矿程序
func (s *IONCMini) StartMining(local bool) error {

	// 从命令行参数中获取矿工账号，如果命令行中不指定一个账户，那么取第一个钱包中的第一个账户
	eb, err := s.Etherbase()
	if err != nil {
		log.Error("Cannot start mining without etherbase", "err", err)
		return fmt.Errorf("etherbase missing: %v", err)
	}
	/*// 如果是POA共识算法
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Etherbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}*/

	// ipos
	if ipos, ok := s.engine.(*ipos.IPos); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Etherbase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		ipos.Authorize(eb, wallet.SignHash)
	}
	go s.miner.Start(eb)
	return nil
}


func (s *IONCMini) StopMining()         { s.miner.Stop() }
func (s *IONCMini) IsMining() bool      { return s.miner.Mining() }
func (s *IONCMini) Miner() *miner.Miner { return s.miner }

func (s *IONCMini) AccountManager() *accounts.Manager { return s.accountManager }
func (s *IONCMini) BlockChain() *core.BlockChain      { return s.blockchain }
func (s *IONCMini) TxPool() *core.TxPool              { return s.txPool }
func (s *IONCMini) ChainDb() ethdb.Database           { return s.chainDb }
func (s *IONCMini) EventMux() *event.TypeMux          { return s.eventMux }
func (s *IONCMini) Engine() consensus.Engine          { return s.engine }
func (s *IONCMini) NetVersion() uint64                { return s.networkId }

func (s *IONCMini) Protocols() []p2p.Protocol {
	return nil
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *IONCMini) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.NetVersion())

	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *IONCMini) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	/*s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}*/
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}