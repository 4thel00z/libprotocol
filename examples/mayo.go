package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/4thel00z/libprotocol/pkg/v1"
	"github.com/4thel00z/libprotocol/pkg/v1/policies/version"
	"log"
	"time"
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
	SupportPolicy v1.Protocol
	Channel       chan []byte
	Ingredients   []Ingredient
	State         v1.State
}

func (m *MayoProtocol) StartTimeout() <-chan time.Time {
	return time.After(3 * time.Second)
}

func (m *MayoProtocol) Timeout() <-chan time.Time {
	return time.After(3 * time.Second)
}

func (m *MayoProtocol) IncrementSequenceNumber() {
	//noop
}

func (m *MayoProtocol) SequenceNumber() int64 {
	//noop
	return -1
}

func (m *MayoProtocol) CurrentState() v1.State {
	return m.State
}

func (m *MayoProtocol) OnEnd(output chan<- []byte) {
	fmt.Println("Mayonese consists of:")

	for _, i := range m.Ingredients {
		fmt.Println(i.Quantity, i.Unit, "of", i.Name)
	}
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
		fmt.Println("Received ingredient")
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

func (m *MayoProtocol) OnNonRecoverableError(e *v1.Error, c chan<- []byte) error {
	log.Fatalln(e.Name, e.Error.Error(), e.Description)
	return e.Error
}

func MockPolicyIO(c chan []byte) {
	req := version.VersionRequest{
		Version: 1.0,
	}

	must(v1.Send(req, c))
	msg := <-c
	fmt.Println(string(msg))

}

func MockMayoIO(c chan []byte) {
	for _, payload := range ingredients {
		fmt.Println("sending", string(payload))
		must(v1.Send(payload, c))
		msg := <-c
		fmt.Println(string(msg))
	}

}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {

	c := make(chan []byte)
	defer close(c)
	protocol := &MayoProtocol{
		Channel:       c,
		SupportPolicy: version.VersionPolicy(false, 1.0),
	}

	go MockPolicyIO(c)
	must(v1.Run(protocol.SupportPolicy, protocol.Channel))
	fmt.Println(protocol.SupportPolicy.CurrentState().Name)

	go MockMayoIO(c)
	must(v1.Run(protocol, protocol.Channel))
}
