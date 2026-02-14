package structs

import (
	"errors"
	"strings"
)

var (
	ErrInvalidUserID    = errors.New("invalid user id")
	ErrInvalidUserName  = errors.New("invalid user name")
	ErrInvalidUserEmail = errors.New("invalid user email")
)

// User 表示基础用户结构。
type User struct {
	ID    int64
	Name  string
	Email string
	Tags  []string
}

// Validate 校验用户字段合法性。
func (u *User) Validate() error {
	if u == nil || u.ID <= 0 {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(u.Name) == "" {
		return ErrInvalidUserName
	}
	if !strings.Contains(u.Email, "@") {
		return ErrInvalidUserEmail
	}
	return nil
}

// Clone 返回用户副本，避免原对象被外部修改。
func (u User) Clone() User {
	clone := u
	if len(u.Tags) > 0 {
		clone.Tags = append([]string(nil), u.Tags...)
	}
	return clone
}

// AddTag 为用户添加标签（幂等）。
func (u *User) AddTag(tag string) {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return
	}
	for _, existing := range u.Tags {
		if existing == tag {
			return
		}
	}
	u.Tags = append(u.Tags, tag)
}
