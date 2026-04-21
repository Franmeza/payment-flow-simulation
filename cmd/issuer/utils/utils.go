package utils

// LastFour returns the last 2 bytes of the UID as a display string.
func LastFour(uid string) string {
	if len(uid) >= 5 {
		return uid[len(uid)-5:]
	}

	return uid
}
