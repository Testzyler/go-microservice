package request

import (
	"reflect"
	"strings"
	"testing"
)

func TestRequestDTO_JSONTagsUseCamelCase(t *testing.T) {
	requestTypes := []any{
		CreateSellerRequest{},
		UpdateSellerRequest{},
		CreateProductRequest{},
	}

	for _, dto := range requestTypes {
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

