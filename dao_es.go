package gozen

import (
	"context"
	"net/http"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"sync"
	"time"

	"gopkg.in/olivere/elastic.v3"
)

type DaoES struct {
	IndexName string
	TypeName  string
	Ctx       context.Context
}

var (
	esClient    *elastic.Client
	esClientMux sync.Mutex
)

func (p *DaoES) GetConnect() (*elastic.Client, error) {
	config := configESGet()
	if esClient == nil {
		esClientMux.Lock()
		defer esClientMux.Unlock()
		if esClient == nil {
			clientHttp := &http.Client{
				Transport: &http.Transport{
					MaxIdleConnsPerHost: config.TransportMaxIdel,
				},
				Timeout: time.Duration(config.Timeout) * time.Millisecond,
			}
			client, err := elastic.NewClient(elastic.SetHttpClient(clientHttp), elastic.SetURL(config.Address...))
			if err != nil {
				LogErrorw(LogNameNet, "es connect error",
					LogKNameCommonErr, err,
					LogKNameCommonData, config,
				)
				return nil, err
			}
			esClient = client
		}
	}
	return esClient, nil
}

func (p *DaoES) CloseConnect(client *elastic.Client) {
	//esPool.ReturnObject(client)
}

func (p *DaoES) Insert(id string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	_, errRes := client.Index().Index(p.IndexName).Type(p.TypeName).Id(id).BodyJson(data).Do()
	SpanErrorFast(span, errRes)
	if errRes != nil {
		UtilLogErrorf("insert error :%s", errRes.Error())
		return errRes
	}
	return nil
}

func (p *DaoES) Update(id string, doc interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	_, errRes := client.Update().Index(p.IndexName).Type(p.TypeName).Id(id).
		Doc(doc).
		Do()
	SpanErrorFast(span, errRes)
	if errRes != nil {
		UtilLogErrorf("daoES Update error :%s", errRes.Error())
		return errRes
	}
	return nil
}

func (p *DaoES) UpdateAppend(id string, name string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	_, errRes := client.Update().Index(p.IndexName).Type(p.TypeName).Id(id).
		Script(elastic.NewScriptFile("append-reply").Param("reply", value)).
		Do()
	SpanErrorFast(span, errRes)
	if errRes != nil {
		UtilLogErrorf("daoES Update error :%s", errRes.Error())
		return err
	}
	return nil
}
