package http

import "time"

type PersonalInformationResponse struct {
	ID                   string    `json:"id"`
	IdentificationNumber string    `json:"identification_number"`
	FirstName            *string   `json:"first_name"`
	LastName             *string   `json:"last_name"`
	PhoneNumber          *string   `json:"phone_number"`
	Email                *string   `json:"email"`
	Address              *string   `json:"address"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ContactResponse struct {
	ID                  string                       `json:"id"`
	ExternalID          string                       `json:"external_id"`
	Label               string                       `json:"label"`
	DisplayName         *string                      `json:"display_name"`
	PersonalInformation *PersonalInformationResponse `json:"personal_information"`
	CreatedAt           time.Time                    `json:"created_at"`
}
