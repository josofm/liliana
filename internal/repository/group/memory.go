package group

import (
	entity "github.com/josofm/liliana/internal/entity/group"
	"sort"
	"sync"
	"time"
)

type inMemoryRepo struct {
	mu      sync.RWMutex
	groups  map[int64]entity.Group
	members map[int64]map[int64]entity.Member
	nextID  int64
}

func NewInMemoryRepo() Repository {
	return &inMemoryRepo{groups: make(map[int64]entity.Group), members: make(map[int64]map[int64]entity.Member), nextID: 1}
}

func (r *inMemoryRepo) Create(g *entity.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g.ID = r.nextID
	r.nextID++
	g.CreatedAt = time.Now().UTC()
	g.UpdatedAt = g.CreatedAt
	r.groups[g.ID] = *g
	r.members[g.ID] = map[int64]entity.Member{g.OwnerID: {GroupID: g.ID, PlayerID: g.OwnerID, Role: entity.RoleOwner, JoinedAt: g.CreatedAt}}
	return nil
}

func (r *inMemoryRepo) GetByID(id int64) (*entity.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.groups[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &g, nil
}

func (r *inMemoryRepo) GetByPlayerID(id int64) ([]*entity.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*entity.Group, 0)
	for groupID, members := range r.members {
		if _, ok := members[id]; ok {
			g := r.groups[groupID]
			result = append(result, &g)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func (r *inMemoryRepo) Update(g *entity.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.groups[g.ID]
	if !ok {
		return ErrNotFound
	}
	stored.Name = g.Name
	stored.Description = g.Description
	stored.UpdatedAt = time.Now().UTC()
	r.groups[g.ID] = stored
	*g = stored
	return nil
}

func (r *inMemoryRepo) GetMembers(id int64) ([]entity.Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.groups[id]; !ok {
		return nil, ErrNotFound
	}
	result := make([]entity.Member, 0, len(r.members[id]))
	for _, m := range r.members[id] {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PlayerID < result[j].PlayerID })
	return result, nil
}

func (r *inMemoryRepo) AddMember(id, playerID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[id]; !ok {
		return ErrNotFound
	}
	if _, ok := r.members[id][playerID]; ok {
		return ErrAlreadyMember
	}
	r.members[id][playerID] = entity.Member{GroupID: id, PlayerID: playerID, Role: entity.RoleMember, JoinedAt: time.Now().UTC()}
	return nil
}

func (r *inMemoryRepo) RemoveMember(id, playerID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	g, ok := r.groups[id]
	if !ok {
		return ErrNotFound
	}
	if g.OwnerID == playerID {
		return ErrOwnerCannotLeave
	}
	if _, ok := r.members[id][playerID]; !ok {
		return ErrMemberNotFound
	}
	delete(r.members[id], playerID)
	return nil
}
