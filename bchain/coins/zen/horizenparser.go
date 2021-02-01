package zen

import (
	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
)

const (
	// MainnetMagic is mainnet network constant
	MainnetMagic wire.BitcoinNet = 0x68736163
	// TestnetMagic is testnet network constant
	TestnetMagic wire.BitcoinNet = 0xe6cdf2bf
)

var (
	// MainNetParams are parser parameters for mainnet
	MainNetParams chaincfg.Params
	// TestNetParams are parser parameters for testnet
	TestNetParams chaincfg.Params
	// RegtestParams are parser parameters for regtest
	RegtestParams chaincfg.Params
)

func init() {
	MainNetParams = chaincfg.MainNetParams
	MainNetParams.Net = MainnetMagic

	// Address encoding magics
	MainNetParams.AddressMagicLen = 2
	MainNetParams.PubKeyHashAddrID = []byte{0x20, 0x89} // base58 prefix: zn
	MainNetParams.ScriptHashAddrID = []byte{0x20, 0x96} // base58 prefix: zs

	TestNetParams = chaincfg.TestNet3Params
	TestNetParams.Net = TestnetMagic

	// Address encoding magics
	TestNetParams.AddressMagicLen = 2
	TestNetParams.PubKeyHashAddrID = []byte{0x20, 0x98} // base58 prefix: zt
	TestNetParams.ScriptHashAddrID = []byte{0x20, 0x92} // base58 prefix: zr
}

// HorizenParser handle
type HorizenParser struct {
	*btc.BitcoinParser
	baseparser *bchain.BaseParser
}

// NewHorizenParser returns new HorizenParser instance
func NewHorizenParser(params *chaincfg.Params, c *btc.Configuration) *HorizenParser {
	return &HorizenParser{
		BitcoinParser: btc.NewBitcoinParser(params, c),
		baseparser:    &bchain.BaseParser{},
	}
}

// GetChainParams contains network parameters for the main Horizen network,
// the test Horizen networkk in this order
func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&MainNetParams) {
		err := chaincfg.Register(&MainNetParams)
		if err == nil {
			err = chaincfg.Register(&TestNetParams)
		}
		if err == nil {
			err = chaincfg.Register(&RegtestParams)
		}
		if err != nil {
			panic(err)
		}
	}
	switch chain {
	case "test":
		return &TestNetParams
	default:
		return &MainNetParams
	}
}

// PackTx packs transaction to byte array using protobuf
func (p *HorizenParser) PackTx(tx *bchain.Tx, height uint32, blockTime int64) ([]byte, error) {
	return p.baseparser.PackTx(tx, height, blockTime)
}

// UnpackTx unpacks transaction from protobuf byte array
func (p *HorizenParser) UnpackTx(buf []byte) (*bchain.Tx, uint32, error) {
	return p.baseparser.UnpackTx(buf)
}
