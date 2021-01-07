package bb2

import (
	"fmt"
	"github.com/rz1226/kits2"
	"github.com/rz1226/serverkit"
	"net/http"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

/**
理想的api是
type SomeBB struct{
	InsertUser BlackBoard `readme:"插入用户信息日志"`
	GaoDeApi  BlackBoard `readme:"调用高德地图api的日志"`
	Db BlackBoard `readme:"调用数据库错误"`

}
每一个blackboard包含info, err, warn 日志， 一个计数器，一个计时器

bb := SomeBB{}
BBinit(&bb, groupName )

type


bb.Info.Info("xx)  .Err()  .Warn()
bb.Info.Inc()  .IncBy(1)
t := bb.Info.Start()
bb.Info.Ends(t)

有类型保护, 调用的时候不容易出错
降低了使用成本

*/

var allbb *allBB //全局所有的bb都在这里统一管理

func init() {
	allbb = sNewAllBB()
}
func ShowAllBBs() string {
	return allbb.showAll()
}
func ShowGroup(groupName string) string {
	return allbb.show(groupName)
}

//  所有的bb
type allBB struct {
	data map[string][]*BlackBoradKit
	mu   *sync.Mutex
}

func sNewAllBB() *allBB {
	a := &allBB{}
	a.mu = &sync.Mutex{}
	a.data = make(map[string][]*BlackBoradKit)
	return a
}
func (a *allBB) add(bb *BlackBoradKit) {
	a.mu.Lock()
	defer a.mu.Unlock()
	groupName := bb.groupname
	_, ok := a.data[groupName]
	if !ok {
		a.data[groupName] = make([]*BlackBoradKit, 0, 5)
	}
	a.data[groupName] = append(a.data[groupName], bb)
}
func (a *allBB) show(groupName string) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	str := ""
	v, ok := a.data[groupName]
	if !ok {
		return "没找到此分组:" + groupName
	}
	for _, v2 := range v {
		str += v2.show()
	}
	return str
}
func (a *allBB) showAll() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	str := ""
	for _, v := range a.data {
		for _, v2 := range v {
			str += v2.show()
		}
	}
	return str
}

//  监控信息黑板

type BlackBoradKit struct {
	logKit           *kits2.LogKit
	logWarnKit       *kits2.LogKit
	logErrKit        *kits2.LogKit
	logPanicKit      *kits2.LogKit
	timerKit         *kits2.TimerKit
	counterKit       *kits2.CounterKit
	bbStartTime      string
	noPrintToConsole bool
	readme           string
	groupname        string
}

func NewBlockBorad(groupName, readme string) *BlackBoradKit {
	bb := &BlackBoradKit{}
	bb.groupname = groupName
	bb.readme = readme
	bb.bbStartTime = time.Now().Format("2006-01-02 15:04:05")
	bb.initLogKit()
	bb.initCounterKit()
	bb.initTimerKit()
	bb.noPrintToConsole = true //  默认不直接打印信息
	allbb.add(bb)
	return bb
}

//  是否同时打印到标准输出
func (bb *BlackBoradKit) SetNoPrintToConsole(result bool) {
	bb.noPrintToConsole = result
}

//  初始化日志kit
func (bb *BlackBoradKit) initLogKit() {
	bb.logKit = kits2.NewLogKit("Log记录：" + bb.readme + "_info")
	bb.logWarnKit = kits2.NewLogKit("warn记录：" + bb.readme + "_warn")
	bb.logErrKit = kits2.NewLogKit("Err记录：" + bb.readme + "_error")
	bb.logPanicKit = kits2.NewLogKit("Panic致命错误记录：" + bb.readme + "_panic")
}

//  初始化计数器kit
func (bb *BlackBoradKit) initCounterKit() {
	bb.counterKit = kits2.NewCounterKit(bb.readme)
}

//  初始化计时器kit
func (bb *BlackBoradKit) initTimerKit() {
	bb.timerKit = kits2.NewTimerKit(bb.readme)
}

/*----------------------------log--------------------------------*/
func (bb *BlackBoradKit) Log(logs ...interface{}) {
	str := bb.logKit.Put(fmt.Sprint(logs...))
	if !bb.noPrintToConsole {
		fmt.Print(str)
	}
}
func (bb *BlackBoradKit) Warn(logs ...interface{}) {
	str := bb.logWarnKit.Put(fmt.Sprint(logs...))
	if !bb.noPrintToConsole {
		fmt.Print(str)
	}
}
func (bb *BlackBoradKit) Err(logs ...interface{}) {
	str := bb.logErrKit.Put(fmt.Sprint(logs...))
	if !bb.noPrintToConsole {
		fmt.Print(str)
	}
}
func (bb *BlackBoradKit) Panic(logs ...interface{}) {
	str := bb.logPanicKit.Put(fmt.Sprint(logs...))
	if !bb.noPrintToConsole {
		fmt.Print(str)
	}
}

/*---------------------------timer---------------------------------*/
func (bb *BlackBoradKit) Start(name string) *kits2.Tick {
	return bb.timerKit.Start(name)
}
func (bb *BlackBoradKit) End(tick *kits2.Tick) {
	bb.timerKit.End(tick)
}

/*---------------------------counter---------------------------------*/
func (bb *BlackBoradKit) Inc() {
	bb.counterKit.Inc()
}
func (bb *BlackBoradKit) IncBy(num int64) {
	bb.counterKit.IncBy(num)
}

/*--------------------------show---------------------------*/
//  获取监控信息
const lines = "\n\n\n"

func (bb *BlackBoradKit) show() string {
	strStart := "\n\n\n########################"
	str := strStart + bb.groupname + ":" + bb.readme + "(" + bb.readme + "):" + " blackboard info #################### : \n\n\n"

	str += "监控启动时间:" + bb.bbStartTime + "\n"
	str += bb.logKit.Show()

	str += lines
	str += bb.logWarnKit.Show()

	str += lines
	str += bb.logErrKit.Show()

	str += lines
	str += bb.logPanicKit.Show()

	str += lines
	str += bb.counterKit.Show()

	str += lines
	str += bb.timerKit.Show()
	return str
}

func httpShowAll(w http.ResponseWriter, r *http.Request) {
	str := ShowAllBBs()
	fmt.Fprintln(w, str)
}

var startedMonitor int32 = 0

func StartMonitor(port string) {
	if atomic.CompareAndSwapInt32(&startedMonitor, 0, 1) {
		go serverkit.NewSimpleHTTPServer().Add("/", httpShowAll).Start(port)
	} else {
		fmt.Println("no action , bb monitor started")
	}
}

/**
自动初始化这样一个结构体， groupname 自己输入参数， 属性名为bb名，readme为bb说明
type SomeBB struct{
	InsertUser *blackboardkit.BlackBoradKit  `readme:"插入用户信息 "`
	GaoDeApi  *blackboardkit.BlackBoradKit `readme:"调用高德地图api "`
	Db *blackboardkit.BlackBoradKit `readme:"调用数据库 "`

}

*/

func BBinit(dstStruct interface{}, groupName string) {
	currentField := ""
	defer func() {
		if co := recover(); co != nil {
			str := "bbinit error:发生panic, field=" + currentField + ":" + fmt.Sprint(co)
			fmt.Println(str)
			os.Exit(1)
		}
	}()

	v := reflect.ValueOf(dstStruct)
	t := v.Type().Elem()
	switch v.Kind() {
	case reflect.Ptr:
		for i := 0; i < v.Elem().NumField(); i++ {
			fieldName := t.Field(i).Name
			tag := t.Field(i).Tag.Get("readme")
			if tag == "" {
				tag = t.Field(i).Tag.Get("tag")
			}

			vType := t.Field(i).Type

			if fmt.Sprint(vType) == "*blackboardkit.BlackBoradKit" {

				bb := NewBlockBorad(groupName, fieldName+tag)
				v.Elem().Field(i).Set(reflect.ValueOf(bb))
			} else {
				panic("bbinit error: 要初始化的结构体的属性的类型必须是*blackboardkit.BlackBoradKit")
			}
		}

	default:
		panic("bbinit error:要初始化的结构体指针")

	}
}
