package conf

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewKubernetesConfigMapSource loads configuration from a Kubernetes ConfigMap
// that has been mounted as a volume.
//
// A prefix may be set to namespace the environment variables that the source
// will be looking at.
func NewKubernetesConfigMapSource(prefix string, dir string) Source {
	base := make([]string, 0, 10)
	if prefix != "" {
		base = append(base, prefix)
	}
	return SourceFunc(func(dst Map) error {
		f, err := os.Open(dir)
		if err != nil {
			return err
		}
		defer f.Close()
		entries, err := f.Readdirnames(0)
		if err != nil {
			return err
		}
		vars := make(map[string]string, 0)
		for _, entry := range entries {
			if len(entry) > 0 && entry[0] == '.' {
				continue
			}
			path := filepath.Join(f.Name(), entry)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			vars[snakecaseUpper(entry)] = string(bytes.TrimSuffix(data, []byte{'\n'}))
		}
		dst.Scan(func(path []string, item MapItem) {
			path = append(base, path...)
			path = append(path, item.Name)

			k := snakecaseUpper(strings.Join(path, "_"))
			if v, ok := vars[k]; ok {
				// this only matches at the very end
				if e := item.Value.Set(v); e != nil {
					err = e
				}
			}
		})
		return nil
	})
}

type Subscriber interface {
	// Subscribe listens for new configuration, invoking the callback when
	// values change. f should be invoked any time Subscribe detects a new key,
	// or an existing key with a new value. If a key is deleted f will be
	// invoked with the value set to the empty string. There is no way to
	// distinguish between a deleted key and an empty key.
	//
	// If the value is retrieved and is empty (file not found), f is invoked
	// with the empty string. At most one instance of f will be invoked at any
	// time per Subscriber instance. If the value cannot be retrieved (read
	// error), f will not be invoked.
	Subscribe(ctx context.Context, f func(key, newValue string))

	// Snapshot returns a copy of the current configuration.
	Snapshot(ctx context.Context) (map[string]string, error)
}

type kubernetesSubscriber struct {
	prefix string
	dir    string
}

func NewKubernetesSubscriber(prefix string, dir string) Subscriber {
	return kubernetesSubscriber{prefix: prefix, dir: dir}
}

// can be overridden in tests
var kubernetesSleepInterval = 30 * time.Second

func (k kubernetesSubscriber) Subscribe(ctx context.Context, f func(key, newValue string)) {
	ticker := time.NewTicker(kubernetesSleepInterval)
	state, initialErr := k.Snapshot(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				newState, err := k.Snapshot(ctx)
				if err != nil {
					continue
				}
				if initialErr != nil {
					initialErr = nil
					// We shouldn't hit any callbacks if we don't have any
					// values to diff
					continue
				}
				newset := make(map[string]bool, len(newState))
				for key, value := range newState {
					newset[key] = true
					oldVal, found := state[key]
					if !found {
						// key has been added
						f(key, value)
						continue
					}
					if oldVal != value {
						// key has been changed.
						f(key, value)
						continue
					}
				}
				for key := range state {
					if !newset[key] {
						// key has been deleted
						f(key, "")
						continue
					}
				}
				state = newState
			}
		}
	}()
}

func (k kubernetesSubscriber) Snapshot(ctx context.Context) (map[string]string, error) {
	f, err := os.Open(k.dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	names, err := f.Readdirnames(10000)
	if err != nil {
		return nil, err
	}
	mp := make(map[string]string, len(names))
	for i := range names {
		data, err := os.ReadFile(filepath.Join(k.dir, names[i]))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		mp[names[i]] = strings.TrimSuffix(string(data), "\n")
	}
	return mp, nil
}
