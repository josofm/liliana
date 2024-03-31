package user

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Sex      string `json:"sex"`
	Age      string `json:"age"`
}
