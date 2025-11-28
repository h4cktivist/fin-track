package swagger

import _ "embed"

//go:embed fin-api.yaml
var finAPISpec []byte

//go:embed fin-analytics.yaml
var finAnalyticsSpec []byte

func FinAPISpec() []byte {
	return finAPISpec
}

func FinAnalyticsSpec() []byte {
	return finAnalyticsSpec
}
