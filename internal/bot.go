package koolo

import (
	"context"
	"fmt"
	"github.com/hectorgimenez/koolo/internal/action"
	"github.com/hectorgimenez/koolo/internal/config"
	"github.com/hectorgimenez/koolo/internal/game"
	"github.com/hectorgimenez/koolo/internal/health"
	"github.com/hectorgimenez/koolo/internal/run"
	"github.com/hectorgimenez/koolo/internal/stats"
	"go.uber.org/zap"
	"time"
)

// Bot will be in charge of running the run loop: create games, traveling, killing bosses, repairing, picking...
type Bot struct {
	logger *zap.Logger
	hm     health.Manager
	ab     action.Builder
}

func NewBot(
	logger *zap.Logger,
	hm health.Manager,
	ab action.Builder,
) Bot {
	return Bot{
		logger: logger,
		hm:     hm,
		ab:     ab,
	}
}

func (b *Bot) Run(ctx context.Context, runs []run.Run) error {
	b.logGameStart(runs)

	gameStartedAt := time.Now()

	for k, r := range runs {
		stats.StartRun(r.Name())
		runStart := time.Now()
		b.logger.Info(fmt.Sprintf("Running: %s", r.Name()))

		actions := []action.Action{
			b.ab.RecoverCorpse(),
			b.ab.Stash(),
			b.ab.VendorRefill(),
			b.ab.ReviveMerc(),
			b.ab.Repair(),
			b.ab.Heal(),
		}

		actions = append(actions, r.BuildActions()...)
		actions = append(actions, b.ab.ItemPickup())

		// Don't return town on last run
		if k != len(runs)-1 {
			actions = append(actions, b.ab.ReturnTown())
		}

		running := true
		for running {
			d := game.Status()
			if err := b.hm.HandleHealthAndMana(d); err != nil {
				return err
			}
			if err := b.shouldEndCurrentGame(gameStartedAt); err != nil {
				return err
			}

			for k, act := range actions {
				if !act.Finished(d) {
					err := act.NextStep(d)
					if err != nil {
						fmt.Println(err)
					}
					break
				}
				if len(actions)-1 == k {
					stats.FinishCurrentRun(stats.EventKill)
					b.logger.Info(fmt.Sprintf("Run %s finished, length: %0.2fs", r.Name(), time.Since(runStart).Seconds()))
					running = false
				}
			}
		}
	}

	return nil
}

func (b *Bot) shouldEndCurrentGame(startAt time.Time) error {
	if time.Since(stats.Status.CurrentRunStart).Seconds() > float64(config.Config.MaxGameLength) {
		return fmt.Errorf(
			"max game length reached, try to exit game: %0.2f",
			time.Since(stats.Status.CurrentRunStart).Seconds(),
		)
	}

	return nil
}

func (b *Bot) logGameStart(runs []run.Run) {
	runNames := ""
	for _, r := range runs {
		runNames += r.Name() + ", "
	}
	b.logger.Info(fmt.Sprintf("Starting Game #%d. Run list: %s", stats.Status.TotalGames+1, runNames[:len(runNames)-2]))
}