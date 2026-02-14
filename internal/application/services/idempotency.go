package services

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sklinkert/go-ddd/internal/domain/entities"
	"github.com/sklinkert/go-ddd/internal/domain/repositories"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

func beginIdempotentOperation(
	ctx context.Context,
	repo repositories.IdempotencyRepository,
	key string,
	request any,
	replayTarget any,
) (*entities.IdempotencyRecord, bool, error) {
	if strings.TrimSpace(key) == "" {
		return nil, false, nil
	}

	requestPayload, err := marshalJSON(request)
	if err != nil {
		return nil, false, err
	}

	existing, err := repo.FindByKey(ctx, key)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return handleExistingIdempotencyRecord(existing, requestPayload, replayTarget)
	}

	record := entities.NewIdempotencyRecord(key, requestPayload)
	created, err := repo.Create(ctx, record)
	if err != nil {
		// Handle race: another request may create this key between FindByKey and Create.
		afterRaceRecord, findErr := repo.FindByKey(ctx, key)
		if findErr == nil && afterRaceRecord != nil {
			return handleExistingIdempotencyRecord(afterRaceRecord, requestPayload, replayTarget)
		}
		return nil, false, err
	}

	return created, false, nil
}

func handleExistingIdempotencyRecord(
	record *entities.IdempotencyRecord,
	requestPayload string,
	replayTarget any,
) (*entities.IdempotencyRecord, bool, error) {
	if strings.TrimSpace(record.Request) != "" && record.Request != requestPayload {
		return nil, false, errorx.Conflict(errorx.CodeIdempotencyKeyReused, errorx.MessageIdempotencyKeyReused)
	}

	if strings.TrimSpace(record.Response) == "" || record.StatusCode == 0 {
		return nil, false, errorx.Conflict(errorx.CodeIdempotencyInProgress, errorx.MessageIdempotencyInProgress)
	}

	if replayTarget != nil {
		if err := json.Unmarshal([]byte(record.Response), replayTarget); err != nil {
			return nil, false, errorx.Internal(errorx.CodeInternal, err)
		}
	}

	return nil, true, nil
}

func finalizeIdempotentOperation(
	ctx context.Context,
	repo repositories.IdempotencyRepository,
	record *entities.IdempotencyRecord,
	response any,
	statusCode int,
) {
	if record == nil {
		return
	}

	responsePayload, err := marshalJSON(response)
	if err != nil {
		_ = repo.Delete(ctx, record.Id)
		return
	}

	record.SetResponse(responsePayload, statusCode)
	if _, err = repo.Update(ctx, record); err != nil {
		_ = repo.Delete(ctx, record.Id)
	}
}

func releaseIdempotentReservation(ctx context.Context, repo repositories.IdempotencyRepository, record *entities.IdempotencyRecord) {
	if record == nil {
		return
	}
	_ = repo.Delete(ctx, record.Id)
}

func marshalJSON(payload any) (string, error) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
