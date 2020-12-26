package v1

type Protocol interface {
	Start(payload []byte, output chan<- []byte) (State, *Error)
	Next(currentState State, payload []byte, output chan<- []byte) (State, *Error)
	OnError(err *Error, output chan<- []byte) (State, *Error)
	OnEnd(output chan<- []byte)
	OnNonRecoverableError(*Error) error
}

type ProtocolBase struct {
	Input         <-chan []byte
	Output        chan<- []byte
	SupportPolicy Protocol
}

func (p *ProtocolBase) Run() error {
	defer close(p.Output)

	nextPayload := <-p.Input

	nextState, err := p.SupportPolicy.Start(nextPayload, p.Output)

	if err.Recoverable {
		nextState, err = p.SupportPolicy.OnError(err, p.Output)
	} else {
		return p.SupportPolicy.OnNonRecoverableError(err)
	}

	for {
		nextPayload := <-p.Input
		nextState, err = p.SupportPolicy.Next(nextState, nextPayload, p.Output)
		if err.Recoverable {
			nextState, err = p.SupportPolicy.OnError(err, p.Output)
		} else {
			return p.SupportPolicy.OnNonRecoverableError(err)
		}
		if nextState.End {
			break
		}
	}
	p.SupportPolicy.OnEnd(p.Output)

	nextPayload = <-p.Input
	nextState, err = p.Start(nextPayload, p.Output)

	if err.Recoverable {
		nextState, err = p.OnError(err, p.Output)
	} else {
		return p.OnNonRecoverableError(err, p.Output)
	}

	for {
		nextPayload := <-p.Input
		nextState, err = p.Next(nextState, nextPayload, p.Output)
		if err.Recoverable {
			nextState, err = p.OnError(err, p.Output)
		} else {
			return p.OnNonRecoverableError(err, p.Output)
		}
		if nextState.End {
			break
		}
	}
	p.OnEnd(p.Output)
	return nil
}

func (p *ProtocolBase) Start(payload []byte, output chan<- []byte) (State, *Error) {
	panic("implement me")
}

func (p *ProtocolBase) Next(currentState State, payload []byte, output chan<- []byte) (State, *Error) {
	panic("implement me")
}

func (p *ProtocolBase) OnError(e *Error, output chan<- []byte) (State, *Error) {
	panic("implement me")
}

func (p *ProtocolBase) OnNonRecoverableError(e *Error, output chan<- []byte) error {
	panic("implement me")
}

func (p *ProtocolBase) OnEnd(chan<- []byte) {
	panic("implement me")
}

type ErrorHandler func(*Error) (State, *Error)

type Error struct {
	Error       error
	Name        string
	Description string
	Recoverable bool
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
	Name  string
	Start bool
	End   bool
}
