package core

import (
	"bytes"
	"context"
	"fmt"
	"runtime/pprof"
	"sync"
	"time"
)

// CPUProfiler implements CPU profiling
type CPUProfiler struct{}

// NewCPUProfiler creates a new CPU profiler
func NewCPUProfiler() Profiler {
	return &CPUProfiler{}
}

// StartProfiling starts CPU profiling
func (c *CPUProfiler) StartProfiling(ctx context.Context, task ProfilingTask) (ProfileSession, error) {
	return NewCPUProfileSession(ctx, task)
}

// GetProfileType returns the profiling type
func (c *CPUProfiler) GetProfileType() string {
	return "cpu"
}

// CPUProfileSession represents a CPU profiling session
type CPUProfileSession struct {
	ctx       context.Context
	task      ProfilingTask
	buffer    *bytes.Buffer
	startTime time.Time
	mu        sync.Mutex
	running   bool
}

// NewCPUProfileSession creates a new CPU profiling session
func NewCPUProfileSession(ctx context.Context, task ProfilingTask) (ProfileSession, error) {
	session := &CPUProfileSession{
		ctx:       ctx,
		task:      task,
		buffer:    new(bytes.Buffer),
		startTime: time.Now(),
		running:   true,
	}

	if err := pprof.StartCPUProfile(session.buffer); err != nil {
		session.running = false
		return nil, fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	// Set up automatic stop after duration
	if task.Duration > 0 {
		go func() {
			timer := time.NewTimer(time.Duration(task.Duration) * time.Second)
			defer timer.Stop()

			select {
			case <-timer.C:
				session.Stop()
			case <-ctx.Done():
				session.Stop()
			}
		}()
	}

	return session, nil
}

// Stop stops the profiling session
func (s *CPUProfileSession) Stop() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return s.buffer.Bytes(), nil
	}

	pprof.StopCPUProfile()
	s.running = false

	return s.buffer.Bytes(), nil
}

// GetStartTime returns when the session started
func (s *CPUProfileSession) GetStartTime() time.Time {
	return s.startTime
}

// IsRunning returns true if the session is still active
func (s *CPUProfileSession) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// HeapProfiler implements heap/memory profiling
type HeapProfiler struct{}

// NewHeapProfiler creates a new heap profiler
func NewHeapProfiler() Profiler {
	return &HeapProfiler{}
}

// StartProfiling starts heap profiling
func (h *HeapProfiler) StartProfiling(ctx context.Context, task ProfilingTask) (ProfileSession, error) {
	return NewHeapProfileSession(ctx, task)
}

// GetProfileType returns the profiling type
func (h *HeapProfiler) GetProfileType() string {
	return "heap"
}

// HeapProfileSession represents a heap profiling session
type HeapProfileSession struct {
	ctx       context.Context
	task      ProfilingTask
	startTime time.Time
	mu        sync.Mutex
	running   bool
	stopped   bool
}

// NewHeapProfileSession creates a new heap profiling session
func NewHeapProfileSession(ctx context.Context, task ProfilingTask) (ProfileSession, error) {
	session := &HeapProfileSession{
		ctx:       ctx,
		task:      task,
		startTime: time.Now(),
		running:   true,
		stopped:   false,
	}

	// Set up automatic stop after duration
	if task.Duration > 0 {
		go func() {
			timer := time.NewTimer(time.Duration(task.Duration) * time.Second)
			defer timer.Stop()

			select {
			case <-timer.C:
				session.Stop()
			case <-ctx.Done():
				session.Stop()
			}
		}()
	}

	return session, nil
}

// Stop stops the profiling session and captures heap profile
func (s *HeapProfileSession) Stop() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil, nil
	}

	s.running = false
	s.stopped = true

	// Capture heap profile
	buffer := new(bytes.Buffer)
	if err := pprof.WriteHeapProfile(buffer); err != nil {
		return nil, fmt.Errorf("failed to write heap profile: %w", err)
	}

	return buffer.Bytes(), nil
}

// GetStartTime returns when the session started
func (s *HeapProfileSession) GetStartTime() time.Time {
	return s.startTime
}

// IsRunning returns true if the session is still active
func (s *HeapProfileSession) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GoroutineProfiler implements goroutine profiling
type GoroutineProfiler struct{}

// NewGoroutineProfiler creates a new goroutine profiler
func NewGoroutineProfiler() Profiler {
	return &GoroutineProfiler{}
}

// StartProfiling starts goroutine profiling
func (g *GoroutineProfiler) StartProfiling(ctx context.Context, task ProfilingTask) (ProfileSession, error) {
	return NewGoroutineProfileSession(ctx, task)
}

// GetProfileType returns the profiling type
func (g *GoroutineProfiler) GetProfileType() string {
	return "goroutine"
}

// GoroutineProfileSession represents a goroutine profiling session
type GoroutineProfileSession struct {
	ctx       context.Context
	task      ProfilingTask
	startTime time.Time
	mu        sync.Mutex
	running   bool
	stopped   bool
}

// NewGoroutineProfileSession creates a new goroutine profiling session
func NewGoroutineProfileSession(ctx context.Context, task ProfilingTask) (ProfileSession, error) {
	session := &GoroutineProfileSession{
		ctx:       ctx,
		task:      task,
		startTime: time.Now(),
		running:   true,
		stopped:   false,
	}

	// Set up automatic stop after duration
	if task.Duration > 0 {
		go func() {
			timer := time.NewTimer(time.Duration(task.Duration) * time.Second)
			defer timer.Stop()

			select {
			case <-timer.C:
				session.Stop()
			case <-ctx.Done():
				session.Stop()
			}
		}()
	}

	return session, nil
}

// Stop stops the profiling session and captures goroutine profile
func (s *GoroutineProfileSession) Stop() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return nil, nil
	}

	s.running = false
	s.stopped = true

	// Capture goroutine profile
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return nil, fmt.Errorf("goroutine profile not found")
	}

	buffer := new(bytes.Buffer)
	if err := profile.WriteTo(buffer, 0); err != nil {
		return nil, fmt.Errorf("failed to write goroutine profile: %w", err)
	}

	return buffer.Bytes(), nil
}

// GetStartTime returns when the session started
func (s *GoroutineProfileSession) GetStartTime() time.Time {
	return s.startTime
}

// IsRunning returns true if the session is still active
func (s *GoroutineProfileSession) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}