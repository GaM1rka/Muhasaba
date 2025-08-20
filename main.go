package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// usersData stores each user's data.
// It's a map with the key being the user's ID (int64) and the value being a pointer to a UserData struct.
var usersData = make(map[int64]*UserData)

// UserData holds data for a specific user.
type UserData struct {
	// contract holds the user's morning contract.
	contract string
	// reportQuestions holds the questions for the evening report.
	reportQuestions []string
	// currentQuestionIndex tracks the current question the user is answering.
	currentQuestionIndex int
	// responses stores the user's answers to the questions.
	responses []string
}

// initUserData initializes data for a new user.
func initUserData(userID int64) *UserData {
	userData := &UserData{
		contract: "",
		reportQuestions: []string{
			"Фард: намазы вовремя? долги, обязательства? права людей (родители/супруг/партнёры)?",
			"Харам: что удалось оставить? где сорвался?",
			"Ният: где просочилась рия? где был «ради людей, не ради Аллаха»?",
			"Сунна/качество: насколько осознанно выполнял?",
			"Язык/взгляд/время: были ли сплетни, пустые скроллы, лишние слова?",
			"Благодарность: за какое доброе дело сегодня сказал «альхамдулиллях»?",
		},
		currentQuestionIndex: -1,
		responses:            make([]string, 0),
	}
	usersData[userID] = userData
	return userData
}

func main() {
	ctx := context.Background()
	// Replace "YOUR_TOKEN" with your token from BotFather.
	token := "8479178091:AAG5lGQUJdiifdPmzt6DbweqlSHu1K5aKAI"

	// Options for creating the bot.
	opts := []bot.Option{
		bot.WithDefaultHandler(handleUpdate),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		log.Fatalf("Ошибка при создании бота: %v", err)
	}

	// Set bot commands.
	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			{Command: "/start", Description: "Начать работу с ботом"},
		},
	})

	// Start a goroutine for sending reminders.
	go sendReminders(b)

	log.Println("Бот запущен...")
	b.Start(ctx)
}

// handleUpdate handles all incoming messages and commands.
func handleUpdate(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message != nil {
		log.Printf("Получено сообщение: %s от пользователя %d", update.Message.Text, update.Message.From.ID)
		handleMessage(ctx, b, update.Message)
	} else if update.CallbackQuery != nil {
		log.Printf("Получен callback: %s от пользователя %d", update.CallbackQuery.Data, update.CallbackQuery.From.ID)
		handleCallbackQuery(ctx, b, update.CallbackQuery)
	}
}

// handleMessage handles text messages.
func handleMessage(ctx context.Context, b *bot.Bot, message *models.Message) {
	userID := message.Chat.ID
	userData, ok := usersData[userID]

	if strings.HasPrefix(message.Text, "/") {
		log.Println("Обрабатывается как команда:", message.Text)
		handleCommand(ctx, b, message)
		return
	}

	// If the user is answering report questions.
	if ok && userData.currentQuestionIndex != -1 {
		userData.responses = append(userData.responses, message.Text)
		userData.currentQuestionIndex++

		if userData.currentQuestionIndex < len(userData.reportQuestions) {
			// Send the next question.
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   userData.reportQuestions[userData.currentQuestionIndex],
			})
		} else {
			// All questions have been asked, send AI recommendation.
			// This is where AI logic would go.
			var aiResponse string
			aiResponse = "..." // AI response placeholder

			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   "Относительно твоих отчетов ИИ выдало вот такую рекомендацию:\n\n" + aiResponse,
			})
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: userID,
				Text:   "Спасибо за твою работу! До встречи завтра. Ассаламу Алейкум!",
			})
			// End the report process.
			userData.currentQuestionIndex = -1
		}
		return
	}

	// If the user is sending their morning contract.
	if ok && userData.contract == "" {
		userData.contract = message.Text

		// Add "Skip Reminders" button here
		markup := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "Пропустить напоминания", CallbackData: "skip_reminders"},
				},
			},
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      userID,
			Text:        "Твой утренний договор сохранен. Тебе придут напоминания в течение дня. БаракаЛлаху фик!",
			ReplyMarkup: markup,
		})
		return
	}

	// Handle other messages.
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   "Извини, я не понял твою команду. Используй /start для начала.",
	})
}

// handleCommand handles commands.
func handleCommand(ctx context.Context, b *bot.Bot, message *models.Message) {
	log.Println("Вызвана функция handleCommand")
	switch message.Text {
	case "/start":
		log.Println("Команда /start получена. Вызываем sendIntroMessage.")
		initUserData(message.Chat.ID)
		sendIntroMessage(ctx, b, message.Chat.ID)
	default:
		log.Println("Неизвестная команда:", message.Text)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: message.Chat.ID,
			Text:   "Неизвестная команда. Пожалуйста, используйте /start.",
		})
	}
}

// sendIntroMessage sends the introductory message.
func sendIntroMessage(ctx context.Context, b *bot.Bot, chatID int64) {
	text := "Ассаламу Алейкум, дорогой пользователь\\!\n\n" +
		"*Мухасаба* — это исламская практика самоанализа, непрерывный цикл, подобный сверке купеческих счетов\\. Этот процесс включает в себя:\n" +
		"1\\. *Мошарата*: Утренний «договор» с самим собой, постановка целей на день\\.\n" +
		"2\\. *Муракаба*: Бдительный контроль в течение дня\\.\n" +
		"3\\. *Мухасаба*: Вечерний разбор своих действий\\.\n" +
		"4\\. *Му’акаба/Му’атиба*: Исправление ошибок, «штраф» или увещевание\\.\n" +
		"5\\. *Муджахада*: Дальнейшая работа над собой\\.\n\n" +
		"*Функционал бота:*\n" +
		"\\(1\\) *Утренний «договор»*: Ты ставишь цели на день\\.\n" +
		"\\(2\\) *Дневная Муракаба*: Бот напомнит о твоих целях в 12:00, 15:00 и 18:00\\.\n" +
		"\\(3\\) *Вечерняя Мухасаба*: Бот поочередно задаст вопросы о твоем дне\\.\n" +
		"\\(4\\) *Получение рекомендации*: На основе твоих ответов ты получишь рекомендацию от ИИ\\."

	// Create a button.
	markup := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Начать Мухасаба", CallbackData: "start_muhasaba"},
			},
		},
	}

	log.Println("Попытка отправки сообщения intro...")

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: markup,
	})

	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	} else {
		log.Println("Сообщение успешно отправлено.")
	}
}

// handleCallbackQuery handles button presses.
func handleCallbackQuery(ctx context.Context, b *bot.Bot, query *models.CallbackQuery) {
	userID := query.From.ID
	switch query.Data {
	case "start_muhasaba":
		log.Println("Получен callback 'start_muhasaba'. Отправляем сообщение с шаблоном.")
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: userID,
			Text: "Новый день с Мухасаба\\! Сделай свой утренний договор, используя этот шаблон:\n\n" +
				"**«О Аллах, ради Тебя выполняю сегодня: \\[фард\\-пункты, рабочие задачи с халяль\\-намерением, заботу о семье\\]\\. С Терпением избегаю: \\[2\\–3 персональные слабости\\]\\. Берегу: язык, взгляд, время\\. Дай искренность и лёгкость»\\.**\n\n" +
				"Просто скопируй и отправь его, заполнив свои пункты\\.",
			ParseMode: models.ParseModeMarkdown,
		})
		if err != nil {
			log.Printf("Ошибка при отправке сообщения с шаблоном: %v", err)
		} else {
			log.Println("Сообщение с шаблоном успешно отправлено.")
		}
	case "start_report":
		startReport(ctx, b, userID)
	case "skip_reminders":
		log.Println("Получен callback 'skip_reminders'. Запускаем отчет.")
		// Immediately start the report
		startReport(ctx, b, userID)
	}
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: query.ID,
	})
	log.Println("Ответ на callback-запрос отправлен.")
}

// startReport sends the first question to the user and begins the report process.
func startReport(ctx context.Context, b *bot.Bot, userID int64) {
	userData := usersData[userID]
	if userData == nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: userID,
			Text:   "Пожалуйста, сначала начните новый день с /start.",
		})
		return
	}
	// Start the report.
	userData.currentQuestionIndex = 0
	userData.responses = make([]string, 0)
	// Send the first question.
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: userID,
		Text:   userData.reportQuestions[userData.currentQuestionIndex],
	})
}

// sendReminders sends reminders throughout the day.
func sendReminders(b *bot.Bot) {
	for {
		now := time.Now()
		// Send reminders at 12:00, 15:00, 18:00
		if now.Hour() == 12 || now.Hour() == 15 || now.Hour() == 18 {
			for userID, userData := range usersData {
				if userData.contract != "" {
					// Add "Skip Reminders" button to reminder messages
					markup := &models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{Text: "Пропустить напоминания", CallbackData: "skip_reminders"},
							},
						},
					}
					b.SendMessage(context.Background(), &bot.SendMessageParams{
						ChatID:      userID,
						Text:        "**Напоминание о твоем утреннем договоре:**\n\n" + userData.contract,
						ParseMode:   models.ParseModeMarkdown,
						ReplyMarkup: markup,
					})
				}
			}
		}
		// Evening Muhasaba at 21:00.
		if now.Hour() == 21 && now.Minute() == 0 {
			for userID, userData := range usersData {
				if userData.contract != "" {
					markup := &models.InlineKeyboardMarkup{
						InlineKeyboard: [][]models.InlineKeyboardButton{
							{
								{Text: "Начать отчет", CallbackData: "start_report"},
							},
						},
					}
					b.SendMessage(context.Background(), &bot.SendMessageParams{
						ChatID:      userID,
						Text:        "Пришло время для вечерней Мухасаба. Нажми кнопку, чтобы начать отчет.",
						ReplyMarkup: markup,
					})
				}
			}
		}

		// Wait for the next hour.
		time.Sleep(1 * time.Hour)
	}
}
