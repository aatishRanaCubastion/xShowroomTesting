package database

import (
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jinzhu/gorm"
)

var (
	// SQL wrapper
	SQL *gorm.DB
	// Database info
	databases Info
)

// Type is the type of database from a Type* constant
type Type string

const (
	// TypeMySQL is MySQL
	TypeMySQL Type = "MySQL"
)

// Info contains the database configurations
type Info struct {
	// Database type
	Type Type
	// MySQL info if used
	MySQL MySQLInfo
}

// MySQLInfo is the details for the database connection
type MySQLInfo struct {
	Username  string
	Password  string
	Name      string
	Hostname  string
	Port      int
	Parameter string
}

// Connect to the database
func Connect(d Info) {
	var err error

	// Store the config
	databases = d

	switch d.Type {
	case TypeMySQL:
		// Connect to MySQL
		SQL, err = gorm.Open("mysql", DSN(d.MySQL))
		if err != nil {
			log.Println("SQL Driver Error", err)
		}

		// Check if is alive
		//if err = SQL.(); err != nil {
		//	log.Println("Database Error", err)
		//}
	default:
		log.Println("No registered database in config")
	}
}

// ReadConfig returns the database information
func ReadConfig() Info {
	return databases
}

// DSN returns the Data Source Name
func DSN(ci MySQLInfo) string {
	// Example: root:@tcp(localhost:3306)/test
	return ci.Username +
		":" +
		ci.Password +
		"@tcp(" +
		ci.Hostname +
		":" +
		fmt.Sprintf("%d", ci.Port) +
		")/" +
		ci.Name + ci.Parameter
}
