package main

import (
	"testing"
	. "gopkg.in/check.v1"
)

func TestUtilMethods(t *testing.T) {
	TestingT(t)
}

type UtilTestSuite struct{}

var _ = Suite(&UtilTestSuite{})

func (suite *UtilTestSuite) SetUpSuite(c *C) {
	initDbFlag := false
	InitApp(testLogFile, &initDbFlag)
	TRACE.Println("UtilTestSuite Setup")
}

func (suite *UtilTestSuite) TearDownSuite(c *C) {
	TRACE.Println("UtilTestSuite TearDown")
	defer cleanUpAfterShutdown()
}

func (suite *UtilTestSuite) TestGetAppProperties(c *C) {
	TRACE.Println("Running test: TestGetAppProperties")

	webtimeOut, err := GetAppProperties("WebTimeOut")
	c.Assert(webtimeOut, Equals, "30")
	c.Assert(err, IsNil)

	invalidProperty, err := GetAppProperties("InvalidProperty")
	c.Assert(invalidProperty, Equals, "")
	c.Assert(err.Error(), Equals, "App property InvalidProperty not set")
}
