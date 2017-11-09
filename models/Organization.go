package models

type Organization struct {
	Id           uint   `gorm:"column:id"`
	Name         string `gorm:"column:name"`
	AddressLine1 string `gorm:"column:address_line_1"`
	AddressLine2 string `gorm:"column:address_line_2"`
	City         string `gorm:"column:city"`
	State        string `gorm:"column:state"`
	Country      string `gorm:"column:country"`
	Users        []User `gorm:"ForeignKey:org_id;AssociationForeignKey:id"`
}

func (Organization) TableName() string {
	return "x_org_ext"
}
