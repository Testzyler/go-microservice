package query

import "github.com/sklinkert/go-ddd/internal/application/common"

type ProductQueryResult struct {
	Result *common.ProductResult
}

func NewProductQueryResult(result *common.ProductResult) *ProductQueryResult {
	return &ProductQueryResult{Result: result}
}

type ProductQueryListResult struct {
	Result []*common.ProductResult
}

func NewProductQueryListResult(result []*common.ProductResult) *ProductQueryListResult {
	if len(result) == 0 {
		return &ProductQueryListResult{Result: []*common.ProductResult{}}
	}

	copied := make([]*common.ProductResult, len(result))
	copy(copied, result)

	return &ProductQueryListResult{Result: copied}
}
