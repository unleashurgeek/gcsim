package zhongli

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
)

var burstFrames []int

const burstHitmark = 101

func init() {
	burstFrames = frames.InitAbilSlice(139) // Q -> N1/E
	burstFrames[action.ActionDash] = 123    // Q -> D
	burstFrames[action.ActionJump] = 123    // Q -> J
	burstFrames[action.ActionSwap] = 138    // Q -> Swap
}

func (c *char) Burst(p map[string]int) action.ActionInfo {
	//deal damage when created
	ai := combat.AttackInfo{
		ActorIndex: c.Index,
		Abil:       "Planet Befall",
		AttackTag:  combat.AttackTagElementalBurst,
		ICDTag:     combat.ICDTagNone,
		ICDGroup:   combat.ICDGroupDefault,
		StrikeType: combat.StrikeTypeBlunt,
		Element:    attributes.Geo,
		Durability: 100,
		Mult:       burst[c.TalentLvlBurst()],
		FlatDmg:    0.33 * c.MaxHP(),
	}
	c.Core.QueueAttack(ai, combat.NewCircleHit(c.Core.Combat.Player(), 5), burstHitmark, burstHitmark)

	if c.Base.Cons >= 2 {
		c.addJadeShield()
	}

	c.SetCD(action.ActionBurst, 720)
	c.ConsumeEnergy(7)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash], // earliest cancel
		State:           action.BurstState,
	}
}
