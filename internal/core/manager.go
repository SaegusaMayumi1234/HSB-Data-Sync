package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// ManagerConfig holds options for the Manager.
type ManagerConfig struct {
    // Logger is used for all internal manager log lines.
    // Defaults to slog.Default() when nil.
    Logger *slog.Logger
}

// Manager owns and orchestrates all registered jobs.
type Manager struct {
    app    *App

    cronRunner *cron.Cron
    wg         sync.WaitGroup
    cancel     context.CancelFunc
    ctx        context.Context

    mu   sync.Mutex
    jobs []registeredJob
}

type registeredJob struct {
    job     Job
    jobType JobType
}

// NewManager creates a new Manager. Call Register* methods, then Start.
func NewManager(app *App) *Manager {
    ctx, cancel := context.WithCancel(context.Background())
    return &Manager{
        app:        app,
        cancel:     cancel,
        ctx:        ctx,
        cronRunner: cron.New(cron.WithSeconds()), // supports 6-field expressions
    }
}

// RegisterCron adds a CronJob to the manager.
func (m *Manager) RegisterCron(j CronJob) error {
    _, err := m.cronRunner.AddFunc(j.CronExpression(), func() {
        m.runOnce(m.ctx, j)
    })
    if err != nil {
        return fmt.Errorf("invalid cron expression for job %q: %w", j.Name(), err)
    }
    m.mu.Lock()
    m.jobs = append(m.jobs, registeredJob{job: j, jobType: JobTypeCron})
    m.mu.Unlock()
    m.app.Logger.Info("registered cron job", "job", j.Name(), "expr", j.CronExpression())
    return nil
}

// RegisterInterval adds an IntervalJob to the manager.
func (m *Manager) RegisterInterval(j IntervalJob) {
    m.mu.Lock()
    m.jobs = append(m.jobs, registeredJob{job: j, jobType: JobTypeInterval})
    m.mu.Unlock()
    m.app.Logger.Info("registered interval job", "job", j.Name(), "interval", j.Interval())
}

// RegisterDynamic adds a DynamicJob to the manager.
func (m *Manager) RegisterDynamic(j DynamicJob) {
    m.mu.Lock()
    m.jobs = append(m.jobs, registeredJob{job: j, jobType: JobTypeDynamic})
    m.mu.Unlock()
    m.app.Logger.Info("registered dynamic job", "job", j.Name())
}

// Start launches all registered jobs. It is non-blocking.
func (m *Manager) Start() {
    m.cronRunner.Start()

    m.mu.Lock()
    jobs := make([]registeredJob, len(m.jobs))
    copy(jobs, m.jobs)
    m.mu.Unlock()

    for _, rj := range jobs {
        switch rj.jobType {
        case JobTypeInterval:
            j := rj.job.(IntervalJob)
            m.wg.Add(1)
            go m.runInterval(m.ctx, j)
        case JobTypeDynamic:
            j := rj.job.(DynamicJob)
            m.wg.Add(1)
            go m.runDynamic(m.ctx, j)
        }
    }
    m.app.Logger.Info("job manager started")
}

// Stop gracefully shuts down the manager, waiting for all in-flight jobs to finish.
func (m *Manager) Stop() {
    m.app.Logger.Info("job manager stopping…")
    m.cancel()
    <-m.cronRunner.Stop().Done()
    m.wg.Wait()
    m.app.Logger.Info("job manager stopped")
}

// ── internal runners ──────────────────────────────────────────────────────────

func (m *Manager) runOnce(ctx context.Context, j Job) {
    log := m.app.Logger.With("job", j.Name())
    log.Info("running job")
    start := time.Now()
    if err := j.Run(ctx, m.app); err != nil {
        log.Error("job failed", "error", err, "duration", time.Since(start))
        return
    }
    log.Info("job completed", "duration", time.Since(start))
}

func (m *Manager) runInterval(ctx context.Context, j IntervalJob) {
    defer m.wg.Done()
    log := m.app.Logger.With("job", j.Name(), "type", "interval")

    // Run immediately on first tick, then wait for the interval.
    m.runOnce(ctx, j)

    ticker := time.NewTicker(j.Interval())
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            log.Info("job stopped")
            return
        case <-ticker.C:
            m.runOnce(ctx, j)
        }
    }
}

func (m *Manager) runDynamic(ctx context.Context, j DynamicJob) {
    defer m.wg.Done()
    log := m.app.Logger.With("job", j.Name(), "type", "dynamic")

    var lastErr error
    for {
        // Ask the job how long to wait before the next run.
        next, err := j.NextInterval(ctx, m.app, lastErr)
        if err != nil {
            log.Error("NextInterval returned error, stopping job", "error", err)
            return
        }
        if next <= 0 {
            log.Info("job requested stop (non-positive interval)")
            return
        }

        log.Info("next run scheduled", "in", next)
        select {
        case <-ctx.Done():
            log.Info("job stopped")
            return
        case <-time.After(next):
        }

        start := time.Now()
        log.Info("running job")
        lastErr = j.Run(ctx, m.app)
        if lastErr != nil {
            log.Error("job failed", "error", lastErr, "duration", time.Since(start))
        } else {
            log.Info("job completed", "duration", time.Since(start))
        }
    }
}
