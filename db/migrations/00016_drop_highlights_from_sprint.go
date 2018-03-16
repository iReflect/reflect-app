package migrations
	
import (
	"database/sql"
	"github.com/pressly/goose"
	"github.com/jinzhu/gorm"
)

//Define only the fields used in this migration and not full model.
//type Category struct {
//	gorm.Model
//	Weight int
//}

func init() {
	goose.AddMigration(Up00016, Down00016)
}

// Up00016 ...
func Up00016(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
	}

	// Drop added columns
	gormdb.Model(&sprint{}).DropColumn("good_highlights")
	gormdb.Model(&sprint{}).DropColumn("okay_highlights")
	gormdb.Model(&sprint{}).DropColumn("bad_highlights")

	return nil
}

// Down00016 ...
func Down00016(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	gormdb, err := gorm.Open("postgres", interface{}(tx).(gorm.SQLCommon))
	if err != nil {
		return err
	}

	type sprint struct {
		GoodHighlights string `gorm:"type:text"`
		OkayHighlights string `gorm:"type:text"`
		BadHighlights  string `gorm:"type:text"`
	}

	// AutoMigrate sprint Model
	gormdb.AutoMigrate(&sprint{})

	return nil
}
