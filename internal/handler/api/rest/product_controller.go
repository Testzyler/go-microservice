package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sklinkert/go-ddd/internal/application/interfaces"
	"github.com/sklinkert/go-ddd/internal/application/query"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/mapper"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/dto/request"
	"github.com/sklinkert/go-ddd/internal/handler/api/rest/httpx"
	"github.com/sklinkert/go-ddd/internal/infrastructure/pkg/errorx"
)

type ProductController struct {
	service interfaces.ProductService
}

func NewProductController(app fiber.Router, service interfaces.ProductService) *ProductController {
	controller := &ProductController{
		service: service,
	}

	app.Post("/api/v1/products", controller.CreateProductController)
	app.Get("/api/v1/products", controller.GetAllProductsController)
	app.Get("/api/v1/products/:id", controller.GetProductByIdController)

	return controller
}

func (pc *ProductController) CreateProductController(c *fiber.Ctx) error {
	var createProductRequest request.CreateProductRequest

	if err := httpx.BindAndValidate(c, &createProductRequest); err != nil {
		return httpx.WriteError(c, err)
	}

	productCommand, err := createProductRequest.ToCreateProductCommand()
	if err != nil {
		return httpx.WriteError(c, err)
	}

	result, err := pc.service.CreateProduct(productCommand)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	response := mapper.ToProductResponse(result.Result)

	return httpx.Created(c, response)
}

func (pc *ProductController) GetAllProductsController(c *fiber.Ctx) error {
	products, err := pc.service.FindAllProducts()
	if err != nil {
		return httpx.WriteError(c, err)
	}

	response := mapper.ToProductListResponse(products.Result)

	return httpx.OK(c, response)
}

func (pc *ProductController) GetProductByIdController(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id", errorx.CodeProductInvalidID, errorx.MessageProductInvalidID)
	if err != nil {
		return httpx.WriteError(c, err)
	}

	product, err := pc.service.FindProductById(query.NewGetProductByIdQuery(id))
	if err != nil {
		return httpx.WriteError(c, err)
	}

	if product == nil {
		return httpx.WriteError(c, errorx.NotFound(errorx.CodeProductNotFound, errorx.MessageProductNotFound))
	}

	response := mapper.ToProductResponse(product.Result)

	return httpx.OK(c, response)
}
