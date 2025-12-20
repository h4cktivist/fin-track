package swagger

import _ "embed"

//go:embed fin-analytics.yaml
var finAnalyticsSpec []byte

func FinAnalyticsSpec() []byte {
	return finAnalyticsSpec
}
