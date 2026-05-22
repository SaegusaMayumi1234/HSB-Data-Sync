package core

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// JobType identifies the scheduling strategy of a job.
type JobType string

const (
    JobTypeCron     JobType = "cron"
    JobTypeInterval JobType = "interval"
    JobTypeDynamic  JobType = "dynamic"
)

type JobConfig struct {
    App    *App
    Logger *slog.Logger
}

// Job is the base interface every job must implement.
type Job interface {
    // Name returns a unique, human-readable identifier for the job.
    Name() string
    // Run executes the job's logic. The context is cancelled when the manager shuts down.
    Run(ctx context.Context, cfg *JobConfig) error
}

// CronJob is scheduled using a cron expression (e.g. "0 * * * *").
type CronJob interface {
    Job
    // CronExpression returns a standard 5-field (or 6-field with seconds) cron string.
    CronExpression() string
    // SetEntryID is called by the Manager after the job is registered with the cron runner.
    SetEntryID(id cron.EntryID)
    // EntryID returns the cron runner entry ID assigned by the Manager.
    EntryID() cron.EntryID
}

// IntervalJob runs repeatedly at a fixed duration after each completion.
type IntervalJob interface {
    Job
    // Interval returns the fixed wait duration between runs.
    Interval() time.Duration
}

// DynamicJob decides its own next run time after each execution.
type DynamicJob interface {
    Job
    // NextInterval is called after every run to determine the wait duration
    // before the next run. Returning 0 or a negative duration stops the job.
    NextInterval(ctx context.Context, cfg *JobConfig, lastErr error) (time.Duration, error)
}

// BaseCronJob provides a default implementation of SetEntryID/EntryID.
// Embed this in your CronJob structs to satisfy the interface automatically.
type BaseCronJob struct {
    entryID cron.EntryID
}

func (b *BaseCronJob) SetEntryID(id cron.EntryID) { b.entryID = id }
func (b *BaseCronJob) EntryID() cron.EntryID      { return b.entryID }
