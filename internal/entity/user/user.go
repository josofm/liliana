package user

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Password string `json:"-"` // Password validation will be handled separately
	Email    string `json:"email" validate:"required,email"`
}
