package models

type Poll struct {
	ID           int64    `json:"id"`
	Question     string   `json:"question"`
	MemberYes    int64    `json:"member_yes"`
	MemberNo     int64    `json:"member_no"`
	NonMemberYes int64    `json:"non_member_yes"`
	NonMemberNo  int64    `json:"non_member_no"`
	TotalVotes   int      `json:"total_votes"`
	WhoVoted     []string `json:"who_voted"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	ExpiresAt    string   `json:"expires_at"`
}
