package serializers

import "github.com/iReflect/reflect-app/apps/user/models"

type UserAuthSerializer struct {
	models.User
	Token string
}
