package model

func safeString(value *string) (result string) {
	if value != nil {
		result = *value
	}
	return
}
