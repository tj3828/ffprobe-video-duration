package main

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"


)

func main() {
	defer ElapsedTime()()

	path :=  os.Args[1]
	if path == "" {
		log.Print("파일 경로 입력하세요.")
		return
	}

	readFile, err := excelize.OpenFile(path)
	if err != nil {
		log.Print(err)
		return
	}

	rows, err := readFile.GetRows("Sheet1")
	if err != nil {
		log.Print(err)
		return
	}

	// CPU 모두 사용
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg := sync.WaitGroup{}
	mutex := &sync.Mutex{}
	curPosition := 0;
	maxGoRoutine := 20;
    if maxGoRoutine >= len(rows) - 1 {
    	maxGoRoutine = 1
	}


	for i := 1; i <= maxGoRoutine; i++ {
		goRoutineIndex := (len(rows) - 1) / maxGoRoutine
		max := goRoutineIndex * i + 1
		min := goRoutineIndex * (i - 1) + 2

		if maxGoRoutine == i {
			max = len(rows)
		}

		log.Printf("GoRoutine%d 시작 : %d행 ~ %d행", i, min, max)
		wg.Add(1)

		go func(rows [][]string, min int, max int, i int) {

			defer wg.Done()

			for n := min; n <= max; n++ {
				row := rows[n - 1]
				path := row[3]

				data, err := GetProbeData(path, 60*time.Second)
				result := ""
				if err != nil {
					result = err.Error()
				} else {
					result = fmt.Sprintf("%f", math.Floor(data.Format.Duration().Seconds()))
				}


				mutex.Lock()
				curPosition++
				if err := readFile.SetCellValue("Sheet1", "E"+strconv.Itoa(n), result); err != nil {
					log.Print(err)
				}
				mutex.Unlock()
				log.Printf("(%d, %d) goroutine%d, path : %v, duration : %v", curPosition, len(rows) - 1, i, path, result)
			}
		}(rows, min, max, i)
	}

	wg.Wait()

	if err := readFile.Save(); err != nil {
		log.Print(err)
	}
}

func ElapsedTime() func() {
	start := time.Now()

	return func() {
		log.Printf("Finish : %s", time.Since(start))
	}
}
