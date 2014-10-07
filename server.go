package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/go-martini/martini"
	"github.com/jinzhu/gorm"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/spf13/viper"
	"github.com/stretchr/graceful"
)

var (
	dbConnection gorm.DB
	configName   = "gondalfConfig"
	logFile      = "gondalf.log"
	testLogFile  = "testLogs.log"
	properties   []AppProperties
	file         *os.File
)

func main() {

	initializeDB := flag.Bool("initdb", false, "initalizing database")
	flag.Parse()

	LoadConfigurationFromFile()

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

	//Validate session token
	m.Post("/validate/token", binding.Bind(ValidateUsernameRequest{}), ValidateSessionTokenHandler)

	//Change password
	m.Post("/user/changePassword", binding.Bind(ChangePasswordRequest{}), ChangePasswordHandler)

	//Check permission
	m.Post("/user/checkPermission", binding.Bind(CheckPermissionRequest{}), CheckPermissionsForUserHandler)

	appGracefulShutdownTimeinSeconds, err := strconv.Atoi(viper.GetString("appGracefulShutdownTimeinSeconds"))
	if err != nil {
		ERROR.Panicln("Cannot start the server, shutdown time missing from config file")
	}

	graceful.Run(":"+viper.GetString("appPort"), time.Duration(appGracefulShutdownTimeinSeconds)*time.Second, m)

	TRACE.Println("DB connection closed")
	defer cleanUpAfterShutdown()
}

func cleanUpAfterShutdown() {
	TRACE.Println("Cleaning up dbConnection and file stream")
	dbConnection.Close()
	file.Close()
}
