package gozen

import (
	"context"
	"log"
	"net/http"
	"os"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"sync"
	"time"

	"gopkg.in/olivere/elastic.v6"
)

type DaoESV6 struct {
	IndexName string
	TypeName  string
	Ctx       context.Context
}

var (
	esv6Client    *elastic.Client
	esv6ClientMux sync.Mutex
)

func (p *DaoESV6) GetConnect() (*elastic.Client, error) {
	config := configESGet()
	if esv6Client == nil {
		esv6ClientMux.Lock()
		defer esv6ClientMux.Unlock()
		if esv6Client == nil {
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
			client, err := elastic.NewClient(options...)
			if err != nil {
				LogErrorw(LogNameEs, "es connect error",
					LogKNameCommonData, config,
					LogKNameCommonErr, err,
				)
				return nil, err
			}
			esv6Client = client
		}
	}
	return esv6Client, nil
}

func (p *DaoESV6) CloseConnect(client *elastic.Client) {
}

func (p *DaoESV6) Insert(id string, data interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.Index().Index(p.IndexName).Type(p.TypeName).Id(id).BodyJson(data).Do(ctx)
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

func (p *DaoESV6) Update(id string, doc interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	defer p.CloseConnect(client)
	ctx := context.Background()
	_, errRes := client.Update().Index(p.IndexName).Type(p.TypeName).Id(id).
		Doc(doc).
		Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "Update DaoESV6 Update error",
			LogKNameCommonId, id,
			LogKNameCommonData, doc,
			LogKNameCommonErr, errRes,
		)
		return errRes
	}
	return nil
}

func (p *DaoESV6) UpdateAppend(id string, name string, value interface{}) error {
	span, _ := ExitSpan(p.Ctx, SpanDaoEs, RunFuncNameUp(), v3.SpanLayer_Database)
	defer SpanEnd(span)
	client, err := p.GetConnect()
	SpanErrorFast(span, err)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, errRes := client.Update().Index(p.IndexName).Type(p.TypeName).Id(id).
		Script(elastic.NewScriptStored("append-reply").Param("reply", value)).
		Do(ctx)
	SpanErrorFast(span, errRes)
	if errRes != nil {
		LogErrorw(LogNameEs, "UpdateAppend DaoESV6 Update error",
			LogKNameCommonId, id,
			LogKNameCommonErr, errRes,
		)
		return err
	}
	return nil
}
