package group

import (
	"errors"
	entity "github.com/josofm/liliana/internal/entity/group"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	repository "github.com/josofm/liliana/internal/repository/group"
	"strings"
	"unicode/utf8"
)

var (
	ErrForbidden   = errors.New("group access denied")
	ErrInvalidName = errors.New("name must contain between 1 and 100 characters")
)

type UserReader interface {
	GetByID(int64) (*userEntity.User, error)
}
type Service struct {
	repo  repository.Repository
	users UserReader
}

func NewService(repo repository.Repository, users UserReader) *Service {
	return &Service{repo: repo, users: users}
}

func normalizeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < 1 || n > 100 {
		return "", ErrInvalidName
	}
	return name, nil
}
func (s *Service) Create(actorID int64, name, description string) (*entity.Group, error) {
	if actorID <= 0 {
		return nil, ErrForbidden
	}
	if _, err := s.users.GetByID(actorID); err != nil {
		return nil, err
	}
	name, err := normalizeName(name)
	if err != nil {
		return nil, err
	}
	g := &entity.Group{Name: name, Description: description, OwnerID: actorID}
	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}
func (s *Service) GetByID(actorID, id int64) (*entity.Group, error) {
	if actorID <= 0 {
		return nil, ErrForbidden
	}
	g, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	members, err := s.repo.GetMembers(id)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		if m.PlayerID == actorID {
			return g, nil
		}
	}
	return nil, ErrForbidden
}
func (s *Service) GetByPlayerID(actorID int64) ([]*entity.Group, error) {
	if actorID <= 0 {
		return nil, ErrForbidden
	}
	return s.repo.GetByPlayerID(actorID)
}
func (s *Service) Update(actorID, id int64, name, description *string) (*entity.Group, error) {
	g, err := s.GetByID(actorID, id)
	if err != nil {
		return nil, err
	}
	if g.OwnerID != actorID {
		return nil, ErrForbidden
	}
	if name != nil {
		g.Name, err = normalizeName(*name)
		if err != nil {
			return nil, err
		}
	}
	if description != nil {
		g.Description = *description
	}
	if err := s.repo.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}
func (s *Service) GetMembers(actorID, id int64) ([]entity.Member, error) {
	if _, err := s.GetByID(actorID, id); err != nil {
		return nil, err
	}
	return s.repo.GetMembers(id)
}
func (s *Service) RemoveMember(actorID, id, playerID int64) error {
	g, err := s.GetByID(actorID, id)
	if err != nil {
		return err
	}
	if actorID != g.OwnerID && actorID != playerID {
		return ErrForbidden
	}
	return s.repo.RemoveMember(id, playerID)
}
