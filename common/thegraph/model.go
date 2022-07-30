package thegraph

import "github.com/hasura/go-graphql-client"

type UniswapPoolDayData struct {
	Date             graphql.Int    `graphql:"date"`
	Volume           graphql.String `graphql:"volumeUSD"`
	TotalValueLocked graphql.String `graphql:"tvlUSD"`
	Fees             graphql.String `graphql:"feesUSD"`
}
