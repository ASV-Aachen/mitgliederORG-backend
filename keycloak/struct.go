package keycloak

type AdminToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	RefreshExpiresIn int `json:"refresh_expires_in"`
	TokenType string `json:"token_type"`
	NotBeforePolicy int `json:"not-before-policy"`
	Scope string `json:"scope"`
}

type GroupToken []struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type UserInfo struct {
	Sub string `json:"sub"`
	EmailVerified bool `json:"email_verified"`
	Name string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Email string `json:"email"`
}