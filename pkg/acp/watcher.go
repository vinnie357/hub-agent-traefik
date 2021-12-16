package acp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/neo-agent/pkg/acp/basicauth"
	"github.com/traefik/neo-agent/pkg/acp/digestauth"
	"github.com/traefik/neo-agent/pkg/acp/jwt"
	"gopkg.in/yaml.v3"
)

// UpdateFunc is a function called when ACP are modified.
type UpdateFunc func(ctx context.Context, cfgs map[string]*Config) error

// Watcher watches access control policy resources and calls an UpdateFunc when there is a change.
type Watcher struct {
	refreshInterval time.Duration
	acpDir          string

	updateFuncs []UpdateFunc
}

// NewWatcher returns a new watcher to track ACP resources.
func NewWatcher(acpDir string, funcs ...UpdateFunc) *Watcher {
	return &Watcher{
		refreshInterval: 5 * time.Second,
		acpDir:          acpDir,
		updateFuncs:     funcs,
	}
}

// Run runs the watcher.
func (w *Watcher) Run(ctx context.Context) {
	t := time.NewTicker(w.refreshInterval)
	defer t.Stop()

	var previous map[string]*Config

	log.Info().Str("directory", w.acpDir).Msg("Starting ACP watcher")

	for {
		select {
		case <-t.C:
			configs, err := readACPDir(w.acpDir)
			if err != nil {
				log.Error().Err(err).Str("directory", w.acpDir).Msg("Unable to read ACP from directory")
				continue
			}

			if reflect.DeepEqual(previous, configs) {
				log.Trace().Msg("Skipping same ACPs configuration")
				continue
			}

			cfgs := make(map[string]*Config, len(configs))
			for k, v := range configs {
				cfgs[k] = v
			}

			log.Debug().Msg("Executing ACP watcher callbacks")

			updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

			for _, fn := range w.updateFuncs {
				if err = fn(updateCtx, cfgs); err != nil {
					log.Error().Err(err).Msg("Unable to execute ACP watcher callback")
					continue
				}
			}

			cancel()

			previous = cfgs

		case <-ctx.Done():
			return
		}
	}
}

func readACPDir(dir string) (map[string]*Config, error) {
	cfgs := make(map[string]*Config)

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		acpName := filepath.Base(strings.TrimSuffix(path, filepath.Ext(path)))
		if _, ok := cfgs[acpName]; ok {
			return fmt.Errorf("multiple ACP named %q defined", acpName)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %q: %w", path, err)
		}

		var cfg Config
		if err = yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("deserialize ACP configuration: %w", err)
		}

		cfgs[acpName] = &cfg

		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return cfgs, nil
}

func buildRoutes(cfgs map[string]*Config) (http.Handler, error) {
	mux := http.NewServeMux()

	for name, cfg := range cfgs {
		switch {
		case cfg.JWT != nil:
			jwtHandler, err := jwt.NewHandler(cfg.JWT, name)
			if err != nil {
				return nil, fmt.Errorf("create %q JWT ACP handler: %w", name, err)
			}

			path := "/" + name

			log.Debug().Str("acp_name", name).Str("path", path).Msg("Registering JWT ACP handler")

			mux.Handle(path, jwtHandler)

		case cfg.BasicAuth != nil:
			h, err := basicauth.NewHandler(cfg.BasicAuth, name)
			if err != nil {
				return nil, fmt.Errorf("create %q basic auth ACP handler: %w", name, err)
			}
			path := "/" + name
			log.Debug().Str("acp_name", name).Str("path", path).Msg("Registering basic auth ACP handler")
			mux.Handle(path, h)

		case cfg.DigestAuth != nil:
			h, err := digestauth.NewHandler(cfg.DigestAuth, name)
			if err != nil {
				return nil, fmt.Errorf("create %q digest auth ACP handler: %w", name, err)
			}
			path := "/" + name
			log.Debug().Str("acp_name", name).Str("path", path).Msg("Registering digest auth ACP handler")
			mux.Handle(path, h)

		default:
			return nil, errors.New("unknown ACP handler type")
		}
	}

	return mux, nil
}