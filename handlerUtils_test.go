package main

import (
	"strconv"
	"testing"

	"code.google.com/p/go.crypto/bcrypt"
	. "gopkg.in/check.v1"
)

func TestHandlerUtils(t *testing.T) {
	TestingT(t)
}

type HandlerUtilsTestSuite struct{}

var _ = Suite(&HandlerUtilsTestSuite{})

func (suite *HandlerUtilsTestSuite) SetUpSuite(c *C) {
	initDbFlag := false
	InitApp(testLogFile, &initDbFlag)
	TRACE.Println("HandlerUtilsTestSuite Setup")
}

func (suite *HandlerUtilsTestSuite) TearDownSuite(c *C) {
	TRACE.Println("HandlerUtilsTestSuite TearDown")
	defer cleanUpAfterShutdown()
}

func (suite *HandlerUtilsTestSuite) TestGenerateSessionToken(c *C) {
	TRACE.Println("Running test: TestGenerateSessionToken")

	expectedSessionToken := "8N7PlLvnGgnE2gFU7+AkSxmAc02cXFkOLlFD5gTuOjo="

	actualSessionToken := generateSessionToken("testMessage", "testKey")

	c.Assert(expectedSessionToken, Equals, actualSessionToken)
}

func (suite *HandlerUtilsTestSuite) TestCreateNewUserWithUniqueUsername(c *C) {
	TRACE.Println("Running test: TestCreateNewUserWithUniqueUsername")

	tx := dbConnection.Begin()

	testString := "UniqueTestUser123321"

	var testCreateUserRequest CreateUserRequest
	testCreateUserRequest.Username = testString
	testCreateUserRequest.LegalName = testString
	testCreateUserRequest.Password = testString

	status, err := CreateNewUser(testCreateUserRequest, tx)

	c.Check(err, IsNil)
	c.Assert(status, Equals, 200)

	var user User

	dbErr := tx.Where(&User{UserName: testString}).First(&user).Error
	c.Assert(dbErr, IsNil)
	comparePasswordErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(testString))
	c.Assert(comparePasswordErr, IsNil)

	var passwordRecord PasswordRecord
	tx.Where(&PasswordRecord{UserId: user.Id}).Find(&passwordRecord)
	c.Assert(passwordRecord.LoginCount, Equals, 0)

	tx.Rollback()
}

func (suite *HandlerUtilsTestSuite) TestCreateNewUserWithDuplicateUsername(c *C) {
	TRACE.Println("Running test: TestCreateNewUserWithDuplicateUsername")

	tx := dbConnection.Begin()

	testString := "UniqueTestUser123321"

	var testCreateUserRequest CreateUserRequest
	testCreateUserRequest.Username = testString
	testCreateUserRequest.LegalName = testString
	testCreateUserRequest.Password = testString

	status, err := CreateNewUser(testCreateUserRequest, tx)

	c.Check(err, IsNil)
	c.Assert(status, Equals, 200)

	dbErr := tx.Where(&User{UserName: testString}).First(&User{}).Error

	c.Assert(dbErr, IsNil)

	statusDuplicateEntry, err := CreateNewUser(testCreateUserRequest, tx)

	c.Check(statusDuplicateEntry, Equals, 409)
	c.Assert(err, NotNil)

	tx.Rollback()
}

func (suite *HandlerUtilsTestSuite) TestCreateNewTokenDbEntry(c *C) {
	TRACE.Println("Running test: TestCreateNewTokenDbEntry")

	tx := dbConnection.Begin()

	testString := "UniqueTestUser123321"

	var testCreateUserRequest = CreateUserRequest{Username: testString,
		LegalName: testString, Password: testString}

	status, err := CreateNewUser(testCreateUserRequest, tx)

	c.Check(err, IsNil)
	c.Assert(status, Equals, 200)

	var user User
	dbErr := tx.Where(&User{UserName: testString}).Find(&user).Error
	c.Assert(dbErr, IsNil)

	var testLoginCredential = LoginCredential{Username: testString, Password: testString,
		DeviceId: web}

	sessionToken, err := CreateNewTokenDbEntry(testLoginCredential, tx)

	var token Token
	dbErr = tx.Where(&Token{Token: sessionToken}).Find(&token).Error
	c.Assert(dbErr, IsNil)

	var activityLog ActivityLog
	dbErr = tx.Where(&ActivityLog{UserId: user.Id, TokenId: token.Id}).Find(&activityLog).Error
	c.Assert(dbErr, IsNil)

	tx.Rollback()
}

func (suite *HandlerUtilsTestSuite) TestGetTimeOutValue(c *C) {

	var err error

	_, err = GetTimeOutValue(web)
	c.Assert(err, IsNil)

	_, err = GetTimeOutValue(mobile)
	c.Assert(err, IsNil)

	_, err = GetTimeOutValue(-100)
	c.Assert(err.Error(), Equals, "Invalid device")
}

func (suite *HandlerUtilsTestSuite) TestAuthenticateUser(c *C) {
	TRACE.Println("Running test: TestAuthenticateUser")
	tx := dbConnection.Begin()

	testString := "UniqueTestUser123321"

	testCreateUserRequest := CreateUserRequest{Username: testString, LegalName: testString, Password: testString}
	status, err := CreateNewUser(testCreateUserRequest, tx)

	c.Assert(status, Equals, 200)
	c.Assert(err, IsNil)

	validLoginCredential := LoginCredential{Username: testString, Password: testString, DeviceId: mobile}
	invalidLoginCredentialWrongPassword := LoginCredential{Username: testString, Password: "invalid", DeviceId: mobile}
	invalidLoginCredentialWrongUsername := LoginCredential{Username: "invalid", Password: testString, DeviceId: mobile}

	c.Assert(AuthenticateUser(validLoginCredential.Username, validLoginCredential.Password, tx), Equals, FirstLoginPasswordChange)
	c.Assert(AuthenticateUser(invalidLoginCredentialWrongPassword.Username, invalidLoginCredentialWrongPassword.Password, tx), Equals, IncorrectPassword)
	c.Assert(AuthenticateUser(invalidLoginCredentialWrongUsername.Username, invalidLoginCredentialWrongUsername.Password, tx), Equals, UnregisteredUser)

	newPassword := "newPassword"
	changePasswordStatus, chagePasswordErr := ChangePassword(validLoginCredential.Username, newPassword, tx)
	c.Assert(changePasswordStatus, Equals, 200)
	c.Assert(chagePasswordErr, IsNil)

	newValidLoginCredential := LoginCredential{Username: testString, Password: newPassword, DeviceId: mobile}
	c.Assert(AuthenticateUser(newValidLoginCredential.Username, newValidLoginCredential.Password, tx), Equals, nil)

	tx.Rollback()
}

func (suite *HandlerUtilsTestSuite) TestGetTimeExtension(c *C) {
	TRACE.Println("Running test: TestGetTimeExtension")
	var timeExtensionFromDb AppProperties

	dbConnection.Where(&AppProperties{PropertyName: "TimeExtension"}).Find(&timeExtensionFromDb)

	timeExtensionFromMethod := GetTimeExtension()

	extension, conversionErr := strconv.Atoi(timeExtensionFromDb.PropertyValue)

	c.Check(conversionErr, IsNil)
	c.Assert(extension, Equals, timeExtensionFromMethod)
}

func (suite *HandlerUtilsTestSuite) TestChangePassword(c *C) {
	TRACE.Println("Running test: TestChangePassword")
	tx := dbConnection.Begin()

	testString := "UniqueTestUser123321"

	testCreateUserRequest := CreateUserRequest{Username: testString, LegalName: testString, Password: testString}
	status, err := CreateNewUser(testCreateUserRequest, tx)

	c.Assert(err, IsNil)
	c.Assert(status, Equals, 200)

	var testUser User

	tx.Where(&User{UserName: testString}).Find(&testUser)

	compareErr := bcrypt.CompareHashAndPassword([]byte(testUser.Password), []byte(testString))
	c.Assert(compareErr, IsNil)

	newPassword := "newPassword"
	changePasswordStatus, changePasswordErr := ChangePassword(testString, newPassword, tx)
	c.Assert(changePasswordErr, IsNil)
	c.Assert(changePasswordStatus, Equals, 200)

	tx.Where(&User{UserName: testString}).Find(&testUser)

	compareErr = bcrypt.CompareHashAndPassword([]byte(testUser.Password), []byte(newPassword))
	c.Assert(compareErr, IsNil)

	changePasswordStatusUnregisteredUser, changePasswordErrUnregisteredUser := ChangePassword("unregisteredUser123321", newPassword, tx)
	c.Assert(changePasswordErrUnregisteredUser, Equals, UnregisteredUser)
	c.Assert(changePasswordStatusUnregisteredUser, Equals, 401)

	tx.Rollback()
}

func (suite *HandlerUtilsTestSuite) TestValidateSessionToken(c *C) {
	TRACE.Println("Running test: TestValidateSessionToken")

	tx := dbConnection.Begin()

	testString := "UniqueTestUser123321"

	var testCreateUserRequest = CreateUserRequest{Username: testString,
		LegalName: testString, Password: testString}

	status, err := CreateNewUser(testCreateUserRequest, tx)

	c.Check(err, IsNil)
	c.Assert(status, Equals, 200)

	var user User
	dbErr := tx.Where(&User{UserName: testString}).Find(&user).Error
	c.Assert(dbErr, IsNil)

	newPassword := "newPassword"
	changePasswordStatus, chagePasswordErr := ChangePassword(user.UserName, newPassword, tx)
	c.Assert(changePasswordStatus, Equals, 200)
	c.Assert(chagePasswordErr, IsNil)

	var testLoginCredential = LoginCredential{Username: testString, Password: newPassword,
		DeviceId: web}

	sessionToken, err := CreateNewTokenDbEntry(testLoginCredential, tx)
	c.Assert(err, IsNil)

	var token Token
	dbErr = tx.Where(&Token{Token: sessionToken}).Find(&token).Error
	c.Assert(dbErr, IsNil)

	var activityLog ActivityLog
	dbErr = tx.Where(&ActivityLog{UserId: user.Id, TokenId: token.Id}).Find(&activityLog).Error
	c.Assert(dbErr, IsNil)

	err, userId := ValidateSessionToken("testSessionToken", tx)
	c.Assert(err, Equals, InvalidSessionToken)
	c.Assert(userId, Equals, int64(-1))

	err, userId = ValidateSessionToken(sessionToken, tx)
	c.Assert(err, IsNil)
	c.Assert(userId, Equals, token.UserId)

	token.Active = false
	tx.Save(&token)

	err, userId = ValidateSessionToken(sessionToken, tx)
	c.Assert(err, Equals, ExpiredSessionToken)
	c.Assert(userId, Equals, int64(-1))

	tx.Rollback()

}
