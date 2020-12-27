package version

import (
	"encoding/json"
	"errors"
	v1 "github.com/4thel00z/libprotocol/pkg/v1"
	"time"
)

const (
	maxInitialSequenceNumber = 10000000
	defaultStartTimeoutTime  = 10
	defaultTimeoutTime       = 10
)

type ProtocolVersionPolicy struct {
	SupportedVersions []float64
	AllowsDowngrades  bool
	state             v1.State
	sequenceNumber    int64
}

func VersionPolicy(allowsDowngrades bool, supportedVersions ...float64) *ProtocolVersionPolicy {

	return &ProtocolVersionPolicy{
		SupportedVersions: supportedVersions,
		AllowsDowngrades:  allowsDowngrades,
		sequenceNumber:    v1.InitialSequenceNumber(maxInitialSequenceNumber),
	}
}

func (p *ProtocolVersionPolicy) CurrentState() v1.State {
	return p.state
}

func (p *ProtocolVersionPolicy) StartTimeout() <-chan time.Time {
	return time.After(time.Second * defaultStartTimeoutTime)
}

func (p *ProtocolVersionPolicy) Timeout() <-chan time.Time {
	return time.After(time.Second * defaultTimeoutTime)
}

func (p *ProtocolVersionPolicy) IncrementSequenceNumber() {
	p.sequenceNumber++
}

func (p *ProtocolVersionPolicy) SequenceNumber() int64 {
	return p.sequenceNumber
}

type VersionRequest struct {
	Version float64
}

type VersionResponse struct {
	Version float64   `json:"version"`
	Error   *v1.Error `json:"error"`
}

// States
var (
	Negotation = v1.State{
		Name:  "Negotation",
		Start: true,
		End:   false,
	}

	SupportsVersion = v1.State{
		Name:  "SupportsVersion",
		Start: false,
		End:   true,
	}
)

// Errors
var (
	NoSupportedVersions = v1.E(
		errors.New("no supported versions"),
		"NoSupportedVersions",
		"Our peer does not support any of the proposed versions",
		false,
	)
	OtherVersionTooHigh = v1.E(
		errors.New("version of the other peer is too high"),
		"OtherVersionTooHigh",
		"The other peer does only support newer protocols than this server.",
		true,
	)

	OtherVersionTooLow = v1.E(
		errors.New("version of the other peer is too low"),
		"OtherVersionTooLow",
		"The other peer does only support lower protocols than this server.",
		true,
	)
	OurVersionTooHigh = v1.E(
		errors.New("version of us is too high"),
		"OurVersionTooHigh",
		"Our version does only support newer protocols than this server.",
		false,
	)
	CouldNotParse = v1.E(
		errors.New("could not parse the message"),
		"CouldNotParse",
		"The message of the other peer is malformed",
		false,
	)
	ConnectionError = v1.E(
		errors.New("could not send a message"),
		"ConnectionError",
		"An I/O issue of some sort",
		false,
	)
)

func (p *ProtocolVersionPolicy) Start(payload []byte, output chan<- []byte) (v1.State, *v1.Error) {
	var req VersionRequest
	err := json.Unmarshal(payload, &req)
	if err != nil {
		return v1.State{}, CouldNotParse
	}

	smaller := []float64{}
	larger := []float64{}

	for _, ourVersion := range p.SupportedVersions {
		if ourVersion == req.Version {
			res := VersionResponse{
				Version: ourVersion,
				Error:   nil,
			}
			err = v1.Send(res, output)
			if err != nil {
				p.state = v1.Abort
				return v1.Abort, ConnectionError.WithError(err)
			}
			return SupportsVersion, nil
		} else if ourVersion < req.Version {
			smaller = append(smaller, ourVersion)
		} else if ourVersion > req.Version {
			larger = append(larger, ourVersion)
		}
	}

	if len(larger) > 0 && len(smaller) == 0 {
		p.state = v1.Abort
		return v1.Abort, OtherVersionTooHigh
	} else if len(larger) == 0 && len(smaller) == 0 {
		p.state = v1.Abort
		return v1.Abort, NoSupportedVersions
	} else if len(smaller) > 0 {
		if !p.AllowsDowngrades {
			p.state = v1.Abort
			return v1.Abort, OtherVersionTooLow
		}
		max, _ := v1.Max(smaller)
		res := VersionResponse{
			Version: max,
			Error:   nil,
		}
		err = v1.Send(res, output)
		if err != nil {
			p.state = v1.Abort

			return v1.Abort, ConnectionError.WithError(err)
		}
		p.state = SupportsVersion
		return SupportsVersion, nil
	}

	p.state = v1.Abort
	return v1.Abort, NoSupportedVersions
}

func (p *ProtocolVersionPolicy) Next(currentState v1.State, payload []byte, output chan<- []byte) (v1.State, *v1.Error) {
	p.state = currentState
	return currentState, nil
}

func (p *ProtocolVersionPolicy) OnError(err *v1.Error, output chan<- []byte) (v1.State, *v1.Error) {
	switch err {
	case OtherVersionTooHigh:
		e := v1.Send(VersionResponse{
			Version: -1,
			Error:   NoSupportedVersions,
		}, output)

		if e != nil {
			p.state = v1.Abort

			return v1.Abort, ConnectionError.WithError(e)
		}

	}

	p.state = v1.Abort
	return v1.Abort, err
}

func (p *ProtocolVersionPolicy) OnEnd(output chan<- []byte) {

}

func (p *ProtocolVersionPolicy) OnNonRecoverableError(err *v1.Error, output chan<- []byte) error {
	switch err {
	case OurVersionTooHigh:
	case NoSupportedVersions:
		e := v1.Send(VersionResponse{
			Version: -1,
			Error:   NoSupportedVersions,
		}, output)

		if e != nil {
			return e
		}
	}
	return err.Error
}
