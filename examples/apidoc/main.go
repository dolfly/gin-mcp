package main

import (
	"net/http"
	"strconv"

	server "github.com/ckanthony/gin-mcp"
	"github.com/gin-gonic/gin"
)

// Product represents a product in our store
type Product struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Price       float64  `json:"price"`
	Tags        []string `json:"tags,omitempty"`
	IsEnabled   bool     `json:"is_enabled"`
}

// In-memory store
var (
	products = make(map[int]*Product)
	nextID   = 1
)

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// Register API routes with apidoc style comments
	registerRoutes(r)

	// Initialize and configure MCP server
	configureMCP(r)

	// Start the server
	r.Run(":8080")
}

func registerRoutes(r *gin.Engine) {
	r.GET("/products", listProducts)
	r.GET("/products/:id", getProduct)
	r.POST("/products", createProduct)
	r.DELETE("/products/:id", deleteProduct)
}

func configureMCP(r *gin.Engine) {
	mcp := server.New(r, &server.Config{
		Name:        "Product API (apidoc style)",
		Description: "RESTful API for managing products with apidoc annotations",
		BaseURL:     "http://localhost:8080",
	})

	// Mount MCP endpoint
	mcp.Mount("/mcp")
}

// @api {get} /products List all products
// @apiName ListProducts
// @apiGroup Product
// @apiDescription Retrieve a paginated list of products with optional filtering
// @apiParam {Number} [page=1] Page number for pagination
// @apiParam {Number} [limit=10] Number of items per page
// @apiSuccess {Object[]} products List of products
// @apiSuccess {Number} products.id Product ID
// @apiSuccess {String} products.name Product name
func listProducts(c *gin.Context) {
	page := 1
	limit := 10

	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	var result []*Product
	for _, product := range products {
		result = append(result, product)
	}

	start := (page - 1) * limit
	if start >= len(result) {
		c.JSON(http.StatusOK, gin.H{"products": []*Product{}, "total": len(result)})
		return
	}

	end := start + limit
	if end > len(result) {
		end = len(result)
	}

	c.JSON(http.StatusOK, gin.H{
		"products": result[start:end],
		"total":    len(result),
	})
}

// @api {get} /products/:id Get product details
// @apiName GetProduct
// @apiGroup Product
// @apiDescription Get detailed information about a specific product
// @apiParam {Number} id Product unique ID
// @apiSuccess {Number} id Product ID
// @apiSuccess {String} name Product name
// @apiSuccess {String} description Product description
// @apiSuccess {Number} price Product price
// @apiError ProductNotFound Product was not found
func getProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if product, exists := products[id]; exists {
		c.JSON(http.StatusOK, product)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
}

// @api {post} /products Create new product
// @apiName CreateProduct
// @apiGroup Product
// @apiDescription Create a new product in the catalog
// @apiParam {String} name Product name (required)
// @apiParam {String} [description] Product description
// @apiParam {Number} price Product price in USD (required)
// @apiParam {String[]} [tags] Product categories
// @apiParam {Boolean} is_enabled Whether product is available (required)
// @apiSuccess {Number} id Created product ID
// @apiSuccess {String} name Product name
func createProduct(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.ID = nextID
	nextID++
	products[product.ID] = &product
	c.JSON(http.StatusCreated, product)
}

// @api {delete} /products/:id Delete product
// @apiName DeleteProduct
// @apiGroup Product
// @apiDescription Permanently remove a product from the catalog
// @apiParam {Number} id Product unique ID
// @apiSuccess (204) {Void} null No content
// @apiError ProductNotFound Product was not found
func deleteProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if _, exists := products[id]; exists {
		delete(products, id)
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
}
