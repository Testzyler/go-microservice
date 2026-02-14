package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/application/common"
	"github.com/sklinkert/go-ddd/internal/application/query"
	"github.com/sklinkert/go-ddd/internal/domain/entities"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
	"testing"
)

// MockProductRepository is a mock implementation of the ProductRepository interface
type MockProductRepository struct {
	products []*entities.ValidatedProduct
}

func (m *MockProductRepository) Create(product *entities.ValidatedProduct) (*entities.Product, error) {
	m.products = append(m.products, product)
	return &product.Product, nil
}

func (m *MockProductRepository) FindAll() ([]*entities.Product, error) {
	var products []*entities.Product
	for _, p := range m.products {
		products = append(products, &p.Product)
	}
	return products, nil
}

func (m *MockProductRepository) Update(product *entities.ValidatedProduct) (*entities.Product, error) {
	for index, p := range m.products {
		if p.Id == product.Id {
			m.products[index] = product
			return &product.Product, nil
		}
	}
	return nil, errors.New("product not found for update")
}

func (m *MockProductRepository) Delete(id uuid.UUID) error {
	for index, p := range m.products {
		if p.Id == id {
			m.products = append(m.products[:index], m.products[index+1:]...)
			return nil
		}
	}
	return errors.New("product not found for delete")
}

func (m *MockProductRepository) FindById(id uuid.UUID) (*entities.Product, error) {
	for _, p := range m.products {
		if p.Id == id {
			return &p.Product, nil
		}
		fmt.Printf("Id: mem:%s - %s\n", p.Id, id)
	}
	return nil, nil
}

// MockIdempotencyRepository is a mock implementation of the IdempotencyRepository interface
type MockIdempotencyRepository struct {
	records map[string]*entities.IdempotencyRecord
}

func NewMockIdempotencyRepository() *MockIdempotencyRepository {
	return &MockIdempotencyRepository{
		records: make(map[string]*entities.IdempotencyRecord),
	}
}

func (m *MockIdempotencyRepository) FindByKey(ctx context.Context, key string) (*entities.IdempotencyRecord, error) {
	if record, exists := m.records[key]; exists {
		return record, nil
	}
	return nil, nil
}

func (m *MockIdempotencyRepository) Create(ctx context.Context, record *entities.IdempotencyRecord) (*entities.IdempotencyRecord, error) {
	m.records[record.Key] = record
	return record, nil
}

func (m *MockIdempotencyRepository) Update(ctx context.Context, record *entities.IdempotencyRecord) (*entities.IdempotencyRecord, error) {
	m.records[record.Key] = record
	return record, nil
}

func (m *MockIdempotencyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	for key, record := range m.records {
		if record.Id == id {
			delete(m.records, key)
			break
		}
	}
	return nil
}

func TestProductService_CreateProduct(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	idempotencyRepo := NewMockIdempotencyRepository()
	service := NewProductService(productRepo, sellerRepo, idempotencyRepo)

	// Create seller
	seller := createPersistedSeller(t, sellerRepo)

	// Create product
	product := entities.NewProduct("Example", 100.0, *seller)
	productCommand := getCreateProductCommand(product)
	_, err := service.CreateProduct(productCommand)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(productRepo.products) != 1 {
		t.Errorf("Expected 1 product in productRepository, but got %d", len(productRepo.products))
	}
}

func TestProductService_GetAllProducts(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	idempotencyRepo := NewMockIdempotencyRepository()
	service := NewProductService(productRepo, sellerRepo, idempotencyRepo)

	// Create seller
	seller := createPersistedSeller(t, sellerRepo)

	// Add two products
	_, _ = service.CreateProduct(getCreateProductCommand(entities.NewProduct("Example1", 100.0, *seller)))
	_, _ = service.CreateProduct(getCreateProductCommand(entities.NewProduct("Example2", 200.0, *seller)))

	products, err := service.FindAllProducts()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if len(products.Result) != 2 {
		t.Errorf("Expected 2 products, but got %d", len(products.Result))
	}
}

func TestProductService_FindProductById(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	idempotencyRepo := NewMockIdempotencyRepository()
	service := NewProductService(productRepo, sellerRepo, idempotencyRepo)

	// Create seller
	seller := createPersistedSeller(t, sellerRepo)

	product := entities.NewProduct("Example", 100.0, *seller)
	result, err := service.CreateProduct(getCreateProductCommand(product))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	foundProduct, err := service.FindProductById(&query.GetProductByIdQuery{Id: result.Result.Id})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if foundProduct.Result.Name != "Example" {
		t.Errorf("Expected product name 'Example', but got %s", foundProduct.Result.Name)
	}

	missingProduct, err := service.FindProductById(&query.GetProductByIdQuery{Id: uuid.New()}) // some non-existent Id
	if err != nil {
		t.Errorf("Expected no error for non-existent product, but got %s", err)
	}
	if missingProduct != nil {
		t.Error("Expected nil product for non-existent Id, but got a value")
	}
}

func TestProductService_CreateProduct_IdempotencyReplay(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	idempotencyRepo := NewMockIdempotencyRepository()
	service := NewProductService(productRepo, sellerRepo, idempotencyRepo)

	key := "idp-replay-key"
	cmd := &command.CreateProductCommand{
		IdempotencyKey: key,
		Name:           "from-cache",
		Price:          99.5,
		SellerId:       uuid.New(),
	}
	expected := command.CreateProductCommandResult{
		Result: &common.ProductResult{
			Id:    uuid.New(),
			Name:  "from-cache",
			Price: 99.5,
		},
	}
	requestPayload, err := marshalJSON(cmd)
	if err != nil {
		t.Fatalf("unexpected marshal request error: %v", err)
	}
	responsePayload, err := marshalJSON(expected)
	if err != nil {
		t.Fatalf("unexpected marshal response error: %v", err)
	}

	record := entities.NewIdempotencyRecord(key, requestPayload)
	record.SetResponse(responsePayload, 201)
	_, _ = idempotencyRepo.Create(context.Background(), record)

	result, err := service.CreateProduct(cmd)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if result == nil || result.Result == nil {
		t.Fatalf("expected cached result")
	}
	if result.Result.Name != "from-cache" {
		t.Fatalf("expected cached name from-cache but got %s", result.Result.Name)
	}
	if len(productRepo.products) != 0 {
		t.Fatalf("expected repository not to be called for replay")
	}
}

func TestProductService_CreateProduct_IdempotencyInProgress(t *testing.T) {
	productRepo := &MockProductRepository{}
	sellerRepo := &MockSellerRepository{}
	idempotencyRepo := NewMockIdempotencyRepository()
	service := NewProductService(productRepo, sellerRepo, idempotencyRepo)

	key := "idp-in-progress-key"
	cmd := &command.CreateProductCommand{
		IdempotencyKey: key,
		Name:           "test",
		Price:          10,
		SellerId:       uuid.New(),
	}
	requestPayload, err := marshalJSON(cmd)
	if err != nil {
		t.Fatalf("unexpected marshal request error: %v", err)
	}

	record := entities.NewIdempotencyRecord(key, requestPayload)
	_, _ = idempotencyRepo.Create(context.Background(), record)

	_, err = service.CreateProduct(cmd)
	if err == nil {
		t.Fatalf("expected conflict error for in-progress record")
	}
	if errorx.KindOf(err) != errorx.KindConflict {
		t.Fatalf("expected conflict kind but got %v", errorx.KindOf(err))
	}
	if errorx.Code(err) != errorx.CodeIdempotencyInProgress {
		t.Fatalf("expected code %s but got %s", errorx.CodeIdempotencyInProgress, errorx.Code(err))
	}
}

func getCreateProductCommand(product *entities.Product) *command.CreateProductCommand {
	return &command.CreateProductCommand{
		Name:     product.Name,
		Price:    product.Price,
		SellerId: product.Seller.Id,
	}
}

func createPersistedSeller(t *testing.T, sellerRepo *MockSellerRepository) *entities.ValidatedSeller {
	seller := entities.NewSeller("John Doe")
	validatedSeller, err := entities.NewValidatedSeller(seller)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	_, err = sellerRepo.Create(validatedSeller)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	return validatedSeller
}
