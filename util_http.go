package gozen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// dynamic 地址

func getDynamicRedisAddress(url string) (err error, address []string) {
	header := []string{"Accept:"}
	var ret []byte
	ret, err = curlGet(url, header)
	if err != nil {
		LogErrorw(LogNameApi, "getDynamicRedisAddress curlGet error",
			LogKNameCommonUrl, url,
			LogKNameCommonErr, err,
		)
		return
	}
	data := new(redisAddressResp)
	err = json.Unmarshal(ret, data)
	if err != nil {
		LogErrorw(LogNameApi, "getDynamicRedisAddress json Unmarshal error",
			LogKNameCommonUrl, url,
			LogKNameCommonErr, err,
			LogKNameCommonRes, string(ret),
		)
		return
	}
	if data.Success {
		return err, data.Content.Result
	} else {
		LogErrorw(LogNameApi, "getDynamicRedisAddress res error",
			LogKNameCommonUrl, url,
			LogKNameCommonErr, err,
			LogKNameCommonRes, data,
		)
	}
	return
}

type redisAddressResp struct {
	Success bool                    `json:"success"`
	Code    string                  `json:"code"`
	Msg     string                  `json:"msg"`
	Content redisAddressRespContent `json:"content"`
}

type redisAddressRespContent struct {
	Result []string `json:"result"`
}

// no proxy 地址 获取

func getNoProxyRedisAddress(reqUrl string, appCode string, bid string, iv string) (info noProxyLocalClusterInfo, err error) {
	header := []string{"Accept:"}
	var ret []byte
	reqUrl = fmt.Sprintf(reqUrl+"?bid=%s&appCode=%s", bid, appCode)
	ret, err = curlGet(reqUrl, header)
	if err != nil {
		LogErrorw(LogNameApi, "getNoProxyRedisAddress curlGet error",
			LogKNameCommonUrl, reqUrl,
			LogKNameCommonErr, err,
		)
		return
	}
	data := new(noProxyRedisAddressResp)
	err = json.Unmarshal(ret, data)
	if err != nil {
		LogErrorw(LogNameApi, "getNoProxyRedisAddress json Unmarshal error",
			LogKNameCommonUrl, reqUrl,
			LogKNameCommonErr, err,
			LogKNameCommonRes, string(ret),
		)
		return
	}
	if data.Success {
		//解密
		info, err = decodeClusterInfo(appCode, bid, iv, data.Content.ClusterInfo)
		if err != nil {
			LogErrorw(LogNameApi, "getNoProxyRedisAddress decodeClusterInfo error",
				LogKNameCommonUrl, reqUrl,
				LogKNameCommonErr, err,
				LogKNameCommonRes, string(ret),
			)
			return
		}
		info.AppCode = appCode
		info.Bid = bid

		return
	} else {
		LogErrorw(LogNameApi, "getNoProxyRedisAddress res error",
			LogKNameCommonUrl, reqUrl,
			LogKNameCommonErr, err,
			LogKNameCommonRes, data,
		)
		err = errors.New("getNoProxyRedisAddress res fail")
	}
	return
}

type noProxyRedisAddressResp struct {
	Success bool                           `json:"success"`
	Code    string                         `json:"code"`
	Msg     string                         `json:"msg"`
	Content noProxyRedisAddressRespContent `json:"content"`
}

type noProxyRedisAddressRespContent struct {
	Bid         string `json:"bid"`
	ClusterInfo string `json:"clusterInfo"` //加密信息
}

/**

{
	"clusterName":"平台集群",
	//密码
	"clusterPassWord": null,
	// 0集群协议   1主从协议
	"clusterProtocol": 0,
	//是否需要密码 0不需要 1需要
	"needAuth": 0,
	//集群node列表 逗号分隔
	"nodeList": "127.0.0.1:8001,127.0.0.1:8101,127.0.0.1:8101,127.0.0.1:8001,127.0.0.1:8001,127.0.0.1:8101"
}

*/

type noProxyClusterInfo struct {
	ClusterName     string `json:"clusterName"`
	ClusterPassWord string `json:"clusterPassWord"`
	ClusterProtocol int    `json:"clusterProtocol"` // 0集群协议   1主从协议
	NeedAuth        int    `json:"needAuth"`        // 是否需要密码 0不需要 1需要
	NodeList        string `json:"nodeList"`
}

// noProxyLocalClusterInfo 本地保存redis信息
type noProxyLocalClusterInfo struct {
	AppCode         string   `json:"app_code"`
	Bid             string   `json:"bid"`
	ClusterInfo     string   `json:"cluster_info"` //加密信息
	ClusterName     string   `json:"cluster_name"`
	ClusterPassWord string   `json:"cluster_password"`
	IsCluster       bool     `json:"is_cluster"`
	NeedAuth        bool     `json:"need_auth"`
	Address         []string `json:"address"`
	AddressLen      int      `json:"address_len"`
}

// no proxy 信息上报 心跳

type noProxyHeartInfo struct {
	Lang          string                        `json:"lang,omitempty"`
	AppCode       string                        `json:"appCode"`
	ClientIp      string                        `json:"clientIp,omitempty"`
	ClientPort    string                        `json:"clientPort,omitempty"`
	ClientVersion string                        `json:"clientVersion,omitempty"`
	ClusterInfos  []noProxyHeartInfoClusterInfo `json:"clusterInfos"`
}

type noProxyHeartInfoClusterInfo struct {
	Bid         string `json:"bid"`
	ExecCount   int64  `json:"execCount"`
	ClusterInfo string `json:"clusterInfo,omitempty"` //加密信息
}

func sendNoProxyRedisHeart(reqUrl string, appCode string, bid string, iv string, info noProxyHeartInfo) (isNew bool, newInfo noProxyLocalClusterInfo, err error) {
	header := []string{"Content-Type:application/json"}
	var ret []byte
	req, _ := json.Marshal(info)
	if ConfigEnvIsDebug() {
		log.Printf("redis 上报心跳 sendNoProxyRedisHeart %v", info)

		r, _ := json.Marshal(info)
		fmt.Println(string(r))
	}
	ret, err = CurlPost(context.Background(), reqUrl, header, string(req), 10*time.Second)
	if err != nil {
		LogErrorw(LogNameApi, "sendNoProxyRedisHeart CurlPost error",
			LogKNameCommonUrl, reqUrl,
			LogKNameCommonErr, err,
		)
		return
	}
	data := new(sendNoProxyRedisHeartResp)
	err = json.Unmarshal(ret, data)
	if err != nil {
		LogErrorw(LogNameApi, "sendNoProxyRedisHeart json Unmarshal error",
			LogKNameCommonUrl, reqUrl,
			LogKNameCommonErr, err,
			LogKNameCommonRes, string(ret),
		)
		return
	}
	if ConfigEnvIsDebug() {
		log.Printf("redis 上报心跳结果 sendNoProxyRedisHeart %v", string(ret))
	}
	if data.Success {
		//根据响应 来判断更新本地no proxy info
		if data.Content.EventType == "0" {
			return
		} else if data.Content.EventType == "1" {
			// 解析内容
			for _, v := range data.Content.NewClusterInfos {
				if v.Bid == bid {
					//解密
					isNew = true
					newInfo, err = decodeClusterInfo(appCode, bid, iv, v.ClusterInfo)
					if err != nil {
						LogErrorw(LogNameApi, "sendNoProxyRedisHeart decodeClusterInfo error",
							LogKNameCommonUrl, reqUrl,
							LogKNameCommonErr, err,
							LogKNameCommonRes, data,
						)
						return
					}
					newInfo.AppCode = appCode
					newInfo.Bid = bid
					return
				}
			}
		}

		return
	} else {
		LogErrorw(LogNameApi, "sendNoProxyRedisHeart res error",
			LogKNameCommonUrl, reqUrl,
			LogKNameCommonErr, err,
			LogKNameCommonRes, data,
		)
		err = errors.New("sendNoProxyRedisHeart res fail")
	}
	return
}

type sendNoProxyRedisHeartResp struct {
	Success bool                             `json:"success"`
	Code    string                           `json:"code"`
	Msg     string                           `json:"msg"`
	Content sendNoProxyRedisHeartRespContent `json:"content"`
}

type sendNoProxyRedisHeartRespContent struct {
	//0:无事件,正常心跳回包
	//1：开启集群切换,返回变更的bid集群信息,需要重新初始化集群信息
	EventType       string                           `json:"eventType"`
	NewClusterInfos []noProxyRedisAddressRespContent `json:"newClusterInfos"`
}

// common

func curlGet(url string, header []string) (ret []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ret, err
	}
	for _, v := range header {
		t := strings.Split(v, ":")
		length := len(t)
		if length == 2 {
			req.Header.Add(t[0], t[1])
		} else if length == 1 {
			req.Header.Add(t[0], "")
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	ret, err = ioutil.ReadAll(resp.Body)

	return ret, err
}

func decodeClusterInfo(appCode string, bid string, iv, origin string) (info noProxyLocalClusterInfo, err error) {
	//解密
	key := UtilCryptoMd5Lower(appCode + ":" + bid)
	var clusterInfoRet []byte
	var clusterInfo noProxyClusterInfo
	clusterInfoRet, err = AesCbcDecryptBase64(key, iv, origin)
	if err != nil {
		LogErrorw(LogNameApi, "decodeClusterInfo AesCbcDecryptBase64 error",
			LogKNameCommonErr, err,
		)
		return
	}
	err = json.Unmarshal(clusterInfoRet, &clusterInfo)
	if err != nil {
		LogErrorw(LogNameApi, "decodeClusterInfo json.Unmarshal error",
			LogKNameCommonErr, err,
		)
		return
	}

	if clusterInfo.NeedAuth == 1 {
		info.NeedAuth = true
	}

	if clusterInfo.ClusterProtocol == 0 {
		info.IsCluster = true
	}
	info.ClusterPassWord = clusterInfo.ClusterPassWord
	info.Address = strings.Split(clusterInfo.NodeList, ",")
	info.AddressLen = len(info.Address)

	info.ClusterName = clusterInfo.ClusterName
	info.ClusterInfo = origin
	return
}
