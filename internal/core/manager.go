package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Manager owns and orchestrates all registered jobs.
type Manager struct {
    app    *App
    logger *slog.Logger

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
        logger:     app.Logger.With("component", "job_manager"),
        cancel:     cancel,
        ctx:        ctx,
        cronRunner: cron.New(cron.WithSeconds()), // supports 6-field expressions
    }
}

// RegisterCron adds a CronJob to the manager.
func (m *Manager) RegisterCron(j CronJob) error {
    entryID, err := m.cronRunner.AddFunc(j.CronExpression(), func() {
        m.runCron(j)
    })
    if err != nil {
        return fmt.Errorf("invalid cron expression for job %q: %w", j.Name(), err)
    }
    j.SetEntryID(entryID)
    m.mu.Lock()
    m.jobs = append(m.jobs, registeredJob{job: j, jobType: JobTypeCron})
    m.mu.Unlock()
    m.logger.Info("registered cron job", "job", j.Name(), "expr", j.CronExpression())
    return nil
}

// RegisterInterval adds an IntervalJob to the manager.
func (m *Manager) RegisterInterval(j IntervalJob) {
    m.mu.Lock()
    m.jobs = append(m.jobs, registeredJob{job: j, jobType: JobTypeInterval})
    m.mu.Unlock()
    m.logger.Info("registered interval job", "job", j.Name(), "interval", j.Interval())
}

// RegisterDynamic adds a DynamicJob to the manager.
func (m *Manager) RegisterDynamic(j DynamicJob) {
    m.mu.Lock()
    m.jobs = append(m.jobs, registeredJob{job: j, jobType: JobTypeDynamic})
    m.mu.Unlock()
    m.logger.Info("registered dynamic job", "job", j.Name())
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
            go m.runInterval(j)
        case JobTypeDynamic:
            j := rj.job.(DynamicJob)
            m.wg.Add(1)
            go m.runDynamic(j)
        }
    }
    m.logger.Info("job manager started")
}

// Stop gracefully shuts down the manager, waiting for all in-flight jobs to finish.
func (m *Manager) Stop() {
    m.logger.Info("job manager stopping…")
    m.cancel()
    <-m.cronRunner.Stop().Done()
    m.wg.Wait()
    m.logger.Info("job manager stopped")
}

// ── internal runners ──────────────────────────────────────────────────────────

// jobLogger returns a logger pre-seeded with the job name and type,
// inheriting the "component" attribute already set on m.logger.
func (m *Manager) jobLogger(j Job, jt JobType) *slog.Logger {
    return m.app.Logger.With("job", j.Name(), "type", string(jt))
}

func (m *Manager) runOnce(j Job, log *slog.Logger) error {
    log.Info("job starting")
    start := time.Now()
    err := j.Run(m.ctx, &JobConfig{App: m.app, Logger: log})
    if err != nil {
        log.Error("job failed", "error", err, "duration", time.Since(start))
        return err
    }
    log.Info("job completed", "duration", time.Since(start))
    return nil
}

func (m *Manager) runCron(j CronJob) {
    log := m.jobLogger(j, JobTypeCron)
    m.runOnce(j, log)
    log.Info("next run scheduled", "at", m.cronRunner.Entry(j.EntryID()).Next)
}

func (m *Manager) runInterval(j IntervalJob) {
    defer m.wg.Done()
    log := m.jobLogger(j, JobTypeInterval)

    // Run immediately on first tick, then wait for the interval.
    m.runOnce(j, log)
    log.Info("next run scheduled", "at", time.Now().Add(j.Interval()))

    ticker := time.NewTicker(j.Interval())
    defer ticker.Stop()
    for {
        select {
        case <-m.ctx.Done():
            log.Info("job stopped", "reason", "context cancelled")
            return
        case <-ticker.C:
            m.runOnce(j, log)
            log.Info("next run scheduled", "at", time.Now().Add(j.Interval()))
        }
    }
}

func (m *Manager) runDynamic(j DynamicJob) {
    defer m.wg.Done()
    log := m.jobLogger(j, JobTypeDynamic)

    var lastErr error
    for {
        // Ask the job how long to wait before the next run.
        next, err := j.NextInterval(m.ctx, &JobConfig{App: m.app, Logger: log}, lastErr)
        if err != nil {
            log.Error("job stopped", "reason", "scheduling error", "error", err)
            return
        }
        if next <= 0 {
            log.Info("job stopped", "reason", "non-positive interval")
            return
        }

        log.Info("next run scheduled", "at", time.Now().Add(next))
        select {
        case <-m.ctx.Done():
            log.Info("job stopped", "reason", "context cancelled")
            return
        case <-time.After(next):
        }

        lastErr = m.runOnce(j, log)
    }
}
