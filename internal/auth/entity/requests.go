package entity

type CreateUserPayload struct {
	UserEmail      string
	UserPassword   string
	UserFCM        string
	UserSocialID   string
	UserSocialType string
}

type LoginPayload struct {
	UserEmail      string
	UserPassword   string
	UserFCM        string
	UserSocialID   string
	UserSocialType string
}

type UserEmailPayload struct {
	UserEmail string
}

type UserPasswordAuthPayload struct {
	UserEmail  string
	AuthNumber string
}

type UpdateUserPayload struct {
	UserNick  *string
	UserImage *string
}

type UpdateUserPasswordPayload struct {
	UserEmail    *string
	UserPassword string
}

type RefreshTokenPayload struct {
	RefreshToken string
}
