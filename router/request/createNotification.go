package request

import "errors"

type CreateNotification struct {
	To    []string          `json:"to"`
	From  string            `json:"from"`
	Event string            `json:"event"`
	Data  map[string]string `json:"data"`
}

func (r *CreateNotification) Validate() error {
	if len(r.To) == 0 {
		return errors.New("empty to")
	}
	if r.From == "" {
		return errors.New("empty from")
	}
	if r.Event == "" {
		return errors.New("empty event")
	}
	if r.Data == nil {
		return errors.New("empty data")
	}
	return nil
}
