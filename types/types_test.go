package types_test

import (
	"encoding/base64"
	"github.com/paust-team/paust-db/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataToRowKey(t *testing.T) {
	//given
	pubKeyBytes, err := base64.StdEncoding.DecodeString("oimd8ZdzgUHzF9CPChJU8gb89VaMYg+1SpX6WT8nQHE=")
	assert.Nil(t, err)

	givenData := types.RealData{Timestamp: 1545982882435375000, UserKey: pubKeyBytes, Qualifier: "Memory", Data: []byte("doNotUse")}
	expectRowKey := []byte{0x15, 0x74, 0x6f, 0x3d, 0x98, 0x65, 0x1f, 0x98, 0xa2, 0x29, 0x9d, 0xf1, 0x97, 0x73, 0x81,
		0x41, 0xf3, 0x17, 0xd0, 0x8f, 0xa, 0x12, 0x54, 0xf2, 0x6, 0xfc, 0xf5, 0x56, 0x8c, 0x62, 0xf, 0xb5, 0x4a, 0x95,
		0xfa, 0x59, 0x3f, 0x27, 0x40, 0x71, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	//when
	actualRowKey := types.DataToRowKey(givenData)

	//then
	assert.Equal(t, expectRowKey, actualRowKey)
}

func TestRowKeyAndValueToData(t *testing.T) {
	//given
	pubKeyBytes, err := base64.StdEncoding.DecodeString("oimd8ZdzgUHzF9CPChJU8gb89VaMYg+1SpX6WT8nQHE=")
	assert.Nil(t, err)

	givenRowKey := []byte{0x15, 0x74, 0x6f, 0x3d, 0x98, 0x65, 0x1f, 0x98, 0xa2, 0x29, 0x9d, 0xf1, 0x97, 0x73, 0x81,
		0x41, 0xf3, 0x17, 0xd0, 0x8f, 0xa, 0x12, 0x54, 0xf2, 0x6, 0xfc, 0xf5, 0x56, 0x8c, 0x62, 0xf, 0xb5, 0x4a, 0x95,
		0xfa, 0x59, 0x3f, 0x27, 0x40, 0x71, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	givenValue := []byte{0x10, 0xff}
	expectData := types.RealData{Timestamp: 1545982882435375000, UserKey: pubKeyBytes, Qualifier: "Memory", Data: []byte{0x10, 0xff}}

	//when
	actualData := types.RowKeyAndValueToData(givenRowKey, givenValue)

	//then
	assert.Equal(t, expectData, actualData)
}

func TestQualifierToByteArr(t *testing.T) {
	//given
	givenQualifier := "Computer"
	expectValue := []byte{0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x72, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0}

	//when
	actualVal := types.QualifierToByteArr(givenQualifier)

	//then
	assert.Equal(t, expectValue, actualVal)
}

func TestQualifierWithoutPadding(t *testing.T) {
	//given
	givenKeySlice := []byte{0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x72, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0}
	expectValue := []byte{0x43, 0x6f, 0x6d, 0x70, 0x75, 0x74, 0x65, 0x72}

	//when
	actualValue := types.QualifierWithoutPadding(givenKeySlice)

	//then
	assert.Equal(t, expectValue, actualValue)
}

func TestMetaDataAndKeyToMetaResponse(t *testing.T) {
	//given
	pubKeyBytes, err := base64.StdEncoding.DecodeString("oimd8ZdzgUHzF9CPChJU8gb89VaMYg+1SpX6WT8nQHE=")
	assert.Nil(t, err)

	givenMetaData := types.MetaData{UserKey: pubKeyBytes, Qualifier: "test"}
	givenKey := []byte{0x15, 0x74, 0x6f, 0x3d, 0x98, 0x65, 0x1f, 0x98, 0xa2, 0x29, 0x9d, 0xf1, 0x97, 0x73, 0x81,
		0x41, 0xf3, 0x17, 0xd0, 0x8f, 0xa, 0x12, 0x54, 0xf2, 0x6, 0xfc, 0xf5, 0x56, 0x8c, 0x62, 0xf, 0xb5, 0x4a, 0x95,
		0xfa, 0x59, 0x3f, 0x27, 0x40, 0x71, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	expectMetaResponse := types.MetaResponse{Timestamp: 1545982882435375000, UserKey: pubKeyBytes, Qualifier: "test"}

	//when
	actualMetaResponse, err := types.MetaDataAndKeyToMetaResponse(givenKey, givenMetaData)
	assert.Nil(t, err)

	//then
	assert.Equal(t, expectMetaResponse, actualMetaResponse)
}
