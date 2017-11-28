package models

import "github.com/jinzhu/gorm"

type ResponseComment struct {
	gorm.Model
	QuestionResponse   QuestionResponse `gorm:"ForeignKey:QuestionResponseID; AssociationForeignKey:ID"`
	QuestionResponseID uint             `gorm:"not null"`
	Comment            string           `gorm:"type:text; not null"`
	Deleted            bool             `gorm:"default:false; not null"`
}
