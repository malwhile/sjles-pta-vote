package models

import (
    "go-sjles-pta-vote/server/db"
)

type Voter struct {
    Email   string `json:"email"`
    IsMember bool   `json:"is_member"`
    YesVote  bool   `json:"yes_vote"`
}

func GetVoters(pollId int64) ([]Voter, error) {
    db_conn, err := db.Connect()
    if err != nil {
        return nil, err
    }
    defer db.Close()

    rows, err := db_conn.Query(`
        SELECT v.voter_email,
               CASE
                   WHEN m.email IS NOT NULL THEN 1
                   ELSE 0
               END AS is_member,
               CASE
                   WHEN p.member_yes_votes + p.non_member_yes_votes > p.member_no_votes + p.non_member_no_votes THEN 1
                   ELSE 0
               END AS yes_vote
        FROM voters v
        LEFT JOIN members m ON v.voter_email = m.email
        LEFT JOIN polls p ON v.poll_id = p.id
        WHERE v.poll_id = $1
    `, pollId)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var voters []Voter
    for rows.Next() {
        var voter Voter
        var isMember int
        var yesVote int
        if err := rows.Scan(&voter.Email, &isMember, &yesVote); err != nil {
            return nil, err
        }
        voter.IsMember = isMember == 1
        voter.YesVote = yesVote == 1
        voters = append(voters, voter)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }

    return voters, nil
}
