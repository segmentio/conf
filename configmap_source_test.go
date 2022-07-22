package conf

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConfigMap(t *testing.T) {
	t.Run("Source", func(t *testing.T) {
		src := NewKubernetesConfigMapSource("", "./testdata/configmap")
		cfg := struct {
			CollectorKinesisEndpoint string
		}{}
		loader := Loader{
			Name:    "collector",
			Args:    []string{},
			Sources: []Source{src},
		}
		if _, _, err := loader.Load(&cfg); err != nil {
			t.Fatal(err)
		}
		if cfg.CollectorKinesisEndpoint != "https://example.com/blah" {
			t.Fatalf("bad value: want example.com/blah got %q", cfg.CollectorKinesisEndpoint)
		}
	})

	t.Run("NestedConfig", func(t *testing.T) {
		a := testConfig{}
		loader := Loader{
			Name: "collector",
			Args: []string{},
			Sources: []Source{
				NewKubernetesConfigMapSource("", "./testdata/configmap"),
			},
		}
		loader.Load(&a)
		if a.Kinesis.StreamName != "segment-logs" {
			t.Errorf("loading nested config did not work correctly")
		}
	})

	t.Run("Prefix", func(t *testing.T) {
		a := struct {
			Kinesis struct {
				Endpoint string
			}
		}{}
		loader := Loader{
			Name: "name",
			Args: []string{},
			Sources: []Source{
				NewKubernetesConfigMapSource("collector", "./testdata/configmap"),
			},
		}
		loader.Load(&a)
		if a.Kinesis.Endpoint != "https://example.com/blah" {
			t.Errorf("loading config with prefix did not work correctly")
		}
	})
}

func TestSubscriber(t *testing.T) {
	tmp, _ := os.MkdirTemp("", "conf-configmap-")
	defer os.RemoveAll(tmp)
	oldInterval := kubernetesSleepInterval
	defer func() {
		kubernetesSleepInterval = oldInterval
	}()
	kubernetesSleepInterval = 3 * time.Millisecond
	t.Run("ValueExists", func(t *testing.T) {
		path := filepath.Join(tmp, "test1")
		if err := os.WriteFile(path, []byte("5\n"), 0640); err != nil {
			t.Fatal(err)
		}
		sc := NewKubernetesSubscriber("", tmp)
		ctx, cancel := context.WithCancel(context.Background())
		count := 0
		sc.Subscribe(ctx, func(key, newValue string) {
			count++
		})
		time.Sleep(10 * time.Millisecond)
		cancel()
		if count != 0 {
			t.Fatalf("expected f to get called zero times, got called %d times", count)
		}
	})

	t.Run("ValueChanges", func(t *testing.T) {
		path := filepath.Join(tmp, "test2")
		if err := os.WriteFile(path, []byte("7\n"), 0640); err != nil {
			t.Fatal(err)
		}
		sc := NewKubernetesSubscriber("", tmp)
		ctx, cancel := context.WithCancel(context.Background())
		count := 0
		value := ""
		var mu sync.Mutex
		sc.Subscribe(ctx, func(key, newValue string) {
			mu.Lock()
			defer mu.Unlock()
			count++
			value = newValue
		})
		go func() {
			time.Sleep(2 * time.Millisecond)
			if err := os.WriteFile(path, []byte("11\n"), 0640); err != nil {
				panic(err)
			}
		}()
		time.Sleep(10 * time.Millisecond)
		cancel()
		mu.Lock()
		defer mu.Unlock()
		if count == 0 {
			t.Fatalf("expected f to get called at least once, got called %d times", count)
		}
		if value != "11" {
			t.Fatalf("bad value: want 11 got %q", value)
		}
	})
}
