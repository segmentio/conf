package conf

import (
	"testing"
)

type kinesisConfig struct {
	StreamName string
}
type testConfig struct {
	Kinesis kinesisConfig `conf:"kinesis"`
}

func TestEnvSource(t *testing.T) {
	t.Run("Struct", func(t *testing.T) {
		src := NewEnvSource("collector", "COLLECTOR_KINESIS_STREAM_NAME=blah")
		a := testConfig{}
		loader := Loader{
			Name:    "collector",
			Args:    []string{},
			Sources: []Source{src},
		}
		if _, _, err := loader.Load(&a); err != nil {
			t.Fatal(err)
		}
		if a.Kinesis.StreamName != "blah" {
			t.Errorf("expected StreamName to get populated, got %q", a.Kinesis.StreamName)
		}
	})

	t.Run("Map", func(t *testing.T) {
		src := NewEnvSource("", "STREAM_NAME=blah")
		cfg := struct {
			StreamName string
		}{}
		loader := Loader{
			Name:    "collector",
			Args:    []string{},
			Sources: []Source{src},
		}
		if _, _, err := loader.Load(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.StreamName != "blah" {
			t.Errorf("expected 'blah' stream name, got %q", cfg.StreamName)
		}
	})
}
