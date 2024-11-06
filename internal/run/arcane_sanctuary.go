package run

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
)

type ArcaneSanctuary struct {
	baseRun
}

func (a ArcaneSanctuary) Name() string {
	return string(config.ArcaneSanctuarylsRun)
}

func (a ArcaneSanctuary) BuildActions() []action.Action {
	openChests := a.CharacterCfg.Game.ArcaneSanctuaryls.OpenChests
	onlyElites := a.CharacterCfg.Game.ArcaneSanctuaryls.FocusOnElitePacks
	filter := data.MonsterAnyFilter()

	if onlyElites {
		filter = data.MonsterEliteFilter()
	}

	actions := []action.Action{
		a.builder.WayPoint(area.ArcaneSanctuary),         // Moving to starting point (Arcane Sanctuary)
		a.builder.MoveToArea(area.AncientTunnels), // Travel to ancient tunnels
	}

	actions = append(actions,
		a.builder.OpenTPIfLeader(),
	)

	// Clear Arcane Sanctuary
	return append(actions, a.builder.ClearArea(openChests, filter))
}