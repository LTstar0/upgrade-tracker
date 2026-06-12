package repo

import (
	"database/sql"
	"fmt"
	"time"
	"upgrade-tracker/internal/model"
)

type ClientRepo struct{ db *sql.DB }

func NewClientRepo(db *sql.DB) *ClientRepo { return &ClientRepo{db: db} }

func (r *ClientRepo) List(search string) ([]*model.Client, *model.Stats, error) {
	query := `
		SELECT c.id, c.name, c.type, COALESCE(c.contact,''), COALESCE(c.note,''),
		       c.current_version, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM upgrade_records u WHERE u.client_id=c.id) AS upgrade_count
		FROM clients c`
	args := []any{}
	if search != "" {
		query += " WHERE c.name LIKE ? OR c.contact LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	query += " ORDER BY c.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var clients []*model.Client
	for rows.Next() {
		c := &model.Client{}
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Contact, &c.Note,
			&c.CurrentVersion, &c.CreatedAt, &c.UpdatedAt, &c.UpgradeCount); err != nil {
			return nil, nil, err
		}
		clients = append(clients, c)
	}
	if clients == nil {
		clients = []*model.Client{}
	}

	stats, err := r.stats()
	return clients, stats, err
}

func (r *ClientRepo) stats() (*model.Stats, error) {
	s := &model.Stats{}
	r.db.QueryRow("SELECT COUNT(*) FROM clients").Scan(&s.TotalClients)
	r.db.QueryRow("SELECT COUNT(*) FROM upgrade_records").Scan(&s.TotalUpgrades)
	month := time.Now().Format("2006-01")
	r.db.QueryRow(
		"SELECT COUNT(*) FROM upgrade_records WHERE DATE_FORMAT(upgrade_date,'%Y-%m')=?", month,
	).Scan(&s.MonthUpgrades)
	return s, nil
}

func (r *ClientRepo) Get(id int) (*model.Client, error) {
	c := &model.Client{}
	err := r.db.QueryRow(`
		SELECT c.id, c.name, c.type, COALESCE(c.contact,''), COALESCE(c.note,''),
		       c.current_version, c.created_at, c.updated_at,
		       (SELECT COUNT(*) FROM upgrade_records u WHERE u.client_id=c.id)
		FROM clients c WHERE c.id=?`, id).
		Scan(&c.ID, &c.Name, &c.Type, &c.Contact, &c.Note,
			&c.CurrentVersion, &c.CreatedAt, &c.UpdatedAt, &c.UpgradeCount)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("client not found")
	}
	return c, err
}

func (r *ClientRepo) Create(name, typ, contact, note, version string) (*model.Client, error) {
	res, err := r.db.Exec(
		"INSERT INTO clients (name, type, contact, note, current_version) VALUES (?,?,?,?,?)",
		name, typ, contact, note, version)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.Get(int(id))
}

func (r *ClientRepo) Update(id int, name, typ, contact, note string) error {
	_, err := r.db.Exec(
		"UPDATE clients SET name=?, type=?, contact=?, note=? WHERE id=?",
		name, typ, contact, note, id)
	return err
}

func (r *ClientRepo) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM clients WHERE id=?", id)
	return err
}

func (r *ClientRepo) SetVersion(id int, version string) error {
	_, err := r.db.Exec("UPDATE clients SET current_version=? WHERE id=?", version, id)
	return err
}
