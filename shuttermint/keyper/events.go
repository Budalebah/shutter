package keyper

import (
	"bytes"
	"fmt"
	"strconv"

	abcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/brainbot-com/shutter/shuttermint/app"
)

// MakePrivkeyGeneratedEvent creates a PrivkeyGeneratedEvent from the given tendermint event of
// type "shutter.privkey-generated"
func MakePrivkeyGeneratedEvent(ev abcitypes.Event) (PrivkeyGeneratedEvent, error) {
	if len(ev.Attributes) < 2 {
		return PrivkeyGeneratedEvent{}, fmt.Errorf("Event contains not enough attributes: %+v", ev)
	}
	if !bytes.Equal(ev.Attributes[0].Key, []byte("BatchIndex")) || !bytes.Equal(ev.Attributes[1].Key, []byte("Privkey")) {
		return PrivkeyGeneratedEvent{}, fmt.Errorf("Bad event attributes: %+v", ev)
	}

	b, err := strconv.Atoi(string(ev.Attributes[0].Value))
	if err != nil {
		return PrivkeyGeneratedEvent{}, err
	}
	privkey, err := app.DecodePrivkeyFromEvent(string(ev.Attributes[1].Value))
	if err != nil {
		return PrivkeyGeneratedEvent{}, err
	}

	return PrivkeyGeneratedEvent{uint64(b), privkey}, nil
}

// MakePubkeyGeneratedEvent creates a PubkeyGeneratedEvent from the given tendermint event of type
// type "shutter.pubkey-generated"
func MakePubkeyGeneratedEvent(ev abcitypes.Event) (PubkeyGeneratedEvent, error) {
	if len(ev.Attributes) < 2 {
		return PubkeyGeneratedEvent{}, fmt.Errorf("Event contains not enough attributes: %+v", ev)
	}
	if !bytes.Equal(ev.Attributes[0].Key, []byte("BatchIndex")) || !bytes.Equal(ev.Attributes[1].Key, []byte("Pubkey")) {
		return PubkeyGeneratedEvent{}, fmt.Errorf("Bad event attributes: %+v", ev)
	}

	b, err := strconv.Atoi(string(ev.Attributes[0].Value))
	if err != nil {
		return PubkeyGeneratedEvent{}, err
	}
	pubkey, err := app.DecodePubkeyFromEvent(string(ev.Attributes[1].Value))
	if err != nil {
		return PubkeyGeneratedEvent{}, err
	}

	return PubkeyGeneratedEvent{uint64(b), pubkey}, nil
}

// MakeBatchConfigEvent creates a BatchConfigEvent from the given tendermint event of type
// "shutter.batch-config"
func MakeBatchConfigEvent(ev abcitypes.Event) (BatchConfigEvent, error) {
	if len(ev.Attributes) < 3 {
		return BatchConfigEvent{}, fmt.Errorf("Event contains not enough attributes: %+v", ev)
	}
	if !bytes.Equal(ev.Attributes[0].Key, []byte("StartBatchIndex")) ||
		!bytes.Equal(ev.Attributes[1].Key, []byte("Threshold")) ||
		!bytes.Equal(ev.Attributes[2].Key, []byte("Keypers")) {
		return BatchConfigEvent{}, fmt.Errorf("Bad event attributes: %+v", ev)
	}

	b, err := strconv.Atoi(string(ev.Attributes[0].Value))
	if err != nil {
		return BatchConfigEvent{}, err
	}

	threshold, err := strconv.Atoi(string(ev.Attributes[1].Value))
	if err != nil {
		return BatchConfigEvent{}, err
	}
	keypers := app.DecodeAddressesFromEvent(string(ev.Attributes[2].Value))
	return BatchConfigEvent{uint64(b), uint32(threshold), keypers}, nil
}

// MakeEvent creates an Event from the given tendermint event. It will return a
// PubkeyGeneratedEvent, PrivkeyGeneratedEvent or BatchConfigEvent based on the event's type.
func MakeEvent(ev abcitypes.Event) (IEvent, error) {
	if ev.Type == "shutter.privkey-generated" {
		res, err := MakePrivkeyGeneratedEvent(ev)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	if ev.Type == "shutter.pubkey-generated" {
		res, err := MakePubkeyGeneratedEvent(ev)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
	if ev.Type == "shutter.batch-config" {
		res, err := MakeBatchConfigEvent(ev)
		if err != nil {
			return nil, err
		}
		return res, nil

	}
	return nil, fmt.Errorf("Cannot make event from %+v", ev)
}