// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package core_ionc

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ionchain/ionchain-core/common"
	"github.com/ionchain/ionchain-core/common/hexutil"
	"github.com/ionchain/ionchain-core/common/math"
	"github.com/ionchain/ionchain-core/params"
)

func (g Genesis) MarshalJSON() ([]byte, error) {
	type Genesis struct {
		Config     *params.ChainConfig                         `json:"config"`
		Timestamp  math.HexOrDecimal64                         `json:"timestamp"`
		ExtraData  hexutil.Bytes                               `json:"extraData"`
		GasLimit   math.HexOrDecimal64                         `json:"gasLimit"   gencodec:"required"`
		Difficulty *math.HexOrDecimal256                       `json:"difficulty" gencodec:"required"`
		Alloc      map[common.UnprefixedAddress]GenesisAccount `json:"alloc"      gencodec:"required"`
		Number     math.HexOrDecimal64                         `json:"number"`
		GasUsed    math.HexOrDecimal64                         `json:"gasUsed"`
		ParentHash common.Hash                                 `json:"parentHash"`

		// 新增字段
		BaseTarget *math.HexOrDecimal256                       `json:"baseTarget"   gencodec:"required"`   // baseTarget
		Coinbase   common.Address                              `json:"coinbase"`
		BlockSignature hexutil.Bytes                           `json:"blockSignature"  gencodec:"required"`   // 区块签名信息
		GenerationSignature hexutil.Bytes                      `json:"generationSignature" gencodec:"required"`   // 生成签名信息
	}
	var enc Genesis
	enc.Config = g.Config
	enc.Timestamp = math.HexOrDecimal64(g.Timestamp)
	enc.ExtraData = g.ExtraData
	enc.GasLimit = math.HexOrDecimal64(g.GasLimit)
	enc.Difficulty = (*math.HexOrDecimal256)(g.Difficulty)
	enc.Coinbase = g.Coinbase
	if g.Alloc != nil {
		enc.Alloc = make(map[common.UnprefixedAddress]GenesisAccount, len(g.Alloc))
		for k, v := range g.Alloc {
			enc.Alloc[common.UnprefixedAddress(k)] = v
		}
	}
	enc.Number = math.HexOrDecimal64(g.Number)
	enc.GasUsed = math.HexOrDecimal64(g.GasUsed)
	enc.ParentHash = g.ParentHash

	enc.BaseTarget = (*math.HexOrDecimal256)(g.BaseTarget)
	enc.BlockSignature = g.BlockSignature
	enc.GenerationSignature = g.GenerationSignature
	return json.Marshal(&enc)
}

func (g *Genesis) UnmarshalJSON(input []byte) error {
	type Genesis struct {
		Config     *params.ChainConfig                         `json:"config"`
		Timestamp  *math.HexOrDecimal64                        `json:"timestamp"`
		ExtraData  hexutil.Bytes                               `json:"extraData"`
		GasLimit   *math.HexOrDecimal64                        `json:"gasLimit"   gencodec:"required"`
		Difficulty *math.HexOrDecimal256                       `json:"difficulty" gencodec:"required"`
		Alloc      map[common.UnprefixedAddress]GenesisAccount `json:"alloc"      gencodec:"required"`
		Number     *math.HexOrDecimal64                        `json:"number"`
		GasUsed    *math.HexOrDecimal64                        `json:"gasUsed"`
		ParentHash *common.Hash                                `json:"parentHash"`
		// 新增字段
		BaseTarget *math.HexOrDecimal256                       `json:"baseTarget"   gencodec:"required"`   // baseTarget
		Coinbase   *common.Address                             `json:"coinbase"`
		BlockSignature hexutil.Bytes                           `json:"blockSignature"  gencodec:"required"`   // 区块签名信息
		GenerationSignature hexutil.Bytes                      `json:"generationSignature" gencodec:"required"`   // 生成签名信息
	}
	var dec Genesis
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Config != nil {
		g.Config = dec.Config
	}

	if dec.Timestamp != nil {
		g.Timestamp = uint64(*dec.Timestamp)
	}
	if dec.ExtraData != nil {
		g.ExtraData = dec.ExtraData
	}
	if dec.GasLimit == nil {
		return errors.New("missing required field 'gasLimit' for Genesis")
	}
	g.GasLimit = uint64(*dec.GasLimit)
	if dec.Difficulty == nil {
		return errors.New("missing required field 'difficulty' for Genesis")
	}
	g.Difficulty = (*big.Int)(dec.Difficulty)

	if dec.Coinbase != nil {
		g.Coinbase = *dec.Coinbase
	}
	if dec.Alloc == nil {
		return errors.New("missing required field 'alloc' for Genesis")
	}
	g.Alloc = make(GenesisAlloc, len(dec.Alloc))
	for k, v := range dec.Alloc {
		g.Alloc[common.Address(k)] = v
	}
	if dec.Number != nil {
		g.Number = uint64(*dec.Number)
	}
	if dec.GasUsed != nil {
		g.GasUsed = uint64(*dec.GasUsed)
	}
	if dec.ParentHash != nil {
		g.ParentHash = *dec.ParentHash
	}
	// 新增字段
	if dec.BaseTarget != nil {
		g.BaseTarget = (*big.Int)(dec.BaseTarget)
	}
	if dec.BlockSignature != nil {
		g.BlockSignature = dec.BlockSignature
	}
	if dec.GenerationSignature != nil {
		g.GenerationSignature = dec.GenerationSignature
	}
	return nil
}
