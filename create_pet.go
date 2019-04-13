package main

import (
	"fmt"
	"time"

	"github.com/yanzay/tbot"
)

func (app *application) createPet(f tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		if u.Message == nil {
			f(u)
			return
		}
		m := u.Message
		pet := app.petStore.Get(m.Chat.ID)
		if !pet.Alive {
			app.petStore.Set(m.Chat.ID, NewPet(m.Chat.ID))
			buttons := tbot.Buttons([][]string{{"Create"}})
			content, err := contentFromTemplate(rootTemplate, pet)
			if err != nil {
				return
			}
			app.client.SendMessage(m.Chat.ID, content, tbot.OptParseModeMarkdown)
			app.client.SendMessage(m.Chat.ID, "Your pet is dead. Create new one?",
				tbot.OptReplyKeyboardMarkup(buttons))
			return
		}

		if pet.Name != "" && pet.Emoji != "" {
			f(u)
			return
		}

		defer app.petStore.Set(m.Chat.ID, pet)
		if pet.AskType {
			switch m.Text {
			case Chicken.String():
				pet.Emoji = Chicken.Emoji
			case Penguin.String():
				pet.Emoji = Penguin.Emoji
			case Dog.String():
				pet.Emoji = Dog.Emoji
			case Monkey.String():
				pet.Emoji = Monkey.Emoji
			case Fox.String():
				pet.Emoji = Fox.Emoji
			case Panda.String():
				pet.Emoji = Panda.Emoji
			case Pig.String():
				pet.Emoji = Pig.Emoji
			case Rabbit.String():
				pet.Emoji = Rabbit.Emoji
			case Mouse.String():
				pet.Emoji = Mouse.Emoji
			default:
				app.client.SendMessage(m.Chat.ID, fmt.Sprintf("Wrong pet type %s", m.Text))
			}
			pet.AskType = false
		}
		if pet.AskName {
			pet.Name = m.Text
			pet.AskName = false
			pet.Born = time.Now()
			pet.Alive = true
			app.petStore.Set(pet.PlayerID, pet)
			app.rootHandler(m)
		}
		if pet.Emoji == "" {
			pet.AskType = true
			pets := tbot.Buttons([][]string{
				{Chicken.String(), Penguin.String(), Dog.String()},
				{Monkey.String(), Fox.String(), Panda.String()},
				{Pig.String(), Rabbit.String(), Mouse.String()},
			})
			pets.OneTimeKeyboard = true
			app.client.SendMessage(m.Chat.ID, "Choose your pet:",
				tbot.OptReplyKeyboardMarkup(pets))
			return
		}
		if pet.Name == "" {
			pet.AskName = true
			app.client.SendMessage(m.Chat.ID, "Name your pet:")
			return
		}
	}
}
