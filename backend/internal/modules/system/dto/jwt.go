package dto

import (
	"github.com/google/uuid"
)

type BaseClaims struct {
	TokenVersion uint64    `json:"tokenVersion"`
	UUID         uuid.UUID `json:"uuid"`
	UserID       uint      `json:"userId"`
	Username     string    `json:"username"`
	NickName     string    `json:"nickName"`
	AuthorityId  uint      `json:"authorityId"`
}
