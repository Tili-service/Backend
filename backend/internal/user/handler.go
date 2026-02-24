package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("", h.Create)       // POST /users
		userRoutes.GET("", h.GetAll)        // GET /users
		userRoutes.GET("/me", h.GetMe)      // GET /users/me
		userRoutes.GET("/:id", h.GetByID)   // GET /users/:id
		userRoutes.PUT("/:id", h.Update)    // PUT /users/:id
		userRoutes.DELETE("/:id", h.Delete) // DELETE /users/:id
	}
}

// Create adds a new user
// @Summary      Create a user
// @Description  Creates a new user in the system
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      201  {object}  map[string]interface{}
// @Router       /users [post]
func (h *Handler) Create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "API call USER POST !"})
}

// GetAll retrieves the list of all users
// @Summary      List users
// @Description  Retrieves the complete list of users
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /users [get]
func (h *Handler) GetAll(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "API call USER GET ALL !"})
}

// GetMe retrieves the currently logged-in user
// @Summary      Current user profile
// @Description  Retrieves the information of the user making the request
// @Tags         users
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "API call USER GET ME !"})
}

// GetByID retrieves a user by their ID
// @Summary      Retrieve a user
// @Description  Retrieves the details of a user using their ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "API call USER GET " + id})
}

// Update modifies an existing user
// @Summary      Update a user
// @Description  Modifies the information of an existing user via their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "API call USER PUT " + id + " !"})
}

// Delete removes a user
// @Summary      Delete a user
// @Description  Deletes a user from the system via their ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "API call USER DELETE " + id + " !"})
}