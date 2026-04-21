package cardutil

// LastFour returns the trailing 5 characters of a card UID for display.
func LastFour(uid string) string {
	if len(uid) >= 5 {
		return uid[len(uid)-5:]
	}
	return uid
}
