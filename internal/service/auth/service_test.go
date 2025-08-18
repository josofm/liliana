package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/josofm/liliana/internal/entity/auth"
	userEntity "github.com/josofm/liliana/internal/entity/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock repository para testes
type mockUserRepo struct {
	users  map[int64]*userEntity.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[int64]*userEntity.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(u *userEntity.User) error {
	u.ID = m.nextID
	m.users[u.ID] = u
	m.nextID++
	return nil
}

func (m *mockUserRepo) GetByID(id int64) (*userEntity.User, error) {
	if u, exists := m.users[id]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepo) GetByEmail(email string) (*userEntity.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepo) GetAll() ([]*userEntity.User, error) {
	var result []*userEntity.User
	for _, u := range m.users {
		result = append(result, u)
	}
	return result, nil
}

func (m *mockUserRepo) Update(id int64, u *userEntity.User) error {
	if _, exists := m.users[id]; !exists {
		return errors.New("user not found")
	}
	u.ID = id
	m.users[id] = u
	return nil
}

func (m *mockUserRepo) Delete(id int64) error {
	if _, exists := m.users[id]; !exists {
		return errors.New("user not found")
	}
	delete(m.users, id)
	return nil
}

func TestNewService(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	service := NewService(userRepo, jwtService)
	assert.NotNil(t, service)
	assert.Equal(t, userRepo, service.userRepo)
	assert.Equal(t, jwtService, service.jwtService)
}

func TestRegister(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	service := NewService(userRepo, jwtService)

	req := &auth.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := service.Register(req)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(1), response.User.ID)
	assert.Equal(t, "Test User", response.User.Name)
	assert.Equal(t, "test@example.com", response.User.Email)

	// Verificar se o usuário foi criado no repositório
	createdUser, err := userRepo.GetByID(1)
	require.NoError(t, err)
	assert.Equal(t, "Test User", createdUser.Name)
	assert.Equal(t, "test@example.com", createdUser.Email)
	assert.NotEqual(t, "password123", createdUser.Password) // Senha deve estar hasheada
}

func TestRegister_DuplicateEmail(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	service := NewService(userRepo, jwtService)

	// Criar primeiro usuário
	req1 := &auth.RegisterRequest{
		Name:     "User 1",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(req1)
	require.NoError(t, err)

	// Tentar criar segundo usuário com mesmo email
	req2 := &auth.RegisterRequest{
		Name:     "User 2",
		Email:    "test@example.com",
		Password: "password456",
	}
	_, err = service.Register(req2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
}

func TestLogin(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	})

	service := NewService(userRepo, jwtService)

	// Criar usuário primeiro
	req := &auth.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(req)
	require.NoError(t, err)

	// Fazer login
	loginReq := &auth.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	response, err := service.Login(loginReq)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, int64(1), response.User.ID)
	assert.Equal(t, "Test User", response.User.Name)
	assert.Equal(t, "test@example.com", response.User.Email)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	service := NewService(userRepo, jwtService)

	// Tentar login com usuário inexistente
	loginReq := &auth.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	_, err := service.Login(loginReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")

	// Criar usuário
	req := &auth.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err = service.Register(req)
	require.NoError(t, err)

	// Tentar login com senha incorreta
	loginReq = &auth.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	_, err = service.Login(loginReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestServiceRefreshToken(t *testing.T) {
	userRepo := newMockUserRepo()
	mockTime := NewMockTimeProvider()
	jwtService := NewJWTServiceWithTimeProvider(JWTConfig{
		SecretKey:     "test-secret",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}, mockTime)

	service := NewService(userRepo, jwtService)

	// Criar usuário e fazer login
	req := &auth.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	registerResponse, err := service.Register(req)
	require.NoError(t, err)

	// Avançar o tempo para garantir timestamps diferentes
	mockTime.Advance(1 * time.Second)

	// Renovar token
	response, err := service.RefreshToken(registerResponse.RefreshToken)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)

	// Verificar que os tokens são diferentes (devido ao timestamp diferente)
	assert.NotEqual(t, registerResponse.AccessToken, response.AccessToken)
	assert.NotEqual(t, registerResponse.RefreshToken, response.RefreshToken)

	// Verificar que os novos tokens são válidos
	claims, err := jwtService.ValidateToken(response.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestServiceRefreshToken_InvalidToken(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	service := NewService(userRepo, jwtService)

	// Tentar renovar com token inválido
	_, err := service.RefreshToken("invalid-token")
	assert.Error(t, err)
}

func TestGetUserByID(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	service := NewService(userRepo, jwtService)

	// Criar usuário
	req := &auth.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	_, err := service.Register(req)
	require.NoError(t, err)

	// Buscar usuário
	user, err := service.GetUserByID(1)
	require.NoError(t, err)
	assert.Equal(t, "Test User", user.Name)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestGetUserByID_NotFound(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtService := NewJWTService(JWTConfig{
		SecretKey: "test-secret",
	})

	service := NewService(userRepo, jwtService)

	// Buscar usuário inexistente
	_, err := service.GetUserByID(999)
	assert.Error(t, err)
}
