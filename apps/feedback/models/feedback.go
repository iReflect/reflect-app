package models

import (
	"github.com/iReflect/reflect-app/db/models/fields"
	"github.com/jinzhu/gorm"
)

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
