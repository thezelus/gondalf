package main

import (
	"testing"
	. "gopkg.in/check.v1"
)

func TestDbMethods(t *testing.T) {
	TestingT(t)
}

type DbMethodsTestSuite struct{}

var _ = Suite(&DbMethodsTestSuite{})

func (suite *DbMethodsTestSuite) SetUpSuite(c *C) {
	initDbFlag := false
	InitApp(testLogFile, &initDbFlag)
	TRACE.Println("DbMethodsTestSuite Setup")
}

func (suite *DbMethodsTestSuite) TearDownSuite(c *C) {
	TRACE.Println("DbMethodsTestSuite TearDown")
	defer cleanUpAfterShutdown()
}

func (suite *DbMethodsTestSuite) TestInsertAppPropertiesMultipleTimes(c *C) {
	TRACE.Println("Running test: TestInsertAppPropertiesMultipleTimes")
	tx := dbConnection.Begin()

	c.Check(InsertAppProperties(tx), Equals, true)

	c.Assert(InsertAppProperties(tx), Equals, true)

	tx.Rollback()
}

func (suite *DbMethodsTestSuite) TestInsertDeviceTypesMultipleTimes(c *C) {
	TRACE.Println("Running test: TestInsertDeviceTypesMultipleTimes")
	tx := dbConnection.Begin()

	c.Check(InsertDeviceTypes(tx), Equals, true)

	c.Assert(InsertDeviceTypes(tx), Equals, true)

	tx.Rollback()
}
