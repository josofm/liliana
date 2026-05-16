package user

import (
	"database/sql"
	"errors"

	userEntity "github.com/josofm/liliana/internal/entity/user"
)

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Create(u *userEntity.User) error {
	const query = `
		INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	return r.db.QueryRow(query, u.Name, u.Email, u.Password).Scan(&u.ID)
}

func (r *postgresRepo) GetAll() ([]*userEntity.User, error) {
	const query = `
		SELECT id, name, email, password
		FROM users
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*userEntity.User, 0)
	for rows.Next() {
		u := &userEntity.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *postgresRepo) GetByID(id int64) (*userEntity.User, error) {
	const query = `
		SELECT id, name, email, password
		FROM users
		WHERE id = $1
	`

	u := &userEntity.User{}
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Name, &u.Email, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *postgresRepo) GetByEmail(email string) (*userEntity.User, error) {
	const query = `
		SELECT id, name, email, password
		FROM users
		WHERE email = $1
	`

	u := &userEntity.User{}
	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Name, &u.Email, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *postgresRepo) Update(id int64, u *userEntity.User) error {
	const query = `
		UPDATE users
		SET name = $1, email = $2, password = $3
		WHERE id = $4
	`

	result, err := r.db.Exec(query, u.Name, u.Email, u.Password, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	u.ID = id
	return nil
}

func (r *postgresRepo) Delete(id int64) error {
	const query = `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
