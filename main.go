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
	Name        string `sql:"type:varchar(30)"  gorm:"column:alias_name"`
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

type RelationType struct {
	ID   uint `sql:"AUTO_INCREMENT"`
	Name string `sql:"type:varchar(30)"`
}

type Relation struct {
	ParentEntityID    uint `sql:"type:int(100)"`
	ParentEntityColID uint `sql:"type:int(100)"`
	ChildEntityID     uint `sql:"type:int(100)"`
	ChildEntityColID  uint `sql:"type:int(100)"`
	InterEntityID     uint `sql:"type:int(100)"`
	RelationTypeID    uint `sql:"type:int(10)"`

	ParentEntity Entity `gorm:"ForeignKey:ParentEntityID"`       //belong to
	ChildEntity  Entity `gorm:"ForeignKey:ChildEntityID"`        //belong to
	InterEntity  Entity `gorm:"ForeignKey:InterEntityID"`        //belong to
	ParentColumn Column `gorm:"ForeignKey:ParentEntityColID"`    //belong to
	ChildColumn  Column `gorm:"ForeignKey:ChildEntityColID"`     //belong to
	RelationType RelationType `gorm:"ForeignKey:RelationTypeID"` //belong to
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

func (RelationType) TableName() string {
	return "c_relation_type"
}

func (Relation) TableName() string {
	return "c_relation"
}

func main() {

	//open data base
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/xshowroomcustom?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// migrate tables
	db.AutoMigrate(&Entity{}, &Column{}, &ColumnType{})

	//fetch all entities
	entities := []Entity{}
	db.Preload("Columns.ColumnType").
		Find(&entities)

	//print all entities
	//for _, entity := range entities {
	//	fmt.Print(entity.Name + " (" + entity.DisplayName + ")\n")
	//	for _, col := range entity.Columns {
	//		fmt.Print("\t", col.Name, " ", col.ColumnType.Type, "(", col.Size, ")\n")
	//	}
	//}

	//creating entity structures
	for _, entity := range entities {
		createEntities(entity, db)
	}

	//create xShowroom.go
	file, err := os.Create("xShowroom.go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//created file
	xShowroom := NewFile("main")

	//write all code
	createXShowRoom(xShowroom)

	//flush xShowroom.go
	fmt.Fprintf(file, "%#v", xShowroom)
	fmt.Println("xShowroom generated!!")
}

//xShowroom generation methods
func createXShowRoom(xShowroom *File) {

	createXShowRoomConfigStruct(xShowroom)

	createXShowRoomInitMethod(xShowroom)

	createXShowRoomMainMethod(xShowroom)
}

func createXShowRoomConfigStruct(xShowroom *File) {

	//add config struct
	xShowroom.Comment("Configuration contains the application settings")
	xShowroom.Type().Id("configuration").Struct(
		Id("Database ").Qual("database", "Info"),
	)

	//add parse method to configuration
	xShowroom.Comment("ParseJSON unmarshals bytes to structs")
	xShowroom.Func().Params(
		Id("c *").Id("configuration"), ).
		Id("ParseJSON").
		Params(Id("b []").Id("byte"), ).Error().Block(
		Return(Qual("encoding/json", "Unmarshal").Call(
			Id("b"),
			Id("&c"),
		)),
	)

	//create an instance of configuration
	xShowroom.Var().Id("config").Op("= &").Id("configuration{}")
}

func createXShowRoomInitMethod(xShowroom *File) {
	//add init method in xShowroom.go
	xShowroom.Func().Id("init").Params().Block(
		Comment(" Use all cpu cores"),
		Qual("runtime", "GOMAXPROCS").Call(Qual("runtime", "NumCPU").Call()),
	)
}

func createXShowRoomMainMethod(xShowroom *File) {
	//add main method in xShowroom.go
	xShowroom.Func().Id("main").Params().Block(

		Comment("Load the configuration file"),
		Qual("jsonconfig", "Load").Call(
			Lit("config").
				Op("+").
				Id("string").
				Op("(").
				Id("os").
				Op(".").
				Id("PathSeparator").
				Op(")").
				Op("+").
				Lit("config.json"),
			Id("config")),

		Empty(),

		Comment("Connect to database"),
		Qual("database", "Connect").Call(
			Id("config").Op(".").Id("Database"),
		),

		Empty(),

		Qual("fmt", "Println").Call(Lit("xShowroom is up and running!!")),
	)
}

//models generation methods
func createEntities(entity Entity, db *gorm.DB) {

	// create entity name from table
	entityName := snakeCaseToCamelCase(entity.DisplayName)

	//create entity file in models sub directory
	file, err := os.Create("models/" + entityName + ".go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	//set package as "models"
	modelFile := NewFile("models")

	//fetch relations of this entity
	relations := []Relation{}
	db.Preload("ChildEntity").
		Preload("ChildColumn").
		Preload("ParentColumn").
		Where("parent_entity_id=?", entity.ID).
		Find(&relations)

	//write structure for entity
	modelFile.Type().Id(entityName).StructFunc(func(g *Group) {

		//write primitive fields
		for _, column := range entity.Columns {
			mapColumnTypes(column, g)
		}

		//write composite fields
		for _, relation := range relations {

			//todo get relation types from db
			name := snakeCaseToCamelCase(relation.ChildEntity.DisplayName)
			childName := string(relation.ChildColumn.Name)
			parentName := string(relation.ParentColumn.Name)

			d := " "
			if entityName == name {
				d = " *" //if name and entityName are same, its a self join, so add *
			}

			switch relation.RelationTypeID {
			case 1: //one to one
				finalId := name + d + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\"`"
				g.Id(finalId)
			case 2: //one to many
				finalId := name + "s []" + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\"`"
				g.Id(finalId)
			case 3: //many to many

			}
		}
	})

	//write table name method for our struct
	modelFile.Func().Params(Id(snakeCaseToCamelCase(entity.DisplayName))).Id("TableName").Params().String().Block(
		Return(Lit(entity.Name)),
	)

	//write getAll method
	modelFile.Func().Id("Fetch" + entityName).Params(
		Id("w").Qual("net/http", "ResponseWriter"),
		Id("req").Op("*").Qual("net/http", "Request"),
	).Block()

	//flush file
	fmt.Fprintf(file, "%#v", modelFile)
}

func mapColumnTypes(col Column, g *Group) {
	if col.ColumnType.Type == "int" {
		finalId := snakeCaseToCamelCase(col.Name) + " uint" + " `gorm:\"column:" + col.Name + "\"`"
		g.Id(finalId)
	} else if col.ColumnType.Type == "varchar" {
		finalId := snakeCaseToCamelCase(col.Name) + " string" + " `gorm:\"column:" + col.Name + "\"`"
		g.Id(finalId)
	} else {
		g.Id(snakeCaseToCamelCase(col.Name)).String() //default string
	}
}

//helper methods
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
