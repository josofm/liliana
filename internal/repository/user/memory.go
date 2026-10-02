package user

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/josofm/liliana/internal/entity/user"
)

type inMemoryRepo struct {
	mu     sync.RWMutex
	users  map[int64]*user.User
	nextID int64
}

func NewInMemoryRepo() Repository {
	return &inMemoryRepo{
		users:  make(map[int64]*user.User),
		nextID: 1,
	}
}

func (r *inMemoryRepo) Create(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u.ID = r.nextID
	r.users[u.ID] = u
	r.nextID++
	return nil
}

func (r *inMemoryRepo) GetAll() ([]*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*user.User
	for _, u := range r.users {
		result = append(result, u)
	}
	return result, nil
}

func (r *inMemoryRepo) GetByID(id int64) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	return u, nil
}

func (r *inMemoryRepo) GetByEmail(email string) (*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *inMemoryRepo) Search(query string) ([]*user.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	query = strings.ToLower(query)
	result := make([]*user.User, 0)
	for _, u := range r.users {
		if strings.Contains(strings.ToLower(u.Email), query) {
			result = append(result, &user.User{ID: u.ID, Name: u.Name, Email: u.Email})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := strings.ToLower(result[i].Name), strings.ToLower(result[j].Name)
		if a == b {
			return result[i].ID < result[j].ID
		}
		return a < b
	})
	if len(result) > 20 {
		result = result[:20]
	}
	return result, nil
}

func (r *inMemoryRepo) Update(id int64, u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[id]; !exists {
		return errors.New("user not found")
	}
	u.ID = id
	r.users[id] = u
	return nil
}

func (r *inMemoryRepo) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}
