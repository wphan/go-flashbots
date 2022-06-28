package account

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"reflect"
	"testing"
)

func Test_LoadPrivateKey(t *testing.T) {
	type args struct {
		privateKeyString string
	}
	tests := []struct {
		name              string
		args              args
		wantPrivateKey    *ecdsa.PrivateKey
		wantPublicAddress common.Address
		wantErr           bool
	}{
		{
			name: "test a key",
			args: args{
				privateKeyString: "0x9c03d71f2cab3ac367e407e25ed213c56b50957a1f75d9f6b4f9be00066d6963",
			},
			wantPrivateKey: func() (p *ecdsa.PrivateKey) {
				p, _ = crypto.HexToECDSA("9c03d71f2cab3ac367e407e25ed213c56b50957a1f75d9f6b4f9be00066d6963")
				return
			}(),
			wantPublicAddress: common.HexToAddress("0xb73c1b61eecdd422a095e619d121c3162fd9fd51"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrivateKey, gotPublicAddress, err := LoadPrivateKeyString(tt.args.privateKeyString)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPrivateKeyString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPrivateKey, tt.wantPrivateKey) {
				t.Errorf("LoadPrivateKeyString() gotPrivateKey= %v, want %v", gotPrivateKey, tt.wantPrivateKey)
			}
			if !reflect.DeepEqual(gotPublicAddress, tt.wantPublicAddress) {
				t.Errorf("LoadPrivateKeyString() gotPublicAddress = %v, want %v", gotPublicAddress, tt.wantPublicAddress)
			}
		})
	}
}
