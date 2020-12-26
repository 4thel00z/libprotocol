package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/4thel00z/libprotocol/pkg/v1"
	"log"
)

func B(p string) []byte {
	return []byte(p)
}

var (
	ingredients = [][]byte{

		B(`
{
	"name":"oil",
	"quantity":"1",
	"unit" : "l"
}
`), B(`
{
	"name":"eggs",
	"quantity":"400",
	"unit" : "g"
}
`), B(`
{
	"name":"salt",
	"quantity":"20",
	"unit" : "g"
}
`),
	}
)

type Ingredient struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Unit     string  `json:"unit"`
}

type MayoProtocol struct {
	Ingredients []Ingredient
	State       v1.State
}

var (
	Cooking = v1.State{
		Name:  "cooking",
		Start: true,
		End:   false,
	}
	Cooked = v1.State{
		Name:  "cooking",
		Start: false,
		End:   true,
	}

	mayoWasAlreadyCookedError = v1.E(
		errors.New("mayo was already cooked"),
		"MayoWasAlreadyCooked",
		"",
		true,
	)

	nonRecoverableMayoError = v1.E(
		errors.New("mayo cooking went terribly wrong"),
		"NonRecoverableMayoError",
		"Probably the kitchen exploded or sth. of that sort. Super sorry ma dude!",
		false,
	)
)

func (m *MayoProtocol) Start(payload []byte, output chan<- []byte) (v1.State, *v1.Error) {
	m.Ingredients = []Ingredient{}
	m.State = Cooking
	return m.Next(m.State, payload, output)
}

func (m *MayoProtocol) Next(currentState v1.State, payload []byte, output chan<- []byte) (v1.State, *v1.Error) {
	switch currentState {
	case Cooking:
		var i Ingredient
		err := json.Unmarshal(payload, &i)

		if err != nil {
			return v1.State{}, mayoWasAlreadyCookedError
		}

		m.Ingredients = append(m.Ingredients, i)

		if len(m.Ingredients) >= 3 {
			return Cooked, nil
		}

		return Cooking, nil

	case Cooked:
	default:
		break
	}

	return v1.State{}, mayoWasAlreadyCookedError
}

func (m *MayoProtocol) OnMayoWasAlreadyCooked(err *v1.Error) (v1.State, *v1.Error) {
	fmt.Println(err.Description)
	return Cooked, nil
}

func (m *MayoProtocol) OnError(e *v1.Error, output chan<- []byte) (v1.State, *v1.Error) {
	switch e {
	case mayoWasAlreadyCookedError:
		return m.OnMayoWasAlreadyCooked(e)
	}

	return v1.State{}, nonRecoverableMayoError.WithError(e.Error)
}

func (m *MayoProtocol) OnNonRecoverableError(e *v1.Error, output chan<- []byte) {
	log.Fatalln(e.Name, e.Error.Error(), e.Description)

}

func main() {

}
