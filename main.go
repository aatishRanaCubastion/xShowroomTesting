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

	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/xshowroomcustom?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	db.AutoMigrate(&Entity{}, &Column{}, &ColumnType{})

	entities := []Entity{}
	db.Preload("Columns.ColumnType").
		Find(&entities)

	//print all tables
	for _, entity := range entities {
		fmt.Print(entity.Name + " (" + entity.DisplayName + ")\n")
		for _, col := range entity.Columns {
			fmt.Print("\t", col.Name, " ", col.ColumnType.Type, "(", col.Size, ")\n")
		}
	}

	file, err := os.Create("xShowroom.go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//created file
	f := NewFile("main")

	//creating structure
	for _, entity := range entities {
		createModel(entity, db)
	}

	//calling method
	f.Func().Id("main").Params().Block(
		Qual("fmt", "Println").Call(Lit("Hello, world")),
	)

	fmt.Fprintf(file, "%#v", f)
}

func createModel(entity Entity, db *gorm.DB) {
	entityName := entity.DisplayName
	file, err := os.Create("models/" + entityName + ".go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()
	//created file
	f := NewFile("models")

	relations := []Relation{}

	db.Preload("ChildEntity").
		Preload("ChildColumn").
		Preload("ParentColumn").
		Where("parent_entity_id=?", entity.ID).
		Find(&relations)
	f.Type().Id(snakeCaseToCamelCase(entity.DisplayName)).StructFunc(func(g *Group) {
		for _, column := range entity.Columns {
			colTypeMapper(column, g)
		}
		for _, relation := range relations {
			//todo get relation types from db
			name := snakeCaseToCamelCase(relation.ChildEntity.DisplayName)
			childName := string(relation.ChildColumn.Name)
			parentName := string(relation.ParentColumn.Name)
			switch relation.RelationTypeID {
			case 1: //one to one
				g.Id(name + " " + name)
			case 2: //one to many
				finalId := name + "s []" + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\"`"
				g.Id(finalId)
			case 3: //many to many

			}
		}
	})

	//add table name
	f.Func().Params(Id(snakeCaseToCamelCase(entity.DisplayName))).Id("TableName").Params().String().Block(
		Return(Lit(entity.Name)),
	)

	fmt.Fprintf(file, "%#v", f)
}

func colTypeMapper(col Column, g *Group) {
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
