package main

import (
	"github.com/rz1226/rzlib/blackboardkit"
	"time"
)

type SomeBB struct {
	InsertUser *blackboardkit.BlackBoradKit `readme:"插入用户信息 "`
	GaoDeApi   *blackboardkit.BlackBoradKit `readme:"调用高德地图api "`
	Db         *blackboardkit.BlackBoradKit `readme:"调用数据库 "`
}

//  每一个blackboard包含info, err, warn 日志， 一个计数器，一个计时器

func main() {

	bb := SomeBB{}

	blackboardkit.BBinit(&bb, "somegroup")

	blackboardkit.StartMonitor("9090")
	for i := 0; i < 5; i++ {
		go add(bb)
	}

	time.Sleep(time.Second * 1000)

}
func add(bb SomeBB) {
	for i := 0; i < 10000000000; i++ {
		t := bb.Db.Start("开始操作db")
		time.Sleep(time.Microsecond)
		bb.Db.Log("这是db日志", i)
		bb.Db.Err("这是db错误日志")
		bb.Db.Inc()
		bb.Db.End(t)

		t2 := bb.InsertUser.Start("开始操作注册")
		time.Sleep(time.Microsecond)
		bb.InsertUser.Log("这是注册用户日志", i)
		bb.InsertUser.Inc()
		bb.InsertUser.End(t2)
	}
}
