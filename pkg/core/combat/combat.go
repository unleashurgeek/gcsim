// Package combat handles all combat related functionalities including
//	- target tracking
//	- target selection
//	- hitbox collision checking
//  - attack queueing
package combat

import (
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
)

type CharHandler interface {
	ByIndex(int) Character
}

type Character interface {
	ApplyAttackMods(a *AttackEvent, t Target) []interface{}
}

type Handler struct {
	log    glog.Logger
	events event.Eventter
	team   CharHandler

	targets     []Target
	TotalDamage float64
	DamageMode  bool
}

func New(log glog.Logger, events event.Eventter, team CharHandler, damageMode bool) *Handler {
	h := &Handler{
		log:        log,
		events:     events,
		team:       team,
		DamageMode: damageMode,
	}
	h.targets = make([]Target, 5)

	return h
}

func (h *Handler) AddTarget(t Target) {
	h.targets = append(h.targets, t)
	t.SetIndex(len(h.targets) - 1)
}

func (h *Handler) Target(i int) Target {
	return h.targets[i]
}

func (h *Handler) SetTargetPos(i int, x, y float64) {
	h.targets[i].SetPos(x, y)
}
