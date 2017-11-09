package main

import fmt "fmt"

type XOrgExt struct {
	Id           int
	Name         string
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	Country      string
	XUsers       []XUser `gorm:"ForeignKey:org_id;AssociationForeignKey:id"`
}
type XUser struct {
	Id           int
	FirstName    string
	LastName     string
	Email        string
	PhoneNumber  string
	AddressLine1 string
	AddressLine2 string
	UserType     string
	OrgId        int
}

func main() {
	fmt.Println("Hello, world")
}
