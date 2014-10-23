package main

import (
	"strconv"

	"github.com/martini-contrib/render"
)

//Returns current server status
func StatusHandler(r render.Render) {
	r.JSON(200, map[string]interface{}{"status": Status()})
}

//Logins a user by checking login credentials, returns a UserToken if login is successful
func LoginHandler(login LoginCredential, r render.Render) {

	authenticateUserErr := AuthenticateUser(login.Username, login.Password, &dbConnection)

	if authenticateUserErr == nil {
		sessionToken, err := CreateNewTokenDbEntry(login, &dbConnection)
		if err != nil {
			ERROR.Println("DB entry failed for token, returning 500")
			r.JSON(500, "Ughh, something wrong with the server")
		} else {
			r.JSON(200, map[string]interface{}{"sessionToken": sessionToken, "error": nil})
		}
	} else if authenticateUserErr == FirstLoginPasswordChange {
		r.JSON(200, map[string]interface{}{"sessionToken": "", "error": FirstLoginPasswordChange.Error()})
	} else {
		r.JSON(401, AuthenticationFailed.Error())
	}
}

func ValidateUsernameHandler(usernameRequest ValidateUsernameRequest, r render.Render) {

	validUsernameFlag := ValidateUniqueUsername(usernameRequest.Username, &dbConnection)

	r.JSON(200, map[string]interface{}{"valid": validUsernameFlag})
}

func CreateUserHandler(request CreateUserRequest, r render.Render) {

	status, err := CreateNewUser(request, &dbConnection)

	if err != nil {
		ERROR.Println(err.Error())
		r.JSON(status, map[string]interface{}{"userCreated": false})
	} else {
		r.JSON(status, map[string]interface{}{"userCreated": true})
	}
}

func ChangePasswordHandler(request ChangePasswordRequest, r render.Render) {

	authenticateUserError := AuthenticateUser(request.Username, request.OldPassword, &dbConnection)
	if authenticateUserError == nil || authenticateUserError == FirstLoginPasswordChange {
		status, changePasswordError := ChangePassword(request.Username, request.NewPassword, &dbConnection)
		if changePasswordError == nil {
			login := LoginCredential{Username: request.Username, Password: request.NewPassword, DeviceId: request.DeviceId}
			sessionToken, err := CreateNewTokenDbEntry(login, &dbConnection)
			if err == nil {
				r.JSON(200, map[string]interface{}{"passwordChanged": true, "sessionToken": sessionToken, "error": nil})
			} else {
				ERROR.Println("DB entry failed for token, returning 500")
				r.JSON(500, map[string]interface{}{"passwordChanged": true, "sessionToken": "", "error": err.Error()})
			}
		} else {
			r.JSON(status, map[string]interface{}{"passwordChanged": false})
		}
	} else {
		r.JSON(401, AuthenticationFailed.Error())
	}

}

func ValidateSessionTokenHandler(request ValidateSessionTokenRequest, r render.Render) {

	err, userId := ValidateSessionToken(request.SessionToken, &dbConnection)

	if err != nil {
		ERROR.Println("Error validating sessionToken: " + request.SessionToken + ", error:" + err.Error())
		r.JSON(500, map[string]interface{}{"userId": userId, "error": err.Error()})
	} else {
		r.JSON(200, map[string]interface{}{"userId": userId, "error": nil})
	}

}

func CheckPermissionsForUserHandler(request CheckPermissionRequest, r render.Render) {

	err := CheckPermissionsForUser(request.UserId, Permission{PermissionDescription: request.PermissionDescription}, &dbConnection)
	if err != nil {
		ERROR.Println("Error checking for permission for userId: " + strconv.FormatInt(request.UserId, 10) + ", permission: " + request.PermissionDescription)
		r.JSON(500, map[string]interface{}{"permissionCheckResult": err.Error()})
	} else {
		r.JSON(200, map[string]interface{}{"permissionCheckResult": nil})
	}
}
