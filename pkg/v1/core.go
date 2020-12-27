package v1

import (
	"errors"
	"time"
)

type Protocol interface {
	Start(payload []byte, output chan<- []byte) (State, *Error)
	Next(currentState State, payload []byte, output chan<- []byte) (State, *Error)
	OnError(err *Error, output chan<- []byte) (State, *Error)
	OnEnd(output chan<- []byte)
	OnNonRecoverableError(*Error, chan<- []byte) error
	CurrentState() State
	StartTimeout() <-chan time.Time
	Timeout() <-chan time.Time
	IncrementSequenceNumber()
	SequenceNumber() int64
}

var (
	Abort = State{
		Name:  "Abort",
		Start: false,
		End:   true,
	}

	Timeout = E(errors.New("timeout"), "Timeout", "To prevent deadlock, the protocol times out after a certain period of time", false)
)

func Run(p Protocol, c chan []byte) error {
	var (
		nextState State
		err       *Error
	)

	select {
	case nextPayload := <-c:
		{
			nextState, err = p.Start(nextPayload, c)
			p.IncrementSequenceNumber()

			if err != nil {
				if err.Recoverable {
					nextState, err = p.OnError(err, c)
				} else {
					return p.OnNonRecoverableError(err, c)
				}
			}
			if nextState.End {
				return nil
			}

		}
	case <-p.StartTimeout():
		{
			return p.OnNonRecoverableError(Timeout, c)
		}
	}

secondLoop:
	for {
		select {
		case nextPayload := <-c:
			{
				nextState, err = p.Next(nextState, nextPayload, c)
				p.IncrementSequenceNumber()
				if err != nil {
					if err.Recoverable {
						nextState, err = p.OnError(err, c)
					} else {
						return p.OnNonRecoverableError(err, c)
					}
				}
				if nextState.End {
					break secondLoop
				}
			}
		case <-p.Timeout():
			return p.OnNonRecoverableError(Timeout, c)
		}

	}
	p.OnEnd(c)
	return nil
}

type ErrorHandler func(*Error) (State, *Error)

type Error struct {
	Error       error  `json:"error"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Recoverable bool   `json:"recoverable"`
}

func E(error error, name, desc string, recoverable bool) *Error {
	return &Error{Error: error, Name: name, Recoverable: recoverable, Description: desc}
}
func (e *Error) WithError(err error) *Error {
	return &Error{
		Error:       err,
		Name:        e.Name,
		Description: e.Description,
		Recoverable: e.Recoverable,
	}
}
func (e *Error) WithRecoverable(r bool) *Error {
	return &Error{
		Error:       e.Error,
		Name:        e.Name,
		Description: e.Description,
		Recoverable: r,
	}
}

type State struct {
	Name  string `json:"name"`
	Start bool   `json:"start"`
	End   bool   `json:"end"`
}
