package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jimeh/envctl"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

const (
	configFileName = "config.json"
)

type (
	Config struct {
		Path     string                 `envconfig:"PATH"`
		LogLevel string                 `envconfig:"LOG_LEVEL"`
		Peer     PeerConfig             `envconfig:"PEER"`
		Extras   map[string]interface{} `envconfig:"EXTRAS"`
	}
	PeerConfig struct {
		PrivateKey           crypto.PrivateKey `envconfig:"PRIVATE_KEY"`
		BindAddress          string            `envconfig:"BIND_ADDRESS"`
		Bootstraps           []peer.Shorthand  `envconfig:"BOOTSTRAPS"`
		ListenOnLocalIPs     bool              `envconfig:"LISTEN_LOCAL"`
		ListenOnPrivateIPs   bool              `envconfig:"LISTEN_PRIVATE"`
		ListenOnExternalPort bool              `envconfig:"LISTEN_EXTERNAL_PORT"`
	}
	otherOtions struct {
		additionalPairs map[string]string
	}
)

func New(opts ...Option) (*Config, error) {
	k, _ := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	cfg := &Config{
		Path:     "~/.nimona",
		LogLevel: "FATAL",
		Peer: PeerConfig{
			PrivateKey:  k,
			BindAddress: "0.0.0.0:0",
			Bootstraps: []peer.Shorthand{
				"bahwqdag4aeqewwlutsgr7kv2iaqsrnppbdcmyykpckqn5uaqczae6fergklclea@tcps:asimov.bootstrap.nimona.io:22581", // nolint: lll
				"bahwqdag4aeqomor45il7jjxlox7y5aj6cigawcljgsfftytwf6ulrpfqtiuzsya@tcps:egan.bootstrap.nimona.io:22581",   // nolint: lll
				"bahwqdag4aeqm5gkdk7dlbzke6wgc7rkm67cnqiv2jctfoxoo3vjmbdpjt5qi6za@tcps:sloan.bootstrap.nimona.io:22581",  // nolint: lll
			},
		},
		Extras: map[string]interface{}{},
	}
	oopts := &otherOtions{}
	for _, opt := range opts {
		opt(cfg, oopts)
	}

	var err error

	envctl.With(
		oopts.additionalPairs,
		func() {
			err = envconfig.Process("nimona", cfg)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error processing env vars, %w", err)
	}

	for k, r := range cfg.Extras {
		envctl.With(
			oopts.additionalPairs,
			func() {
				err = envconfig.Process("nimona_extras_"+k, r)
			},
		)
		if err != nil {
			return nil, fmt.Errorf("error processing extra env vars, %w", err)
		}
	}

	newPath, err := homedir.Expand(cfg.Path)
	if err != nil {
		return nil, err
	}

	cfg.Path = newPath

	if err := os.MkdirAll(cfg.Path, 0700); err != nil {
		return nil, fmt.Errorf("error creating directory, %w", err)
	}

	return cfg, nil
}

var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

var gatherRegexp = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
var acronymRegexp = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")

func ToEnvPairs(prefix string, spec interface{}) (map[string]string, error) {
	s := reflect.ValueOf(spec)

	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidSpecification
	}
	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidSpecification
	}
	typeOfSpec := s.Type()

	cfg := map[string]string{}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)
		if isTrue(ftype.Tag.Get("ignored")) {
			continue
		}

		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					break
				}
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		key := ftype.Name
		alt := strings.ToUpper(ftype.Tag.Get("envconfig"))

		if isTrue(ftype.Tag.Get("split_words")) {
			key = rewordKey(key)
		}
		if alt != "" {
			key = alt
		}
		if prefix != "" {
			key = fmt.Sprintf("%s_%s", prefix, key)
		}
		key = strings.ToUpper(key)

		if f.Kind() == reflect.Struct {
			_, isStringer := f.Interface().(fmt.Stringer)
			if !isStringer {
				innerPrefix := prefix
				if !ftype.Anonymous {
					innerPrefix = key
				}

				embeddedPtr := f.Addr().Interface()
				embeddedCfg, err := ToEnvPairs(innerPrefix, embeddedPtr)
				if err != nil {
					return nil, err
				}
				for k, v := range embeddedCfg {
					cfg[k] = v
				}
				continue
			}
		}
		if f.IsZero() {
			continue
		}
		if f.Kind() == reflect.Map {
			iter := f.MapRange()
			for iter.Next() {
				mapPrefix, err := toString(iter.Key())
				if err != nil {
					return nil, err
				}
				mapPrefix = strings.ToUpper(
					fmt.Sprintf("%s_%s", key, mapPrefix),
				)
				vs, err := ToEnvPairs(mapPrefix, iter.Value().Interface())
				if err != nil {
					return nil, err
				}
				for k, v := range vs {
					cfg[k] = v
				}
			}
			continue
		}
		v, err := toString(f)
		if err != nil {
			return nil, err
		}
		cfg[key] = v
	}
	return cfg, nil
}

func toString(f reflect.Value) (string, error) {
	switch f.Kind() {
	case reflect.Struct:
		if s, ok := f.Interface().(fmt.Stringer); ok {
			return s.String(), nil
		}
	case reflect.String:
		return f.String(), nil
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return fmt.Sprintf("%d", f.Int()), nil
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		return fmt.Sprintf("%d", f.Uint()), nil
	case reflect.Float32,
		reflect.Float64:
		return fmt.Sprintf("%f", f.Float()), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", f.Bool()), nil
	case reflect.Slice, reflect.Array:
		vs := []string{}
		for i := 0; i < f.Len(); i++ {
			v, err := toString(f.Index(i))
			if err != nil {
				return "", err
			}
			vs = append(vs, v)
		}
		return strings.Join(vs, ","), nil
	}
	return "", fmt.Errorf("unknown kind " + f.Kind().String())
}

func isTrue(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func rewordKey(key string) string {
	words := gatherRegexp.FindAllStringSubmatch(key, -1)
	if len(words) == 1 {
		return key
	}
	var name []string
	for _, words := range words {
		if m := acronymRegexp.FindStringSubmatch(words[0]); len(m) == 3 {
			name = append(name, m[1], m[2])
		} else {
			name = append(name, words[0])
		}
	}
	return strings.Join(name, "_")
}
