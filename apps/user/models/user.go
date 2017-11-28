package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Email     string `gorm:"type:varchar(100); not null; unique_index"`
	FirstName string `gorm:"type:varchar(30); not null"`
	LastName  string `gorm:"type:varchar(30)"`
	Role      Role   `gorm:"ForeignKey:RoleID;AssociationForeignKey:ID"`
	RoleID    uint   `gorm:"not null"`
	Active    bool   `gorm:"default:true; not null"`
}
