package user

import entity "github.com/josofm/liliana/internal/entity/user"

// strpos treats %, _ and backslashes as literal input, not LIKE wildcards.
func (r *postgresRepo) Search(query string) ([]*entity.User, error) {
	rows, err := r.db.Query(`SELECT id,name,email FROM users WHERE strpos(lower(email),lower($1))>0 ORDER BY lower(name),id LIMIT 20`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]*entity.User, 0)
	for rows.Next() {
		u := &entity.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
