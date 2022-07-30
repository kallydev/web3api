package contract

//go:generate abigen --abi ./erc20/erc20.abi --pkg erc20 --type ERC20 --out ./erc20/erc20.go
//go:generate abigen --abi ./uniswap/pool_v3.abi --pkg uniswap --type PoolV3 --out ./uniswap/pool_v3.go
