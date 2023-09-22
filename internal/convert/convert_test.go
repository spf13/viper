package convert

import (
	"testing"
)

func TestConvert(t *testing.T) {
	type Tmp1 struct {
		Str    string                 `viper:"str"`
		I8     int8                   `viper:"i8"`
		Int16  int16                  `viper:"i16"`
		Int32  int32                  `viper:"i32"`
		Int64  int64                  `viper:"i64"`
		I      int                    `viper:"i"`
		U8     int8                   `viper:"u8"`
		Uint16 int16                  `viper:"u16"`
		Uint32 int32                  `viper:"u32"`
		Uint64 int64                  `viper:"u64"`
		U      int                    `viper:"u"`
		F32    float32                `viper:"f32"`
		F64    float64                `viper:"f64"`
		TF     bool                   `viper:"tf"`
		M      map[string]interface{} `viper:"m"`
		S      []interface{}          `viper:"s"`
	}
	tc := map[string]interface{}{
		"str": "Hello world",
		"i8":  -8,
		"i16": -16,
		"i32": -32,
		"i64": -64,
		"i":   -1,
		"u8":  8,
		"u16": 16,
		"u32": 32,
		"u64": 64,
		"u":   1,
		"f32": 3.32,
		"f64": 3.64,
		"tf":  true,
		"m": map[string]interface{}{
			"im": 123,
		},
		"s": []interface{}{
			"1234",
			1.23,
		},
	}

	var tmp Tmp1
	err := Convert(tc, &tmp)
	if err != nil {
		t.Error(err)
	}
	t.Error(tmp)

}

func BenchmarkConvert(b *testing.B) {
	type Tmp1 struct {
		Str    string                 `viper:"str"`
		I8     int8                   `viper:"i8"`
		Int16  int16                  `viper:"i16"`
		Int32  int32                  `viper:"i32"`
		Int64  int64                  `viper:"i64"`
		I      int                    `viper:"i"`
		U8     int8                   `viper:"u8"`
		Uint16 int16                  `viper:"u16"`
		Uint32 int32                  `viper:"u32"`
		Uint64 int64                  `viper:"u64"`
		U      int                    `viper:"u"`
		F32    float32                `viper:"f32"`
		F64    float64                `viper:"f64"`
		TF     bool                   `viper:"tf"`
		M      map[string]interface{} `viper:"m"`
		S      []interface{}          `viper:"s"`
	}
	tc := map[string]interface{}{
		"str": "Hello world",
		"i8":  -8,
		"i16": -16,
		"i32": -32,
		"i64": -64,
		"i":   -1,
		"u8":  8,
		"u16": 16,
		"u32": 32,
		"u64": 64,
		"u":   1,
		"f32": 3.32,
		"f64": 3.64,
		"tf":  true,
		"m": map[string]interface{}{
			"im": 123,
		},
		"s": []interface{}{
			"1234",
			1.23,
		},
	}
	for i := 0; i < b.N; i++ {
		var tmp Tmp1
		err := Convert(tc, &tmp)
		if err != nil {
			b.Error(err)
		}
	}
}
