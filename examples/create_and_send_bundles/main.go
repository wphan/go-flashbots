package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wphan/go-flashbots"
	"github.com/wphan/go-flashbots/account"
)

func main() {
	// sample private key
	pkey, pubAddr, _ := account.LoadPrivateKey("0x9c03d71f2cab3ac367e407e25ed213c56b50957a1f75d9f6b4f9be00066d6963")
	r, err := flashbots.NewRelayClient(pkey, "https://relay.flashbots.net")
	if err != nil {
		panic(err)
	}

	// create some signed transactions
	tx, err := types.SignNewTx(pkey, types.NewEIP2930Signer(big.NewInt(1)), &types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    0,
		GasPrice: nil,
		Gas:      5000000,
		To:       &pubAddr,
		Value:    big.NewInt(0),
		Data:     nil,
	})
	tx2, err := types.SignNewTx(pkey, types.NewEIP2930Signer(big.NewInt(1)), &types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    1,
		GasPrice: nil,
		Gas:      5000000,
		To:       &pubAddr,
		Value:    big.NewInt(0),
		Data:     nil,
	})

	// create the bundle
	b, err := NewBundle([]*types.Transaction{tx, tx2}, 12639480, 0, nil, nil, nil)
	if err != nil {
		panic(err)
	}

	// simulate bundle, send bundle is similar
	resp, duration, err := r.SimulateBundle(b)
	// took: 264.911015ms
	fmt.Printf("took: %+v\n", duration)
	// response: map[id:1 jsonrpc:2.0 result:map[bundleGasPrice:0 bundleHash:0xacaf0c77e88712d83e26ffad84b7f2ed5690451d5fd92fa6adbf7fd7b53530be coinbaseDiff:0 ethSentToCoinbase:0 gasFees:0 results:[map[coinbaseDiff:0 ethSentToCoinbase:0 fromAddress:0xb73C1b61eECdD422A095E619d121C3162fd9fD51 gasFees:0 gasPrice:0 gasUsed:21000 toAddress:0xb73C1b61eECdD422A095E619d121C3162fd9fD51 txHash:0x442bc407a878ec5144f3d9d57f043b416f750fb69344c85d99dbe516c2735931 value:0x] map[coinbaseDiff:0 ethSentToCoinbase:0 fromAddress:0xb73C1b61eECdD422A095E619d121C3162fd9fD51 gasFees:0 gasPrice:0 gasUsed:21000 toAddress:0xb73C1b61eECdD422A095E619d121C3162fd9fD51 txHash:0x62790e732190d74b8fb2c302b5f90a9bcb124abef2b945bc570cf977f4414e93 value:0x]] stateBlockNumber:1.2639768e+07 totalGasUsed:42000]]
	fmt.Printf("response: %+v\n", resp)

	if err != nil {
		panic(err)
	}
}
