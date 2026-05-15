package auth

import (
	"fmt"

	"github.com/josofm/liliana/internal/entity/auth"
	"github.com/josofm/liliana/internal/entity/user"
	userRepo "github.com/josofm/liliana/internal/repository/user"
)

type Service struct {
	userRepo   userRepo.Repository
	jwtService *JWTService
}

func NewService(userRepo userRepo.Repository, jwtService *JWTService) *Service {
	return &Service{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

// Register registra um novo usuário
func (s *Service) Register(req *auth.RegisterRequest) (*auth.AuthResponse, error) {
	// Verificar se o email já existe
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash da senha
	hashedPassword, err := s.jwtService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Criar usuário
	newUser := &user.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = s.userRepo.Create(newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Gerar tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(newUser.ID, newUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &auth.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(tokenPair.ExpiresAt.Unix()),
		User: auth.User{
			ID:    newUser.ID,
			Name:  newUser.Name,
			Email: newUser.Email,
		},
	}, nil
}

// Login autentica um usuário existente
func (s *Service) Login(req *auth.LoginRequest) (*auth.AuthResponse, error) {
	// Buscar usuário por email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verificar senha
	if !s.jwtService.CheckPassword(req.Password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Gerar tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &auth.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(tokenPair.ExpiresAt.Unix()),
		User: auth.User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// RefreshToken renova um access token usando o refresh token
func (s *Service) RefreshToken(refreshToken string) (*auth.AuthResponse, error) {
	// Validar refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Buscar usuário
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Gerar novos tokens
	tokenPair, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &auth.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(tokenPair.ExpiresAt.Unix()),
		User: auth.User{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	}, nil
}

// ValidateToken valida um access token e retorna as claims
func (s *Service) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// GetUserByID busca um usuário por ID
func (s *Service) GetUserByID(userID int64) (*user.User, error) {
	return s.userRepo.GetByID(userID)
}
