package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/application/common"
	"github.com/sklinkert/go-ddd/internal/application/interfaces"
	"github.com/sklinkert/go-ddd/internal/application/mapper"
	"github.com/sklinkert/go-ddd/internal/application/query"
	"github.com/sklinkert/go-ddd/internal/domain/entities"
	"github.com/sklinkert/go-ddd/internal/domain/repositories"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

type ProductService struct {
	productRepository repositories.ProductRepository
	sellerRepository  repositories.SellerRepository
	idempotencyRepo   repositories.IdempotencyRepository
}

func NewProductService(
	productRepository repositories.ProductRepository,
	sellerRepository repositories.SellerRepository,
	idempotencyRepo repositories.IdempotencyRepository,
) interfaces.ProductService {
	return &ProductService{
		productRepository: productRepository,
		sellerRepository:  sellerRepository,
		idempotencyRepo:   idempotencyRepo,
	}
}

func (s *ProductService) CreateProduct(productCommand *command.CreateProductCommand) (*command.CreateProductCommandResult, error) {
	ctx := context.Background()

	var replayResult command.CreateProductCommandResult
	idempotencyRecord, replayed, err := beginIdempotentOperation(
		ctx,
		s.idempotencyRepo,
		productCommand.IdempotencyKey,
		productCommand,
		&replayResult,
	)
	if err != nil {
		return nil, err
	}
	if replayed {
		return &replayResult, nil
	}

	writeCompleted := false
	defer func() {
		if !writeCompleted {
			releaseIdempotentReservation(ctx, s.idempotencyRepo, idempotencyRecord)
		}
	}()

	storedSeller, err := s.sellerRepository.FindById(productCommand.SellerId)
	if err != nil {
		return nil, err
	}

	if storedSeller == nil {
		return nil, errorx.NotFound(errorx.CodeSellerNotFound, errorx.MessageSellerNotFound)
	}

	validatedSeller, err := entities.NewValidatedSeller(storedSeller)
	if err != nil {
		return nil, err
	}

	var newProduct = entities.NewProduct(
		productCommand.Name,
		productCommand.Price,
		*validatedSeller,
	)

	validatedProduct, err := entities.NewValidatedProduct(newProduct)
	if err != nil {
		return nil, err
	}

	_, err = s.productRepository.Create(validatedProduct)
	if err != nil {
		return nil, err
	}

	result := command.CreateProductCommandResult{
		Result: mapper.NewProductResultFromValidatedEntity(validatedProduct),
	}

	finalizeIdempotentOperation(ctx, s.idempotencyRepo, idempotencyRecord, result, 201)
	writeCompleted = true

	return &result, nil
}

func (s *ProductService) FindAllProducts() (*query.GetAllProductsQueryResult, error) {
	storedProducts, err := s.productRepository.FindAll()
	if err != nil {
		return nil, err
	}

	queryListResult := make([]*common.ProductResult, 0, len(storedProducts))
	for _, product := range storedProducts {
		queryListResult = append(queryListResult, mapper.NewProductResultFromEntity(product))
	}

	return query.NewProductQueryListResult(queryListResult), nil
}

func (s *ProductService) FindProductById(productQuery *query.GetProductByIdQuery) (*query.GetProductByIdQueryResult, error) {
	if productQuery == nil {
		return nil, errorx.InvalidArgument(errorx.CodeProductQueryRequired, errorx.MessageProductQueryRequired)
	}
	if productQuery.Id == uuid.Nil {
		return nil, errorx.InvalidArgument(errorx.CodeProductIDRequired, errorx.MessageProductIDRequired)
	}

	storedProduct, err := s.productRepository.FindById(productQuery.Id)
	if err != nil {
		return nil, err
	}
	if storedProduct == nil {
		return nil, nil
	}

	return query.NewProductQueryResult(mapper.NewProductResultFromEntity(storedProduct)), nil
}

func (s *ProductService) UpdateProduct(productCommand *command.UpdateProductCommand) (*command.UpdateProductCommandResult, error) {
	ctx := context.Background()

	var replayResult command.UpdateProductCommandResult
	idempotencyRecord, replayed, err := beginIdempotentOperation(
		ctx,
		s.idempotencyRepo,
		productCommand.IdempotencyKey,
		productCommand,
		&replayResult,
	)
	if err != nil {
		return nil, err
	}
	if replayed {
		return &replayResult, nil
	}

	writeCompleted := false
	defer func() {
		if !writeCompleted {
			releaseIdempotentReservation(ctx, s.idempotencyRepo, idempotencyRecord)
		}
	}()

	// Find existing product
	existingProduct, err := s.productRepository.FindById(productCommand.Id)
	if err != nil {
		return nil, err
	}

	if existingProduct == nil {
		return nil, errorx.NotFound(errorx.CodeProductNotFound, errorx.MessageProductNotFound)
	}

	// Find seller if different
	if productCommand.SellerId != existingProduct.Seller.Id {
		storedSeller, err := s.sellerRepository.FindById(productCommand.SellerId)
		if err != nil {
			return nil, err
		}

		if storedSeller == nil {
			return nil, errorx.NotFound(errorx.CodeSellerNotFound, errorx.MessageSellerNotFound)
		}

		validatedSeller, err := entities.NewValidatedSeller(storedSeller)
		if err != nil {
			return nil, err
		}
		existingProduct.Seller = validatedSeller.Seller
	}

	// Update product fields
	if err := existingProduct.UpdateName(productCommand.Name); err != nil {
		return nil, err
	}

	if err := existingProduct.UpdatePrice(productCommand.Price); err != nil {
		return nil, err
	}

	validatedProduct, err := entities.NewValidatedProduct(existingProduct)
	if err != nil {
		return nil, err
	}

	_, err = s.productRepository.Update(validatedProduct)
	if err != nil {
		return nil, err
	}

	result := command.UpdateProductCommandResult{
		Result: mapper.NewProductResultFromValidatedEntity(validatedProduct),
	}

	finalizeIdempotentOperation(ctx, s.idempotencyRepo, idempotencyRecord, result, 200)
	writeCompleted = true

	return &result, nil
}
func (s *ProductService) DeleteProduct(productCommand *command.DeleteProductCommand) (*command.DeleteProductCommandResult, error) {
	ctx := context.Background()

	var replayResult command.DeleteProductCommandResult
	idempotencyRecord, replayed, err := beginIdempotentOperation(
		ctx,
		s.idempotencyRepo,
		productCommand.IdempotencyKey,
		productCommand,
		&replayResult,
	)
	if err != nil {
		return nil, err
	}
	if replayed {
		return &replayResult, nil
	}

	writeCompleted := false
	defer func() {
		if !writeCompleted {
			releaseIdempotentReservation(ctx, s.idempotencyRepo, idempotencyRecord)
		}
	}()

	// Check if product exists
	existingProduct, err := s.productRepository.FindById(productCommand.Id)
	if err != nil {
		return nil, err
	}

	if existingProduct == nil {
		return nil, errorx.NotFound(errorx.CodeProductNotFound, errorx.MessageProductNotFound)
	}

	// Delete product
	err = s.productRepository.Delete(productCommand.Id)
	if err != nil {
		return nil, err
	}

	result := command.DeleteProductCommandResult{
		Success: true,
	}

	finalizeIdempotentOperation(ctx, s.idempotencyRepo, idempotencyRecord, result, 204)
	writeCompleted = true

	return &result, nil
}
