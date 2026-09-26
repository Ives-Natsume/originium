package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TaskStatus int 
const (
	Pending TaskStatus = iota
	Running
	Completed
	Failed
)

func (t TaskStatus) String() string {
	switch t {
	case Pending:
		return "Pending"
	case Running:
		return "Running"
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	default:
		return "Undefined"
	}
}

type TaskType int
const (
	Insert		TaskType = iota
	Delete
	Modify
	Query
	Invalid
)

func (t TaskType) String() string {
    switch t {
    case Insert:
        return "Insert"
    case Delete:
        return "Delete"
    case Modify:
        return "Modify"
    case Query:
        return "Query"
    default:
        return "Invalid"
    }
}

type TaskResult struct {
	Output 		string
	Err 		error
}

type Task struct {
	ID			uint64
	Status 		TaskStatus
	Type 		TaskType
	Result 		TaskResult
	CreatedAt 	time.Time
}

func main() {
	fmt.Println("hello from projects/x-zoot")
	// 后续可以通过对id的编号排序实现执行优先级
	// 或者直接用队列fifo？
	var idList []uint64

	var req1 = "Query"
	var req2 = "Insert"
	var req3 = "Login"

	var reqList = []string{req1, req2, req3}
	var taskList []Task

	for _, req := range reqList {
		task, err := parseRequest(req, idList)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		idList = append(idList, task.ID)
		taskList = append(taskList, task)
	}
	
	for _, task := range taskList {
		err := handleTask(task)
		if err != nil {
			fmt.Println(err.Error())
		}
		printTask(&task)
	}
}

func parseRequest(taskType string, idList []uint64) (Task, error) {
	var parsedType TaskType
	var task Task
	switch taskType {
	case "Insert":
		parsedType = Insert
	case "Delete":
		parsedType = Delete
	case "Modify":
		parsedType = Modify
	case "Query":
		parsedType = Query
	default:
		return task, errors.New("parseRequest: not a valid task type")
	}

	task.CreatedAt = time.Now()
	task.Type = parsedType
	task.Status = Pending

	var lastID = uint64(0)
	if len(idList) != 0 {
		lastID = idList[len(idList) - 1]
	}

	task.ID = lastID + 1
	return task, nil
}

func printTask(task *Task) {
	fmt.Println()
	fmt.Println("Task: ", task.ID)
	fmt.Println(". Created at: ", task.CreatedAt)
	fmt.Println(". Type: ", task.Type)
	fmt.Println(". Status: ", task.Status)
	fmt.Println(". Result: ", task.Result)
	fmt.Println()
}

func handleTask(task Task) error {
	switch task.Type {
	case Query:
		err := queryHandler()
		if err != nil {
			task.Status = Failed
			task.Result.Err = err
			return err
		}
	default:
		fmt.Println("handleTask: Not impled yet")
		task.Status = Completed
	}
	return nil
}

func queryHandler() error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"http://localhost:8080/operators", nil)
	if err != nil {
		return err
	}

	
}

func insertHandler() error {

	return nil
}
