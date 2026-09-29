package httputil

import "strings"

// ContactLabel resuelve la etiqueta de presentación de un contacto:
// nombre completo de la persona, si no el display_name, si no el identificador
// externo del contacto.
func ContactLabel(firstName, lastName, displayName *string, externalContactID string) string {
	if firstName != nil || lastName != nil {
		fullName := strings.TrimSpace(joinName(firstName, lastName))
		if fullName != "" {
			return fullName
		}
	}

	if displayName != nil {
		value := strings.TrimSpace(*displayName)
		if value != "" {
			return value
		}
	}

	return externalContactID
}

func joinName(firstName, lastName *string) string {
	var first, last string
	if firstName != nil {
		first = *firstName
	}
	if lastName != nil {
		last = *lastName
	}

	if first == "" {
		return last
	}

	if last == "" {
		return first
	}

	return first + " " + last
}
