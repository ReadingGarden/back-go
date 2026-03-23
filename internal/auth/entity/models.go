package entity

import "time"

type User struct {
	UserNo         int64
	UserNick       string
	UserEmail      string
	UserPassword   string
	UserFCM        string
	UserSocialID   string
	UserSocialType string
	UserImage      string
	UserAuthNumber *string
	UserCreatedAt  time.Time
}

type RefreshToken struct {
	ID     int64
	UserNo int64
	Token  string
	Exp    time.Time
}

type Garden struct {
	GardenNo    int64
	GardenTitle string
	GardenInfo  string
	GardenColor string
}

type GardenUser struct {
	ID             int64
	GardenNo       int64
	UserNo         int64
	GardenLeader   bool
	GardenMain     bool
	GardenSignDate time.Time
}

type Push struct {
	UserNo     int64
	PushAppOK  bool
	PushBookOK bool
	PushTime   *time.Time
}

type Book struct {
	BookNo int64
	UserNo int64
}

type BookImage struct {
	ID        int64
	BookNo    int64
	ImageName string
	ImageURL  string
}

type Memo struct {
	ID     int64
	UserNo int64
}

type MemoImage struct {
	ID        int64
	MemoNo    int64
	ImageName string
	ImageURL  string
}

type UserProfile struct {
	UserNo         int64     `json:"user_no"`
	UserNick       string    `json:"user_nick"`
	UserEmail      string    `json:"user_email"`
	UserFCM        string    `json:"user_fcm"`
	UserSocialID   string    `json:"user_social_id"`
	UserSocialType string    `json:"user_social_type"`
	UserImage      string    `json:"user_image"`
	UserAuthNumber *string   `json:"user_auth_number"`
	UserCreatedAt  time.Time `json:"user_created_at"`
}

func NewUserProfile(user User) UserProfile {
	return UserProfile{
		UserNo:         user.UserNo,
		UserNick:       user.UserNick,
		UserEmail:      user.UserEmail,
		UserFCM:        user.UserFCM,
		UserSocialID:   user.UserSocialID,
		UserSocialType: user.UserSocialType,
		UserImage:      user.UserImage,
		UserAuthNumber: user.UserAuthNumber,
		UserCreatedAt:  user.UserCreatedAt,
	}
}
