package user

import (
	"errors"
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
		return nil, errors.New("user not found")
	}
	return u, nil
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
