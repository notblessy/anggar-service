package repository

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notblessy/anggar-service/model"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var waitingRooms = make(map[int64]bool)

type capitalBotRepository struct {
	db     *gorm.DB
	openAi model.RecognizerRepository
	bot    *tgbotapi.BotAPI
}

func NewCapitalBotRepository(db *gorm.DB, bot *tgbotapi.BotAPI, openAi model.RecognizerRepository) *capitalBotRepository {
	return &capitalBotRepository{
		db:     db,
		bot:    bot,
		openAi: openAi,
	}
}

func (c *capitalBotRepository) ListenMessage(ctx context.Context) error {
	if c == nil {
		return fmt.Errorf("bot repository is missing")
	}

	if c.bot == nil {
		return fmt.Errorf("no telegram bot instance")
	}

	updates := c.bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Timeout: 60,
	})

	for update := range updates {
		if update.Message != nil {
			c.handleMessage(ctx, update.Message)
		}
	}

	return nil
}

func (c *capitalBotRepository) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	logger := logrus.WithContext(ctx).WithField("message", message.Text)

	switch {
	case message.Text == "/start":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome! Send me a message like:\n\nmakan ayam: 50000\n\nI'll track it as an expense.")
		c.bot.Send(msg)
		return
	case message.Text == "/login":
		if _, exists := waitingRooms[message.Chat.ID]; exists {
			msg := tgbotapi.NewMessage(message.Chat.ID, "You are already in the waiting room.")
			c.bot.Send(msg)
			return
		}

		waitingRooms[message.Chat.ID] = true
		msg := tgbotapi.NewMessage(message.Chat.ID, "You are now in the waiting room. Please send your email address to log in.")
		c.bot.Send(msg)
		return
	case message.Text == "/cancel":
		if _, exists := waitingRooms[message.Chat.ID]; !exists {
			msg := tgbotapi.NewMessage(message.Chat.ID, "You are not in the waiting room.")
			c.bot.Send(msg)
			return
		}
		delete(waitingRooms, message.Chat.ID)
		msg := tgbotapi.NewMessage(message.Chat.ID, "You have been removed from the waiting room.")
		c.bot.Send(msg)
		return
	case waitingRooms[message.Chat.ID]:
		if message.Text == "/cancel" {
			msg := tgbotapi.NewMessage(message.Chat.ID, "You are not in the waiting room.")
			c.bot.Send(msg)
			return
		}

		email := ""
		err := c.db.WithContext(ctx).Where("email = ?", message.Text).First(&model.User{}).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				msg := tgbotapi.NewMessage(message.Chat.ID, "Email not found. Please try again.")
				c.bot.Send(msg)
				return
			}

			logger.Error("failed to find user: ", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while processing your request.")
			c.bot.Send(msg)
			return
		}

		err = c.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Update("telegram_id", message.Chat.ID).Error
		if err != nil {
			logger.Error("failed to update user: ", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while processing your request.")
			c.bot.Send(msg)
			return
		}

		delete(waitingRooms, message.Chat.ID)
		msg := tgbotapi.NewMessage(message.Chat.ID, "You have been successfully logged in. You can now send me messages to track your expenses.")
		c.bot.Send(msg)
		return
	case message.Text == "/help":
		msg := tgbotapi.NewMessage(message.Chat.ID, "Here are some commands you can use:\n\n/start - Start the bot\n/login - Log in to your account\n/cancel - Cancel the current operation\n/help - Show this help message")
		c.bot.Send(msg)
		return
	case message.Text == "/whoami":
		var user model.User
		err := c.db.WithContext(ctx).Where("telegram_id = ?", message.Chat.ID).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				msg := tgbotapi.NewMessage(message.Chat.ID, "You are not logged in.")
				c.bot.Send(msg)
				return
			}
			logger.Error("failed to find user: ", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while processing your request.")
			c.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("You are logged in as %s (%s)", user.Name, user.Email))
		c.bot.Send(msg)

		return
	default:
		var loggedUser model.User

		err := c.db.WithContext(ctx).Where("telegram_id = ?", message.Chat.ID).First(&loggedUser).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				msg := tgbotapi.NewMessage(message.Chat.ID, "You are not logged in. Please log in first.")
				c.bot.Send(msg)
				return
			}
			logger.Error("failed to find user: ", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while processing your request.")
			c.bot.Send(msg)
			return
		}

		var potentialShareUser model.User

		err = c.db.WithContext(ctx).Where("email <> ?", loggedUser.Email).First(&potentialShareUser).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Error("failed to find potential share user: ", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while processing your request.")
			c.bot.Send(msg)
			return
		}

		withPrompt := model.SystemPrompt(loggedUser.ID, potentialShareUser.ID)

		transaction, err := c.openAi.RecognizeTransaction(ctx, withPrompt, message.Text)
		if err != nil {
			logger.Error("failed to recognize transaction: ", err)
			reply := tgbotapi.NewMessage(message.Chat.ID, "Sorry, I couldn't understand that.")
			c.bot.Send(reply)
			return
		}

		transaction.UserID = loggedUser.ID

		reply := tgbotapi.NewMessage(message.Chat.ID, replyMessage(transaction))
		c.bot.Send(reply)
	}
}

func replyMessage(transaction model.Transaction) string {
	// Format the reply message with the transaction details
	// You can customize this format as per your requirements
	return fmt.Sprintf(
		"✅ Transaction Recognized:\n"+
			"Category: %s\n"+
			"Type: %s\n"+
			"Description: %s\n"+
			"Amount: %s\n"+
			"Spent At: %s\n"+
			"Shared: %t",
		transaction.Category,
		transaction.TransactionType,
		transaction.Description,
		transaction.Amount.StringFixed(2),
		transaction.SpentAt.Format("2006-01-02"),
		transaction.IsShared,
	)
}
