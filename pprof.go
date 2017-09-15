package conf

import "runtime"

// PPROF is a confiuration struct which can be used to configure the runtime
// profilers of programs.
//
//	config := struct{
//		PPROF `conf:"pprof"`
//	}
//	conf.Load(&config)
//	conf.SetPPROF(config.PPROF)
//
type PPROF struct {
	BlockProfileRate     float64 `conf:"block-profile-rate"     help:"Sets the mutex profile fraction to enable runtime profiling of lock contention, zero disables mutex profiling" validate:"min=0,max=1"`
	MutexProfileFraction float64 `conf:"mutex-profile-fraction" help:"Sets the mutex profile fraction to enable runtime profiling of lock contention, zero disables mutex profiling" validate:"min=0,max=1"`
}

// DefaultPPROF returns the default value of a PPROF struct. Note that the
// zero-value is valid, DefaultPPROF differs because it captures the current
// configuration of the program's runtime.
func DefaultPPROF() PPROF {
	return PPROF{
		MutexProfileFraction: 1 / float64(runtime.SetMutexProfileFraction(-1)),
	}
}

// SetPPROF configures the runtime profilers based on the given PPROF config.
func SetPPROF(config PPROF) {
	runtime.SetBlockProfileRate(int(1 / config.BlockProfileRate))
	runtime.SetMutexProfileFraction(int(1 / config.MutexProfileFraction))
}
