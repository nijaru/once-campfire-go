package database

import (
	"context"
	"time"
)

func (d *DB) SessionUser(ctx context.Context, token string) (User, error) {
	return userRow(d.Read.QueryRowContext(ctx,
		"SELECT "+userColumns+" FROM users u JOIN sessions s ON s.user_id=u.id WHERE s.token=? AND u.status=0", token))
}

// AuthenticateSession observes the user and hourly activity threshold in one
// fresh read. The conditional update elects at most one cookie refresher when
// concurrent requests cross that threshold; ordinary authentication never writes.
func (d *DB) AuthenticateSession(ctx context.Context, token, agent, ip string) (User, bool, error) {
	var u User
	var active time.Time
	err := d.Read.QueryRowContext(ctx,
		"SELECT "+userColumns+",s.last_active_at FROM users u JOIN sessions s ON s.user_id=u.id WHERE s.token=? AND u.status=0", token).
		Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Role, &u.Status, &u.Bio,
			timestamp{&u.UpdatedAt}, &u.BotToken, timestamp{&active})
	if err != nil {
		return u, false, err
	}
	now := d.Now()
	if !active.Before(now.Add(-time.Hour)) {
		return u, false, nil
	}
	r, err := d.Write.ExecContext(ctx,
		"UPDATE sessions SET last_active_at=?,updated_at=?,user_agent=?,ip_address=? WHERE token=? AND last_active_at<?",
		Stamp(now), Stamp(now), agent, ip, token, Stamp(now.Add(-time.Hour)))
	if err != nil {
		return u, false, err
	}
	n, err := r.RowsAffected()
	return u, n > 0, err
}

func (d *DB) StartSession(ctx context.Context, user int64, agent, ip string) (string, error) {
	token, now := Token(), Stamp(d.Now())
	_, err := d.Write.ExecContext(ctx,
		"INSERT INTO sessions(token,user_id,user_agent,ip_address,last_active_at,created_at,updated_at) VALUES (?,?,?,?,?,?,?)",
		token, user, agent, ip, now, now, now)
	return token, err
}
