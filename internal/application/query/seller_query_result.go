package query

import "github.com/sklinkert/go-ddd/internal/application/common"

type SellerQueryResult struct {
	Result *common.SellerResult
}

func NewSellerQueryResult(result *common.SellerResult) *SellerQueryResult {
	return &SellerQueryResult{Result: result}
}

type SellerQueryListResult struct {
	Result []*common.SellerResult
}

func NewSellerQueryListResult(result []*common.SellerResult) *SellerQueryListResult {
	if len(result) == 0 {
		return &SellerQueryListResult{Result: []*common.SellerResult{}}
	}

	copied := make([]*common.SellerResult, len(result))
	copy(copied, result)

	return &SellerQueryListResult{Result: copied}
}
