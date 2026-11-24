// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package config

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// decodeHook is viper's default hook with one more step before the split of
// a string into a list: a string becomes a list of signing keys.
func decodeHook() viper.DecoderConfigOption {
	return viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		stringToSigningKeys,
		stringToSlice,
	))
}

// stringToSigningKeys reads controller.api.security.signing_keys from the
// environment, where a variable cannot hold a YAML list: id=key pairs with
// a comma between them, such as 2026-11=<key>,2026-08=<key>. A pair without
// = is a key with no ID, which the config's validation refuses.
func stringToSigningKeys(
	f reflect.Type,
	t reflect.Type,
	data any,
) (any, error) {
	if f.Kind() != reflect.String || t != reflect.TypeFor[[]SigningKey]() {
		return data, nil
	}

	raw := data.(string)
	if raw == "" {
		return []SigningKey{}, nil
	}
	pairs := strings.Split(raw, ",")
	keys := make([]SigningKey, 0, len(pairs))
	for _, pair := range pairs {
		id, key, ok := strings.Cut(pair, "=")
		if !ok {
			id, key = "", pair
		}
		keys = append(keys, SigningKey{ID: id, Key: key})
	}

	return keys, nil
}

// stringToSlice is the split of viper's default hook: a string becomes a
// list at each comma, of any element type.
func stringToSlice(
	f reflect.Type,
	t reflect.Type,
	data any,
) (any, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.Slice {
		return data, nil
	}

	raw := data.(string)
	if raw == "" {
		return []string{}, nil
	}

	return strings.Split(raw, ","), nil
}
