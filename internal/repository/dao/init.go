package dao

import "gorm.io/gorm"

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Article{},
		&PublishedArticle{},
		&Interactive{},
		&UserLikeBiz{},
		&UserCollectionBiz{},
		&AsyncSMS{},
		&Job{},
		&Comment{},
		&FollowRelation{},
		&Tag{},
		&TagBiz{},
	)
}
