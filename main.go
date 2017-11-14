package main

import (
	. "github.com/dave/jennifer/jen"
	"os"
	"log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"strings"
	"fmt"
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

	allModels := make([]string, 0)
	//creating entity structures
	for _, entity := range entities {
		allModels = append(allModels, createEntities(entity, db))
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
	createXShowRoom(xShowroom, allModels)

	//flush xShowroom.go
	fmt.Fprintf(file, "%#v", xShowroom)
	fmt.Println("xShowroom generated!!")
}

//xShowroom generation methods
func createXShowRoom(xShowroom *File, allModels []string) {

	createXShowRoomConfigStruct(xShowroom)

	createXShowRoomInitMethod(xShowroom)

	createXShowRoomMainMethod(xShowroom, allModels)
}

func createXShowRoomConfigStruct(xShowroom *File) {

	//add config struct
	xShowroom.Comment("Configuration contains the application settings")
	xShowroom.Type().Id("configuration").Struct(
		Id("Database ").Qual("shared/database", "Info"),
		Id("Server").Qual("shared/server", "Server"),
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

func createXShowRoomMainMethod(xShowroom *File, allModels []string) {

	//add main method in xShowroom.go
	xShowroom.Func().Id("main").Params().Block(

		Comment("Load the configuration file"),
		Qual("shared/jsonconfig", "Load").Call(
			Lit("config").
				Op("+").
				Id("string").
				Op("(").
				Qual("os", "PathSeparator").
				Op(")").
				Op("+").
				Lit("config.json"),
			Id("config")),

		Empty(),

		Comment("Connect to database"),
		Qual("shared/database", "Connect").Call(
			Id("config").Op(".").Id("Database"),
		),

		Empty(),

		Comment("Load the controller routes"),
		Qual("models", "Load").Call(),

		Empty(),

		Comment("Auto migrate all models"),
		Qual("shared/database", "SQL.AutoMigrate").CallFunc(func(g *Group) {
			for _, value := range allModels {
				g.Id("&" + "models." + value + "{}")
			}
		}),

		Empty(),

		Comment("Start the listener"),
		Qual("shared/server", "Run").Call(
			Qual("shared/route", "LoadHTTP").Call(),
			Qual("shared/route", "LoadHTTPS").Call(),
			Id("config").Op(".").Id("Server"),
		),
	)
}

//models generation methods
func createEntities(entity Entity, db *gorm.DB) string {

	// create entity name from table
	entityName := snakeCaseToCamelCase(entity.DisplayName)

	//create entity file in models sub directory
	file, err := os.Create("vendor/models/" + entityName + ".go")
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
				finalId := name + d + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\" json:\"" + relation.ChildEntity.DisplayName + ",omitempty\"`"
				g.Id(finalId)
			case 2: //one to many
				finalId := name + "s []" + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\" json:\"" + relation.ChildEntity.DisplayName + "s,omitempty\"`"
				g.Id(finalId)
			case 3: //many to many

			}
		}
	})

	//write table name method for our struct
	modelFile.Func().Params(Id(snakeCaseToCamelCase(entity.DisplayName))).Id("TableName").Params().String().Block(
		Return(Lit(entity.Name)),
	)

	getAllMethodName := "GetAll" + entityName + "s"
	getByIdMethodName := "Get" + entityName
	postMethodName := "Post" + entityName
	putMethodName := "Put" + entityName
	deleteMethodName := "Delete" + entityName

	modelFile.Empty()
	//write routes in init method
	modelFile.Comment("Routes related to " + entityName)
	modelFile.Func().Id("init").Params().Block(
		Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)), Id(getAllMethodName)),
		Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(getByIdMethodName)),
		Qual("shared/router", "Post").Call(Lit("/"+strings.ToLower(entityName)), Id(postMethodName)),
		Qual("shared/router", "Put").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(putMethodName)),
		Qual("shared/router", "Delete").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(deleteMethodName)),
	)

	createEntitiesGetAllMethod(modelFile, entityName, getAllMethodName)

	createEntitiesGetMethod(modelFile, entityName, getByIdMethodName)

	createEntitiesPostMethod(modelFile, entityName, postMethodName)

	createEntitiesPutMethod(modelFile, entityName, putMethodName)

	createEntitiesDeleteMethod(modelFile, entityName, deleteMethodName)

	fmt.Fprintf(file, "%#v", modelFile)

	return entityName
}

func createEntitiesGetAllMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write getAll method
	modelFile.Comment("This method will return a list of all " + entityName + "s")
	modelFile.Func().Id(methodName).Params(handlerRequestParams()).Block(
		Id("data").Op(":=").Op("[]").Id(entityName).Op("{}"),
		Qual("shared/database", "SQL.Find").Call(Id("&").Id("data")),
		setJsonHeader(),
		sendResponse(2000, "Data fetched successfully", Id("data")),
	)
}

func createEntitiesGetMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write getOne method
	modelFile.Comment("This method will return one " + entityName + " based on id")
	modelFile.Func().Id(methodName).Params(handlerRequestParams()).Block(
		Comment("Get the parameter id"),
		Id("params").Op(":=").Qual("shared/router", "Params").Call(Id("req")),
		Id("ID").Op(":=").Qual("", "params.ByName").Call(Lit("id")),
		Id("data").Op(":=").Id(entityName).Op("{}"),
		Qual("shared/database", "SQL.First").Call(Id("&").Id("data"), Id("ID")),
		setJsonHeader(),
		sendResponse(2000, "Data fetched successfully", Id("data")),
	)
}

func createEntitiesPostMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write insert method
	modelFile.Comment("This method will insert one " + entityName + " in db")
	modelFile.Func().Id(methodName).Params(handlerRequestParams()).Block(
		Id("decoder").Op(":=").Qual("encoding/json", "NewDecoder").Call(Id("req").Op(".").Id("Body")),
		Var().Id("data").Id(entityName),
		Id("err").Op(":=").Qual("", "decoder.Decode").Call(Id("&").Id("data")),
		If(Id("err").Op("!=").Nil()).Block(
			setJsonHeader(),
			sendResponse(9999, "Data not saved", "invalid data"),
			Return(),
		),
		Defer().Qual("", "req.Body.Close").Call(),
		Qual("shared/database", "SQL.Create").Call(Id("&").Id("data")),
		setJsonHeader(),
		sendResponse(2000, "Data saved successfully", "data saved"),
	)
}

func createEntitiesPutMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write update method
	modelFile.Comment("This method will update " + entityName + " based on id")
	modelFile.Func().Id(methodName).Params(handlerRequestParams()).Block(
		Comment("Get the parameter id"),
		Id("params").Op(":=").Qual("shared/router", "Params").Call(Id("req")),
		Id("ID").Op(",").Id("_").Op(":=").Qual("strconv", "ParseUint").Call(
			Qual("", "params.ByName").Call(Lit("id")),
			Id("10"),
			Id("0"),
		),
		Id("oldData").Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("uint(ID)").Op("}"),
		Empty(),

		Comment("Get new data from request"),
		Id("decoder").Op(":=").Qual("encoding/json", "NewDecoder").Call(Id("req").Op(".").Id("Body")),
		Var().Id("newData").Id(entityName),
		Id("err").Op(":=").Qual("", "decoder.Decode").Call(Id("&").Id("newData")),
		If(Id("err").Op("!=").Nil()).Block(
			setJsonHeader(),
			sendResponse(9999, "Data not saved", "invalid data"),
			Return(),
		),
		Defer().Qual("", "req.Body.Close").Call(),

		Empty(),
		Comment("Update record"),
		Qual("shared/database", "SQL.Model").Call(Id("&oldData")).Op(".").Id("Updates").Call(Id("newData")),
		setJsonHeader(),
		sendResponse(2000, "Data updated successfully", Id("nil")),

	)
}

func createEntitiesDeleteMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write delete method
	modelFile.Comment("This method will delete " + entityName + " based on id")
	modelFile.Func().Id(methodName).Params(handlerRequestParams()).Block(
		Comment("Get the parameter id"),
		Id("params").Op(":=").Qual("shared/router", "Params").Call(Id("req")),
		Id("ID").Op(",").Id("_").Op(":=").Qual("strconv", "ParseUint").Call(
			Qual("", "params.ByName").Call(Lit("id")),
			Id("10"),
			Id("0"),
		),
		Id("data").Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("uint(ID)").Op("}"),
		Qual("shared/database", "SQL.Delete").Call(Id("&").Id("data")),
		setJsonHeader(),
		sendResponse(2000, "Data deleted successfully", Id("nil")),
	)
}

func mapColumnTypes(col Column, g *Group) {
	if col.ColumnType.Type == "int" {
		finalId := snakeCaseToCamelCase(col.Name) + " uint" + " `gorm:\"column:" + col.Name + "\" json:\"" + col.Name + ",omitempty\"`"
		g.Id(finalId)
	} else if col.ColumnType.Type == "varchar" {
		finalId := snakeCaseToCamelCase(col.Name) + " string" + " `gorm:\"column:" + col.Name + "\" json:\"" + col.Name + ",omitempty\"`"
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

func handlerRequestParams() (Code, Code) {
	return Id("w").Qual("net/http", "ResponseWriter"), Id("req").Op("*").Qual("net/http", "Request")
}

func setJsonHeader() Code {
	return Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
}

func sendResponse(statusCode uint, statusMsg string, data interface{}) Code {
	return Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
		Op("{").
		Lit(statusCode).Op(",").
		Lit(statusMsg).Op(",").
		Lit(data).
		Op("}"))
}
