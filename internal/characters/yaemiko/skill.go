package yaemiko

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

var skillFrames []int

// kitsune spawn frame
const skillStart = 34

func init() {
	skillFrames = frames.InitAbilSlice(37) // E -> N1/E
	skillFrames[action.ActionCharge] = 36  // E -> CA
	skillFrames[action.ActionBurst] = 36   // E -> Q
	skillFrames[action.ActionDash] = 20    // E -> D
	skillFrames[action.ActionJump] = 20    // E -> J
	skillFrames[action.ActionSwap] = 20    // E -> Swap
}

func (c *char) Skill(p map[string]int) action.ActionInfo {

	c.Core.Tasks.Add(func() { c.makeKitsune() }, skillStart)
	c.SetCDWithDelay(action.ActionSkill, 4*60, 16)

	return action.ActionInfo{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash], // earliest cancel
		State:           action.SkillState,
	}
}
