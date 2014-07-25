package main

type ConnectionParameters struct {
	username string
	password string
	host     string
	port     string
	dbname   string
	sslmode  string
}

type LoginCredential struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceId int    `json:"deviceId" binding:"required"`
}

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required"`
	LegalName string `json:"legalName" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type ValidateUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type ChangePasswordRequest struct {
	Username    string `json:"username" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
	DeviceId    int    `json:"deviceId" binding:"required"`
}
