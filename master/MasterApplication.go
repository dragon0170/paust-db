package master

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/paust-team/paust-db/consts"
	"github.com/paust-team/paust-db/libs/db"
	"github.com/paust-team/paust-db/libs/log"
	"github.com/paust-team/paust-db/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/abci/example/code"
	abciTypes "github.com/tendermint/tendermint/abci/types"
	"math/rand"
	"os"
	"strings"
)

type MasterApplication struct {
	abciTypes.BaseApplication

	hash   []byte
	serial bool
	db     *db.CRocksDB
	wb     db.Batch
	mwb    db.Batch

	logger log.Logger
}

func NewMasterApplication(serial bool, dir string, option log.Option) (*MasterApplication, error) {
	hash := make([]byte, 8)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "make directory failed")
	}
	database, err := db.NewCRocksDB(consts.DBName, dir)
	if err != nil {
		return nil, errors.Wrap(err, "NewCRocksDB err")
	}

	binary.BigEndian.PutUint64(hash, rand.Uint64())
	return &MasterApplication{
		serial: serial,
		hash:   hash,
		db:     database,
		logger: log.NewFilter(log.NewPDBLogger(log.NewSyncWriter(os.Stdout)), option),
	}, nil
}

func (app *MasterApplication) Info(req abciTypes.RequestInfo) abciTypes.ResponseInfo {
	return abciTypes.ResponseInfo{
		Data: fmt.Sprintf("---- Info"),
	}
}

func (app *MasterApplication) CheckTx(tx []byte) abciTypes.ResponseCheckTx {
	var baseDataObjs []types.BaseDataObj
	err := json.Unmarshal(tx, &baseDataObjs)
	if err != nil {
		return abciTypes.ResponseCheckTx{Code: code.CodeTypeEncodingError, Log: err.Error()}
	}

	return abciTypes.ResponseCheckTx{Code: code.CodeTypeOK}
}

func (app *MasterApplication) InitChain(req abciTypes.RequestInitChain) abciTypes.ResponseInitChain {
	app.wb = app.db.NewBatch()
	app.mwb = app.db.NewBatch()

	return abciTypes.ResponseInitChain{}
}

func (app *MasterApplication) BeginBlock(req abciTypes.RequestBeginBlock) abciTypes.ResponseBeginBlock {
	return abciTypes.ResponseBeginBlock{}
}

func (app *MasterApplication) DeliverTx(tx []byte) abciTypes.ResponseDeliverTx {
	//Unmarshal tx to baseDataObjs
	var baseDataObjs []types.BaseDataObj
	if err := json.Unmarshal(tx, &baseDataObjs); err != nil {
		app.logger.Error("Error unmarshaling BaseDataObj", "state", "DeliverTx", "err", err)
		return abciTypes.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: err.Error()}
	}

	//meta와 real 나누어 batch에 담는다
	for i := 0; i < len(baseDataObjs); i++ {
		var metaValue struct {
			OwnerId   string `json:"ownerId"`
			Qualifier []byte `json:"qualifier"`
		}
		metaValue.OwnerId = baseDataObjs[i].MetaData.OwnerId
		metaValue.Qualifier = baseDataObjs[i].MetaData.Qualifier

		metaData, err := json.Marshal(metaValue)
		if err != nil {
			app.logger.Error("Error marshaling metaValue", "state", "DeliverTx", "err", err)
			return abciTypes.ResponseDeliverTx{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}
		app.mwb.SetColumnFamily(app.db.ColumnFamilyHandles()[consts.MetaCFNum], baseDataObjs[i].MetaData.RowKey, metaData)
		app.wb.SetColumnFamily(app.db.ColumnFamilyHandles()[consts.RealCFNum], baseDataObjs[i].RealData.RowKey, baseDataObjs[i].RealData.Data)
	}

	app.logger.Info("Put success", "state", "DeliverTx", "size", len(baseDataObjs), "tx", tx)
	return abciTypes.ResponseDeliverTx{Code: code.CodeTypeOK}
}

func (app *MasterApplication) EndBlock(req abciTypes.RequestEndBlock) abciTypes.ResponseEndBlock {
	return abciTypes.ResponseEndBlock{}
}

func (app *MasterApplication) Commit() (resp abciTypes.ResponseCommit) {
	//resp.Data = app.hash
	count, err := app.mwb.Write()
	if err != nil {
		app.logger.Error("Error writing batch", "state", "Commit", "err", err)
		return
	} else if count > 0 {
		app.logger.Info("Flush metadata", "state", "Commit", "size", count)
	}

	count, err = app.wb.Write()
	if err != nil {
		app.logger.Error("Error writing batch", "state", "Commit", "err", err)
		return
	} else if count > 0 {
		app.logger.Info("Flush realdata", "state", "Commit", "size", count)
	}

	app.mwb = app.db.NewBatch()
	app.wb = app.db.NewBatch()

	return
}

func (app *MasterApplication) Query(reqQuery abciTypes.RequestQuery) abciTypes.ResponseQuery {
	var responseValue []byte
	switch reqQuery.Path {
	case consts.QueryPath:
		var queryObj = types.QueryObj{}
		if err := json.Unmarshal(reqQuery.Data, &queryObj); err != nil {
			app.logger.Error("Error unmarshaling QueryObj", "state", "Query", "err", err)
			return abciTypes.ResponseQuery{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}

		if queryObj.Start >= queryObj.End {
			err := errors.New("query end must be greater than start ")
			return abciTypes.ResponseQuery{Code: code.CodeTypeUnknownError, Log: err.Error()}
		}

		metaDataObjs, err := app.metaDataQuery(queryObj)
		if err != nil {
			app.logger.Error("Error processing queryObj", "state", "Query", "err", err)
			return abciTypes.ResponseQuery{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}
		responseValue, err = json.Marshal(metaDataObjs)
		if err != nil {
			app.logger.Error("Error marshaling metaDataObj", "state", "Query", "err", err)
			return abciTypes.ResponseQuery{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}
		app.logger.Info("Query success", "state", "Query", "path", reqQuery.Path, "data", reqQuery.Data)

	case consts.FetchPath:
		var fetchObj = types.FetchObj{}
		if err := json.Unmarshal(reqQuery.Data, &fetchObj); err != nil {
			app.logger.Error("Error unmarshaling FetchObj", "state", "Query", "err", err)
			return abciTypes.ResponseQuery{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}

		realDataObjs, err := app.realDataFetch(fetchObj)
		if err != nil {
			app.logger.Error("Error processing fetchObj", "state", "Query", "err", err)
			return abciTypes.ResponseQuery{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}
		responseValue, err = json.Marshal(realDataObjs)
		if err != nil {
			app.logger.Error("Error marshaling realDataObj", "state", "Query", "err", err)
			return abciTypes.ResponseQuery{Code: code.CodeTypeEncodingError, Log: err.Error()}
		}
		app.logger.Info("Fetch success", "state", "Query", "path", reqQuery.Path, "data", reqQuery.Data)

	}

	return abciTypes.ResponseQuery{Code: code.CodeTypeOK, Value: responseValue}
}

func (app *MasterApplication) metaDataQuery(queryObj types.QueryObj) ([]types.MetaDataObj, error) {
	var rawMetaDataObjs []types.MetaDataObj
	var metaDataObjs []types.MetaDataObj

	// query field nil error 처리
	if queryObj.Qualifier == nil {
		return nil, errors.Errorf("Qualifier must not be nil")
	}

	if len(queryObj.OwnerId) > consts.OwnerIdLenLimit {
		return nil, errors.Errorf("OwnerId must be %v or below", consts.OwnerIdLenLimit)
	}

	// create start and end for iterator
	salt := uint16(0)

	startByte := types.GetRowKey(queryObj.Start, salt)
	endByte := types.GetRowKey(queryObj.End, salt)

	itr := app.db.IteratorColumnFamily(startByte, endByte, app.db.ColumnFamilyHandles()[consts.MetaCFNum])
	defer itr.Close()

	// time range에 해당하는 모든 데이터를 가져온다
	for itr.Seek(startByte); itr.Valid() && bytes.Compare(itr.Key(), endByte) == -1; itr.Next() {
		var metaObj = types.MetaDataObj{}

		var metaValue struct {
			OwnerId   string `json:"ownerId"`
			Qualifier []byte `json:"qualifier"`
		}
		if err := json.Unmarshal(itr.Value(), &metaValue); err != nil {
			return nil, errors.Wrap(err, "metaValue unmarshal err")
		}

		metaObj.RowKey = make([]byte, len(itr.Key()))
		copy(metaObj.RowKey, itr.Key())
		metaObj.OwnerId = metaValue.OwnerId
		metaObj.Qualifier = metaValue.Qualifier

		rawMetaDataObjs = append(rawMetaDataObjs, metaObj)

	}

	// 가져온 데이터를 제한사항에 맞게 거른다
	switch {
	case strings.Compare(queryObj.OwnerId, "") == 0 && len(queryObj.Qualifier) == 0:
		metaDataObjs = rawMetaDataObjs
	case strings.Compare(queryObj.OwnerId, "") == 0:
		for i, metaObj := range rawMetaDataObjs {
			if bytes.Compare(metaObj.Qualifier, queryObj.Qualifier) == 0 {
				metaDataObjs = append(metaDataObjs, rawMetaDataObjs[i])
			}
		}
	case len(queryObj.Qualifier) == 0:
		for i, metaObj := range rawMetaDataObjs {
			if strings.Compare(metaObj.OwnerId, queryObj.OwnerId) == 0 {
				metaDataObjs = append(metaDataObjs, rawMetaDataObjs[i])
			}
		}
	default:
		for i, metaObj := range rawMetaDataObjs {
			if bytes.Compare(metaObj.Qualifier, queryObj.Qualifier) == 0 && strings.Compare(metaObj.OwnerId, queryObj.OwnerId) == 0 {
				metaDataObjs = append(metaDataObjs, rawMetaDataObjs[i])
			}
		}
	}
	return metaDataObjs, nil

}

func (app *MasterApplication) realDataFetch(fetchObj types.FetchObj) ([]types.RealDataObj, error) {
	var realDataObjs []types.RealDataObj

	for _, rowKey := range fetchObj.RowKeys {
		var realDataObj types.RealDataObj

		realDataObj.RowKey = rowKey
		valueSlice, err := app.db.GetDataFromColumnFamily(consts.RealCFNum, rowKey)
		if err != nil {
			return nil, errors.Wrap(err, "GetDataFromColumnFamily err")
		}
		realDataObj.Data = make([]byte, valueSlice.Size())
		copy(realDataObj.Data, valueSlice.Data())
		realDataObjs = append(realDataObjs, realDataObj)

		valueSlice.Free()

	}

	return realDataObjs, nil
}

func (app *MasterApplication) Destroy() {
	app.db.Close()
}
