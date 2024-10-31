package gozen

import (
	"fmt"
	"testing"
)

func TestDaoGRPC_GetConn(t *testing.T) {
	daoGrpc := &DaoGRPC{}
	daoGrpc.ServerName = "test"
	conn, err := daoGrpc.GetConn()
	if err != nil {
		t.Errorf("get failed:%s", err.Error())
	} else {
		fmt.Printf("conn:%v\n", conn)
		defer conn.Close()
	}
	daoGrpc2 := &DaoGRPC{}
	daoGrpc2.ServerName = "test2"
	conn2, err2 := daoGrpc2.GetConn()
	if err2 != nil {
		t.Errorf("get failed:%s", err2.Error())
	} else {
		fmt.Printf("conn:%v\n", conn2)
		defer conn2.Close()
	}
	if conn == conn2 {
		t.Errorf("conn is replicate")
	}
}

func BenchmarkDaoGRPC_GetConn(b *testing.B) {
	daoGrpc := &DaoGRPC{}
	daoGrpc.ServerName = "test"
	for i := 0; i < b.N; i++ {
		conn, err := daoGrpc.GetConn()
		if err != nil {
			b.Errorf("get failed:%s", err.Error())
		} else {
			fmt.Printf("conn:%v\n", conn)
			conn.Close()
		}
	}
}
