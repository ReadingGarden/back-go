package entity

import "time"

type HTTPResp struct {
	RespCode int    `json:"resp_code"`
	RespMsg  string `json:"resp_msg"`
}

type DataResp struct {
	RespCode int         `json:"resp_code"`
	RespMsg  string      `json:"resp_msg"`
	Data     interface{} `json:"data"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type GetUserData struct {
	UserNo         int64     `json:"user_no"`
	UserNick       string    `json:"user_nick"`
	UserEmail      string    `json:"user_email"`
	UserSocialType string    `json:"user_social_type"`
	UserImage      string    `json:"user_image"`
	UserCreatedAt  time.Time `json:"user_created_at"`
	GardenCount    int       `json:"garden_count"`
	ReadBookCount  int       `json:"read_book_count"`
	LikeBookCount  int       `json:"like_book_count"`
}
