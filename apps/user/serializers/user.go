package serializers

import "github.com/iReflect/reflect-app/apps/user/models"

type UserAuthSerializer struct {
	models.User
	Token string
}

type User struct {
	ID        uint
	Email     string
	FirstName string
	LastName  string
	Active    bool
}

// MemberSerializer ...
type MembersSerializer struct {
	Members       []User
}
