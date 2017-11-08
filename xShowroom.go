package main

import fmt "fmt"

type XOrgExt struct {
	Id           int
	Name         int
	AddressLine1 int
	AddressLine2 int
	City         int
	State        int
	Country      int
}
type XUser struct {
	Id           int
	FirstName    int
	LastName     int
	Email        int
	PhoneNumber  int
	AddressLine1 int
	AddressLine2 int
	UserType     int
}

func add(a int, b int) int {
	return a + b
}
func main() {
	fmt.Println("Hello, world")
	fmt.Println("Aatish Here")
	fmt.Println(add(2, 3))
}
