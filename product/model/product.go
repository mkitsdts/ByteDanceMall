package model

type Product struct {
	ID          uint64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string  `json:"name" gorm:"type:varchar(255);index:idx_product_name;not null"`
	Description string  `json:"description" gorm:"type:text"`
	Price       float32 `json:"price" gorm:"type:decimal(10,2);not null;default:0.00"`
	ImageURL    string  `json:"image_url" gorm:"type:varchar(512)"`
	CategoryID  uint64  `json:"category_id" gorm:"index:idx_product_category"`
	Status      int8    `json:"status" gorm:"default:1;comment:'1:在售 2:下架 3:删除'"`
	LabelID     uint64  `json:"label_id" gorm:"type:varchar(255);comment:'标签，以逗号分隔'"`
}
