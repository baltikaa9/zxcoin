package dto

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func validateHexField(field string, value string, expectedBytes int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}

	decoded, err := hex.DecodeString(value)

	if err != nil {
		return fmt.Errorf("%s must be valid hex: %w", field, err)
	}

	if len(decoded) != expectedBytes {
		return fmt.Errorf(
			"%s must contain %d bytes, got %d",
			field,
			expectedBytes,
			len(decoded),
		)
	}

	return nil
}
