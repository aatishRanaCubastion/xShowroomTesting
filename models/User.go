package models

type User struct {
	Id           uint   `gorm:"column:id"`
	FirstName    string `gorm:"column:first_name"`
	LastName     string `gorm:"column:last_name"`
	Email        string `gorm:"column:email"`
	PhoneNumber  string `gorm:"column:phone_number"`
	AddressLine1 string `gorm:"column:address_line_1"`
	AddressLine2 string `gorm:"column:address_line_2"`
	UserType     string `gorm:"column:user_type"`
	OrgId        uint   `gorm:"column:org_id"`
}

func (User) TableName() string {
	return "x_user"
}
