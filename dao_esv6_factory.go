package gozen

import (
	"github.com/jolestar/go-commons-pool"
	"gopkg.in/olivere/elastic.v6"
	"net/http"
	"time"
)

type DaoESV6Factory struct {
}

func (f *DaoESV6Factory) MakeObject() (*pool.PooledObject, error) {
	client, err := f.MakeClient()
	return pool.NewPooledObject(client), err
}

func (f *DaoESV6Factory) MakeClient() (*elastic.Client, error) {
	config := configESGet()

	if esTransport == nil {
		esTransportMux.Lock()
		defer esTransportMux.Unlock()
		if esTransport == nil {
			esTransport = &http.Transport{
				MaxIdleConnsPerHost: config.TransportMaxIdel,
			}
		}
	}
	clientHttp := &http.Client{
		Transport: esTransport,
		Timeout:   time.Duration(config.Timeout) * time.Millisecond,
	}

	client, err := elastic.NewClient(elastic.SetHttpClient(clientHttp), elastic.SetURL(config.Address...))

	if err != nil {
		// Handle error

		UtilLogErrorf("es connect error :%s,address:%v", err.Error(), config.Address)

		return nil, err
	}
	return client, err
}

func (f *DaoESV6Factory) DestroyObject(object *pool.PooledObject) error {
	//do destroy

	return nil
}

func (f *DaoESV6Factory) ValidateObject(object *pool.PooledObject) bool {
	//do validate
	esClient, ok := object.Object.(*elastic.Client)

	if !ok {
		UtilLogInfo("es pool validate object failed,convert to client failed")
		return false
	}
	if !esClient.IsRunning() {
		UtilLogInfo("es pool validate object failed,convert to client failed")
		return false
	}

	return true
}

func (f *DaoESV6Factory) ActivateObject(object *pool.PooledObject) error {
	//do activate
	return nil
}

func (f *DaoESV6Factory) PassivateObject(object *pool.PooledObject) error {
	//do passivate
	return nil
}
