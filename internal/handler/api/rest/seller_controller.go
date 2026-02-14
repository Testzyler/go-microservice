package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sklinkert/go-ddd/internal/application/command"
	"github.com/sklinkert/go-ddd/internal/application/interfaces"
	"github.com/sklinkert/go-ddd/internal/application/query"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/mapper"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/request"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/httpx"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

type SellerController struct {
	service interfaces.SellerService
}

func NewSellerController(app fiber.Router, service interfaces.SellerService) *SellerController {
	controller := &SellerController{
		service: service,
	}

	app.Post("/api/v1/sellers", controller.CreateSellerController)
	app.Get("/api/v1/sellers", controller.GetAllSellersController)
	app.Get("/api/v1/sellers/:id", controller.GetSellerByIdController)
	app.Put("/api/v1/sellers", controller.PutSellerController)
	app.Delete("/api/v1/sellers/:id", controller.DeleteSellerController)

	return controller
}

func (sc *SellerController) CreateSellerController(c *fiber.Ctx) error {
	var createSellerRequest request.CreateSellerRequest

	if err := httpx.BindAndValidate(c, &createSellerRequest); err != nil {
		return httpx.WriteError(c, err)
	}

	sellerCommand, err := createSellerRequest.ToCreateSellerCommand()
	if err != nil {
		return httpx.WriteError(c, err)
	}

	commandResult, err := sc.service.CreateSeller(sellerCommand)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	response := mapper.ToSellerResponse(commandResult.Result)

	return httpx.Created(c, response)
}

func (sc *SellerController) GetAllSellersController(c *fiber.Ctx) error {
	sellers, err := sc.service.FindAllSellers()
	if err != nil {
		return httpx.WriteError(c, err)
	}

	response := mapper.ToSellerListResponse(sellers.Result)

	return httpx.OK(c, response)
}

func (sc *SellerController) GetSellerByIdController(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id", errorx.CodeSellerInvalidID, errorx.MessageSellerInvalidID)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	seller, err := sc.service.FindSellerById(query.NewGetSellerByIdQuery(id))
	if err != nil {
		return httpx.WriteError(c, err)
	}

	if seller == nil {
		return httpx.WriteError(c, errorx.NotFound(errorx.CodeSellerNotFound, errorx.MessageSellerNotFound))
	}

	response := mapper.ToSellerResponse(seller.Result)

	return httpx.OK(c, response)
}

func (sc *SellerController) PutSellerController(c *fiber.Ctx) error {
	var updateSellerRequest request.UpdateSellerRequest

	if err := httpx.BindAndValidate(c, &updateSellerRequest); err != nil {
		return httpx.WriteError(c, err)
	}

	updateSellerCommand, err := updateSellerRequest.ToUpdateSellerCommand()
	if err != nil {
		return httpx.WriteError(c, err)
	}

	commandResult, err := sc.service.UpdateSeller(updateSellerCommand)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	response := mapper.ToSellerResponse(commandResult.Result)

	return httpx.OK(c, response)
}

func (sc *SellerController) DeleteSellerController(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id", errorx.CodeSellerInvalidID, errorx.MessageSellerInvalidID)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	_, err = sc.service.DeleteSeller(&command.DeleteSellerCommand{Id: id})
	if err != nil {
		return httpx.WriteError(c, err)
	}

	return httpx.NoContent(c)
}
