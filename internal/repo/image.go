package repo

import (
	"database/sql"
	"fmt"
	"upgrade-tracker/internal/model"
)

type ImageRepo struct{ db *sql.DB }

func NewImageRepo(db *sql.DB) *ImageRepo { return &ImageRepo{db: db} }

func (r *ImageRepo) List(search string) ([]*model.ProductImage, error) {
	query := `
		SELECT id, name, version, type, public_url, internal_url, config_guide, description, created_at, updated_at
		FROM product_images`
	args := []any{}
	if search != "" {
		query += " WHERE name LIKE ? OR description LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*model.ProductImage
	for rows.Next() {
		i := &model.ProductImage{}
		if err := rows.Scan(&i.ID, &i.Name, &i.Version, &i.Type, &i.PublicURL, &i.InternalURL, &i.ConfigGuide, &i.Description, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		images = append(images, i)
	}
	if images == nil {
		images = []*model.ProductImage{}
	}
	return images, nil
}

func (r *ImageRepo) Get(id int) (*model.ProductImage, error) {
	i := &model.ProductImage{}
	err := r.db.QueryRow(`
		SELECT id, name, version, type, public_url, internal_url, config_guide, description, created_at, updated_at
		FROM product_images WHERE id=?`, id).
		Scan(&i.ID, &i.Name, &i.Version, &i.Type, &i.PublicURL, &i.InternalURL, &i.ConfigGuide, &i.Description, &i.CreatedAt, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("image not found")
	}
	return i, err
}

func (r *ImageRepo) Create(name, version, imgType, publicURL, internalURL, configGuide, description string) (*model.ProductImage, error) {
	res, err := r.db.Exec(
		"INSERT INTO product_images (name, version, type, public_url, internal_url, config_guide, description) VALUES (?,?,?,?,?,?,?)",
		name, version, imgType, publicURL, internalURL, configGuide, description)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.Get(int(id))
}

func (r *ImageRepo) Update(id int, name, version, imgType, publicURL, internalURL, configGuide, description string) error {
	_, err := r.db.Exec(
		"UPDATE product_images SET name=?, version=?, type=?, public_url=?, internal_url=?, config_guide=?, description=? WHERE id=?",
		name, version, imgType, publicURL, internalURL, configGuide, description, id)
	return err
}

func (r *ImageRepo) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM product_images WHERE id=?", id)
	return err
}
