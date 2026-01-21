package redis

import "fmt"

func KeySession(token string) string {
	return fmt.Sprintf("SES:%s", token)
}

func KeyPreferences(userID string) string {
	return fmt.Sprintf("PREF:%s", userID)
}

func KeyList(name string) string {
	return fmt.Sprintf("L:%s", name)
}

func KeySet(name string) string {
	return fmt.Sprintf("S:%s", name)
}

func KeyCounter(name string) string {
	return fmt.Sprintf("CTR:%s", name)
}
