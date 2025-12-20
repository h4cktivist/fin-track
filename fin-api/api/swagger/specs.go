package swagger

import _ "embed"

//go:embed fin-api.yaml
var finAPISpec []byte

func FinAPISpec() []byte {
	return finAPISpec
}
