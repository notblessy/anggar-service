package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sirupsen/logrus"
)

func (h *httpService) findBudgetOverviews(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	overviews, err := h.budgetRepo.FindOverviews(c.Request().Context(), session.ID)
	if err != nil {
		logger.Errorf("Error getting overviews: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    overviews,
	})
}

func (h *httpService) findAllBudgetHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var query model.BudgetQueryInput

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

	budgets, total, err := h.budgetRepo.FindAll(c.Request().Context(), query)
	if err != nil {
		logger.Errorf("Error getting budgets: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    withPaging(budgets, total, query.PageOrDefault(), query.SizeOrDefault()),
	})
}

func (h *httpService) createBudgetHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var budget model.Budget

	if err := c.Bind(&budget); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	budget.UserID = session.ID

	if err := h.budgetRepo.Create(c.Request().Context(), &budget); err != nil {
		logger.Errorf("Error creating budget: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, response{
		Success: true,
		Data:    budget,
	})
}

func (h *httpService) findBudgetByIDHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	budget, err := h.budgetRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error getting budget: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	if budget.UserID != session.ID {
		return c.JSON(http.StatusForbidden, response{Message: "forbidden"})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    budget,
	})
}

func (h *httpService) updateBudgetHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	var budget model.Budget

	if err := c.Bind(&budget); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	budget.UserID = session.ID

	if err := h.budgetRepo.Update(c.Request().Context(), id, budget); err != nil {
		logger.Errorf("Error updating budget: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true})
}

func (h *httpService) deleteBudgetHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	budget, err := h.budgetRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error getting budget: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	if budget.UserID != session.ID {
		return c.JSON(http.StatusForbidden, response{Message: "forbidden"})
	}

	if err := h.budgetRepo.Delete(c.Request().Context(), id); err != nil {
		logger.Errorf("Error deleting budget: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true})
}
