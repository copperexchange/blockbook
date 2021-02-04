package zen

import (
  "encoding/hex"

	"github.com/martinboehm/btcd/wire"
	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/btc"
	"github.com/martinboehm/btcutil/txscript"
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

// GetAddrDescFromVout returns internal address representation of a given transaction output.
func (p *HorizenParser) GetAddrDescFromVout(output *bchain.Vout) (bchain.AddressDescriptor, error) {
	script, err := hex.DecodeString(output.ScriptPubKey.Hex)
	if err != nil {
		return nil, err
	}

	scriptClass, addresses, _, err := txscript.ExtractPkScriptAddrs(script, p.Params)
	if err != nil {
	  // bip115 P2PKH
	  pubKeyWithBlock := extractPubKeyHashWithBlock(script)

	  if (pubKeyWithBlock != nil)  {
	    return bchain.AddressDescriptor(pubKeyWithBlock), nil
	  }

	  // bip115 P2SH
	  scriptKeyWithBlock := extractScriptHashWithBlock(script)

	  if (scriptKeyWithBlock != nil) {
	    return bchain.AddressDescriptor(scriptKeyWithBlock), nil
	  }

		return nil, err
	}

	if scriptClass.String() == "nulldata" {
		if parsedOPReturn := p.BitcoinParser.TryParseOPReturn(script); parsedOPReturn != "" {
			return []byte(parsedOPReturn), nil
		}
	}

	var addressByte []byte
	for i := range addresses {
		addressByte = append(addressByte, addresses[i].String()...)
	}
	return bchain.AddressDescriptor(addressByte), nil
}

// extractPubKeyHashWithBlock extracts the public key hash from the passed script if it
// is a standard pay-to-pubkey-hash script with bip115. It will return nil otherwise.
func extractPubKeyHashWithBlock(script []byte) []byte {
	// A pay-to-pubkey-hash script is of the form:
	//  OP_DUP OP_HASH160 <20-byte hash> OP_EQUALVERIFY OP_CHECKSIG <32-byte block hash> <block-number> OP_CHECKBLOCKATHEIGHT
	if len(script) > 59 &&
		script[0] == 0x76 && // OP_DUP
		script[1] == 0xa9 && // OP_HASH160
		script[2] == 0x14 && // OP_DATA_20
		script[23] == 0x88 && // OP_EQUALVERIFY
		script[24] == 0xac && // OP_CHECKSIG
		script[25] == 0x20 && // OP_DATA_32
		script[len(script) - 1] == 0xb4 { // OP_CHECKBLOCKATHEIGHT

		return script[3:23]
	}

	return nil
}

// extractPubKeyHashWithBlock extracts the public key hash from the passed script if it
// is a standard pay-to-script-hash script with bip115. It will return nil otherwise.
func extractScriptHashWithBlock(script []byte) []byte {
	// A pay-to-script-hash script is of the form:
	//  OP_HASH160 <20-byte scripthash> OP_EQUAL
	if len(script) > 57 &&
		script[0] == 0xa9 && // OP_HASH160
		script[1] == 0x14 && // OP_DATA_20
		script[22] == 0x87 && //OP_EQUAL
		script[23] == 0x20 && // OP_DATA_32
		script[len(script) - 1] == 0xb4 { // OP_CHECKBLOCKATHEIGHT

		return script[2:22]
	}

	return nil
}
