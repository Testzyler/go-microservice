package response

import (
	"reflect"
	"strings"
	"testing"
)

func TestResponseDTO_JSONTagsUseCamelCase(t *testing.T) {
	responseTypes := []any{
		SellerResponse{},
		ListSellersResponse{},
		ProductResponse{},
		ListProductsResponse{},
	}

	for _, dto := range responseTypes {
		dtoType := reflect.TypeOf(dto)
		for i := 0; i < dtoType.NumField(); i++ {
			field := dtoType.Field(i)
			tag := field.Tag.Get("json")
			tagName := strings.Split(tag, ",")[0]
			if tagName == "" || tagName == "-" {
				continue
			}
			if strings.Contains(tagName, "_") {
				t.Fatalf("%s.%s has non-camelCase json tag: %q", dtoType.Name(), field.Name, tagName)
			}
		}
	}
}

