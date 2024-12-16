package utils

import (
	"strings"

	"github.com/mendersoftware/mender-server/services/useradm/model"
)

func CheckIfPassSimilarToEmailRaw(mail string, pass string) bool {
	if len(mail) == 0 || len(pass) == 0 {
		return false
	}
	lmail := strings.ToLower(mail)
	lpass := strings.ToLower(pass)
	return strings.Contains(lmail, lpass)
}
func CheckIfPassSimilarToEmail(user *model.User, userUpdate *model.UserUpdate) bool {
	if len(userUpdate.Password) > 0 {
		pwd := userUpdate.Password
		// Check if similar to updated email
		if userUpdate.Email != "" &&
			CheckIfPassSimilarToEmailRaw(string(userUpdate.Email), pwd) {
			return true
		}
		// Check if similar to (old) email
		return CheckIfPassSimilarToEmailRaw(string(user.Email), pwd)
	} else if len(userUpdate.Email) > 0 {
		// check when changing the email
		return CheckIfPassSimilarToEmailRaw(
			string(userUpdate.Email),
			userUpdate.CurrentPassword)
	}
	return false
}
