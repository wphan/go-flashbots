package flashbots

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type Bundle struct {
	Transactions     []*types.Transaction `json:"-"`
	TxsByteString    []string             `json:"txs"`                        // TxsByteString is the same as Transactions, but in raw hex formatted strings
	BlockNumber      string               `json:"blockNumber"`                // BlockNumber is the earliest block number where this bundle will be valid (stored as hex string)
	StateBlockNumber string               `json:"stateBlockNumber,omitempty"` // StateBlockNumber must be provided if simulating bundle (stored as hex string)
	MinTimestamp     *int                 `json:"minTimestamp,omitempty"`     // MinTimestamp in which the bundle will be valid, nil to allow any time
	MaxTimestamp     *int                 `json:"maxTimestamp,omitempty"`     // MaxTimestamp in which the bundle will be valid, nil to allow any time

	// RevertingTxHashes contain list of transaction hashes that are "allowed to revert" - bundle will still land on chain if these transactions revert
	RevertingTxHashes []string `json:"revertingTxHashes,omitempty"`
}

// NewBundle creates a new bundle.
// blockNumber:       block number for which this bundle is valid on
// stateBlockNumber:  block number or a block tag for which state to base this simulation on. will be "latest" if 0
// minTimestamp:      min timestamp that bundle will be valid, nil for any
// maxTimestamp:      max timestamp that bundle will be valid, nil for any
// revertingTxHashes: list of transaction hashes that are "allowed to fail"
func NewBundle(transactions []*types.Transaction, blockNumber, stateBlockNumber uint64,
	minTimestamp *int, maxTimestamp *int, revertingTxHashes []common.Hash) (b Bundle, retErr error) {

	b.Transactions = transactions

	b.TxsByteString = make([]string, len(b.Transactions))
	for i, tx := range b.Transactions {
		txBytes, err := tx.MarshalBinary()
		if err != nil {
			retErr = fmt.Errorf("failed tx.MarshalBinary() for tx: %s\nerrors: %w", tx.Hash().Hex(), err)
			return
		}
		b.TxsByteString[i] = "0x" + common.Bytes2Hex(txBytes)
	}

	b.BlockNumber = "0x" + strconv.FormatUint(blockNumber, 16)
	if stateBlockNumber == 0 {
		b.StateBlockNumber = "latest"
	} else {
		b.StateBlockNumber = "0x" + strconv.FormatUint(stateBlockNumber, 16)
	}
	b.MinTimestamp = minTimestamp
	b.MaxTimestamp = maxTimestamp
	b.RevertingTxHashes = make([]string, len(revertingTxHashes))
	for i, tx := range revertingTxHashes {
		b.RevertingTxHashes[i] = tx.Hex()
	}

	return
}

// AddTransaction to existing bundle
func (b *Bundle) AddTransaction(tx *types.Transaction) {
	b.Transactions = append(b.Transactions, tx)
}

type rpcPaylod struct {
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int64       `json:"id"`
}

type RelayClient struct {
	name                 string            // name used to identify this RelayClient
	signingPrivateKey    *ecdsa.PrivateKey // signingPrivateKey signs bundles for flashbots
	signingPublicAddress common.Address    // signingPublicAddress is the public Ethereum address of signingPrivateKey
	mainEndpoint         string            // mainEndpoint of the relay server, bundles are sent to this server

	// simulationEndpoint is used for simulating bundles
	// this is useful if you run a local version of mev-geth and don't want to wait for the slow public relays to respond
	simulationEndpoint string
}

// NewRelayClient creates a new relay client
// signingPrivateKey:   the key used to sign bundles (this can be any valid private key)
// mainEndpoint:        the relay server endpoint used for sending bundles
// simulationEndpoint:  the relay server endpoint used for bundle simulation
func NewRelayClient(signingPrivateKey *ecdsa.PrivateKey, name, mainEndpoint, simulationEndpoint string) (r *RelayClient, retErr error) {
	if signingPrivateKey == nil {
		retErr = errors.New("must provide a signingPrivateKey")
		return
	}
	r = &RelayClient{
		name:                 name,
		signingPrivateKey:    signingPrivateKey,
		signingPublicAddress: crypto.PubkeyToAddress(signingPrivateKey.PublicKey),
		mainEndpoint:         mainEndpoint,
		simulationEndpoint:   simulationEndpoint,
	}

	return
}

func (r RelayClient) Name() string                   { return r.name }
func (r RelayClient) SigningAddress() common.Address { return r.signingPublicAddress }
func (r RelayClient) MainEndpoint() string           { return r.mainEndpoint }
func (r RelayClient) SimulationEndpoint() string     { return r.simulationEndpoint }

func (r *RelayClient) prepareBundlePayload(b Bundle, method string) (payloadBytes []byte, retErr error) {

	payload := rpcPaylod{
		JsonRPC: "2.0",
		Method:  method,
		Params:  []Bundle{b},
		ID:      1,
	}

	payloadBytes, retErr = json.Marshal(payload)
	return
}

// prepareBundleStatsPayload prepares a payload for requesting stats for a single bundle. The signign address must be
// the same as the one who submitted the bundle
func (r *RelayClient) prepareBundleStatsPayload(bundleHash, blockNumber, method string) (payloadBytes []byte, retErr error) {

	payload := rpcPaylod{
		JsonRPC: "2.0",
		Method:  method,
		Params: []map[string]string{
			{
				"bundleHash":  bundleHash,
				"blockNumber": blockNumber,
			},
		},
		ID: 1,
	}

	payloadBytes, retErr = json.Marshal(payload)
	return
}

func (r *RelayClient) signPayload(payload []byte) (signature string, retErr error) {
	hashedBody := crypto.Keccak256Hash(payload).Hex()
	payloadHash := crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(hashedBody)) + hashedBody))
	signatureBytes, err := crypto.Sign(payloadHash, r.signingPrivateKey)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signatureBytes), nil
}

func (r *RelayClient) fbRequest(endpoint string, payload []byte) (responseBytes []byte, duration time.Duration, retErr error) {
	signature, err := r.signPayload(payload)
	if err != nil {
		retErr = err
		return
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		retErr = err
		return
	}
	req.Header.Add("X-Flashbots-Signature", r.signingPublicAddress.Hex()+":"+signature)
	req.Header.Set("Content-Type", "application/json")
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	duration = time.Since(start)
	if err != nil {
		retErr = err
		return
	}
	responseBytes, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		retErr = err
		return
	}

	return
}

func unmarshalBodyBytesToGenericMap(bytes []byte) (response map[string]interface{}, retErr error) {
	err := json.Unmarshal(bytes, &response)
	if err != nil {
		retErr = fmt.Errorf("failed to unmarshal body to a generic map: %w", err)
		return
	}
	return
}

type SendBundleResponse struct {
	ResponseBytes []byte
	Duration      time.Duration
	Error         error
}

// SendBundle sends a Bundle on RelayClient.
func (r *RelayClient) SendBundle(b Bundle) (resp SendBundleResponse) {
	payload, err := r.prepareBundlePayload(b, "eth_sendBundle")
	if err != nil {
		return SendBundleResponse{
			Error: err,
		}
	}

	responseBytes, duration, err := r.fbRequest(r.mainEndpoint, payload)

	return SendBundleResponse{
		ResponseBytes: responseBytes,
		Duration:      duration,
		Error:         err,
	}
}

func (r *RelayClient) SimulateBundle(b Bundle) (responseBytes []byte, duration time.Duration, retErr error) {
	if r.simulationEndpoint == "" {
		retErr = errors.New("no simulation endpoint for relay " + r.name)
		return
	}
	if b.StateBlockNumber == "" || b.StateBlockNumber == "0x0" {
		b.StateBlockNumber = "latest"
	}
	payload, err := r.prepareBundlePayload(b, "eth_callBundle")
	if err != nil {
		retErr = err
		return
	}
	return r.fbRequest(r.simulationEndpoint, payload)
}

type BundleStats struct {
	ID      int    `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		IsHighPriority bool      `json:"isHighPriority"`
		IsSentToMiners bool      `json:"isSentToMiners"`
		IsSimulated    bool      `json:"isSimulated"`
		SentToMinersAt time.Time `json:"sentToMinersAt"`
		SimulatedAt    time.Time `json:"simulatedAt"`
		SubmittedAt    time.Time `json:"submittedAt"`
	} `json:"result"`
}

// GetBundleStats queries flashbots_getBundleStats for stats on a single bundle. BundleHash and blockNumber must be a hexadecimal strings
func (r *RelayClient) GetBundleStats(bundleHash, blockNumber string) (bundleStats BundleStats, duration time.Duration, retErr error) {
	payload, err := r.prepareBundleStatsPayload(bundleHash, blockNumber, "flashbots_getBundleStats")
	if err != nil {
		retErr = err
		return
	}

	var bodyBytes []byte
	bodyBytes, duration, err = r.fbRequest(r.MainEndpoint(), payload)
	if err != nil {
		retErr = fmt.Errorf("failed to make fbRequest: %w", err)
		return
	}

	err = json.Unmarshal(bodyBytes, &bundleStats)
	if err != nil {
		retErr = fmt.Errorf("failed to unmarshal into BundleStats: %s\nerror: %w", string(bodyBytes), err)
		return
	}

	return
}

func ExtractGasUsedFromBundleResponse(bytes []byte) (gasUsed float64) {
	resp, err := unmarshalBodyBytesToGenericMap(bytes)
	if err != nil {
		panic(fmt.Errorf("body bytes: %s, %w", string(bytes), err))
	}
	gasUsed = resp["result"].(map[string]interface{})["totalGasUsed"].(float64)
	return
}

func ExtractBundleHashFromBundleResponse(bytes []byte) (bundleHash string) {
	resp, err := unmarshalBodyBytesToGenericMap(bytes)
	if err != nil {
		panic(fmt.Errorf("body bytes: %s, %w", string(bytes), err))
	}
	bundleHash = resp["result"].(map[string]interface{})["bundleHash"].(string)
	return
}

func ExtractExecutionErrorFromSendBundleResponse(bytes []byte) (errs []error) {
	errs = make([]error, 0)

	resp, err := unmarshalBodyBytesToGenericMap(bytes)
	if err != nil {
		errs = append(
			errs,
			fmt.Errorf("failed to unmarshal response: %s, %w", string(bytes), err))
		return
	}

	var resultOk bool
	var txResultsI map[string]interface{}
	txResultsI, resultOk = resp["result"].(map[string]interface{})
	if !resultOk {
		errResponse, errorOk := resp["error"]
		if errorOk {
			errs = append(
				errs,
				fmt.Errorf("%s", errResponse))
			return
		}
		errs = append(
			errs,
			errors.New(txResultsI["message"].(string)))
		return
	}
	txResults, ok := txResultsI["results"].([]interface{})
	if !ok {
		return
	}
	for _, txI := range txResults {
		tx := txI.(map[string]interface{})
		errString, errOk := tx["error"]
		revertString, revertOk := tx["revert"]
		if errOk || revertOk {
			errs = append(
				errs,
				fmt.Errorf("err: %s, revertString: %s", errString, revertString))
		}
	}

	return
}

type BatchRelayClient struct {
	relayClients []*RelayClient
}

func NewBatchRelayClient(
	signingKeys []*ecdsa.PrivateKey,
	names, mainEndpoints []string,
) (b *BatchRelayClient, retErr error) {
	if (len(signingKeys) != len(names)) || (len(signingKeys) != len(mainEndpoints)) {
		retErr = errors.New("must initialize with same length slices")
		return
	}

	r := make([]*RelayClient, 0)
	for idx := range signingKeys {
		c, err := NewRelayClient(
			signingKeys[idx],
			names[idx],
			mainEndpoints[idx],
			"",
		)
		if err != nil {
			retErr = fmt.Errorf(
				"failed to initialize relay client, name: %s, endpoint: %s, error: %w",
				names[idx],
				mainEndpoints[idx],
				err,
			)
			return
		}

		r = append(r, c)
	}

	b = &BatchRelayClient{
		relayClients: r,
	}
	return
}

// BatchSendBundle sends a Bundle on all connected relay clients
func (r *BatchRelayClient) BatchSendBundle(b Bundle) (resps map[string]SendBundleResponse) {

	resps = make(map[string]SendBundleResponse)

	for _, client := range r.relayClients {
		resp := client.SendBundle(b)
		resps[client.Name()] = resp
	}

	return
}
