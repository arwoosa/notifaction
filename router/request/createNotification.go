package request

type CreateNotification struct {
	To    []string          `json:"to"`
	From  string            `json:"from"`
	Event string            `json:"event"`
	Data  map[string]string `json:"data"`
}

func (r *CreateNotification) Validate() error {
	// TODO: complete
	return nil
}
