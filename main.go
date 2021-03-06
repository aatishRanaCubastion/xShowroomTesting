package main

import (
	. "github.com/dave/jennifer/jen"
	"os"
	"log"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"strings"
	"fmt"
	"strconv"
	"bytes"
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
	ID                uint `sql:"AUTO_INCREMENT"`
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

type EntityRelation struct {
	Type             string
	SubEntityName    string
	SubEntityColName string
}

type EntityRelationMethod struct {
	MethodName       string
	Type             string
	SubEntityName    string
	SubEntityColName string
}

func main() {

	//open data base
	db, err := gorm.Open("mysql", "root:@tcp(127.0.0.1:3306)/xshowroomcustom?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	// migrate tables
	db.AutoMigrate(&Entity{}, &Column{}, &ColumnType{}, &Relation{}, &RelationType{})

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
	fmt.Println("=========================")
	fmt.Println("xShowroom generated!!!")
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

	//entity relations stored to generate routes and their methods for each sub entities ((parent to child) and (child to parent))
	entityRelationsForEachEndpoint := []EntityRelation{}

	//entity relations stored to generate one route to access all sub entities depending on query params(parent to child only)
	entityRelationsForAllEndpoint := []EntityRelation{}

	//create entity file in models sub directory
	file, err := os.Create("vendor/models/" + entityName + ".go")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	//set package as "models"
	modelFile := NewFile("models")

	//fetch relations of this entity matching parent
	relationsParent := []Relation{}
	db.Preload("InterEntity").
		Preload("ChildEntity").
		Preload("ChildColumn").
		Preload("ParentColumn").
		Where("parent_entity_id=?", entity.ID).
		Find(&relationsParent)

	//fetch relations of this entity matching child
	relationsChild := []Relation{}
	db.Preload("InterEntity").
		Preload("ParentEntity").
		Preload("ChildColumn").
		Preload("ParentColumn").
		Where("child_entity_id=?", entity.ID).
		Find(&relationsChild)

	//write structure for entity
	modelFile.Type().Id(entityName).StructFunc(func(g *Group) {

		//write primitive fields
		for _, column := range entity.Columns {
			mapColumnTypes(column, g)
		}

		//write composite fields while looking at parent
		for _, relation := range relationsParent {
			//fmt.Println("parent ", relation)
			name := snakeCaseToCamelCase(relation.ChildEntity.DisplayName)
			childName := string(relation.ChildColumn.Name)
			parentName := string(relation.ParentColumn.Name)

			d := " "
			relType := "_normal"
			if entityName == name {
				d = "*" //if name and entityName are same, its a self join, so add *
				relType = "_self"
			}

			switch relation.RelationTypeID {
			case 1: //one to one
				relationName := name
				finalId := relationName + " " + d + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\" json:\"" + relation.ChildEntity.DisplayName + ",omitempty\"`"
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"OneToOne" + relType, name, childName})
				entityRelationsForAllEndpoint = append(entityRelationsForAllEndpoint, EntityRelation{"OneToOne" + relType, relationName, childName})
				g.Id(finalId)
			case 2: //one to many
				relationName := name + "s"
				finalId := relationName + " []" + name + " `gorm:\"ForeignKey:" + childName + ";AssociationForeignKey:" + parentName + "\" json:\"" + relation.ChildEntity.DisplayName + "s,omitempty\"`"
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"OneToMany", name, childName})
				entityRelationsForAllEndpoint = append(entityRelationsForAllEndpoint, EntityRelation{"OneToMany", relationName, childName})
				g.Id(finalId)
			case 3: //many to many
				relationName := name + "s"
				finalId := relationName + " []" + name + " `gorm:\"many2many:" + relation.InterEntity.Name + "\" json:\"" + relation.ChildEntity.DisplayName + "s,omitempty\"`"
				g.Id(finalId)
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"ManyToMany", name, childName})
				//handle for many to many(all child entities)
				//entityRelationsForAllEndpoint = append(entityRelationsForAllEndpoint, EntityRelation{"ManyToMany", relationName, childName})
			}
		}

		//write composite fields while looking at child
		for _, relation := range relationsChild {
			//fmt.Println("child ", relation)
			name := snakeCaseToCamelCase(relation.ParentEntity.DisplayName)
			childName := string(relation.ChildColumn.Name)
			//parentName := string(relation.ParentColumn.Name)

			switch relation.RelationTypeID {
			case 1: //ont to one
				// means current entity's one item belongs to
				if name != entityName { // if check to exclude self join
					entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"OneToOne_reverse", name, childName})
					//todo no need, two way association not allowed
					//finalId := name + " " + name + " `gorm:\"ForeignKey:" + snakeCaseToCamelCase(childName) + "\" json:\"" + name + ",omitempty\"`"
					//g.Id(finalId)
				}
			case 2: //one to many
				// means current entity's many items belongs to
				finalId := name + " " + name + " `gorm:\"ForeignKey:" + snakeCaseToCamelCase(childName) + "\" json:\"" + name + ",omitempty\"`"
				entityRelationsForEachEndpoint = append(entityRelationsForEachEndpoint, EntityRelation{"ManyToOne", name, childName})
				g.Id(finalId)
			case 3: //many to many
				// add two record in relation table to create many to many or uncomment this and add relation here
				//fmt.Println("\t\t many to many " + relation.InterEntity.DisplayName + " for " + entityName + " from child")
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

	allMethodName := "GetAll" + entityName + "sSubEntities"
	allMethodExist := false

	specialMethods := []EntityRelationMethod{}

	modelFile.Empty()
	//write routes in init method
	modelFile.Comment("Routes related to " + entityName)
	modelFile.Func().Id("init").Params().BlockFunc(func(g *Group) {

		g.Empty()
		g.Comment("Standard routes")
		g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)), Id(getAllMethodName))
		g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(getByIdMethodName))
		g.Qual("shared/router", "Post").Call(Lit("/"+strings.ToLower(entityName)), Id(postMethodName))
		g.Qual("shared/router", "Put").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(putMethodName))
		g.Qual("shared/router", "Delete").Call(Lit("/"+strings.ToLower(entityName)+"/:id"), Id(deleteMethodName))

		if len(entityRelationsForEachEndpoint) > 0 {
			g.Empty()
			g.Comment("Sub entities routes")
			for _, entRel := range entityRelationsForEachEndpoint {

				if entRel.Type == "OneToMany" {
					methodName := "Get" + entityName + entRel.SubEntityName + "s"
					specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
					g.Empty()
					g.Comment("has many")
					g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName+"s")), Id(methodName))
				} else if entRel.Type == "OneToOne_normal" || entRel.Type == "OneToOne_self" || entRel.Type == "OneToOne_reverse" {
					methodName := "Get" + entityName + entRel.SubEntityName
					specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
					g.Empty()
					g.Comment("has one")
					g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName)), Id(methodName))
				} else if entRel.Type == "ManyToOne" {
					methodName := "Get" + entityName + entRel.SubEntityName + ""
					specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
					g.Empty()
					g.Comment("belongs to")
					g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName)), Id(methodName))
				} else if entRel.Type == "ManyToMany" {
					methodName := "Get" + entityName + entRel.SubEntityName + "s"
					specialMethods = append(specialMethods, EntityRelationMethod{methodName, entRel.Type, entRel.SubEntityName, entRel.SubEntityColName})
					g.Empty()
					g.Comment("has many to many")
					g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/"+strings.ToLower(entRel.SubEntityName)), Id(methodName))
				}

			}
		}

		if len(entityRelationsForAllEndpoint) > 0 {
			allMethodExist = true
			g.Empty()
			g.Comment("Super sonic route")
			g.Qual("shared/router", "Get").Call(Lit("/"+strings.ToLower(entityName)+"/:id/all"), Id(allMethodName))
		}
	})

	createEntitiesChildSlice(modelFile, entityName, entityRelationsForAllEndpoint)

	createEntitiesGetAllMethod(modelFile, entityName, getAllMethodName)

	createEntitiesGetMethod(modelFile, entityName, getByIdMethodName)

	createEntitiesPostMethod(modelFile, entityName, postMethodName)

	createEntitiesPutMethod(modelFile, entityName, putMethodName)

	createEntitiesDeleteMethod(modelFile, entityName, deleteMethodName)

	if len(specialMethods) > 0 {
		for _, method := range specialMethods {
			modelFile.Empty()
			modelFile.Func().Id(method.MethodName).Params(handlerRequestParams()).BlockFunc(func(g *Group) {
				g.Empty()
				g.Comment("Get the parameter id")
				g.Id("params").Op(":=").Qual("shared/router", "Params").Call(Id("req"))
				g.Id("ID").Op(",").Id("_").Op(":=").Qual("strconv", "ParseUint").Call(
					Qual("", "params.ByName").Call(Lit("id")),
					Id("10"),
					Id("0"),
				)

				if method.Type == "OneToMany" || method.Type == "OneToOne_normal" {
					g.Id("data").Op(":= []").Id(method.SubEntityName).Id("{}")
					g.Qual("shared/database", "SQL.Find").Call(Id("&").Id("data"), Lit(" "+method.SubEntityColName+" = ?"), Id("ID"))
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}

				if method.Type == "ManyToOne" || method.Type == "OneToOne_reverse" {
					g.Id(strings.ToLower(entityName)).Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("uint(").Id("ID").Op(")}")

					g.Id("data").Op(":= ").Id(method.SubEntityName).Id("{}")
					g.Qual("shared/database", "SQL.Find").Call(
						Id("&").Id("data"), Lit(" id = (?)"),
						Qual("shared/database", "SQL.Select").Call(Lit(method.SubEntityColName)).Op(".").Id("First").Call(Id("&").Id(strings.ToLower(entityName))).Op(".").Id("QueryExpr").Call(),
					)
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}

				if method.Type == "OneToOne_self" {
					g.Id("data").Op(":= ").Id(method.SubEntityName).Id("{}")
					g.Qual("shared/database", "SQL.Find").Call(Id("&").Id("data"), Lit(" "+method.SubEntityColName+" = ?"), Id("ID"))
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}

				if method.Type == "ManyToMany" {

					relation := method.SubEntityName + "s"

					g.Id("data").Op(":=").Id(entityName).Id("{}")
					g.Qual("shared/database", "SQL.Find").Call(Id("&").Id("data"), Id("ID"))
					g.Qual("shared/database", "SQL.Model").Call(Id("&").Id("data")).Op(".").Id("Association").Call(Lit(relation)).
						Op(".").Id("Find").Call(Id("&").Id("data").Op(".").Id(relation))
					g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
					g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
						Op("{").
						Id("2000").Op(",").
						Lit("Data fetched successfully").Op(",").
						Id("data").
						Op("}"))
				}
			})
		}
	}

	if allMethodExist {
		createEntitiesAllChildMethod(modelFile, entityName, allMethodName, entityRelationsForAllEndpoint)
	}

	fmt.Fprintf(file, "%#v", modelFile)

	fmt.Println(entityName + " generated")
	return entityName
}

func createEntitiesChildSlice(modelFile *File, entityName string, entityRelationsForAllEndpoint []EntityRelation) {
	allChildren := []string{}
	for _, value := range entityRelationsForAllEndpoint {
		allChildren = append(allChildren, value.SubEntityName)
	}

	modelFile.Empty()
	modelFile.Comment("Child entities")
	modelFile.Var().Id(entityName + "Children").Op("=").Lit(allChildren)
}

func createEntitiesGetAllMethod(modelFile *File, entityName string, methodName string) {
	modelFile.Empty()
	//write getAll method
	modelFile.Comment("This method will return a list of all " + entityName + "s")
	modelFile.Func().Id(methodName).Params(handlerRequestParams()).Block(
		modelFile.Empty(),
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
		modelFile.Empty(),
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
		modelFile.Empty(),
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
		modelFile.Empty(),
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
		modelFile.Empty(),
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

func createEntitiesAllChildMethod(modelFile *File, entityName string, allMethodName string, entityRelationsForAllEndpoint []EntityRelation) {
	modelFile.Empty()
	modelFile.Func().Id(allMethodName).Params(handlerRequestParams()).BlockFunc(func(g *Group) {
		g.Empty()
		g.Comment("Get the parameter id")
		g.Id("params").Op(":=").Qual("shared/router", "Params").Call(Id("req"))
		g.Id("ID").Op(",").Id("_").Op(":=").Qual("strconv", "ParseUint").Call(
			Qual("", "params.ByName").Call(Lit("id")),
			Id("10"),
			Id("0"),
		)
		g.Id("data").Op(":=").Id(entityName).Op("{").Id("Id").Op(":").Id("uint(ID)").Op("}")
		g.Empty()
		g.Var().Id("relations ").Op("[").Id(strconv.Itoa(len(entityRelationsForAllEndpoint))).Op("]").Id("string")
		g.Id("children").Op(":=").Qual("", "req.URL.Query().Get").Call(Lit("child"))
		g.If(Id("children").Op("!= \"\"")).
			Block(
			Var().Id("neededChildren ").Op("[]").Id("string"),

			For(Id("_,child").Op(":=").Id("range").Id(entityName + "Children")).
				Block(
				If(Qual("", "isValueInList").
					Call(
					Id("child"),
					Qual("strings", "Split").
						Call(
						Id("children"), Id("sep"),
					),
				).
					Block(
					Id("neededChildren").Op("=").Qual("", "append").Call(Id("neededChildren"), Id("child")),
				),
				), ),

			Empty(),

			For(Id("i").Op(":=").Id("range").Id("neededChildren")).
				Block(
				Id("relations").Op("[").Id("i").Op("]").Op("=").Id("neededChildren").Op("[").Id("i").Op("]"),
			),
		).Else().
			Block(
			For(Id("i").Op(":=").Id("range").Id(entityName + "Children")).
				Block(
				Id("relations").Op("[").Id("i").Op("]").Op("=").Id(entityName + "Children").Op("[").Id("i").Op("]"),
			),
		)
		g.If(Qual("", "len").Call(Id("relations")).Op(">0")).BlockFunc(func(g *Group) {

			var buffer bytes.Buffer
			buffer.WriteString("SQL.")
			for i := range entityRelationsForAllEndpoint {
				buffer.WriteString("Preload(relations[" + strconv.Itoa(i) + "]).")
			}
			buffer.WriteString("First")
			g.Qual("shared/database", buffer.String()).Call(Op("&").Id("data"))
		})
		g.Qual("", "w.Header().Set").Call(Lit("Content-Type"), Lit("application/json"))
		g.Qual("encoding/json", "NewEncoder").Call(Id("w")).Op(".").Id("Encode").Call(Id("Response").
			Op("{").
			Id("2000").Op(",").
			Lit("Data fetched successfully").Op(",").
			Id("data").
			Op("}"))
	})
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
