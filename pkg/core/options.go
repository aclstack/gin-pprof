package core

import "time"

// Options represents configuration options for the profiling manager
type Options struct {
	// MaxConcurrent is the maximum number of concurrent profiling sessions
	MaxConcurrent int `yaml:"max_concurrent" json:"max_concurrent"`
	
	// DefaultDuration is the default profiling duration
	DefaultDuration time.Duration `yaml:"default_duration" json:"default_duration"`
	
	// CleanupInterval is the interval for cleaning up old profiles
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
	
	// MaxFileAge is the maximum age for profile files before cleanup
	MaxFileAge time.Duration `yaml:"max_file_age" json:"max_file_age"`
	
	// Enabled controls whether profiling is enabled
	Enabled bool `yaml:"enabled" json:"enabled"`
	
	// ProfileDir is the directory to store profile files (for file storage)
	ProfileDir string `yaml:"profile_dir" json:"profile_dir"`
	
	// DefaultSampleRate is the default sample rate for profiling
	DefaultSampleRate int `yaml:"default_sample_rate" json:"default_sample_rate"`
}

// DefaultOptions returns default configuration options
func DefaultOptions() Options {
	return Options{
		MaxConcurrent:     3,
		DefaultDuration:   30 * time.Second,
		CleanupInterval:   10 * time.Minute,
		MaxFileAge:        24 * time.Hour,
		Enabled:           true,
		ProfileDir:        "./profiles",
		DefaultSampleRate: 1,
	}
}