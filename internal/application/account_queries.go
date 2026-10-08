package application

import (
	"context"

	"github.com/basecamp/once-campfire-go/internal/database"
)

type AccountQueries struct{ DB *database.DB }

type AccountPeople struct {
	Administrators, Users []database.AccountMember
	NextPage              int64
}

func accountPage(number int64, count int) (int64, int64) {
	number = max(1, min(number, 1_000_000_000))
	last := int64(max(1, (count+499)/500))
	next := number + 1
	if number == last {
		next = 0
	}
	return number, next
}

func (q *AccountQueries) Settings(ctx context.Context, role int, number int64) (AccountPeople, error) {
	users, err := q.DB.AccountUsers(ctx, role == 1)
	if err != nil {
		return AccountPeople{}, err
	}
	var data AccountPeople
	_, data.NextPage = accountPage(number, len(users))
	for _, user := range users {
		if user.Role == 1 {
			data.Administrators = append(data.Administrators, user)
		} else {
			data.Users = append(data.Users, user)
		}
	}
	return data, nil
}

func (q *AccountQueries) Members(ctx context.Context, number int64) (AccountPeople, error) {
	users, err := q.DB.AccountUsers(ctx, false)
	if err != nil {
		return AccountPeople{}, err
	}
	number, next := accountPage(number, len(users))
	start := min(int((number-1)*500), len(users))
	return AccountPeople{Users: users[start:min(start+500, len(users))], NextPage: next}, nil
}
