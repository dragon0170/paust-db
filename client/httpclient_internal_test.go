package client

import (
	"encoding/binary"
	"encoding/json"
	"github.com/paust-team/paust-db/types"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	TestOwnerId   = "ownertest"
	TestQualifier = "testQualifier"
)

func TestHTTPClient_deSerializeKeyObj(t *testing.T) {
	require := require.New(t)
	var timestamp1 uint64 = 1547772882435375000
	var timestamp2 uint64 = 1547772960049177000
	timestampBytes1 := make([]byte, 8)
	timestampBytes2 := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes1, timestamp1)
	binary.BigEndian.PutUint64(timestampBytes2, timestamp2)
	salt := make([]byte, 2)
	binary.BigEndian.PutUint16(salt, 0)
	rowKey1, err := json.Marshal(types.KeyObj{Timestamp: timestampBytes1, Salt: salt})
	require.Nil(err, "json marshal err: %+v", err)
	rowKey2, err := json.Marshal(types.KeyObj{Timestamp: timestampBytes2, Salt: salt})
	require.Nil(err, "json marshal err: %+v", err)

	// MetaDataResObj deserialize
	metaDataObjs, err := json.Marshal([]types.MetaDataObj{{RowKey: rowKey1, OwnerId: TestOwnerId, Qualifier: []byte(TestQualifier)}, {RowKey: rowKey2, OwnerId: TestOwnerId, Qualifier: []byte(TestQualifier)}})
	require.Nil(err, "json marshal err: %+v", err)
	outputQueryObjs, err := json.Marshal([]OutputQueryObj{{Id: rowKey1, Timestamp: timestamp1, OwnerId: TestOwnerId, Qualifier: TestQualifier}, {Id: rowKey2, Timestamp: timestamp2, OwnerId: TestOwnerId, Qualifier: TestQualifier}})
	require.Nil(err, "json marshal err: %+v", err)

	deserializedBytes, err := deSerializeKeyObj(metaDataObjs, true)
	require.Nil(err, "SerializeKeyObj err: %+v", err)

	require.EqualValues(outputQueryObjs, deserializedBytes)

	// RealDataResObj deserialize
	realDataObjs, err := json.Marshal([]types.RealDataObj{{RowKey: rowKey1, Data: []byte("testData1")}, {RowKey: rowKey2, Data: []byte("testData2")}})
	require.Nil(err, "json marshal err: %+v", err)
	outputFetchObjs, err := json.Marshal([]OutputFetchObj{{Id: rowKey1, Timestamp: timestamp1, Data: []byte("testData1")}, {Id: rowKey2, Timestamp: timestamp2, Data: []byte("testData2")}})
	require.Nil(err, "json marshal err: %+v", err)

	deserializedBytes, err = deSerializeKeyObj(realDataObjs, false)
	require.Nil(err, "SerializeKeyObj err: %+v", err)

	require.EqualValues(outputFetchObjs, deserializedBytes)
}
