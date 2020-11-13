package main

import (
	"fmt"
	"log"
	"pipeline/pipeline"
	"time"
)

// Lesson From https://segmentfault.com/a/1190000014788594
// modify/learn from https://github.com/jamesblockk/Start-Go/blob/6bf886f9c72998de2427f3797be2c9c18f5a6b4a/patten/pipeline/main.go
/*
	流水线：
	* 通过channel容量来控制某工序的并发数
	* 流水线各工序的数据传递直接通过全局变量操作【存在数据竞争？】
*/

func loadCheckpoint() int {
	return 1
}

func extractReviewsFromA(check *int, num int) ([]int, error) {
	fmt.Println("extractReviewsFromA")

	return []int{1, 1123}, nil
}

func jobA(datas []int) error {
	datas[0]++
	//fmt.Println("jobA", datas)
	return nil
}

func jobB(datas []int) error {
	datas[0]++
	//fmt.Println("jobB", datas)
	return nil
}

func jobC(datas []int) error {
	datas[0]++
	//fmt.Println("jobC", datas)
	return nil
}

func saveCheckpoint(checkpoint int) error {
	// fmt.Println("saveCheckpoint", checkpoint)
	return nil
}

func main() {
	t1 := time.Now() //   測試運行時間用的

	i := 0
	for i < 1 {
		startPipeLine()
		i++
	}
	elapsed := time.Since(t1)    //   測試運時間用的
	fmt.Println("time", elapsed) //印出時間
}

func startPipeLine() {
	checkpoint := loadCheckpoint()

	//工序(1)在pipeline外执行，最后一个工序是保存checkpoint
	//pipeline := pipeline.NewPipeline(8, 32, 2, 1)
	pipeline := pipeline.NewPipeline(8, 4, 2, 1)

	//(1)
	//加载100条数据，并修改变量checkpoint
	//data是数组，每个元素是一条评论，之后的联表、NLP都直接修改data里的每条记录。
	data, err := extractReviewsFromA(&checkpoint, 100)
	if err != nil {
		log.Print(err)
	}

	for {
		curCheckpoint := checkpoint

		ok := pipeline.Async(func() error {
			//(2)
			return jobA(data)
		}, func() error {
			//(3)
			return jobB(data)
		}, func() error {
			//(4)
			return jobC(data)
		}, func() error {
			//(5)保存checkpoint
			// log.Print("done:", curCheckpoint)
			return saveCheckpoint(curCheckpoint)
		})

		if !ok {
			break
		}

		if data[0] > 10000 {
			break
		} //done

	}
	err = pipeline.Wait()
	if err != nil {
		log.Print(err)
	}
}
