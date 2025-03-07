package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sirupsen/logrus"
)

func (h *httpService) findScopeOverviews(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	overviews, err := h.scopeRepo.FindOverviews(c.Request().Context(), session.ID)
	if err != nil {
		logger.Errorf("Error getting overviews: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    overviews,
	})
}

func (h *httpService) findAllScopeHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var query model.ScopeQueryInput

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

	scopes, total, err := h.scopeRepo.FindAll(c.Request().Context(), query)
	if err != nil {
		logger.Errorf("Error getting scopes: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    withPaging(scopes, total, query.PageOrDefault(), query.SizeOrDefault()),
	})
}

func (h *httpService) createScopeHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var scope model.ScopeInput

	if err := c.Bind(&scope); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	scope.UserID = session.ID

	result, err := h.scopeRepo.Create(c.Request().Context(), scope)
	if err != nil {
		logger.Errorf("Error creating scope: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, response{
		Success: true,
		Data:    result,
	})
}

func (h *httpService) findScopeByIDHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	scope, err := h.scopeRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error getting scope: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	if scope.UserID != session.ID {
		return c.JSON(http.StatusForbidden, response{Message: "forbidden"})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    scope,
	})
}

func (h *httpService) updateScopeHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	var scope model.Scope

	if err := c.Bind(&scope); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	scope.UserID = session.ID

	if err := h.scopeRepo.Update(c.Request().Context(), id, scope); err != nil {
		logger.Errorf("Error updating scope: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true})
}

func (h *httpService) deleteScopeHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	scope, err := h.scopeRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error getting scope: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	if scope.UserID != session.ID {
		return c.JSON(http.StatusForbidden, response{Message: "forbidden"})
	}

	if err := h.scopeRepo.Delete(c.Request().Context(), id); err != nil {
		logger.Errorf("Error deleting scope: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true})
}
