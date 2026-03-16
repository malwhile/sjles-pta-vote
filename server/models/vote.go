package models

type Vote struct {
	PollId int64 `json:"poll_id"`
	Vote bool `json:"vote"`
	Email string `json:"email"`
}