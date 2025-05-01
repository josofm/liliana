package user

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Password string `json:"-"`
	Email    string `json:"email"`
}
