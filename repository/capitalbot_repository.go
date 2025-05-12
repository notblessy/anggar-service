package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/notblessy/anggar-service/model"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

		user := model.User{}
		err := c.db.WithContext(ctx).Where("email = ?", message.Text).First(&user).Error
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

		err = c.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", user.Email).Debug().Update("telegram_id", message.Chat.ID).Error
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
		transaction.SpentAt = time.Now()

		err = c.db.WithContext(ctx).Create(&transaction).Error
		if err != nil {
			logger.Error("failed to save transaction: ", err)
			reply := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while saving your transaction.")
			c.bot.Send(reply)
			return
		}

		transactionIndex := model.Transaction{}

		err = c.db.WithContext(ctx).Where("id = ?", transaction.ID).Preload("TransactionShares.User").First(&transactionIndex).Error
		if err != nil {
			logger.Error("failed to find transaction: ", err)
			reply := tgbotapi.NewMessage(message.Chat.ID, "An error occurred while retrieving your transaction.")
			c.bot.Send(reply)
			return
		}

		reply := tgbotapi.NewMessage(message.Chat.ID, replyMessage(transactionIndex))
		reply.ParseMode = tgbotapi.ModeMarkdownV2
		c.bot.Send(reply)
	}
}

func replyMessage(transaction model.Transaction) string {
	var b strings.Builder
	titleCaser := cases.Title(language.English)

	b.WriteString("✅ *Transaction Recognized*\n\n")
	b.WriteString(fmt.Sprintf("*Category:* %s\n", escapeMarkdownV2(titleCaser.String(transaction.Category))))
	b.WriteString(fmt.Sprintf("*Type:* %s\n", escapeMarkdownV2(titleCaser.String(strings.ToLower(transaction.TransactionType)))))
	b.WriteString(fmt.Sprintf("*Description:* %s\n", escapeMarkdownV2(transaction.Description)))
	b.WriteString(fmt.Sprintf("*Amount:* %s\n", escapeMarkdownV2(formatRupiah(transaction.Amount))))
	b.WriteString(fmt.Sprintf("*Date:* %s\n", escapeMarkdownV2(transaction.SpentAt.Format("2 Jan 2006"))))
	b.WriteString(fmt.Sprintf("*Shared:* %s\n", map[bool]string{true: "Yes", false: "No"}[transaction.IsShared]))

	if transaction.IsShared && len(transaction.TransactionShares) > 0 {
		b.WriteString("\n*Breakdown:*\n")
		for _, share := range transaction.TransactionShares {
			if share.Percentage.GreaterThan(decimal.Zero) {
				b.WriteString(fmt.Sprintf("• %s: %s%% — %s\n",
					escapeMarkdownV2(share.User.Name),
					escapeMarkdownV2(share.Percentage.StringFixed(2)),
					escapeMarkdownV2(formatRupiah(share.Amount)),
				))
			} else {
				b.WriteString(fmt.Sprintf("• %s: %s\n",
					escapeMarkdownV2(share.User.Name),
					escapeMarkdownV2(formatRupiah(share.Amount)),
				))
			}
		}
	}

	return b.String()
}

func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func formatRupiah(amount decimal.Decimal) string {
	// Get integer part
	intPart := amount.IntPart()

	// Format with dot as thousand separator
	formatted := strconv.FormatInt(intPart, 10)
	n := len(formatted)
	if n <= 3 {
		return "Rp" + formatted
	}

	var result []string
	for i := n; i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		result = append([]string{formatted[start:i]}, result...)
	}
	return "Rp" + strings.Join(result, ".")
}
