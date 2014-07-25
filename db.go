package main

import (
	"time"

	"github.com/gosimple/conf"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

//Method to initialize by automigrating new tables and adding columns to old ones.
//This method doesn't change any existing column types or delete columns.
//Corresponding DB structures can be found in dbStructs.go
//App properties entries are also initialized here
func InitDB() {

	dbConnection.AutoMigrate(User{})
	dbConnection.AutoMigrate(Permission{})
	dbConnection.AutoMigrate(Group{})
	dbConnection.AutoMigrate(GroupPermission{})
	dbConnection.AutoMigrate(UserGroup{})

	dbConnection.AutoMigrate(AppProperties{})
	dbConnection.AutoMigrate(ActivityLog{})

	dbConnection.AutoMigrate(Token{})
	dbConnection.AutoMigrate(DeviceType{})
	dbConnection.AutoMigrate(PasswordRecord{})

	InsertAppProperties(&dbConnection)
	InsertDeviceTypes(&dbConnection)
}

func InsertAppProperties(db *gorm.DB) bool {

	webTimeOut := AppProperties{PropertyName: "WebTimeOut", PropertyValue: "30", UpdatedAt: time.Now().UTC()}
	mobileTimeOut := AppProperties{PropertyName: "MobileTimeOut", PropertyValue: "720", UpdatedAt: time.Now().UTC()}
	dbDebugLogs := AppProperties{PropertyName: "DbDebugLogs", PropertyValue: "false", UpdatedAt: time.Now().UTC()}
	timeExtension := AppProperties{PropertyName: "TimeExtension", PropertyValue: "5", UpdatedAt: time.Now().UTC()}

	propertiesSlice := []AppProperties{webTimeOut, mobileTimeOut, dbDebugLogs, timeExtension}

	for i := range propertiesSlice {
		existsErr := db.Where(&AppProperties{PropertyName: propertiesSlice[i].PropertyName}).Find(&AppProperties{}).Error
		if existsErr == gorm.RecordNotFound {
			db.Save(&propertiesSlice[i])
		}
	}

	return true
}

func InsertDeviceTypes(db *gorm.DB) bool {
	webDevice := DeviceType{Device: "web", DeviceCode: web}
	mobileDevice := DeviceType{Device: "mobile", DeviceCode: mobile}

	devicesSlice := []DeviceType{webDevice, mobileDevice}

	for i := range devicesSlice {
		existsErr := db.Where(&devicesSlice[i]).Find(&DeviceType{}).Error
		if existsErr == gorm.RecordNotFound {
			db.Save(&devicesSlice[i])
		}
	}

	return true
}

//Returns a database connection with connection pooling
func GetDBConnection() gorm.DB {

	TRACE.Println("Connection properties will be read from " + configFile)
	connParam := GetConnectionProperties(configFile)

	source := "user=" + connParam.username +
		" password=" + connParam.password +
		" dbname=" + connParam.dbname +
		" port=" + connParam.port +
		" host=" + connParam.host +
		" sslmode=" + connParam.sslmode

	db, err := gorm.Open("postgres", source)

	if err != nil {
		ERROR.Panicln("Error opening DB connection")
	}

	TRACE.Println("DB Connection opened")

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.DB().Ping()

	return db
}

//Method to get DB connection properties from the specified file
func GetConnectionProperties(filename string) ConnectionParameters {

	var conParam ConnectionParameters

	c, err := conf.ReadFile(filename)

	if err != nil {
		ERROR.Panicln(err)
	}

	conParam.username, err = c.String("default", "username")
	if err != nil {
		ERROR.Panicln(err)
	}

	conParam.password, err = c.String("default", "password")
	if err != nil {
		ERROR.Panicln(err)
	}

	conParam.host, err = c.String("default", "host")
	if err != nil {
		ERROR.Panicln(err)
	}

	conParam.port, err = c.String("default", "port")
	if err != nil {
		ERROR.Panicln(err)
	}

	conParam.dbname, err = c.String("default", "dbname")
	if err != nil {
		ERROR.Panicln(err)
	}

	conParam.sslmode, err = c.String("default", "sslmode")
	if err != nil {
		ERROR.Panicln(err)
	}

	return conParam
}
