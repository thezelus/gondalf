package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/go.crypto/bcrypt"
	"github.com/jinzhu/gorm"
)

//Authenticates the user by checking username/password against the DB values
func AuthenticateUser(username string, password string, db *gorm.DB) error {
	var user User

	dbErr := db.Where(&User{UserName: username}).Find(&user).Error
	if dbErr != nil {
		if dbErr == gorm.RecordNotFound {
			WARNING.Println(UnregisteredUser.Error() + " ,username: " + username)
			return UnregisteredUser
		} else {
			ERROR.Println(dbErr.Error())
			return DatabaseError
		}
	}

	TRACE.Println("Login attempt by userId " + strconv.FormatInt(user.Id, 10))
	passwordCompareErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if passwordCompareErr != nil {
		WARNING.Println(passwordCompareErr.Error() + ", userId: " + strconv.FormatInt(user.Id, 10))
		return IncorrectPassword
	}

	var passwordRecord PasswordRecord

	dbPasswordRecordErr := db.Where(&PasswordRecord{UserId: user.Id}).Find(&passwordRecord).Error
	if dbPasswordRecordErr != nil {
		ERROR.Println(dbPasswordRecordErr.Error() + ", userId: " + strconv.FormatInt(user.Id, 10))
		return DatabaseError
	}

	if dbPasswordRecordErr == nil && passwordRecord.LoginCount == 0 {
		TRACE.Println(FirstLoginPasswordChange.Error() + ", userId: " + strconv.FormatInt(user.Id, 10))
		return FirstLoginPasswordChange
	}

	TRACE.Println("User authenticated, userId: " + strconv.FormatInt(user.Id, 10))
	return nil
}

//Create new token entry in the database
func CreateNewTokenDbEntry(login LoginCredential, db *gorm.DB) (string, error) {
	user := User{}
	db.Where(&User{UserName: login.Username}).Find(&user)

	token := Token{}
	token.UserId = user.Id
	token.Key = uuid.New()
	token.CreatedAt = time.Now().UTC()
	token.LastAccessedAt = time.Now().UTC()
	token.Token = generateSessionToken(login.Username+token.CreatedAt.String(), token.Key)
	token.Active = true
	expiryTime, err := GetTimeOutValue(login.DeviceId)
	if err != nil {
		ERROR.Println(err.Error())
		return "", err
	}

	token.ExpiresAt = expiryTime

	var device DeviceType
	db.Where(&DeviceType{DeviceCode: login.DeviceId}).Find(&device)

	token.DeviceTypeId = device.Id

	db.Save(&token)

	db.Save(&ActivityLog{UserId: token.UserId, TokenId: token.Id, ActivityTime: time.Now().UTC(), Event: LOGIN})

	return token.Token, nil
}

//Reads app property to get time out value based on the device ID
func GetTimeOutValue(deviceId int) (time.Time, error) {

	var propName string

	if deviceId == web {
		propName = "WebTimeOut"
	} else if deviceId == mobile {
		propName = "MobileTimeOut"
	} else {
		ERROR.Println("Invalid deviceId")
		return time.Time{}, errors.New("Invalid device")
	}

	timeOutString, err := GetAppProperties(propName)

	if err != nil {
		ERROR.Println(err.Error())
		return time.Time{}, err
	}

	timeOutValue, err := strconv.Atoi(timeOutString)

	if err != nil {
		ERROR.Println("String to integer conversion failed for " + propName)
		return time.Time{}, err
	}

	expiryTime := time.Now().UTC().Add(time.Duration(timeOutValue) * time.Minute)

	return expiryTime, nil
}

//Generates Session Token based on string and key provided
func generateSessionToken(message string, key string) string {
	keyBytes := []byte(key)
	h := hmac.New(sha256.New, keyBytes)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func CreateNewUser(request CreateUserRequest, db *gorm.DB) (int, error) {

	var err error

	if ValidateUniqueUsername(request.Username, db) {

		var user User

		user.UserName = request.Username
		user.LegalName = request.LegalName
		user.Password, err = EncryptPassword(request.Password)
		user.UpdatedAt = time.Now().UTC()
		user.Active = false

		if err != nil {
			ERROR.Println(err.Error())
			return 500, EncryptionError
		}

		db.Save(&user)
		TRACE.Println("New user created, userId: " + strconv.FormatInt(user.Id, 10))

		db.Save(&PasswordRecord{UserId: user.Id, LoginCount: 0})
		TRACE.Println("New password record created, userId: " + strconv.FormatInt(user.Id, 10))

		return 200, nil
	}

	ERROR.Println("Duplicate username " + request.Username + " server validation in create user")
	return 409, DuplicateUsernameError
}

//Returns true if the username is valid i.e. doesn't already exist, else returns false.
func ValidateUniqueUsername(username string, db *gorm.DB) bool {

	TRACE.Println("Validating username " + username)
	err := db.Where(&User{UserName: username}).First(&User{}).Error

	if err == gorm.RecordNotFound {
		return true
	}
	WARNING.Println("Username " + username + " already present")
	return false
}

func EncryptPassword(password string) (string, error) {

	encryptedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(encryptedPasswordBytes), nil
}

func Status() string {
	return "alive at " + time.Now().String()
}

func ValidateSessionToken(sessionToken string, db *gorm.DB) (error, int64) {
	var token Token

	dbErr := db.Where(&Token{Token: sessionToken}).Find(&token).Error
	if dbErr != nil {
		if dbErr == gorm.RecordNotFound {
			WARNING.Println(InvalidSessionToken.Error() + ", sessionToken: " + sessionToken)
			return InvalidSessionToken, -1
		}
		ERROR.Println(dbErr.Error() + ", sessionToken: " + sessionToken)
		return dbErr, -1
	}

	if token.Active && token.ExpiresAt.After(time.Now().UTC()) {
		TRACE.Println("SessionToken validated: " + sessionToken)
		timeLeft := token.ExpiresAt.Sub(time.Now().UTC())
		if timeLeft.Minutes() < 5 {
			token.ExpiresAt = token.ExpiresAt.Add(time.Duration(GetTimeExtension()) * time.Minute)
		}
		token.LastAccessedAt = time.Now().UTC()
		db.Save(&token)
		return nil, token.UserId
	}

	return ExpiredSessionToken, -1
}

func GetTimeExtension() int {

	timeExtensionString, err := GetAppProperties("TimeExtension")
	if err != nil {
		ERROR.Println(err.Error())
		return 0
	}

	extension, conversionErr := strconv.Atoi(timeExtensionString)
	if conversionErr != nil {
		ERROR.Println(conversionErr.Error())
		return 0
	}

	return extension
}

func ChangePassword(username string, newpassword string, db *gorm.DB) (int, error) {
	var user User
	var err error

	dbErr := db.Where(&User{UserName: username}).Find(&user).Error
	if dbErr != nil {
		if dbErr == gorm.RecordNotFound {
			WARNING.Println(UnregisteredUser.Error() + " ,username: " + username)
			return 401, UnregisteredUser
		} else {
			ERROR.Println(dbErr.Error())
			return 500, DatabaseError
		}
	}

	user.Password, err = EncryptPassword(newpassword)
	if err != nil {
		ERROR.Println(err.Error())
		return 500, EncryptionError
	}

	db.Save(&user)
	updatePasswordErr := UpdatePasswordRecordLoginCount(user.Id, false, db)

	if updatePasswordErr != nil {
		ERROR.Println("Password record not updated for userId: " + strconv.FormatInt(user.Id, 10) + ", Error details: " + updatePasswordErr.Error())
		return 500, DatabaseError
	}

	db.Save(&ActivityLog{UserId: user.Id, TokenId: -1, ActivityTime: time.Now().UTC(), Event: PASSWORD_CHANGE})
	TRACE.Println("Password changed for userId: " + strconv.FormatInt(user.Id, 10))

	return 200, nil
}

//UpdatePasswordRecordLoginCount if resetFlag = true, LoginCount it reset to 0 else it is incremented by 1
func UpdatePasswordRecordLoginCount(userid int64, resetFlag bool, db *gorm.DB) error {
	var record PasswordRecord
	dbErr := db.Where(&PasswordRecord{UserId: userid}).Find(&record).Error
	if dbErr != nil {
		return dbErr
	}

	if resetFlag {
		record.LoginCount = 0
	} else {
		record.LoginCount++
	}

	db.Save(&record)

	return nil
}

//CheckPermissions for user
//TODO: Refactor with raw sql if performance bottleneck
func CheckPermissionsForUser(userid int64, permission Permission, db *gorm.DB) error {
	if userid == int64(-1) {
		return PermissionDenied
	}

	err := db.Where(&Permission{PermissionDescription: permission.PermissionDescription}).Find(&permission).Error
	if err != nil {
		ERROR.Println(err.Error() + " while checking permission for: " + permission.PermissionDescription)
		return err
	}

	var groupPermission GroupPermission
	err = db.Where(&GroupPermission{PermissionId: permission.Id}).Find(&groupPermission).Error
	if err != nil {
		ERROR.Println(err.Error() + " while finding groupPermission for: " + strconv.FormatInt(permission.Id, 10))
		return err
	}

	var userGroup UserGroup
	err = db.Where(&UserGroup{UserId: userid, GroupId: groupPermission.GroupId}).Find(&userGroup).Error
	if err == gorm.RecordNotFound {
		WARNING.Println(PermissionDenied.Error() + " for user: " + strconv.FormatInt(userid, 10) +
			" for Permission: " + permission.PermissionDescription)
		return PermissionDenied
	} else if err != nil {
		ERROR.Println(err.Error() + " in UserGroup while searching for userId: " + strconv.FormatInt(userid, 10) +
			" and groupId: " + strconv.FormatInt(groupPermission.GroupId, 10))
		return err
	}

	TRACE.Println("Permission verified for userId: " + strconv.FormatInt(userid, 10) + " , permission: " + permission.PermissionDescription)
	return nil
}

//End session for logout
