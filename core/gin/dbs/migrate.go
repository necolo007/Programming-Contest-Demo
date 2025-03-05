package dbs

import (
	"Programming-Demo/internal/app/File/file_entity"
	"Programming-Demo/internal/app/user/user_entity"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&user_entity.User{},
		&user_entity.Token{},
		&file_entity.File{},
	)
	return err
}
