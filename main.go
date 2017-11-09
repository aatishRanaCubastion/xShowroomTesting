package main

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
	"os"
	"log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"strings"
)

type Entity struct {
	ID          uint `sql:"AUTO_INCREMENT"`
	Name        string `sql:"type:varchar(30)"`
	DisplayName string `sql:"type:varchar(30)"`
	Columns     []Column `gorm:"ForeignKey:entity_id;AssociationForeignKey:id"` // one to many, has many columns
}

type ColumnType struct {
	ID      uint    `sql:"AUTO_INCREMENT"`
	Type    string `sql:"type:varchar(30)"`
	Columns []Column `gorm:"ForeignKey:type_id;AssociationForeignKey:id"` //one to many, has many columns
}

type Column struct {
	ID         uint `sql:"AUTO_INCREMENT"`
	Name       string `sql:"type:varchar(30)"`
	Size       uint `sql:"type:int(30)"`
	TypeID     uint `sql:"type:int(30)"`
	EntityID   uint `sql:"type:int(100)"`
	ColumnType ColumnType `gorm:"ForeignKey:TypeID"` //belong to (for reverse access)
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

	entities := []Entity{}

	db.Debug().Preload("Columns.ColumnType").Find(&entities)

	file, err := os.Create("xShowroom.go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//created file
	f := NewFile("main")

	for _, entity := range entities {
		f.Type().Id(snakeCaseToCamelCase(entity.Name)).StructFunc(func(g *Group) {
			for _, column := range entity.Columns {
				colTypeMapper(column, g)
			}
		})
	}

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
}

func colTypeMapper(col Column, g *Group) {
	if col.ColumnType.Type == "int" {
		g.Id(snakeCaseToCamelCase(col.Name)).Int()
	} else if col.ColumnType.Type == "varchar" {
		g.Id(snakeCaseToCamelCase(col.Name)).String()
	} else {
		g.Id(snakeCaseToCamelCase(col.Name)).String() //default string
	}
}

func snakeCaseToCamelCase(inputUnderScoreStr string) (camelCase string) {
	//snake_case to camelCase

	isToUpper := false

	for k, v := range inputUnderScoreStr {
		if k == 0 {
			camelCase = strings.ToUpper(string(inputUnderScoreStr[0]))
		} else {
			if isToUpper {
				camelCase += strings.ToUpper(string(v))
				isToUpper = false
			} else {
				if v == '_' {
					isToUpper = true
				} else {
					camelCase += string(v)
				}
			}
		}
	}
	return

}
