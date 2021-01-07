package bb2

//默认bb，方便使用，预先定义了集中比较常用的场景
type defaultbb struct {
	Mongo *BlackBoradKit `readme:"mongo操作日志"`
	API   *BlackBoradKit `readme:"对外接口日志"`
	Db    *BlackBoradKit `readme:"Mysql数据库操作日志"`
	Buz   *BlackBoradKit `readme:"业务日志"`
}

var DefaultBB = defaultbb{}

func init() {
	BBinit(&DefaultBB, "default bb")
}
