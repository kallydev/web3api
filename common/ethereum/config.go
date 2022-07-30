package ethereum

type Config struct {
	Network *Network `json:"network"`
}

type Network struct {
	Ethereum *Endpoint `json:"ethereum"`
	Polygon  *Endpoint `json:"polygon"`
	Optimism *Endpoint `json:"optimism"`
	Arbitrum *Endpoint `json:"arbitrum"`
}

type Endpoint struct {
	HTTP      string `json:"http"`
	WebSocket string `json:"websocket"`
}
