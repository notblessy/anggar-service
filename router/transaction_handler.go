package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notblessy/anggar-service/model"
	"github.com/notblessy/anggar-service/utils"
	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

func (h *httpService) findAllTransactionHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var query model.TransactionQueryInput
	err := c.Bind(&query)
	if err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	transactions, total, err := h.transactionRepo.FindAll(c.Request().Context(), query)
	if err != nil {
		logger.Errorf("Error getting transactions: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	results := make(map[string][]model.Transaction)

	for _, tx := range transactions {
		dateKey := tx.SpentAt.Format("2006-01-02")
		results[dateKey] = append(results[dateKey], tx)
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    withPaging(results, total, query.PageOrDefault(), query.SizeOrDefault()),
	})
}

func (h *httpService) createTransactionHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	var transaction model.Transaction
	err := c.Bind(&transaction)
	if err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	transaction.UserID = session.ID

	transaction.ID = ulid.Make().String()

	for i := range transaction.TransactionShares {
		transaction.TransactionShares[i].ID = ulid.Make().String()
		transaction.TransactionShares[i].TransactionID = transaction.ID
	}

	err = h.transactionRepo.Create(c.Request().Context(), &transaction)
	if err != nil {
		logger.Errorf("Error creating transaction: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    transaction,
	})
}

func (h *httpService) findTransactionByIDHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	transaction, err := h.transactionRepo.FindByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error finding transaction: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    transaction,
	})
}

func (h *httpService) updateTransactionHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	var transaction model.Transaction
	err := c.Bind(&transaction)
	if err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	transaction.UserID = session.ID

	err = h.transactionRepo.Update(c.Request().Context(), id, transaction)
	if err != nil {
		logger.Errorf("Error updating transaction: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true, Data: transaction})
}

func (h *httpService) deleteTransactionHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	id := utils.ParseID(c.Param("id"))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	if session.ID == "" {
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	err = h.transactionRepo.Delete(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Error deleting transaction: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{Success: true})
}

func (h *httpService) currentMonthSummaryHandler(c echo.Context) error {
	logger := logrus.WithField("ctx", utils.Dump(c.Request().Context()))

	session, err := authSession(c)
	if err != nil {
		logger.Errorf("Error getting session: %v", err)
		return c.JSON(http.StatusUnauthorized, response{Message: "unauthorized"})
	}

	var query model.SummaryQueryInput

	if err := c.Bind(&query); err != nil {
		logger.Errorf("Error parsing request: %v", err)
		return c.JSON(http.StatusBadRequest, response{Message: err.Error()})
	}

	query.UserID = session.ID

	summary, err := h.transactionRepo.CurrentMonthSummary(c.Request().Context(), query)
	if err != nil {
		logger.Errorf("Error getting current month summary: %v", err)
		return c.JSON(http.StatusInternalServerError, response{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, response{
		Success: true,
		Data:    summary,
	})
}
