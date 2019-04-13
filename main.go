package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"text/template"
	"time"

	"github.com/yanzay/tbot"
)

var local = flag.Bool("local", false, "Launch bot without webhook")
var dataFile = flag.String("data", "tamago.db", "Database file")

type application struct {
	petStore     *PetStorage
	historyStore *PetStorage
	client       *tbot.Client
}

func main() {
	flag.Parse()
	app := &application{}
	storage := NewStorage(*dataFile)
	app.petStore = storage.PetStorage()
	app.historyStore = storage.HistoryStorage()
	app.resetPlays()
	go app.gameStats()

	defer storage.Close()
	token := os.Getenv("TELEGRAM_TOKEN")
	var bot *tbot.Server
	if *local {
		bot = tbot.New(token, tbot.WithLogger(tbot.BasicLogger{}))
	} else {
		bot = tbot.New(token,
			tbot.WithWebhook("https://tamago.yanzay.com/"+token, "0.0.0.0:8014"))
	}
	app.client = bot.Client()

	bot.Use(app.createPet)
	bot.Use(app.sleep)

	bot.HandleMessage(HomeButton, app.rootHandler)
	bot.HandleMessage(FeedButton, app.feedHandler)
	bot.HandleMessage(FoodPizza, app.fullMealHandler)
	bot.HandleMessage(FoodMeat, app.fullMealHandler)
	bot.HandleMessage(FoodSalad, app.smallMealHandler)
	bot.HandleMessage(FoodPopcorn, app.smallMealHandler)
	bot.HandleMessage(PlayButton, app.playHandler)

	bot.HandleMessage(GameVideo, app.playGameHandler)
	bot.HandleMessage(GameBoard, app.playGameHandler)
	bot.HandleMessage(GameTennis, app.playGameHandler)
	bot.HandleMessage(GameGuitar, app.playGameHandler)

	bot.HandleMessage(HealButton, app.healHandler)
	bot.HandleMessage(PillButton, app.pillHandler)
	bot.HandleMessage(InjectionButton, app.injectionHandler)
	bot.HandleMessage(SleepButton, app.sleepHandler)
	bot.HandleMessage(Sleep5m, app.sleep5mHandler)
	bot.HandleMessage(Sleep1h, app.sleep1hHandler)
	bot.HandleMessage(Sleep8h, app.sleep8hHandler)
	bot.HandleMessage(TopButton, app.topHandler)
	bot.HandleMessage(AliveButton, app.topAliveHandler)
	bot.HandleMessage(AllButton, app.topAllHandler)
	bot.HandleMessage("", app.defaultHandler)
	go app.mainLoop()
	go app.sleepLoop()
	log.Fatal(bot.Start())
}

func (app *application) resetPlays() {
	pets := app.petStore.Alive()
	for _, pet := range pets {
		app.petStore.Update(pet.PlayerID, func(p *Pet) {
			p.Play = false
		})
	}
}

func (app *application) defaultHandler(m *tbot.Message) {
	app.client.SendMessage(m.Chat.ID, "hm?")
}

func (app *application) rootHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	content, err := contentFromTemplate(rootTemplate, pet)
	if err != nil {
		return
	}
	buttons := tbot.Buttons([][]string{
		{HomeButton, FeedButton, PlayButton},
		{HealButton, SleepButton, TopButton},
	})
	app.client.SendMessage(m.Chat.ID, content,
		tbot.OptReplyKeyboardMarkup(buttons),
		tbot.OptParseModeMarkdown)
}

func (app *application) feedHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	content, err := contentFromTemplate(feedTemplate, pet)
	if err != nil {
		return
	}
	buttons := tbot.Buttons([][]string{
		{FoodSalad, FoodMeat},
		{FoodPopcorn, FoodPizza},
		{HomeButton},
	})
	app.client.SendMessage(m.Chat.ID, content,
		tbot.OptReplyKeyboardMarkup(buttons),
		tbot.OptParseModeMarkdown)
}

func (app *application) fullMealHandler(m *tbot.Message) {
	message := "Om-nom-nom..."
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		if pet.Food == 200 {
			message = "I can't eat more."
		}
		pet.Food += 10
		if pet.Food > 200 {
			pet.Food = 200
		}
	})
	app.client.SendMessage(m.Chat.ID, message)
	app.feedHandler(m)
}

func (app *application) smallMealHandler(m *tbot.Message) {
	message := "Om-nom..."
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		if pet.Food == 200 {
			message = "I can't eat more."
		}
		pet.Food += 5
		if pet.Food > 200 {
			pet.Food = 200
		}
	})
	app.client.SendMessage(m.Chat.ID, message)
	app.feedHandler(m)
}

func (app *application) playHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	content, err := contentFromTemplate(playTemplate, pet)
	if err != nil {
		return
	}
	buttons := tbot.Buttons([][]string{
		{GameVideo, GameBoard},
		{GameTennis, GameGuitar},
		{HomeButton},
	})
	app.client.SendMessage(m.Chat.ID, content,
		tbot.OptReplyKeyboardMarkup(buttons),
		tbot.OptParseModeMarkdown)
}

func (app *application) playGameHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	if pet.Play {
		app.client.SendMessage(m.Chat.ID, "You pet is already playing. Keep calm.")
		return
	}
	pets := app.petStore.Alive()
	randomPet := pets[rand.Intn(len(pets))]
	if fmt.Sprint(randomPet.PlayerID) != m.Chat.ID {
		app.client.SendMessage(m.Chat.ID,
			fmt.Sprintf("Your pet started to play %s with %s", m.Text, randomPet.String()))
	} else {
		app.client.SendMessage(m.Chat.ID,
			fmt.Sprintf("Your pet plays %s with himself.", m.Text))
	}
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Play = true
	})
	time.Sleep(5 * time.Second)
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Play = false
		if pet.Happy < 120 {
			pet.XP += 100
		}
		pet.Happy += 10
		if pet.Happy > 120 {
			pet.Happy = 120
		}
	})
	app.client.SendMessage(m.Chat.ID, "Weeeee! It was fun!")
}

func (app *application) healHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	content, err := contentFromTemplate(healTemplate, pet)
	if err != nil {
		return
	}
	buttons := tbot.Buttons([][]string{
		{PillButton, InjectionButton},
		{HomeButton},
	})
	app.client.SendMessage(m.Chat.ID, content,
		tbot.OptReplyKeyboardMarkup(buttons),
		tbot.OptParseModeMarkdown)
}

func (app *application) pillHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	if pet.Health == 100 {
		app.client.SendMessage(m.Chat.ID, "I'm not sick!")
		return
	}
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Health += 40
		pet.Happy -= 10
		if pet.Health > 100 {
			pet.Health = 100
		}
		if pet.Happy < 0 {
			pet.Happy = 0
		}
	})
	app.client.SendMessage(m.Chat.ID, "Ugh!")
	app.healHandler(m)
}

func (app *application) injectionHandler(m *tbot.Message) {
	pet := app.petStore.Get(m.Chat.ID)
	if pet.Health == 100 {
		app.client.SendMessage(m.Chat.ID, "I'm not sick!")
		return
	}
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Health = 100
		if pet.Happy > 10 {
			pet.Happy = 10
		}
	})
	app.client.SendMessage(m.Chat.ID, "Ouch!")
	app.healHandler(m)
}

func (app *application) sleepHandler(m *tbot.Message) {
	buttons := tbot.Buttons([][]string{
		{Sleep5m, Sleep1h, Sleep8h},
		{HomeButton},
	})
	app.client.SendMessage(m.Chat.ID, "How much to sleep?",
		tbot.OptReplyKeyboardMarkup(buttons),
		tbot.OptParseModeMarkdown)
}

func (app *application) sleep5mHandler(m *tbot.Message) {
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Sleep = true
		pet.AwakeTime = time.Now().Add(5 * time.Minute)
	})
	app.client.SendMessage(m.Chat.ID, "Zzz...")
}

func (app *application) sleep1hHandler(m *tbot.Message) {
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Sleep = true
		pet.AwakeTime = time.Now().Add(1 * time.Hour)
	})
	app.client.SendMessage(m.Chat.ID, "Zzz...")
}

func (app *application) sleep8hHandler(m *tbot.Message) {
	app.petStore.Update(m.Chat.ID, func(pet *Pet) {
		pet.Sleep = true
		pet.AwakeTime = time.Now().Add(8 * time.Hour)
	})
	app.client.SendMessage(m.Chat.ID, "Zzz...")
}

func (app *application) topHandler(m *tbot.Message) {
	buttons := tbot.Buttons([][]string{
		{AliveButton, AllButton},
		{HomeButton},
	})
	app.client.SendMessage(m.Chat.ID, "Choose top",
		tbot.OptReplyKeyboardMarkup(buttons))
}

func (app *application) topAliveHandler(m *tbot.Message) {
	pets := app.petStore.Alive()
	sort.Slice(pets, func(i, j int) bool {
		return pets[i].XP > pets[j].XP
	})
	b := &bytes.Buffer{}
	if len(pets) > 10 {
		pets = pets[:10]
	}
	err := topTemplate.Execute(b, pets)
	if err != nil {
		log.Printf("Can't render topTemplate: %q", err)
	}
	content := "```\n" + b.String() + "```"
	app.client.SendMessage(m.Chat.ID, content, tbot.OptParseModeMarkdown)
}

func (app *application) topAllHandler(m *tbot.Message) {
	pets := app.petStore.Alive()
	pets = append(pets, app.historyStore.All()...)
	sort.Slice(pets, func(i, j int) bool {
		return pets[i].XP > pets[j].XP
	})
	b := &bytes.Buffer{}
	if len(pets) > 10 {
		pets = pets[:10]
	}
	err := topTemplate.Execute(b, pets)
	if err != nil {
		log.Printf("Can't render topTemplate: %q", err)
	}
	content := "```\n" + b.String() + "```"
	app.client.SendMessage(m.Chat.ID, content, tbot.OptParseModeMarkdown)
}

func contentFromTemplate(tpl *template.Template, pet *Pet) (string, error) {
	b := &bytes.Buffer{}
	err := tpl.Execute(b, pet)
	if err != nil {
		log.Printf("Can't render template %v: %q", tpl, err)
		return "", err
	}
	return "```\n" + b.String() + "```", nil
}

func (app *application) gameStats() {
	for {
		pets := app.petStore.All()
		alive := app.petStore.Alive()
		log.Printf("Players: %d, alive: %d", len(pets), len(alive))
		time.Sleep(60 * time.Second)
	}
}
