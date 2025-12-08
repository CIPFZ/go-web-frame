package model

type JwtBlacklist struct {
	BaseModel
	Jwt string `gorm:"type:text;comment:jwt"`
}
