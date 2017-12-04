package base_models

import "github.com/jinzhu/gorm"

type Category struct {
	gorm.Model
	Title       string `gorm:"type:varchar(255); not null"`
	Description string `gorm:"type:text"`
}
