package base

import "github.com/jinzhu/gorm"
import "github.com/iReflect/reflect-app/db/models/fields"

type User struct {
	gorm.Model
	Email string `gorm:"type:varchar(100);unique_index"`
	Name  string
	Type  string
}

type Group struct {
	gorm.Model
	Name string
}

type Role struct {
	gorm.Model
	Name    string
	UserID  uint `gorm:"index"`
	GroupID uint `gorm:"index"`
	Type    string
}

type Category struct {
	gorm.Model
	Title    string `gorm:"type:varchar(100);unique_index"`
	Archived bool
	Items    []Item
}

type ItemType struct {
	gorm.Model
	Title  string
	Weight int
}

type Item struct {
	gorm.Model
	Title      string
	Weight     int
	CategoryID uint         `gorm:"index"`
	ItemTypeID uint         `gorm:"index"`
	Data       fields.JSONB `type: jsonb not null default '{}'::jsonb`
	Category   Category
	ItemType   ItemType
}
