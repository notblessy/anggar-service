package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sirupsen/logrus"
)

func (h *httpService) findAllWalletHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var query model.WalletQueryInput

	if err := c.Bind(&query); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	query.UserID = session.ID

	wallets, total, err := h.walletRepo.FindAll(c.Request().Context(), query)
	if err != nil {
		logger.Errorf("Error getting wallets: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    withPaging(wallets, total, query.PageOrDefault(), query.SizeOrDefault()),
	})
}

func (h *httpService) createWalletHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var wallet model.Wallet

	if err := c.Bind(&wallet); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	wallet.UserID = session.ID

	if err := h.walletRepo.Create(c.Request().Context(), &wallet); err != nil {
		logger.Errorf("Error creating wallet: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true, Data: wallet})
}

func (h *httpService) findWalletByIDHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := c.Param("id")

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	wallet, err := h.walletRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error getting wallet: %v", err)
		return c.JSON(http.StatusNotFound, response{Message: err.Error()})
	}

	if wallet.UserID != session.ID {
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	return c.JSON(http.StatusOK, response{Success: true, Data: wallet})
}

func (h *httpService) updateWalletHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := c.Param("id")

	var wallet model.Wallet

	if err := c.Bind(&wallet); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	wallet.ID = id
	wallet.UserID = session.ID

	if err := h.walletRepo.Update(c.Request().Context(), id, wallet); err != nil {
		logger.Errorf("Error updating wallet: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true, Data: wallet})
}

func (h *httpService) deleteWalletHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := c.Param("id")

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	wallet, err := h.walletRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error getting wallet: %v", err)
		return c.JSON(http.StatusNotFound, response{Message: err.Error()})
	}

	if wallet.UserID != session.ID {
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	if err := h.walletRepo.Delete(c.Request().Context(), id); err != nil {
		logger.Errorf("Error deleting wallet: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true})
}

func (h *httpService) findWalletOptionHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	options, err := h.walletRepo.Option(c.Request().Context(), session.ID)
	if err != nil {
		logger.Errorf("Error getting wallet options: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true, Data: options})
}
