package ethereum

import "github.com/pkg/errors"

var ErrorUnsupportedNetwork = errors.New("unsupported network")

const (
	NetworkEthereum = "ethereum"
	NetworkPolygon  = "polygon"
)
