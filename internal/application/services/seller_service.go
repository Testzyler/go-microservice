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

type SellerService struct {
	repo            repositories.SellerRepository
	idempotencyRepo repositories.IdempotencyRepository
}

// NewSellerService - Constructor for the service
func NewSellerService(repo repositories.SellerRepository, idempotencyRepo repositories.IdempotencyRepository) interfaces.SellerService {
	return &SellerService{
		repo:            repo,
		idempotencyRepo: idempotencyRepo,
	}
}

// CreateSeller saves a new seller
func (s *SellerService) CreateSeller(sellerCommand *command.CreateSellerCommand) (*command.CreateSellerCommandResult, error) {
	ctx := context.Background()

	var replayResult command.CreateSellerCommandResult
	idempotencyRecord, replayed, err := beginIdempotentOperation(
		ctx,
		s.idempotencyRepo,
		sellerCommand.IdempotencyKey,
		sellerCommand,
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

	var newSeller = entities.NewSeller(sellerCommand.Name)

	validatedSeller, err := entities.NewValidatedSeller(newSeller)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Create(validatedSeller)
	if err != nil {
		return nil, err
	}

	result := command.CreateSellerCommandResult{
		Result: mapper.NewSellerResultFromValidatedEntity(validatedSeller),
	}

	finalizeIdempotentOperation(ctx, s.idempotencyRepo, idempotencyRecord, result, 201)
	writeCompleted = true

	return &result, nil
}

// FindAllSellers fetches all sellers
func (s *SellerService) FindAllSellers() (*query.GetAllSellersQueryResult, error) {
	storedSellers, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	queryResult := make([]*common.SellerResult, 0, len(storedSellers))
	for _, seller := range storedSellers {
		queryResult = append(queryResult, mapper.NewSellerResultFromEntity(seller))
	}

	return query.NewSellerQueryListResult(queryResult), nil
}

// FindSellerById fetches a specific seller by Id
func (s *SellerService) FindSellerById(sellerQuery *query.GetSellerByIdQuery) (*query.GetSellerByIdQueryResult, error) {
	if sellerQuery == nil {
		return nil, errorx.InvalidArgument(errorx.CodeSellerQueryRequired, errorx.MessageSellerQueryRequired)
	}
	if sellerQuery.Id == uuid.Nil {
		return nil, errorx.InvalidArgument(errorx.CodeSellerIDRequired, errorx.MessageSellerIDRequired)
	}

	storedSeller, err := s.repo.FindById(sellerQuery.Id)
	if err != nil {
		return nil, err
	}
	if storedSeller == nil {
		return nil, nil
	}

	return query.NewSellerQueryResult(mapper.NewSellerResultFromEntity(storedSeller)), nil
}

// UpdateSeller updates a seller
func (s *SellerService) UpdateSeller(updateCommand *command.UpdateSellerCommand) (*command.UpdateSellerCommandResult, error) {
	ctx := context.Background()

	var replayResult command.UpdateSellerCommandResult
	idempotencyRecord, replayed, err := beginIdempotentOperation(
		ctx,
		s.idempotencyRepo,
		updateCommand.IdempotencyKey,
		updateCommand,
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

	seller, err := s.repo.FindById(updateCommand.Id)
	if err != nil {
		return nil, err
	}

	if seller == nil {
		return nil, errorx.NotFound(errorx.CodeSellerNotFound, errorx.MessageSellerNotFound)
	}

	if err := seller.UpdateName(updateCommand.Name); err != nil {
		return nil, err
	}

	validatedUpdatedSeller, err := entities.NewValidatedSeller(seller)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.Update(validatedUpdatedSeller)
	if err != nil {
		return nil, err
	}

	result := command.UpdateSellerCommandResult{
		Result: mapper.NewSellerResultFromEntity(seller),
	}

	finalizeIdempotentOperation(ctx, s.idempotencyRepo, idempotencyRecord, result, 200)
	writeCompleted = true

	return &result, nil
}

func (s *SellerService) DeleteSeller(sellerCommand *command.DeleteSellerCommand) (*command.DeleteSellerCommandResult, error) {
	ctx := context.Background()

	var replayResult command.DeleteSellerCommandResult
	idempotencyRecord, replayed, err := beginIdempotentOperation(
		ctx,
		s.idempotencyRepo,
		sellerCommand.IdempotencyKey,
		sellerCommand,
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

	// Check if seller exists
	existingSeller, err := s.repo.FindById(sellerCommand.Id)
	if err != nil {
		return nil, err
	}

	if existingSeller == nil {
		return nil, errorx.NotFound(errorx.CodeSellerNotFound, errorx.MessageSellerNotFound)
	}

	// Delete seller
	err = s.repo.Delete(sellerCommand.Id)
	if err != nil {
		return nil, err
	}

	result := command.DeleteSellerCommandResult{
		Success: true,
	}

	finalizeIdempotentOperation(ctx, s.idempotencyRepo, idempotencyRecord, result, 204)
	writeCompleted = true

	return &result, nil
}
