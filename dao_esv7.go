package gozen

import (
	"context"
	"log"
	"net/http"
	"os"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
)

type DaoESV7 struct {
	IndexName string
	Ctx       context.Context
}

var (
	esv7Client    *elastic.Client
	esv7ClientMux sync.Mutex
)

func (p *DaoESV7) GetConnect() (*elastic.Client, error) {
	config := configESGet()
	if esv7Client == nil {
		esv7ClientMux.Lock()
		defer esv7ClientMux.Unlock()
		if esv7Client == nil {
			clientHttp := &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: config.TransportMaxIdel,
				},
				Timeout: time.Duration(config.Timeout) * time.Millisecond,
			}
			options := []elastic.ClientOptionFunc{
				elastic.SetHttpClient(clientHttp),
				elastic.SetURL(config.Address...),
				elastic.SetHealthcheckInterval(time.Duration(config.HealthcheckInterval) * time.Second),
				elastic.SetHealthcheckTimeout(time.Duration(config.HealthcheckTimeout) * time.Second),
				elastic.SetSniff(config.SnifferEnabled),
				elastic.SetHealthcheck(config.HealthcheckEnabled),
			}
			if ConfigEnvIsDev() {
				// 开发环境显示es查询日志
				options = append(options, elastic.SetTraceLog(log.New(os.Stderr, "[[ELASTIC]]", 0)))
			}
			// 添加用户名和密码
			if config.Password != "" && config.Account != "" {
				options = append(options, elastic.SetBasicAuth(config.Account, config.Password))
			}
			client, err := elastic.NewClient(options...)
			if err != nil {
				LogErrorw(LogNameEs, "es connect error",
					LogKNameCommonData, config,
					LogKNameCommonErr, err,
				)
				return nil, err
			}
			esv7Client = client
		}
	}
	return esv7Client, nil
}

func (p *DaoESV7) CloseConnect(client *elastic.Client) {
}

func (p *DaoESV7) SetIndexName(indexName string) {
	p.IndexName = indexName
}

func (p *DaoESV7) NewMatchAllQuery() elastic.Query {
	return elastic.NewMatchAllQuery()
}

func (p *DaoESV7) NewIdsQuery() *elastic.IdsQuery {
	return elastic.NewIdsQuery()
}

func (p *DaoESV7) NewQueryStringQuery(str string) *elastic.QueryStringQuery {
	return elastic.NewQueryStringQuery(str)
}

func (p *DaoESV7) CreateIndex(indexName string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.CreateIndex(indexName).BodyJson(data).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "CreateIndex error",
			LogKNameCommonData, data,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) DelIndex(indexName string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.Delete().Index(indexName).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "DelIndex error",
			LogKNameCommonData, indexName,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) IndexPutSettings(indexName string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.IndexPutSettings(indexName).BodyJson(data).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "IndexPutSettings error",
			LogKNameCommonData, data,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) AddIndexField(indexName string, data map[string]interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.PutMapping().Index(indexName).BodyJson(data).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "AddIndexField error",
			LogKNameCommonData, data,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) Insert(id string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.Index().Index(p.IndexName).Id(id).BodyJson(data).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "insert error",
			LogKNameCommonId, id,
			LogKNameCommonData, data,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) Bulk(upsertData map[string]interface{}, deleteData []string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	var requests []elastic.BulkableRequest
	for _, v := range deleteData {
		index1Req := elastic.NewBulkDeleteRequest().Index(p.IndexName).Id(v)
		requests = append(requests, index1Req)
	}
	for k, v := range upsertData {
		index1Req := elastic.NewBulkUpdateRequest().Index(p.IndexName).Id(k).Upsert(v).RetryOnConflict(3).Doc(v)
		requests = append(requests, index1Req)
	}
	_, errRes := client.Bulk().Index(p.IndexName).Add(requests...).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "Bulk error",
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) BulkDelete(idS []string) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	var requests []elastic.BulkableRequest
	for _, v := range idS {
		index1Req := elastic.NewBulkDeleteRequest().Index(p.IndexName).Id(v)
		requests = append(requests, index1Req)
	}
	_, errRes := client.Bulk().Index(p.IndexName).Add(requests...).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "BulkDelete error",
			LogKNameCommonId, idS,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) BulkUpsert(idS []string, dataS []interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	var requests []elastic.BulkableRequest
	for k, v := range idS {
		index1Req := elastic.NewBulkUpdateRequest().Index(p.IndexName).Id(v).
			Upsert(dataS[k]).RetryOnConflict(3).Doc(dataS[k])
		requests = append(requests, index1Req)
	}
	_, errRes := client.Bulk().Index(p.IndexName).Add(requests...).Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "BulkUpsert error",
			LogKNameCommonId, idS,
			LogKNameCommonData, dataS,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) Update(id string, doc interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.Update().Index(p.IndexName).Id(id).
		Doc(doc).
		Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "Update DaoESV7 Update error",
			LogKNameCommonId, id,
			LogKNameCommonData, doc,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) Upsert(id string, doc interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.Update().Index(p.IndexName).Id(id).
		Upsert(doc).
		Doc(doc).
		Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "Upsert DaoESV7 Update error",
			LogKNameCommonId, id,
			LogKNameCommonData, doc,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) UpsertUseScript(id string, doc interface{}, script *elastic.Script) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()

	_, errRes := client.Update().Index(p.IndexName).Id(id).
		Script(script).
		Upsert(doc).
		Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "DaoESV7 UpsertUseScript error",
			LogKNameCommonId, id,
			LogKNameCommonData, doc,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV7) UpdateAppend(id string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, errRes := client.Update().Index(p.IndexName).Id(id).
		Script(elastic.NewScriptStored("append-reply").Param("reply", value)).
		Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "UpdateAppend DaoESV7 Update error",
			LogKNameCommonId, id,
			LogKNameCommonErr, errRes,
		)
		return err
	}
	return nil
}
