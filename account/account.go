package account

import (
	"crypto/ecdsa"
	"strings"

	"errors"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// LoadPrivateKeyString takes a string private key and returns an ecdsa private key and public ehtereum address
func LoadPrivateKeyString(privateKeyString string) (privateKey *ecdsa.PrivateKey, publicAddress ethcommon.Address, retErr error) {
	if strings.HasPrefix(strings.ToLower(privateKeyString), "0x") {
		privateKeyString = privateKeyString[2:]
	}

	privateKey, retErr = crypto.HexToECDSA(privateKeyString)
	if privateKey == nil || retErr != nil {
		return
	}
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		retErr = errors.New("ownerPrivateKey.Public is not (*ecdsa.PublicKey)")
		return
	}

	publicAddress = crypto.PubkeyToAddress(*publicKeyECDSA)

	return
}
