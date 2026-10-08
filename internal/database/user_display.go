package database

import (
	"context"
	"database/sql"
)

func (d *DB) UserDisplays(ctx context.Context, ids []int64) (map[int64]UserDisplay, error) {
	unique := make(map[int64]bool, len(ids))
	for _, id := range ids {
		unique[id] = true
	}
	if len(unique) == 0 {
		return map[int64]UserDisplay{}, nil
	}
	tx, err := d.Read.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	users, err := userDisplays(ctx, tx, unique)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return users, nil
}

func userDisplays(ctx context.Context, tx *sql.Tx, ids map[int64]bool) (map[int64]UserDisplay, error) {
	users := make(map[int64]UserDisplay, len(ids))
	if len(ids) == 0 {
		return users, nil
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,name,coalesce(bio,''),updated_at FROM users WHERE id IN (SELECT value FROM json_each(?))", displayIDs(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user UserDisplay
		if err := rows.Scan(&user.ID, &user.Name, &user.Bio, timestamp{&user.UpdatedAt}); err != nil {
			return nil, err
		}
		users[user.ID] = user
	}
	return users, rows.Err()
}
