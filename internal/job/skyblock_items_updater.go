package job

import (
	"context"

	"github.com/robfig/cron/v3"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/core"
)

type SkyblockItemsUpdater struct {
	entryID cron.EntryID
}

func NewSkyblockItemsUpdater() *SkyblockItemsUpdater {
	return &SkyblockItemsUpdater{}
}

func (j *SkyblockItemsUpdater) Name() string {
	return "skyblock_items_updater"
}

func (j *SkyblockItemsUpdater) CronExpression() string {
	return "0 * * * * *"
}

func (b *SkyblockItemsUpdater) SetEntryID(id cron.EntryID) {
	b.entryID = id
}

func (b *SkyblockItemsUpdater) EntryID() cron.EntryID {
	return b.entryID
}

func (j *SkyblockItemsUpdater) Run(ctx context.Context, jc *core.JobConfig) error {
	jc.Logger.Info("starting skyblock items update")
	jc.Logger.Info("done")
	return nil
}
