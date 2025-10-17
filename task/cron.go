package task

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"
	"trade/middleware"
	"trade/services"
)

var (
	db *sql.DB
)

type Job struct {
	Name           string
	CronExpression string
	FunctionName   string
	Package        string
}

func LoadJobs() ([]Job, error) {
	var jobs []Job

	err := middleware.DB.Table("scheduled_tasks").Where("status = ?", 1).Select("name, cron_expression, function_name, package").Scan(&jobs).Error
	if err != nil {
		log.Fatal("Failed to load tasks:", err)
		return nil, err
	}
	for _, job := range jobs {
		fn := getFunction(job.Package, job.FunctionName)
		fmt.Println(job.FunctionName)
		if fn.IsValid() {
			taskFunc := func() {
				fn.Call(nil)
			}
			RegisterTask(job.Name, taskFunc)

		} else {
			log.Printf("Function %s not found in package %s", job.FunctionName, job.Package)
		}
	}
	return jobs, nil
}

func ExecuteWithLock(taskName string) {
	lockKey := "lock:" + taskName

	var expiration time.Duration
	switch taskName {
	case "PushBoxAsset":
		expiration = 6 * time.Minute
	default:
		expiration = 1 * time.Minute
	}

	identifier, acquired := middleware.AcquireLock(lockKey, expiration)
	if !acquired {
		log.Printf("任务 %s 获取锁失败，可能正在执行中", taskName)
		return
	}
	defer middleware.ReleaseLock(lockKey, identifier)

	ExecuteTask(taskName)
}
func getFunction(pkgName, funcName string) reflect.Value {
	switch pkgName {
	case "services":

		manager := services.CronService{}
		return reflect.ValueOf(&manager).MethodByName(funcName)
	default:
		return reflect.Value{}
	}
}
