package mysqlClient

import (
	"database/sql"
	"fmt"
	"game/define"
	_ "github.com/go-sql-driver/mysql"
	"github.com/liangdas/mqant/log"
	"sync"
	"systemConf"
	"time"
)


var mysqlDB *sql.DB
var systems map[string]bool
var locker *sync.RWMutex
var gDBName string

func Initialize() {
	if mysqlDB != nil {
		return
	}

	conf := systemConf.SystemConfMgr.MysqlConf
	log.Info("Mysql connect to address:%v; database:%v", conf.Address, conf.Database)
	dbName := conf.Database
	gDBName = dbName

	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", conf.Account, conf.Password, conf.Address)
	mysqlDB, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
		return
	}

	mysqlDB.SetMaxIdleConns(10)
	mysqlDB.SetMaxOpenConns(80)
	mysqlDB.SetConnMaxLifetime(time.Minute * 30)
	if e := mysqlDB.Ping(); e != nil {
		panic(e)
	}
	_, err = mysqlDB.Exec(fmt.Sprintf("CREATE DATABASE if not exists %s DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;", dbName))
	if err != nil {
		panic(err)
		return
	}

	_, err = mysqlDB.Exec(fmt.Sprintf("USE %v", dbName))
	systems = make(map[string]bool)

	locker = new(sync.RWMutex)
	for _, tableName := range define.USER_DATA_NAME{
		CreateTable(tableName)
	}
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	return mysqlDB.Exec(query, args...)
}

func CreateTable(tableName string) {
	_, err := mysqlDB.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (", tableName) +
		"`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '角色id'," +
		"`data` blob COMMENT '数据'," +
		"`submission_date` DATE," +
		"PRIMARY KEY ( `id` )" +
		")ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='玩家数据表';")
	if err != nil {
		panic(err)
	}
	locker.Lock()
	systems[tableName] = true
	locker.Unlock()
}

//保存用户数据
//uid 玩家名字
//systemName 系统名字
//data 需要保存的数据
func SaveUserData(uid int64, systemName string, data []byte) error {
	_, _ = mysqlDB.Exec(fmt.Sprintf("USE %v", gDBName))
	_, err := mysqlDB.Exec(fmt.Sprintf("INSERT INTO `%s` (id,data) VALUES ",systemName)+
		"(?,COMPRESS(?)) ON DUPLICATE KEY UPDATE data=COMPRESS(?);",
		uid,
		data,
		data)
	return err
}

//获取用户数据
//uid 玩家名字
//systemName 系统名字
func GetUserData(uid int64, systemName string) (data []byte, err error) {
	_, _ = mysqlDB.Exec(fmt.Sprintf("USE %v", gDBName))
	row := mysqlDB.QueryRow(fmt.Sprintf("select UNCOMPRESS(data) from `%s` where id=?",systemName), uid)
	if row == nil {
		return nil, fmt.Errorf("Get user data from mysql db Error;uid=%v,systemName=%v", uid, systemName)
	}

	err = row.Scan(&data)
	if err != nil {
		return nil, nil
	}
	log.Info("userid:%d get data from mysql,len=%d",uid,len(data))
	return data, nil
}
