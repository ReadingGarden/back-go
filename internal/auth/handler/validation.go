package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
)

var createUserEmailPattern = regexp.MustCompile(`^\w+@[a-zA-Z_]+?\.[a-zA-Z]{2,3}$`)

type validationErrorDetail struct {
	Type string   `json:"type"`
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
}

type validationErrorResponse struct {
	Detail []validationErrorDetail `json:"detail"`
}

func decodeCreateUserPayload(c *gin.Context) (entity.CreateUserPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.CreateUserPayload{}, false
	}

	payload := entity.CreateUserPayload{
		UserEmail:      "",
		UserPassword:   "",
		UserFCM:        "",
		UserSocialID:   "",
		UserSocialType: "",
	}

	var details []validationErrorDetail
	if value, exists := body["user_email"]; exists {
		str, valid := decodeStringValue(value)
		if !valid {
			details = append(details, stringTypeError("user_email"))
		} else {
			payload.UserEmail = str
			if !createUserEmailPattern.MatchString(str) {
				details = append(details, validationErrorDetail{
					Type: "value_error.str.regex",
					Loc:  []string{"body", "user_email"},
					Msg:  `string does not match regex "^\w+@[a-zA-Z_]+?\.[a-zA-Z]{2,3}$"`,
				})
			}
		}
	}
	if value, exists := body["user_password"]; exists {
		str, valid := decodeStringValue(value)
		if !valid {
			details = append(details, stringTypeError("user_password"))
		} else {
			payload.UserPassword = str
		}
	}
	if value, exists := body["user_fcm"]; exists {
		str, valid := decodeStringValue(value)
		if !valid {
			details = append(details, stringTypeError("user_fcm"))
		} else {
			payload.UserFCM = str
		}
	}
	if value, exists := body["user_social_id"]; exists {
		str, valid := decodeStringValue(value)
		if !valid {
			details = append(details, stringTypeError("user_social_id"))
		} else {
			payload.UserSocialID = str
		}
	}
	if value, exists := body["user_social_type"]; exists {
		str, valid := decodeStringValue(value)
		if !valid {
			details = append(details, stringTypeError("user_social_type"))
		} else {
			payload.UserSocialType = str
		}
	}

	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.CreateUserPayload{}, false
	}

	return payload, true
}

func decodeLoginPayload(c *gin.Context) (entity.LoginPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.LoginPayload{}, false
	}

	payload := entity.LoginPayload{UserSocialID: "", UserSocialType: ""}
	var details []validationErrorDetail

	payload.UserEmail, details = requiredString(body, "user_email", details)
	payload.UserPassword, details = requiredString(body, "user_password", details)
	payload.UserFCM, details = requiredString(body, "user_fcm", details)
	if value, exists := body["user_social_id"]; exists {
		if str, valid := decodeStringValue(value); valid {
			payload.UserSocialID = str
		} else {
			details = append(details, stringTypeError("user_social_id"))
		}
	}
	if value, exists := body["user_social_type"]; exists {
		if str, valid := decodeStringValue(value); valid {
			payload.UserSocialType = str
		} else {
			details = append(details, stringTypeError("user_social_type"))
		}
	}

	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.LoginPayload{}, false
	}

	return payload, true
}

func decodeUserEmailPayload(c *gin.Context) (entity.UserEmailPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.UserEmailPayload{}, false
	}

	payload := entity.UserEmailPayload{}
	var details []validationErrorDetail
	payload.UserEmail, details = requiredString(body, "user_email", details)
	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.UserEmailPayload{}, false
	}

	return payload, true
}

func decodeUserPasswordAuthPayload(c *gin.Context) (entity.UserPasswordAuthPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.UserPasswordAuthPayload{}, false
	}

	payload := entity.UserPasswordAuthPayload{}
	var details []validationErrorDetail
	payload.UserEmail, details = requiredString(body, "user_email", details)
	payload.AuthNumber, details = requiredString(body, "auth_number", details)
	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.UserPasswordAuthPayload{}, false
	}

	return payload, true
}

func decodeUpdateUserPasswordPayload(c *gin.Context) (entity.UpdateUserPasswordPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.UpdateUserPasswordPayload{}, false
	}

	payload := entity.UpdateUserPasswordPayload{}
	var details []validationErrorDetail
	if value, exists := body["user_email"]; exists {
		if str, valid := decodeStringValue(value); valid {
			payload.UserEmail = &str
		} else {
			details = append(details, stringTypeError("user_email"))
		}
	}
	payload.UserPassword, details = requiredString(body, "user_password", details)
	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.UpdateUserPasswordPayload{}, false
	}

	return payload, true
}

func decodeRefreshTokenPayload(c *gin.Context) (entity.RefreshTokenPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.RefreshTokenPayload{}, false
	}

	payload := entity.RefreshTokenPayload{}
	var details []validationErrorDetail
	payload.RefreshToken, details = requiredString(body, "refresh_token", details)
	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.RefreshTokenPayload{}, false
	}

	return payload, true
}

func decodeUpdateUserPayload(c *gin.Context) (entity.UpdateUserPayload, bool) {
	body, _, ok := decodeJSONBody(c)
	if !ok {
		return entity.UpdateUserPayload{}, false
	}

	payload := entity.UpdateUserPayload{}
	var details []validationErrorDetail
	if value, exists := body["user_nick"]; exists {
		if str, valid := decodeStringValue(value); valid {
			payload.UserNick = &str
		} else {
			details = append(details, stringTypeError("user_nick"))
		}
	}
	if value, exists := body["user_image"]; exists {
		if str, valid := decodeStringValue(value); valid {
			payload.UserImage = &str
		} else {
			details = append(details, stringTypeError("user_image"))
		}
	}

	if len(details) > 0 {
		writeValidationError(c, details)
		return entity.UpdateUserPayload{}, false
	}

	return payload, true
}

func decodeJSONBody(c *gin.Context) (map[string]json.RawMessage, map[string]interface{}, bool) {
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return nil, nil, false
	}

	rawBytes, err := json.Marshal(raw)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return nil, nil, false
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rawBytes, &body); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return nil, nil, false
	}

	return body, raw, true
}

func requiredString(body map[string]json.RawMessage, field string, details []validationErrorDetail) (string, []validationErrorDetail) {
	value, exists := body[field]
	if !exists {
		details = append(details, validationErrorDetail{
			Type: "value_error.missing",
			Loc:  []string{"body", field},
			Msg:  "field required",
		})
		return "", details
	}

	str, valid := decodeStringValue(value)
	if !valid {
		details = append(details, stringTypeError(field))
		return "", details
	}

	return str, details
}

func decodeStringValue(value json.RawMessage) (string, bool) {
	var str string
	if err := json.Unmarshal(value, &str); err != nil {
		return "", false
	}
	return str, true
}

func stringTypeError(field string) validationErrorDetail {
	return validationErrorDetail{
		Type: "type_error.str",
		Loc:  []string{"body", field},
		Msg:  "str type expected",
	}
}

func writeValidationError(c *gin.Context, details []validationErrorDetail) {
	c.JSON(http.StatusUnprocessableEntity, validationErrorResponse{Detail: details})
}
