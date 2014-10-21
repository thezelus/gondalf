package main

import (
	"github.com/spf13/viper"
	"strconv"
	"time"
)

func StartScheduledJobs() {
	JobRefreshAppProperties()
	JobArchiveExpiredSessionToken()
}

func JobRefreshAppProperties() {

	LoadConfigurationFromFile()
	appPropertiesRefreshTimeinMinutes, err := strconv.Atoi(viper.GetString("appPropertiesRefreshTimeinMinutes"))
	if err != nil {
		ERROR.Panicln("appPropertiesRefreshTimeinMinutes missing from the config file")
	}

	ticker := time.NewTicker(time.Duration(appPropertiesRefreshTimeinMinutes) * time.Minute)
	TRACE.Println("Starting goroutine for refreshing app properties")
	go func() {
		for {
			select {
			case <-ticker.C:
				LoadAppPropertiesFromDb()
			case <-quit:
				TRACE.Println("Quit signal: JobRefreshAppProperties")
				ticker.Stop()
			}
		}
	}()
}

func JobArchiveExpiredSessionToken() {

	tokenCleanUpFrequencyString, err := GetAppProperties("TokenCleanUpFrequency")
	if err != nil {
		ERROR.Println(err.Error())
	}

	tokenCleanUpFrequencyInteger, err := strconv.Atoi(tokenCleanUpFrequencyString)
	if err != nil {
		ERROR.Println("String to integer conversion failed for TokenCleanUpFrequency reverting to default value of 180 minutes")
		tokenCleanUpFrequencyInteger = 180
	}

	ticker := time.NewTicker(time.Duration(tokenCleanUpFrequencyInteger) * time.Minute)
	TRACE.Println("Starting goroutine for token archiving")
	go func() {
		for {
			select {
			case <-ticker.C:
				ArchiveTokenAfterCutOffTime(&dbConnection)
			case <-quit:
				TRACE.Println("Quit signal: JobArchiveExpiredSessionToken")
			}
		}
	}()

}
