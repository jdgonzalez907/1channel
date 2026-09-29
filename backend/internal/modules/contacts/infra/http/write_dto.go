package http

type SavePersonalInformationRequest struct {
	IdentificationNumber string  `json:"identification_number"`
	FirstName            *string `json:"first_name"`
	LastName             *string `json:"last_name"`
	PhoneNumber          *string `json:"phone_number"`
	Email                *string `json:"email"`
	Address              *string `json:"address"`
}
