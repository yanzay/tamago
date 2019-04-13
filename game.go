package main

import (
	"fmt"
	"time"
)

const (
	SpeedFood         = 2
	SpeedHappy        = 1
	SpeedHealth       = 2
	SpeedNormalWeight = 1
	SpeedOverWeight   = 2
	NormalWeight      = 42
)

var (
	moveDuration       = 60 * time.Second
	sleepCheckDuration = 5 * time.Second
)

func (app *application) mainLoop() {
	tick := time.Tick(moveDuration)
	for range tick {
		for _, pet := range app.petStore.Alive() {
			if !pet.Sleep {
				app.petStore.Update(pet.PlayerID, func(pet *Pet) {
					app.decreaseFood(pet)
					app.decreaseHappy(pet)
					died := app.decreaseHealth(pet)
					if died {
						go app.historyStore.Create(pet)
					}
					pet.Weight += getWeightDelta(pet)
					if pet.Weight < 2 {
						pet.Weight = 1
					}
				})
			}
		}
	}
}

func (app *application) sleepLoop() {
	tick := time.Tick(sleepCheckDuration)
	for range tick {
		for _, pet := range app.petStore.Alive() {
			if pet.Sleep && pet.AwakeTime.Before(time.Now()) {
				app.petStore.Update(pet.PlayerID, func(p *Pet) {
					p.Sleep = false
					app.notifyPlayer(p.PlayerID, "Good morning!")
				})
			}
		}
	}
}

func (app *application) notifyPlayer(playerID, text string) {
	app.client.SendMessage(playerID, "💬 "+text)
}

func (app *application) decreaseFood(pet *Pet) {
	if pet.Food > SpeedFood {
		pet.Food -= SpeedFood
	} else {
		if pet.Food > 0 {
			app.notifyPlayer(pet.PlayerID, "Hey! I am hungry!")
		}
		pet.Food = 0
	}
}

func (app *application) decreaseHappy(pet *Pet) {
	speed := SpeedHappy
	if pet.Food == 0 {
		speed *= 2
	}
	if pet.Happy > speed {
		pet.Happy -= speed
	} else {
		if pet.Happy > 0 {
			app.notifyPlayer(pet.PlayerID, "Hey! I am bored!")
		}
		pet.Happy = 0
	}
}

func (app *application) decreaseHealth(pet *Pet) bool {
	if pet.Happy == 0 || pet.Food == 0 {
		if pet.Health < 10 {
			app.notifyPlayer(pet.PlayerID, "I'm dying! Please help me!")
		}
		if pet.Health > SpeedHealth {
			pet.Health -= SpeedHealth
		} else {
			pet.Die()
			app.notifyPlayer(pet.PlayerID, fmt.Sprintf("Oh no! your pet %s died.", pet.String()))
			return true
		}
	}
	return false
}

func getWeightDelta(pet *Pet) int {
	switch {
	case pet.Food > 120:
		return SpeedOverWeight
	case pet.Food > 80 && pet.Weight < NormalWeight:
		return SpeedNormalWeight
	case pet.Food > 80 && pet.Weight > NormalWeight:
		return -SpeedNormalWeight
	case pet.Food > 20 && pet.Weight > NormalWeight:
		return -SpeedNormalWeight
	case pet.Food < 20:
		return -SpeedOverWeight
	}
	return 0
}
