package httputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContactLabel(t *testing.T) {
	str := func(value string) *string { return &value }

	tests := []struct {
		name              string
		firstName         *string
		lastName          *string
		displayName       *string
		externalContactID string
		want              string
	}{
		{
			name:              "personal information full name wins",
			firstName:         str("Juan"),
			lastName:          str("Perez"),
			displayName:       str("Cliente 1001"),
			externalContactID: "54911",
			want:              "Juan Perez",
		},
		{
			name:              "personal information with only first name",
			firstName:         str("Juan"),
			displayName:       str("Cliente 1001"),
			externalContactID: "54911",
			want:              "Juan",
		},
		{
			name:              "falls back to display name",
			displayName:       str("Cliente 1001"),
			externalContactID: "54911",
			want:              "Cliente 1001",
		},
		{
			name:              "falls back to external contact id",
			externalContactID: "54911",
			want:              "54911",
		},
		{
			name:              "ignores blank display name",
			displayName:       str("   "),
			externalContactID: "54911",
			want:              "54911",
		},
		{
			name:              "ignores blank personal information",
			firstName:         str("  "),
			lastName:          str("  "),
			displayName:       str("Cliente 1001"),
			externalContactID: "54911",
			want:              "Cliente 1001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := ContactLabel(tt.firstName, tt.lastName, tt.displayName, tt.externalContactID)

			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}
