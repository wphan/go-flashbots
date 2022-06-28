package flashbots

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/wphan/go-flashbots/account"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var fbLiveTest = flag.Bool("fbLiveTest", false, "perform test against live Flashbots endpoints")

func TestNewBundle(t *testing.T) {
	type args struct {
		txBytes           []string
		blockNumber       uint64
		stateBlockNumber  uint64
		minTimestamp      *int
		maxTimestamp      *int
		revertingTxHashes []common.Hash
	}
	tests := []struct {
		name    string
		args    args
		wantB   Bundle
		wantErr bool
	}{
		{
			name: "test minimum fields",
			args: args{
				txBytes:           []string{"0xf903068080830c350094111111111111111111111111111111111111111180b902a4589b65e900000000000000000000000000000000000000000000000000000000000000200000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e0000000000000000000000006b5194d22231a3b030ddad0668db8984833926d6000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b77ede0000000000000000000000000000000000000000000000000011e0d95d7f154a000000000000000000000000e592427a0aece92de3edee1f18e0157c0586156400000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000104db3e2198000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000000000000000000000000000000000000000000bb80000000000000000000000006951b5bd815043e3f842c1b026b0fa888cc2dd850000000000000000000000000000000000000000000000000000000060c251690000000000000000000000000000000000000000000000000000000000b77ee20000000000000000000000000000000000000000000000000011e0d95d7f154a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025a03ed2623ec91dec66a8167fd2e93928d04c2f5b763c060e18f7f5ed636f84bc29a00322ef1ae2dbccffdf35a86f7fc06f5a2490086136c5854517945bf03b690d08"},
				blockNumber:       1000,
				stateBlockNumber:  0,
				minTimestamp:      nil,
				maxTimestamp:      nil,
				revertingTxHashes: nil,
			},
			wantB: Bundle{
				//Transactions:     []string{"0xf903068080830c350094111111111111111111111111111111111111111180b902a4589b65e900000000000000000000000000000000000000000000000000000000000000200000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e0000000000000000000000006b5194d22231a3b030ddad0668db8984833926d6000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b77ede0000000000000000000000000000000000000000000000000011e0d95d7f154a000000000000000000000000e592427a0aece92de3edee1f18e0157c0586156400000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000104db3e2198000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000000000000000000000000000000000000000000bb80000000000000000000000006951b5bd815043e3f842c1b026b0fa888cc2dd850000000000000000000000000000000000000000000000000000000060c251690000000000000000000000000000000000000000000000000000000000b77ee20000000000000000000000000000000000000000000000000011e0d95d7f154a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025a03ed2623ec91dec66a8167fd2e93928d04c2f5b763c060e18f7f5ed636f84bc29a00322ef1ae2dbccffdf35a86f7fc06f5a2490086136c5854517945bf03b690d08"},
				BlockNumber:      "0x3e8",
				StateBlockNumber: "latest",
			},
		},
		{
			name: "test with timesetamp",
			args: args{
				txBytes:           []string{"0xf903068080830c350094111111111111111111111111111111111111111180b902a4589b65e900000000000000000000000000000000000000000000000000000000000000200000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e0000000000000000000000006b5194d22231a3b030ddad0668db8984833926d6000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b77ede0000000000000000000000000000000000000000000000000011e0d95d7f154a000000000000000000000000e592427a0aece92de3edee1f18e0157c0586156400000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000104db3e2198000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000000000000000000000000000000000000000000bb80000000000000000000000006951b5bd815043e3f842c1b026b0fa888cc2dd850000000000000000000000000000000000000000000000000000000060c251690000000000000000000000000000000000000000000000000000000000b77ee20000000000000000000000000000000000000000000000000011e0d95d7f154a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025a03ed2623ec91dec66a8167fd2e93928d04c2f5b763c060e18f7f5ed636f84bc29a00322ef1ae2dbccffdf35a86f7fc06f5a2490086136c5854517945bf03b690d08"},
				blockNumber:       1000,
				stateBlockNumber:  0,
				minTimestamp:      func() *int { a := 23123; return &a }(),
				maxTimestamp:      nil,
				revertingTxHashes: nil,
			},
			wantB: Bundle{
				//Transactions:     []string{"0xf903068080830c350094111111111111111111111111111111111111111180b902a4589b65e900000000000000000000000000000000000000000000000000000000000000200000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e0000000000000000000000006b5194d22231a3b030ddad0668db8984833926d6000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b77ede0000000000000000000000000000000000000000000000000011e0d95d7f154a000000000000000000000000e592427a0aece92de3edee1f18e0157c0586156400000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000104db3e2198000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000000000000000000000000000000000000000000bb80000000000000000000000006951b5bd815043e3f842c1b026b0fa888cc2dd850000000000000000000000000000000000000000000000000000000060c251690000000000000000000000000000000000000000000000000000000000b77ee20000000000000000000000000000000000000000000000000011e0d95d7f154a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025a03ed2623ec91dec66a8167fd2e93928d04c2f5b763c060e18f7f5ed636f84bc29a00322ef1ae2dbccffdf35a86f7fc06f5a2490086136c5854517945bf03b690d08"},
				BlockNumber:      "0x3e8",
				StateBlockNumber: "latest",
				MinTimestamp:     func() *int { a := 23123; return &a }(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs := make([]*types.Transaction, len(tt.args.txBytes))

			tt.wantB.Transactions = make([]*types.Transaction, len(tt.args.txBytes))
			for i, _ := range tt.args.txBytes {
				tx := types.NewTx(&types.AccessListTx{
					ChainID: big.NewInt(1),
				})
				err := tx.UnmarshalBinary(common.FromHex(tt.args.txBytes[i]))
				if err != nil {
					t.Errorf("failed to unmarshal bytes to tx: %+v\n", err)
					return
				}
				txs[i] = tx
				tt.wantB.Transactions[i] = tx
			}
			gotB, err := NewBundle(txs, tt.args.blockNumber, tt.args.stateBlockNumber, tt.args.minTimestamp, tt.args.maxTimestamp, tt.args.revertingTxHashes)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBundle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotB.Transactions, tt.wantB.Transactions) {
				t.Errorf("NewBundle() Transactions\ngotB: %v\nwant: %v", gotB.Transactions, tt.wantB.Transactions)
			}
			if !reflect.DeepEqual(gotB.BlockNumber, tt.wantB.BlockNumber) {
				t.Errorf("NewBundle() BlockNumber\ngotB: %v\nwant: %v", gotB.BlockNumber, tt.wantB.BlockNumber)
			}
			if !reflect.DeepEqual(gotB.StateBlockNumber, tt.wantB.StateBlockNumber) {
				t.Errorf("NewBundle() StateBlockNumber\ngotB: %v\nwant: %v", gotB.StateBlockNumber, tt.wantB.StateBlockNumber)
			}
			if !reflect.DeepEqual(gotB.MinTimestamp, tt.wantB.MinTimestamp) {
				t.Errorf("NewBundle() MinTimestamp\ngotB: %v\nwant: %v", gotB.MinTimestamp, tt.wantB.MinTimestamp)
			}
			if !reflect.DeepEqual(gotB.MaxTimestamp, tt.wantB.MaxTimestamp) {
				t.Errorf("NewBundle() MaxTimestamp\ngotB: %v\nwant: %v", gotB.MaxTimestamp, tt.wantB.MaxTimestamp)
			}
			if len(gotB.RevertingTxHashes) != len(tt.wantB.RevertingTxHashes) {
				t.Errorf("NewBundle() len(RevertingTxHashes)\ngotB: %v\nwant: %v", len(gotB.RevertingTxHashes), len(tt.wantB.RevertingTxHashes))
			}
			for i, _ := range gotB.RevertingTxHashes {
				if gotB.RevertingTxHashes[i] != tt.wantB.RevertingTxHashes[i] {
					t.Errorf("NewBundle() RevertingTxHashes[%d]\ngotB: %v\nwant: %v", i, gotB.RevertingTxHashes[i], tt.wantB.RevertingTxHashes[i])
				}
			}
		})
	}
}

func TestNewBundleJSON(t *testing.T) {
	txBytes := []string{"0xf903068080830c350094111111111111111111111111111111111111111180b902a4589b65e900000000000000000000000000000000000000000000000000000000000000200000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e0000000000000000000000006b5194d22231a3b030ddad0668db8984833926d6000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b77ede0000000000000000000000000000000000000000000000000011e0d95d7f154a000000000000000000000000e592427a0aece92de3edee1f18e0157c0586156400000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000104db3e2198000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000000000000000000000000000000000000000000bb80000000000000000000000006951b5bd815043e3f842c1b026b0fa888cc2dd850000000000000000000000000000000000000000000000000000000060c251690000000000000000000000000000000000000000000000000000000000b77ee20000000000000000000000000000000000000000000000000011e0d95d7f154a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025a03ed2623ec91dec66a8167fd2e93928d04c2f5b763c060e18f7f5ed636f84bc29a00322ef1ae2dbccffdf35a86f7fc06f5a2490086136c5854517945bf03b690d08"}
	txs := make([]*types.Transaction, len(txBytes))
	for i, _ := range txBytes {
		tx := types.NewTx(&types.AccessListTx{
			ChainID: big.NewInt(1),
		})
		err := tx.UnmarshalBinary(common.FromHex(txBytes[i]))
		if err != nil {
			t.Errorf("failed to unmarshal bytes to tx: %+v\n", err)
			return
		}
		txs[i] = tx
	}
	ts := 555555
	b, err := NewBundle(txs, 12345, 0, &ts, nil, nil)
	if err != nil {
		t.Errorf("failed to make new bundle: %+v\n", err)
		return
	}

	bundleJSON, err := json.Marshal(b)
	if err != nil {
		t.Errorf("failed to make bundle json: %+v\n", err)
		return
	}

	wantBundleJSON := fmt.Sprintf(`{"txs":["%s"],"blockNumber":"0x3039","stateBlockNumber":"latest","minTimestamp":555555}`, txBytes[0])
	if !reflect.DeepEqual(string(bundleJSON), wantBundleJSON) {
		t.Errorf("wrong json\nwant: %+v\ngot:  %+v\n", wantBundleJSON, string(bundleJSON))
		return
	}
}

func Test_preparePayload(t *testing.T) {
	txBytes := []string{"0xf903068080830c350094111111111111111111111111111111111111111180b902a4589b65e900000000000000000000000000000000000000000000000000000000000000200000000000000000000000001e0447b19bb6ecfdae1e4ae1694b0c3659614e4e0000000000000000000000006b5194d22231a3b030ddad0668db8984833926d6000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b77ede0000000000000000000000000000000000000000000000000011e0d95d7f154a000000000000000000000000e592427a0aece92de3edee1f18e0157c0586156400000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000000104db3e2198000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480000000000000000000000000000000000000000000000000000000000000bb80000000000000000000000006951b5bd815043e3f842c1b026b0fa888cc2dd850000000000000000000000000000000000000000000000000000000060c251690000000000000000000000000000000000000000000000000000000000b77ee20000000000000000000000000000000000000000000000000011e0d95d7f154a00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000025a03ed2623ec91dec66a8167fd2e93928d04c2f5b763c060e18f7f5ed636f84bc29a00322ef1ae2dbccffdf35a86f7fc06f5a2490086136c5854517945bf03b690d08"}
	txs := make([]*types.Transaction, len(txBytes))
	for i, _ := range txBytes {
		tx := types.NewTx(&types.AccessListTx{
			ChainID: big.NewInt(1),
		})
		err := tx.UnmarshalBinary(common.FromHex(txBytes[i]))
		if err != nil {
			t.Errorf("failed to unmarshal bytes to tx: %+v\n", err)
			return
		}
		txs[i] = tx
	}
	ts := 555555
	b, err := NewBundle(txs, 12345, 0, &ts, nil, nil)
	if err != nil {
		t.Errorf("failed to make new bundle: %+v\n", err)
		return
	}
	pkey, _, _ := account.LoadPrivateKeyString("0x9c03d71f2cab3ac367e407e25ed213c56b50957a1f75d9f6b4f9be00066d6963")
	r, _ := NewRelayClient(pkey, "", "", "")
	payloadBytes, err := r.prepareBundlePayload(b, "eth_sendBundle")
	if err != nil {
		t.Errorf("failed prepareBundlePayload: %+v\n", err)
		return
	}

	wantPayloadJSON := fmt.Sprintf(`{"txs":["%s"],"blockNumber":"0x3039","stateBlockNumber":"latest","minTimestamp":555555}`, txBytes[0])
	wantFinalPayload := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_sendBundle","params":[%s],"id":1}`, wantPayloadJSON)
	if !reflect.DeepEqual(string(payloadBytes), string(wantFinalPayload)) {
		t.Errorf("wrong json\nwant: %+v\ngot:  %+v\n", string(payloadBytes), string(wantFinalPayload))
		return
	}
}

func TestRelayClient_SendBundle(t *testing.T) {
	if !*fbLiveTest {
		t.SkipNow()
	}

	pkey, pubAddr, _ := account.LoadPrivateKeyString("0x9c03d71f2cab3ac367e407e25ed213c56b50957a1f75d9f6b4f9be00066d6963")
	r, err := NewRelayClient(pkey, "test-client", "https://relay.flashbots.net", "https://relay.flashbots.net")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := types.SignNewTx(pkey, types.NewEIP2930Signer(big.NewInt(1)), &types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    0,
		GasPrice: nil,
		Gas:      500,
		To:       &pubAddr,
		Value:    big.NewInt(1000000000),
		Data:     nil,
	})

	fmt.Printf("tx hash: %+v\n", tx.Hash())
	fmt.Printf("tx to: %+v\n", tx.To())
	fmt.Printf("tx type: %+v\n", tx.Type())
	fmt.Printf("tx value: %+v\n", tx.Value())

	b, err := NewBundle([]*types.Transaction{}, 12639450, 0, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := r.SendBundle(b)
	fmt.Printf("took: %+v\n", resp.Duration)
	fmt.Printf("response: %+v\n", resp.ResponseBytes)
	if resp.Error != nil {
		t.Fatalf("%+v", resp.Error)
	}
}

func TestRelayClient_SimulateBundle(t *testing.T) {
	if !*fbLiveTest {
		t.SkipNow()
	}

	pkey, pubAddr, _ := account.LoadPrivateKeyString("0x9c03d71f2cab3ac367e407e25ed213c56b50957a1f75d9f6b4f9be00066d6963")
	r, err := NewRelayClient(pkey, "test-client", "https://relay.flashbots.net", "https://relay.flashbots.net")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := types.SignNewTx(pkey, types.NewEIP2930Signer(big.NewInt(1)), &types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    0,
		GasPrice: big.NewInt(100000000),
		Gas:      5000000,
		To:       &pubAddr,
		Value:    big.NewInt(0),
		Data:     nil,
	})
	tx2, err := types.SignNewTx(pkey, types.NewEIP2930Signer(big.NewInt(1)), &types.AccessListTx{
		ChainID:  big.NewInt(1),
		Nonce:    1,
		GasPrice: big.NewInt(1111),
		Gas:      5000000,
		To:       &pubAddr,
		Value:    big.NewInt(0),
		Data:     nil,
	})

	fmt.Printf("tx hash: %+v\n", tx.Hash())
	fmt.Printf("tx to: %+v\n", tx.To())
	fmt.Printf("tx type: %+v\n", tx.Type())
	fmt.Printf("tx value: %+v\n", tx.Value())

	b, err := NewBundle([]*types.Transaction{tx, tx2}, 12639480, 0, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	respBytes, duration, err := r.SimulateBundle(b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("took: %+v\n", duration)
	fmt.Printf("response: %s\n", string(respBytes))
	errs := ExtractExecutionErrorFromSendBundleResponse(respBytes)
	if len(errs) > 0 {
		t.Fatalf("errors: %+v", errs)
	}

	gasUsed := ExtractGasUsedFromBundleResponse(respBytes)
	fmt.Printf("GasUsed: %+v\n", gasUsed)

	if err != nil {
		t.Fatal(err)
	}
}

func TestExtractGasUsedFromBundleResponse(t *testing.T) {
	type args struct {
		jsonResponseStr string
	}
	tests := []struct {
		name        string
		args        args
		wantGasUsed float64
	}{
		{
			name: "generic reponse",
			args: args{
				jsonResponseStr: `{"id":1,"jsonrpc":2.0,"result":{"bundleHash":"0x0d1b53154e2910960564190ad0c5ef34c49befb865e3d56374adbf2b1160aa65","results":[{"txHash":"0x0a9e21a9c0dd6b868b1d26d9bc6b11a549fac1fb70bc2c10ebf925c43def862c"}],"totalGasUsed":243051}}`,
			},
			wantGasUsed: 243051,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotGasUsed := ExtractGasUsedFromBundleResponse([]byte(tt.args.jsonResponseStr)); gotGasUsed != tt.wantGasUsed {
				t.Errorf("ExtractGasUsedFromBundleResponse() = %v, want %v", gotGasUsed, tt.wantGasUsed)
			}
		})
	}
}

func TestExtractExecutionErrorFromBundleResponse(t *testing.T) {
	type args struct {
		jsonResponseStr string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []error
	}{
		{
			name: "Test with no error",
			args: args{
				jsonResponseStr: `{"id":1,"jsonrpc":2.0,"result":{"bundleHash":"0x0d1b53154e2910960564190ad0c5ef34c49befb865e3d56374adbf2b1160aa65","results":[{"txHash":"0x0a9e21a9c0dd6b868b1d26d9bc6b11a549fac1fb70bc2c10ebf925c43def862c"}],"totalGasUsed":243051}}`,
			},
			wantErrs: []error{},
		},
		{
			name: "Test with error",
			args: args{
				jsonResponseStr: `{"id":1,"jsonrpc":2.0,"result":{"bundleGasPrice":0,"bundleHash":"0x0ccf11afd8f1aaeb05d5057d79395a612e35d589ead6bb63e5caef2d5e7b670f","coinbaseDiff":0,"ethSentToCoinbase":0,"gasFees":0,"results":[{"coinbaseDiff":0,"error":"execution reverted","ethSentToCoinbase":0,"fromAddress":"0x3cA43755058a2294Fb280DfF9127db6F9c2216EA","gasFees":0,"gasPrice":0,"gasUsed":240600,"revert":"y","toAddress":"0x162Ab7D33ab2f61A5c380a37F7b516EDaFd77913","txHash":"0xabc8eb8ca3f66072aba73063332edc8d86904febd5f85923cc44d289ecaf2623"}],"stateBlockNumber":1.3051998e+07,"totalGasUsed":240600}}`,
			},
			wantErrs: []error{errors.New("err: execution reverted, revertString: y")},
		},
		{
			name: "Test simulation error",
			args: args{
				jsonResponseStr: `{"error":{"code":-32000, "message":"err: nonce too low: address 0x3cA43755058a2294Fb280DfF9127db6F9c2216EA, tx: 31 state: 32; txhash 0xb5fba72f1163ec32218697b50e39ab30039fde4ee894e4ffc233753f4ecb82d7"},"id":1,"jsonrpc":2.0}`,
			},
			wantErrs: []error{errors.New("map[code:%!s(float64=-32000) message:err: nonce too low: address 0x3cA43755058a2294Fb280DfF9127db6F9c2216EA, tx: 31 state: 32; txhash 0xb5fba72f1163ec32218697b50e39ab30039fde4ee894e4ffc233753f4ecb82d7]")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErrs := ExtractExecutionErrorFromSendBundleResponse([]byte(tt.args.jsonResponseStr))
			for i, _ := range gotErrs {
				gotErr := gotErrs[i].Error()
				wantErr := tt.wantErrs[i].Error()

				if !reflect.DeepEqual(gotErr, wantErr) {
					t.Errorf("ExtractExecutionErrorFromSendBundleResponse()\ngot:  %+v\nwant: %+v", gotErr, wantErr)
				}
			}
		})
	}
}
func TestRelayClient_GetBundleStats(t *testing.T) {
	if !*fbLiveTest {
		t.SkipNow()
	}

	pkey, pubAddr, _ := account.LoadPrivateKeyString("")
	r, err := NewRelayClient(pkey, "test-client", "https://relay.flashbots.net", "https://relay.flashbots.net")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(pubAddr)
	bundleHash := "0x48f1df898a9bde45e92b21736cda94841e4bae6b2da6abcca1d42b96b47c0ecd"
	blockNumber := "0xc73a46"
	iResp, duration, err := r.GetBundleStats(bundleHash, blockNumber)
	fmt.Printf("took: %+v\n", duration)
	fmt.Printf("response: %+v\n", iResp)
	if err != nil {
		t.Fatal(err)
	}
}
