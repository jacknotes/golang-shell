/*
connlog需要每年底创建下一年数据库，此脚本用于生成创建和删除数据库SQL命令
此文件具备AlwaysOn集群新建数据库及全备的脚本生成功能
*/
package main

import (
	"fmt"
	"sort"
)

type Month int

const (
	YEAR   int    = 2023
	Prefix string = "Log"
)

const (
	January Month = iota + 1
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var (
	d               = 1
	Slist           = []string{}
	DATA_DIR        = "D:\\SQLData"
	SIZE            = "8192KB"
	FILEGROWTH      = "65536KB"
	PRIMARY_MAXSIZE = "UNLIMITED"
	LOG_MAXSIZE     = "2048GB"
	BACKUP_DIR      = "D:\\tmp"
)

// 判断是否为润年，润年2月29天，非润年2月28天
func IsRunNian(year int) bool {
	if year > 0 {
		result := year % 4
		if result == 0 {
			return true
		} else {
			return false
		}
	} else {
		panic("[ERROR]: 年份不合法")
	}
}

// 输出大小月日期
func OutputDate() {
	MaxMonthList := [7]Month{January, March, May, July, August, October, December}
	MinMonthList := [4]Month{April, June, September, November}
	for _, v := range MaxMonthList {
		if v < October {
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%s%d", Prefix, YEAR, "0", v, "0", d))
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", v, d+10))
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", v, d+20))
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", v, d+30))
		} else {
			Slist = append(Slist, fmt.Sprintf("%s%d%d%s%d", Prefix, YEAR, v, "0", d))
			Slist = append(Slist, fmt.Sprintf("%s%d%d%d", Prefix, YEAR, v, d+10))
			Slist = append(Slist, fmt.Sprintf("%s%d%d%d", Prefix, YEAR, v, d+20))
			Slist = append(Slist, fmt.Sprintf("%s%d%d%d", Prefix, YEAR, v, d+30))
		}
	}
	for _, v := range MinMonthList {
		if v < November {
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%s%d", Prefix, YEAR, "0", v, "0", d))
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", v, d+10))
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", v, d+20))
			Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", v, d+29))
		} else {
			Slist = append(Slist, fmt.Sprintf("%s%d%d%s%d", Prefix, YEAR, v, "0", d))
			Slist = append(Slist, fmt.Sprintf("%s%d%d%d", Prefix, YEAR, v, d+10))
			Slist = append(Slist, fmt.Sprintf("%s%d%d%d", Prefix, YEAR, v, d+20))
			Slist = append(Slist, fmt.Sprintf("%s%d%d%d", Prefix, YEAR, v, d+29))
		}
	}
}

// 加入2月日期
func GenDate(b bool) {
	if b {
		//润年，2月29天
		OutputDate()
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%s%d", Prefix, YEAR, "0", February, "0", d))
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", February, d+10))
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", February, d+20))
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", February, d+28))
	} else {
		//非润年，2月28天
		OutputDate()
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%s%d", Prefix, YEAR, "0", February, "0", d))
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", February, d+10))
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", February, d+20))
		Slist = append(Slist, fmt.Sprintf("%s%d%s%d%d", Prefix, YEAR, "0", February, d+27))
	}
}

func AlwaysOnClusterCreateDB() {
	// 生成新建数据库脚本
	fmt.Println("---- AlwaysOn集群创建数据库")
	for _, v := range Slist {
		fmt.Printf("-- %s\n", v)
		fmt.Printf("CREATE DATABASE %s\nON PRIMARY\n(\nNAME = N'%s',\nFILENAME = N'%s\\%s.mdf',\nSIZE = %s,\nFILEGROWTH = %s,\nMAXSIZE = %s\n)\nLOG ON\n(\nNAME = N'%s_log',\nFILENAME = N'%s\\%s_log.ldf',\nSIZE = %s,\nFILEGROWTH = %s,\nMAXSIZE = %s\n)\nGO\n",
			v, v, DATA_DIR, v, SIZE, FILEGROWTH, PRIMARY_MAXSIZE, v, DATA_DIR, v, SIZE, FILEGROWTH, LOG_MAXSIZE)
		fmt.Println()
	}

	// 生成完全备份数据库脚本
	fmt.Println("---- AlwaysOn集群完全备份数据库")
	for _, v := range Slist {
		fmt.Printf("-- %s\n", v)
		fmt.Printf("BACKUP DATABASE [%s] TO DISK=N'%s\\%s_full.bak'\nWITH NOFORMAT, NOINIT, NAME=N'%s-完整 数据库 备份', SKIP, NOREWIND, NOUNLOAD, STATS=10\n", v, BACKUP_DIR, v, v)
		fmt.Println()
	}

	// 生成恢复数据库脚本-RECOVERY
	fmt.Println("---- AlwaysOn集群恢复数据库-RECOVERY")
	for _, v := range Slist {
		fmt.Printf("-- %s\n", v)
		fmt.Printf("RESTORE DATABASE [%s]\nFROM\nDISK=N'%s\\%s_full.bak'\nWITH MOVE '%s' TO N'%s\\%s.mdf',\nMOVE '%s_log' TO N'%s\\%s_log.ldf',\nSTATS = 10, REPLACE,RECOVERY\nGO\n",
			v, BACKUP_DIR, v, v, DATA_DIR, v, v, DATA_DIR, v)
		fmt.Println()
	}

	// 生成恢复数据库脚本-NORECOVERY
	fmt.Println("---- AlwaysOn集群恢复数据库-NORECOVERY")
	for _, v := range Slist {
		fmt.Printf("-- %s\n", v)
		fmt.Printf("RESTORE DATABASE [%s]\nFROM\nDISK=N'%s\\%s_full.bak'\nWITH MOVE '%s' TO N'%s\\%s.mdf',\nMOVE '%s_log' TO N'%s\\%s_log.ldf',\nSTATS = 10, REPLACE,NORECOVERY\nGO\n",
			v, BACKUP_DIR, v, v, DATA_DIR, v, v, DATA_DIR, v)
		fmt.Println()
	}

}

func CreateDBStatement() {
	for _, v := range Slist {
		fmt.Printf("CREATE DATABASE %s\n", v)
	}
}

func DropDBStatement() {
	for _, v := range Slist {
		fmt.Printf("DROP DATABASE %s\n", v)
	}
}

func main() {
	result := IsRunNian(YEAR)
	GenDate(result)

	// 对slice进行排序
	sort.Strings(Slist)

	fmt.Println("-------- AlwaysOn集群创建数据库和完全备份语句")
	AlwaysOnClusterCreateDB()

	// fmt.Println("-- 创建数据库语句")
	// CreateDBStatement()

	fmt.Println("-- 删除数据库语句")
	DropDBStatement()
}
