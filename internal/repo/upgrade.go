package repo

import (
	"database/sql"
	"fmt"
	"upgrade-tracker/internal/model"
)

type UpgradeRepo struct{ db *sql.DB }

func NewUpgradeRepo(db *sql.DB) *UpgradeRepo { return &UpgradeRepo{db: db} }

func (r *UpgradeRepo) ListByClient(clientID int) ([]*model.UpgradeRecord, error) {
	rows, err := r.db.Query(`
		SELECT id, client_id, version, DATE_FORMAT(upgrade_date,'%Y-%m-%d'),
		       operator, COALESCE(tags,''), COALESCE(description,''), COALESCE(files,''), created_at
		FROM upgrade_records WHERE client_id=?
		ORDER BY upgrade_date DESC, id DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.UpgradeRecord
	for rows.Next() {
		u := &model.UpgradeRecord{}
		var tags, files string
		if err := rows.Scan(&u.ID, &u.ClientID, &u.Version, &u.UpgradeDate,
			&u.Operator, &tags, &u.Description, &files, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Tags = model.SplitTrim(tags)
		u.Files = model.SplitTrim(files)
		list = append(list, u)
	}
	if list == nil {
		list = []*model.UpgradeRecord{}
	}
	return list, nil
}

func (r *UpgradeRepo) Create(clientID int, version, date, operator, tags, desc, files string) (*model.UpgradeRecord, error) {
	res, err := r.db.Exec(
		`INSERT INTO upgrade_records (client_id, version, upgrade_date, operator, tags, description, files)
		 VALUES (?,?,?,?,?,?,?)`,
		clientID, version, date, operator, tags, desc, files)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.Get(int(id))
}

func (r *UpgradeRepo) Get(id int) (*model.UpgradeRecord, error) {
	u := &model.UpgradeRecord{}
	var tags, files string
	err := r.db.QueryRow(`
		SELECT id, client_id, version, DATE_FORMAT(upgrade_date,'%Y-%m-%d'),
		       operator, COALESCE(tags,''), COALESCE(description,''), COALESCE(files,''), created_at
		FROM upgrade_records WHERE id=?`, id).
		Scan(&u.ID, &u.ClientID, &u.Version, &u.UpgradeDate,
			&u.Operator, &tags, &u.Description, &files, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("record not found")
	}
	if err != nil {
		return nil, err
	}
	u.Tags = model.SplitTrim(tags)
	u.Files = model.SplitTrim(files)
	return u, nil
}

func (r *UpgradeRepo) Delete(id int) (int, error) {
	var clientID int
	err := r.db.QueryRow("SELECT client_id FROM upgrade_records WHERE id=?", id).Scan(&clientID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("record not found")
	}
	if err != nil {
		return 0, err
	}
	_, err = r.db.Exec("DELETE FROM upgrade_records WHERE id=?", id)
	return clientID, err
}

// LatestVersion returns the most recent version for a client (empty string if none)
func (r *UpgradeRepo) LatestVersion(clientID int) string {
	var v string
	r.db.QueryRow(
		"SELECT version FROM upgrade_records WHERE client_id=? ORDER BY upgrade_date DESC, id DESC LIMIT 1",
		clientID).Scan(&v)
	return v
}
