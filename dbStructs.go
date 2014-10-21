package main

import "time"

type User struct {
	Id        int64
	UserName  string    `sql:"size:256;not null;unique"`
	LegalName string    `sql:"size:256;not null"`
	Password  string    `sql:"not null"`
	UpdatedAt time.Time `sql:"not null"`
	Active    bool      `sql:"not null"`
}

type Permission struct {
	Id                    int64
	PermissionDescription string `sql:"not null; size: 512"`
}

type Group struct {
	Id               int64
	GroupDescription string `sql:"not null; size: 512"`
}

type GroupPermission struct {
	Id           int64
	GroupId      int64 `sql:"not null"`
	PermissionId int64 `sql:"not null"`
}

type UserGroup struct {
	Id      int64
	UserId  int64 `sql:"not null"`
	GroupId int64 `sql:"not null"`
}

type AppProperties struct {
	Id            int64
	PropertyName  string    `sql:"not null;unique"`
	PropertyValue string    `sql:"type:varchar(256);not null"`
	UpdatedAt     time.Time `sql:"not null"`
}

type ActivityLog struct {
	Id           int64
	UserId       int64     `sql:"not null"`
	TokenId      int64     `sql:"not null"`
	ActivityTime time.Time `sql:"not null"`
	Event        string    `sql:"size:255"`
}

type Token struct {
	Id             int64
	Token          string `sql:"type:varchar(256);not null;unique"`
	UserId         int64
	Key            string `sql:"type:varchar(256);not null"`
	CreatedAt      time.Time
	LastAccessedAt time.Time
	ExpiresAt      time.Time
	DeviceTypeId   int
	Active         bool `sql:not null`
}

type ArchivedToken struct {
	Id             int64
	Token          string `sql:"type:varchar(256);not null;unique"`
	UserId         int64
	Key            string `sql:"type:varchar(256);not null"`
	CreatedAt      time.Time
	LastAccessedAt time.Time
	ExpiresAt      time.Time
	DeviceTypeId   int
	Active         bool `sql:not null`
}

type DeviceType struct {
	Id         int
	Device     string `sql:"not null"`
	DeviceCode int    `sql:"not null;unique"`
}

type PasswordRecord struct {
	Id         int
	UserId     int64
	LoginCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
