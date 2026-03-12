package model

type Message struct {
	Type string `json:"type"` // broadcast | private | group
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
}
