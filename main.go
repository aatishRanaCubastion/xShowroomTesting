package main

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"os"
	"log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
)

type Entity struct {
	ID          uint `sql:"AUTO_INCREMENT"`
	Name        string `sql:"type:varchar(30)"`
	DisplayName string `sql:"type:varchar(30)"`
}

type ColumnType struct {
	ID   uint    `sql:"AUTO_INCREMENT"`
	Type string `sql:"type:varchar(30)"`
}

type Column struct {
	ID       uint `sql:"AUTO_INCREMENT"`
	Name     string `sql:"type:varchar(30)"`
	Size     uint `sql:"type:int(30)"`
	TypeID   uint `sql:"type:int(30)"`
	EntityID uint `sql:"type:int(100)"`
}

func (Entity) TableName() string {
	return "c_entity"
}

func (ColumnType) TableName() string {
	return "c_column_type"
}

func (Column) TableName() string {
	return "c_column"
}

func main() {

	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/xshowroomcustom?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	db.AutoMigrate(&Entity{}, &Column{}, &ColumnType{})

	//get all entities
	entities := []Entity{}
	db.Find(&entities)

	allColumns := []Column{}
	db.Find(&allColumns)

	file, err := os.Create("xShowroom.go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//created file
	f := NewFile("main")

	//declaring method
	f.Func().Id("add").Params(
		Id("a").Int(),
		Id("b").Int(),
	).Int().Block(
		Return(Id("a").Op("+").Id("b")),
	)

	//calling method
	f.Func().Id("main").Params().Block(
		Qual("fmt", "Println").Call(Lit("Hello, world")),
		Qual("fmt", "Println").Call(Lit("Aatish Here")),
		Qual("fmt", "Println").Call(Qual("", "add").Call(Id("2").Op(",").Id("3"))),
	)

	fmt.Fprintf(file, "%#v", f)

	fmt.Println("Entities")
	for _, value := range entities {
		fmt.Println("\t", value)
	}

	fmt.Println("Columns")
	for _, value := range allColumns {
		fmt.Println("\t", value)
	}

}
