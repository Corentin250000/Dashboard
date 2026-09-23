package auth

type LoginSuccessResponse struct {
	Message string `json:"message" example:"Logged in"`
}

type LoginBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid request body"`
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
	Message string `json:"message" example:"Server error"`
}

type RegisterSuccessResponse struct {
	Message string `json:"message" example:"Account created, please check your email"`
}

type RegisterBadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Invalid fields"`
}

type RegisterConflictResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"This email is already in use"`
}

type RegisterServerErrorResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Server error"`
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
	Message string `json:"message" example:"database error"`
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
	Message string `json:"message" example:"database error"`
}

type ResendUnavailableResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"database not configured"`
}
