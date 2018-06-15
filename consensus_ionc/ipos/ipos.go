package ipos

import(
	types "github.com/ionchain/ionchain-core/core/types_ionc"
	"github.com/ionchain/ionchain-core/common"
	consensus "github.com/ionchain/ionchain-core/consensus_ionc"
	"github.com/ionchain/ionchain-core/rpc"
	"github.com/ionchain/ionchain-core/core/state"
)

type IPos struct {

}

// Author retrieves the Ethereum address of the account that minted the given
// block, which may be different from the header's coinbase if a consensus
// engine is based on signatures.
func (c *IPos) Author(header *types.Header) (common.Address, error) {
	return common.Address{},nil
}

// VerifyHeader checks whether a header conforms to the consensus rules.
func (c *IPos) VerifyHeader(chain consensus.ChainReader, header *types.Header, seal bool) error {
	return nil
}

// VerifyHeaders is similar to VerifyHeader, but verifies a batch of headers. The
// method returns a quit channel to abort the operations and a results channel to
// retrieve the async verifications (the order is that of the input slice).
func (c *IPos) VerifyHeaders(chain consensus.ChainReader, headers []*types.Header, seals []bool) (chan<- struct{}, <-chan error) {

	return nil, nil
}

// VerifySeal implements consensus.Engine, checking whether the signature contained
// in the header satisfies the consensus protocol requirements.
func (c *IPos) VerifySeal(chain consensus.ChainReader, header *types.Header) error {
	return nil
}



// Prepare implements consensus.Engine, preparing all the consensus fields of the
// header for running the transactions on top.
func (c *IPos) Prepare(chain consensus.ChainReader, header *types.Header) error {

	return nil
}

// Finalize implements consensus.Engine, ensuring no uncles are set, nor block
// rewards given, and returns the final block.
func (c *IPos) Finalize(chain consensus.ChainReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {

	return nil, nil
}


// Seal implements consensus.Engine, attempting to create a sealed block using
// the local signing credentials.
func (c *IPos) Seal(chain consensus.ChainReader, block *types.Block, stop <-chan struct{}) (*types.Block, error) {

	return nil, nil
}

// APIs implements consensus.Engine, returning the user facing RPC API to allow
// controlling the signer voting.
func (c *IPos) APIs(chain consensus.ChainReader) []rpc.API {
	return []rpc.API{{}}
}