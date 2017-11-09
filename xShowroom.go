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
}

func add(a int, b int) int {
	return a + b
}
func main() {
	fmt.Println("Hello, world")
	fmt.Println("Aatish Here")
	fmt.Println(add(2, 3))
}
