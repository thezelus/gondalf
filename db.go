package main

import (
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
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
	dbConnection.AutoMigrate(ArchivedToken{})
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
	tokenCutOffTime := AppProperties{PropertyName: "TokenCutOffTime", PropertyValue: "30", UpdatedAt: time.Now().UTC()}
	tokenCleanUpFrequency := AppProperties{PropertyName: "TokenCleanUpFrequency", PropertyValue: "180", UpdatedAt: time.Now().UTC()}

	propertiesSlice := []AppProperties{webTimeOut, mobileTimeOut, dbDebugLogs, timeExtension, tokenCutOffTime, tokenCleanUpFrequency}

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

	LoadConfigurationFromFile()

	var connParam ConnectionParameters

	connParam.username = viper.GetString("dbUsername")
	connParam.password = viper.GetString("dbPassword")
	connParam.host = viper.GetString("dbHost")
	connParam.port = viper.GetString("dbPort")
	connParam.dbname = viper.GetString("dbName")
	connParam.sslmode = viper.GetString("dbSSLmode")

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

	maxIdleConnections, _ := strconv.Atoi(viper.GetString("dbMaxIdleConnections"))
	maxOpenConnections, _ := strconv.Atoi(viper.GetString("dbMaxOpenConnections"))

	db.DB().SetMaxIdleConns(maxIdleConnections)
	db.DB().SetMaxOpenConns(maxOpenConnections)
	db.DB().Ping()

	return db
}
