package main

import (
	"fmt"
	"time"

	"github.com/yanzay/tbot"
)

func (app *application) sleep(f tbot.UpdateHandler) tbot.UpdateHandler {
	return func(u *tbot.Update) {
		if u.Message == nil {
			f(u)
			return
		}
		m := u.Message
		pet := app.petStore.Get(m.Chat.ID)
		if pet.Sleep {
			if time.Until(pet.AwakeTime) > 5*time.Second {
				app.client.SendMessage(m.Chat.ID,
					fmt.Sprintf("Your pet is sleeping. Time to wake up: %s", roundDuration(time.Until(pet.AwakeTime))))
			} else {
				app.client.SendMessage(m.Chat.ID, "Your pet will wake up soon.")
			}
			return
		}
		f(u)
	}
}
