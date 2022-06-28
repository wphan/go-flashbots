# Flashbots Go Client

go-flashbots is slightly more opinionated client for the flashbots rpc

* https://github.com/flashbots/mev-relay-js
* https://docs.flashbots.net/flashbots-core/searchers/advanced/rpc-endpoint
* other Golang providers: https://docs-staging.flashbots.net/flashbots-auction/searchers/libraries/golang

# Features

* takes in `[]*types.Transaction` to create bundle
* returns a `time.Duration` to track response times of relays (useful for identifying when relay may be congested)
* allow bulk sending bundles (send to multiple relays) via `BatchRelayClient` and `BatchSendBundle`