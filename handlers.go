package main

import (
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
			ERROR.Println(DatabaseError.Error() + " :" + login.Username)
			response := ErrorResponse{Status: "Internal Server Error", Message: SystemError.Error(), Description: ""}
			r.JSON(500, response)
		} else {
			r.JSON(200, map[string]interface{}{"sessionToken": sessionToken})
		}
	} else if authenticateUserErr == FirstLoginPasswordChange {
		response := ErrorResponse{Status: "Forbidden", Message: FirstLoginPasswordChange.Error(), Description: ""}
		r.JSON(403, response)
	} else {
		response := ErrorResponse{Status: "Unauthorized", Message: AuthenticationFailed.Error(), Description: ""}
		r.JSON(401, response)
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
		var response ErrorResponse
		if status == 500 {
			response = ErrorResponse{Status: "Internal Server Error", Message: SystemError.Error(), Description: ""}
		} else if status == 409 {
			response = ErrorResponse{Status: "Conflict", Message: err.Error(), Description: ""}
		}
		r.JSON(status, response)
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
				r.JSON(200, map[string]interface{}{"passwordChanged": true, "sessionToken": sessionToken})
			} else {
				ERROR.Println("DB entry failed for token, returning 500")
				response := ErrorResponse{Status: "Internal Server Error", Message: "Password changed with but encountered: " + SystemError.Error(), Description: ""}
				r.JSON(500, response)
			}
		} else {
			var response ErrorResponse
			if status == 401 {
				response = ErrorResponse{Status: "Unauthorized", Message: AuthenticationFailed.Error(), Description: ""}
			} else if status == 500 {
				response = ErrorResponse{Status: "Internal Server Error", Message: SystemError.Error(), Description: ""}
			}
			r.JSON(status, response)
		}
	} else {
		response := ErrorResponse{Status: "Unauthorized", Message: AuthenticationFailed.Error(), Description: ""}
		r.JSON(401, response)
	}

}

func ValidateSessionTokenHandler(request ValidateSessionTokenRequest, r render.Render) {

	err, userId := ValidateSessionToken(request.SessionToken, &dbConnection)

	if err == nil {
		r.JSON(200, map[string]interface{}{"userId": userId})
	} else {
		ERROR.Println("Error validating sessionToken: " + request.SessionToken + ", error:" + err.Error())
		if err == InvalidSessionToken {
			response := ErrorResponse{Status: "Unauthorized", Message: InvalidSessionToken.Error(), Description: ""}
			r.JSON(401, response)
		} else if err == ExpiredSessionToken {
			response := ErrorResponse{Status: "Forbidden", Message: ExpiredSessionToken.Error(), Description: ""}
			r.JSON(403, response)
		} else {
			response := ErrorResponse{Status: "Internal Server Error", Message: SystemError.Error(), Description: ""}
			r.JSON(500, response)
		}
	}
}

func CheckPermissionsForUserHandler(request CheckPermissionRequest, r render.Render) {

	err := CheckPermissionsForUser(request.UserId, Permission{PermissionDescription: request.PermissionDescription}, &dbConnection)

	if err == nil {
		r.JSON(200, map[string]interface{}{"permissionCheck": true})
	} else if err == PermissionDenied {
		response := ErrorResponse{Status: "Unauthorized", Message: PermissionDenied.Error(), Description: ""}
		r.JSON(401, response)
	} else {
		response := ErrorResponse{Status: "Internal Server Error", Message: SystemError.Error(), Description: ""}
		r.JSON(500, response)
	}

}
