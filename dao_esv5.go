package gozen

import (
	"context"
	"gopkg.in/olivere/elastic.v5"
	"net/http"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"sync"
	"time"
)

type DaoESV5 struct {
	IndexName string
	TypeName  string
	Ctx       context.Context
}

var (
	esv5Client    *elastic.Client
	esv5ClientMux sync.Mutex
)

func (p *DaoESV5) GetConnect() (*elastic.Client, error) {
	config := configESGet()
	if esv5Client == nil {
		esv5ClientMux.Lock()
		defer esv5ClientMux.Unlock()
		if esv5Client == nil {
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
			esv5Client = client
		}
	}
	return esv5Client, nil
}

func (p *DaoESV5) CloseConnect(client *elastic.Client) {
	//esPool.ReturnObject(client)
}

func (p *DaoESV5) Insert(id string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, err = client.Index().Index(p.IndexName).Type(p.TypeName).Id(id).BodyJson(data).Do(ctx)
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameEs, "insert error",
			LogKNameCommonErr, err,
		)
		return err
	}
	return nil
}

func (p *DaoESV5) Update(id string, doc interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, err = client.Update().Index(p.IndexName).Type(p.TypeName).Id(id).
		Doc(doc).
		Do(ctx)
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameEs, "update error", LogKNameCommonErr, err)
		return err
	}
	return nil
}

func (p *DaoESV5) UpdateAppend(id string, name string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, err = client.Update().Index(p.IndexName).Type(p.TypeName).Id(id).
		Script(elastic.NewScriptFile("append-reply").Param("reply", value)).
		Do(ctx)
	SpanErrorFast(span, err)
	if err != nil {
		LogErrorw(LogNameEs, "update append error", LogKNameCommonErr, err)
		return err
	}
	return nil
}
