package gozen

import (
	"fmt"
	"testing"
)

var (
	UserNumTopic = "rollbak:a"
)

type Location struct {
	Area string
	City string
}

func Detail(age int, name string, l Location) {
	fmt.Println("我的名字：", name)
	fmt.Println("我的年龄：", age)
	fmt.Println("我的住址：", l.Area)
	fmt.Println("我的城市：", l.City)
}

func Test_EventBus(t *testing.T) {
	RollBus.Subscribe(UserNumTopic, Detail)
	defer RollBus.Unsubscribe(UserNumTopic, Detail)
	l := Location{
		Area: "启迪 方洲",
		City: "南京",
	}
	RollBus.Publish(UserNumTopic, 2, "董再东", l)

	RollBus.Publish(UserNumTopic, 3, "董再东3", l)
}
