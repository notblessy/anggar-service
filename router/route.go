package router

import (
	"github.com/labstack/echo/v4"
	"github.com/notblessy/anggar-service/model"
	"gorm.io/gorm"
)

type httpService struct {
	db              *gorm.DB
	userRepo        model.UserRepository
	walletRepo      model.WalletRepository
	scopeRepo       model.ScopeRepository
	transactionRepo model.TransactionRepository
}

func NewHTTPService() *httpService {
	return &httpService{}
}

func (h *httpService) RegisterPostgres(db *gorm.DB) {
	h.db = db
}

func (h *httpService) RegisterUserRepository(repo model.UserRepository) {
	h.userRepo = repo
}

func (h *httpService) RegisterWalletRepository(repo model.WalletRepository) {
	h.walletRepo = repo
}

func (h *httpService) RegisterScopeRepository(repo model.ScopeRepository) {
	h.scopeRepo = repo
}

func (h *httpService) RegisterTransactionRepository(repo model.TransactionRepository) {
	h.transactionRepo = repo
}

func (h *httpService) Router(e *echo.Echo) {
	e.GET("/ping", h.ping)
	e.GET("/health", h.health)

	v1 := e.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.POST("/google", h.loginWithGoogleHandler)

	protected := v1.Group("")
	protected.Use(NewJWTMiddleware().ValidateJWT)
	users := protected.Group("/users")
	users.GET("/me", h.profileHandler)
	users.GET("/options", h.findUserOptionHandler)

	scope := protected.Group("/scopes")
	scope.GET("/overviews", h.findScopeOverviews)
	scope.GET("", h.findAllScopeHandler)
	scope.POST("", h.createScopeHandler)
	scope.GET("/:id", h.findScopeByIDHandler)
	scope.PUT("/:id", h.updateScopeHandler)
	scope.DELETE("/:id", h.deleteScopeHandler)

	wallet := protected.Group("/wallets")
	wallet.GET("", h.findAllWalletHandler)
	wallet.POST("", h.createWalletHandler)
	wallet.GET("/:id", h.findWalletByIDHandler)
	wallet.PUT("/:id", h.updateWalletHandler)
	wallet.DELETE("/:id", h.deleteWalletHandler)
	wallet.GET("/options", h.findWalletOptionHandler)

	transaction := protected.Group("/transactions")
	transaction.GET("", h.findAllTransactionHandler)
	transaction.POST("", h.createTransactionHandler)
	transaction.GET("/:id", h.findTransactionByIDHandler)
	transaction.PUT("/:id", h.updateTransactionHandler)
	transaction.DELETE("/:id", h.deleteTransactionHandler)
}

func (h *httpService) ping(c echo.Context) error {
	return c.JSON(200, response{Data: "pong"})
}

func (h *httpService) health(c echo.Context) error {
	err := h.db.Raw("SELECT 1").Error
	if err != nil {
		return c.JSON(500, response{Message: err.Error()})
	}

	return c.JSON(200, "OK")
}
