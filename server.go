package main

import (
	"flag"
	"os"
	"time"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/stretchr/graceful"
)

var (
	dbConnection gorm.DB
	configFile   = "gondalf.config"
	logFile      = "gondalf.log"
	testLogFile  = "testLogs.log"
	properties   []AppProperties
	file         *os.File
)

func main() {

	initializeDB := flag.Bool("initdb", false, "initalizing database")
	flag.Parse()

	InitApp(logFile, initializeDB)

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Use(martini.Recovery())

	//status
	m.Get("/status", StatusHandler)

	//Login end point
	m.Post("/auth/login", binding.Bind(LoginCredential{}), LoginHandler)

	//Create new user
	m.Post("/user/create", binding.Bind(CreateUserRequest{}), CreateUserHandler)

	//Validate unique username
	m.Post("/validate/username", binding.Bind(ValidateUsernameRequest{}), ValidateUsernameHandler)

	//Change password
	m.Post("/user/changePassword", binding.Bind(ChangePasswordRequest{}), ChangePasswordHandler)

	graceful.Run(":3000", 10*time.Second, m)

	TRACE.Println("DB connection closed")
	defer cleanUpAfterShutdown()
}

func cleanUpAfterShutdown() {
	TRACE.Println("Cleaning up dbConnection and file stream")
	dbConnection.Close()
	file.Close()
}
