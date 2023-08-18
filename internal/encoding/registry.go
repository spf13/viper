package encoding

import (
	"sync"

	"github.com/spf13/viper/internal/encoding/codec"
	"github.com/spf13/viper/internal/encoding/dotenv"
	"github.com/spf13/viper/internal/encoding/hcl"
	"github.com/spf13/viper/internal/encoding/ini"
	"github.com/spf13/viper/internal/encoding/javaproperties"
	"github.com/spf13/viper/internal/encoding/json"
	"github.com/spf13/viper/internal/encoding/toml"
	"github.com/spf13/viper/internal/encoding/yaml"
)

const (
	// ErrCodecNotFound is returned when there is no codec registered for a format.
	ErrCodecNotFound = encodingError("codec not found for this format")

	// ErrCodecFormatAlreadyRegistered is returned when a codec is already registered for a format.
	ErrCodecFormatAlreadyRegistered = encodingError("codec already registered for this format")
)

// supportedCodecFormats stores all supported codec, the empty pointers are used to construct a corresponding
// codec object without reflection.
var supportedCodecFormats = map[string]func(args ...interface{}) codec.Codec{
	"yaml":       yaml.New,
	"yml":        yaml.New,
	"json":       json.New,
	"toml":       toml.New,
	"hcl":        hcl.New,
	"tfvars":     hcl.New,
	"ini":        ini.New,
	"properties": javaproperties.New,
	"props":      javaproperties.New,
	"prop":       javaproperties.New,
	"dotenv":     dotenv.New,
	"env":        dotenv.New,
}

type CodecRegistry struct {
	codecs map[string]codec.Codec
	mu     sync.RWMutex

	keyDelim       string
	iniLoadOptions ini.LoadOptions
}

// NewCodecRegistry returns a new, initialized CodecRegistry.
func NewCodecRegistry(keyDelim string, iniLoadOptions ini.LoadOptions) *CodecRegistry {
	return &CodecRegistry{
		codecs:         make(map[string]codec.Codec),
		keyDelim:       keyDelim,
		iniLoadOptions: iniLoadOptions,
	}
}

func (e *CodecRegistry) getCodecLazily(format string) (codec.Codec, error) {
	e.mu.RLock()
	c, ok := e.codecs[format]
	e.mu.RUnlock()
	if ok {
		return c, nil
	}

	newCodecFn, ok := supportedCodecFormats[format]
	if !ok {
		return nil, ErrCodecNotFound
	}

	switch format {
	case "ini":
		c = newCodecFn(e.keyDelim, e.iniLoadOptions)
	case "properties", "props", "prop":
		c = newCodecFn(e.keyDelim)
	default:
		c = newCodecFn()
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.codecs[format] = c
	return c, nil
}

func (e *CodecRegistry) Decode(format string, b []byte, v map[string]interface{}) error {
	decoder, err := e.getCodecLazily(format)
	if err != nil {
		return err
	}
	return decoder.Decode(b, v)
}

func (e *CodecRegistry) Encode(format string, v map[string]interface{}) ([]byte, error) {
	decoder, err := e.getCodecLazily(format)
	if err != nil {
		return nil, err
	}
	return decoder.Encode(v)
}

// RegisterCodec registers a Codec for a format.
// Registering a Codec for an already existing format is not supported.
func (e *CodecRegistry) RegisterCodec(format string, codec codec.Codec) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.codecs[format]; ok {
		return ErrCodecFormatAlreadyRegistered
	}

	e.codecs[format] = codec

	return nil
}
