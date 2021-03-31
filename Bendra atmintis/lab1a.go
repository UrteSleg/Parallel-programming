package main

import (
	"fmt"
	"sync"
)

var (
	DataSize = 100
)

type ResultMonitor struct {
	DataArray []Result
	Count     int
	cond      *sync.Cond
	mutex     *sync.Mutex
}

type DataMonitor struct {
	DataArray         []Employee
	Count             int
	hasFinishedAdding bool
	cond              *sync.Cond
	mutex             *sync.Mutex
}

func CreateDataMonitor() *DataMonitor {
	mutex := sync.Mutex{}
	return &DataMonitor{Count: 0, cond: sync.NewCond(&mutex), mutex: &mutex, DataArray: make([]Employee, DataSize/2)}
}

func CreateResultMonitor() *ResultMonitor {
	mutex := sync.Mutex{}
	return &ResultMonitor{Count: 0, cond: sync.NewCond(&mutex), mutex: &mutex, DataArray: make([]Result, DataSize)}
}

func (masyvas *ResultMonitor) Add(result *Result) {

	masyvas.mutex.Lock()
	defer masyvas.mutex.Unlock()

	for index := 0; index < masyvas.Count; index++ {
		if masyvas.DataArray[index].Employee.Compare(result.Employee) {

			var oldRez Result
			newRez := *result
			for i := index; i < masyvas.Count+1; i++ {
				oldRez = masyvas.DataArray[i]
				masyvas.DataArray[i] = newRez
				newRez = oldRez
			}
			masyvas.Count++
			return
		}
	}
	masyvas.DataArray[masyvas.Count] = *result
	masyvas.Count++
}

func (masyvas *DataMonitor) Add(employee *Employee) {

	masyvas.mutex.Lock()
	defer masyvas.mutex.Unlock()
	defer masyvas.cond.Signal()

	for masyvas.Count >= len(masyvas.DataArray) {
		// Put thread to sleep to wait for data to be taken out
		masyvas.cond.Wait()
	}

	masyvas.DataArray[masyvas.Count] = *employee
	masyvas.Count++

}

func (masyvas *DataMonitor) Take() *Employee {

	masyvas.mutex.Lock()

	defer masyvas.mutex.Unlock()
	defer masyvas.cond.Signal() // Notify a waiting thread about a taken value

	for masyvas.Count == 0 {
		if masyvas.hasFinishedAdding {
			// Indicates that no more values will be added so it will return with a signal to finish threads
			return nil
		}
		masyvas.cond.Wait()
	}

	emp := masyvas.DataArray[masyvas.Count-1]
	masyvas.Count--

	return &emp
}

func (employees *Employees) ReadJsonEmployees(fileName string) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	err = json.Unmarshal(byteValue, &employees)
	if err != nil {
		panic(err)
	}
}

func Hashing(data *DataMonitor, result *ResultMonitor, waitGroup *sync.WaitGroup) {

	defer waitGroup.Done()

	for { // Loops till the ResultMonitor is empty and no longer being added to
		emp := data.Take()
		if emp == nil {
			// If there is no more work to be done, it exits
			break
		}

		stringToHash := fmt.Sprintf("%v %v %v", emp.Name, emp.Age, emp.Salary)
		var hash [32]byte
		hash = sha256.Sum256([]byte(stringToHash))
		for i := 0; i < 1000; i++ {
			hash = sha256.Sum256([]byte(fmt.Sprintf("%v%x", i, hash)))
		}

		if emp.Salary >= 1000 {
			result.Add(&Result{Employee: emp, ResultValue: hash})
		}

	}
}

func main() {

	var employees Employees
	dataMonitor := CreateDataMonitor()
	resultMonitor := CreateResultMonitor()
	employees.ReadJsonEmployees("IFF813_SlegaityteU_L1_dat_2.json")
	waitGroup := sync.WaitGroup{}
	threadCount := DataSize / 4
	waitGroup.Add(threadCount)

	for i := 0; i < threadCount; i++ {
		go Hashing(dataMonitor, resultMonitor, &waitGroup)
	}

	for _, emp := range employees.Employees { // Adding values to the DataMonitor
		dataMonitor.Add(&emp)
	}
	dataMonitor.hasFinishedAdding = true // Setting a flag that no more values will be added

	waitGroup.Wait()

	resultFile, err := os.OpenFile("IFF813_SlegaityteU_L1_rez.txt", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer resultFile.Close()

	resultFile.WriteString(fmt.Sprint(resultMonitor.Count))
	resultFile.WriteString(fmt.Sprintf("Name: %-12v |Age: %-5v |Salary: %-5v |Hash: %v\n", "", "", "", ""))
	for i := 0; i < resultMonitor.Count; i++ {
		resultFile.WriteString(fmt.Sprintf("%-18v |%-11v |%-12v |%x\n",
			resultMonitor.DataArray[i].Employee.Name,
			resultMonitor.DataArray[i].Employee.Age,
			resultMonitor.DataArray[i].Employee.Salary,
			resultMonitor.DataArray[i].ResultValue))
	}
}
