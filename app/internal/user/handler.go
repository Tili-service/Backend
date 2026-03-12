package user

import (
	"net/http"
	"strconv"

	"tili/app/internal/middleware"
	"tili/app/internal/token"

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
		userRoutes.POST("", h.Create)      // POST /users
		userRoutes.POST("/login", h.login) // POST /users/login

		protected := userRoutes.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/me", h.me) // GET /users/me

			managerRoutes := protected.Group("")
			managerRoutes.Use(middleware.LevelAccessRequired(token.Manager))
			{
				managerRoutes.PUT("/:id", h.Update) // PUT /users/:id
			}

			adminRoutes := protected.Group("")
			adminRoutes.Use(middleware.LevelAccessRequired(token.Admin))
			{
				adminRoutes.GET("", h.GetAll)        // GET /users
				adminRoutes.GET("/:id", h.GetByID)   // GET /users/:id
				adminRoutes.DELETE("/:id", h.Delete) // DELETE /users/:id
			}

			superAdminRoutes := protected.Group("")
			superAdminRoutes.Use(middleware.LevelAccessRequired(token.SuperAdmin))
			{
			}
		}
	}
}

// Create adds a new user
// @Summary      Create a user
// @Description  Creates a new user in the system
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body      CreateUserInput true "User payload"
// @Success      201  {object}  User
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users [post]
func (h *Handler) Create(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}

// GetAll retrieves the list of all users
// @Summary      List users
// @Description  Retrieves the complete list of users
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   User
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users [get]
func (h *Handler) GetAll(c *gin.Context) {
	users, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetByID retrieves a user by their ID
// @Summary      Retrieve a user
// @Description  Retrieves the details of a user using their ID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  User
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /users/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	u, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Update modifies an existing user
// @Summary      Update a user
// @Description  Modifies the information of an existing user via their ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int             true "User ID"
// @Param        body body      UpdateUserInput true "User update payload"
// @Success      200  {object}  User
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

// Delete removes a user
// @Summary      Delete a user
// @Description  Deletes a user from the system via their ID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Login authenticates a user and returns a JWT token
// @Summary      Login user
// @Description  Authenticates a user with email and password, returns a JWT token
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body      LoginInput true "Login credentials"
// @Success      200  {object}  map[string]interface{} "Token response"
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/login [post]
func (h *Handler) login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Login(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	tokenString, err := token.Create(u.UserID, u.Name, u.Email, u.AccessLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Me retrieves the current authenticated user's information
// @Summary      Get current user
// @Description  Retrieves the profile information of the currently authenticated user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  User
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /users/me [get]
func (h *Handler) me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, ok := userID.(int)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}
	u, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	c.JSON(http.StatusOK, u)
}
