package main

import ()


// ------------------------------------------------------------------------------------------------------------
// SQL Structs
type jsonStruct struct {
	Website        []website        `json:"WEBSITE"`
	Keycloak       []keycloak_user       `json:"KEYCLOAK"`
	Arbeitsstunden []arbeitsstunden `json:"ARBEITSSTUNDEN"`
}

type website struct {
	ID            string `json:"id"`
	USERNAME      string `json:"username"`
	FIRST_NAME    string `json:"first_name"`
	LAST_NAME     string `json:"last_name"`
	EMAIL         string `json:"email"`
	IS_ACTIVE     bool   `json:"is_active"`
	PROFILE_IMAGE string `json:"profile_image"`
	STATUS        string `json:"status"`
}

type keycloak_user struct {
	ID             string `json:"id"`
	EMAIL          string `json:"Email"`
	EMAIL_VERIFIED []byte `json:"Email_verified"`
	ENABLED        []byte `json:"enabled"`
	FIRST_NAME     string `json:"first_name"`
	LAST_NAME      string `json:"last_name"`
	USERNAME       string `json:"username"`
}

type arbeitsstunden struct {
	FIRST_NAME string `json:"first_name"`
	LAST_NAME  string `json:"last_name"`
	EMAIL      string `json:"email"`
}


