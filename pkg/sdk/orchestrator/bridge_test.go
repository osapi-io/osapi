package orchestrator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type BridgeTestSuite struct {
	suite.Suite
}

func (s *BridgeTestSuite) TestStructToMapUnmarshalError() {
	original := jsonUnmarshalFn
	defer func() { jsonUnmarshalFn = original }()

	jsonUnmarshalFn = func(
		_ []byte,
		_ any,
	) error {
		return fmt.Errorf("forced unmarshal error")
	}

	result := StructToMap(struct {
		Name string `json:"name"`
	}{Name: "test"})

	s.Nil(result)

	// Restore and verify normal behavior.
	jsonUnmarshalFn = json.Unmarshal

	result = StructToMap(struct {
		Name string `json:"name"`
	}{Name: "test"})
	s.NotNil(result)
	s.Equal("test", result["name"])
}

func TestBridgeTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeTestSuite))
}
