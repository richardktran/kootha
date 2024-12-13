package model

import "github.com/richardktran/realtime-quiz/gen"

type IDGenerator struct {
	ID     string
	Entity string
}

func IDGeneratorFromProto(p *gen.IDGenerator) *IDGenerator {
	return &IDGenerator{
		ID:     p.GetId(),
		Entity: p.GetEntity(),
	}
}

func IDGeneratorToProto(u *IDGenerator) *gen.IDGenerator {
	return &gen.IDGenerator{
		Id:     u.ID,
		Entity: u.Entity,
	}
}
