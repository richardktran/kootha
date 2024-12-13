package model

import "github.com/richardktran/realtime-quiz/gen"

type User struct {
	ID   string
	Name string
}

func UserFromProto(p *gen.User) *User {
	return &User{
		ID:   p.GetId(),
		Name: p.GetName(),
	}
}

func UserToProto(u *User) *gen.User {
	return &gen.User{
		Id:   u.ID,
		Name: u.Name,
	}
}
