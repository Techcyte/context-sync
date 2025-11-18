package model

type ContextKey string

const (
	CaseNumber ContextKey = "case"
)

type ContextItem struct {
	Key   ContextKey `json:"key"`
	Value string     `json:"value"`
}
