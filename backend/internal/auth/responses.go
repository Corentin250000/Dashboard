package auth

type LoginSuccessResponse struct {
	Message string `json:"message" example:"Logged in"`
}

type LoginBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid request body: email and password are required."`
}

type LoginUnauthorizedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid credentials"`
}

type LoginExternalResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"This account uses an external login"`
}

type LoginUnconfirmedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Account not confirmed, please check your email"`
}

type LoginUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"database not configured"`
}

type LoginServerErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Server error: unable to process the login."`
}

type LogoutSuccessResponse struct {
	Message string `json:"message" example:"Logged out"`
}

type DeleteAccountSuccessResponse struct {
	Message string `json:"message" example:"Account deleted"`
}

type DeleteAccountErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to delete account"`
}

type DeleteAccountUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"database not configured"`
}

type CurrentUserResponse struct {
	ID    string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"user@example.com"`
}

type OAuthErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth provider is not configured"`
}

type SessionUnauthorizedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Authentication required."`
}

type SessionExpiredResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Session expired or invalid."`
}

type SessionErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to validate session."`
}

type OAuthProviderNotFoundResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth provider not supported."`
}

type OAuthNotConfiguredResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth provider is not configured. Set CLIENT_ID_<PROVIDER> and CLIENT_SECRET_<PROVIDER>."`
}

type OAuthStartErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to start OAuth login."`
}

type OAuthInvalidStateResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid OAuth state. Restart the login from /auth/{provider}."`
}

type OAuthSessionExpiredResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth session expired. Restart the login from /auth/{provider}."`
}

type OAuthAuthorizationFailedResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth authorization failed."`
}

type OAuthProfileErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to fetch OAuth profile."`
}

type OAuthNoEmailResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"OAuth provider did not return an email."`
}

type OAuthAccountErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to create account."`
}

type OAuthSessionCreateErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unable to create session."`
}

type RegisterSuccessResponse struct {
	Message string `json:"message" example:"Account created, please check your email"`
}

type RegisterBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid fields: a valid email and a password of at least 8 characters are required."`
}

type RegisterConflictResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"This email is already in use"`
}

type RegisterServerErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Server error: unable to create the account."`
}

type RegisterUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"database not configured"`
}

type ConfirmedResponse struct {
	Status string `json:"status" example:"confirmed"`
}

type AlreadyConfirmedResponse struct {
	Status string `json:"status" example:"already_confirmed"`
}

type ExpiredResponse struct {
	Status string `json:"status" example:"expired"`
}

type InvalidTokenResponse struct {
	Status  string `json:"status" example:"invalid"`
	Message string `json:"message" example:"invalid or already used token"`
}

type ConfirmErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Database error: unable to confirm the account."`
}

type ConfirmBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"missing token"`
}

type ResendSuccessResponse struct {
	Status string `json:"status" example:"ok"`
}

type ResendBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"token required"`
}

type ResendNotFoundResponse struct {
	Status  string `json:"status" example:"invalid"`
	Message string `json:"message" example:"unknown token"`
}

type ResendErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Database error: unable to rotate the token."`
}

type ResendUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"database not configured"`
}
