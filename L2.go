package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
)

var dataNumber = 25

func decreaseData() int {
	dataNumber--
	return dataNumber
}

type Employees struct {
	Employees []Employee `json:"employees"`
}

type Employee struct {
	Name        string  `json:"name"`
	Age         int     `json:"age"`
	Salary      float32 `json:"salary"`
	ResultValue [32]byte
}

func (e *Employees) DeleteEmployee(index int) Employee {
	employee := e.Employees[index]
	copy(e.Employees[index:], e.Employees[index+1:])
	e.Employees[len(e.Employees)-1] = Employee{}
	return employee
}

func (e *Employees) AddEmployee(employee Employee, index int) {
	e.Employees[index] = employee
}

func (employees *Employees) ReadData(fileName string) {
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

type SimpleUnboundedCounter struct {
	count int
}

func (counter *SimpleUnboundedCounter) Increase() {
	counter.count++
}

func (counter *SimpleUnboundedCounter) Decrease() {
	counter.count--
}

const workerThreads = 4

func main() {

	var employees Employees
	employees.ReadData("IFF813_SlegaityteU_L1_dat_2.json") 

	fmt.Println("Duomenys:")

	for i := 0; i < len(employees.Employees); i++ {
		fmt.Print("Name: ", employees.Employees[i].Name)
		fmt.Print(", Age: ", employees.Employees[i].Age)
		fmt.Print(", Salary: ", employees.Employees[i].Salary)
		fmt.Println()
	}

	fmt.Println()

	finisherChannel, actionChannel, resultChannel, increaserChannel, decreaserChannel, resultAddChannel := make(chan Employee), make(chan Employee), make(chan Employees), make(chan Employee), make(chan Employee), make(chan Employee)
	messageChannel := make(chan int)

	go func() {
		lenght := 25
		min := 0
		max := 25
		data := Employees{Employees: make([]Employee, lenght)}
		var counter SimpleUnboundedCounter

	K:
		for {
			activeChannels := []chan Employee{finisherChannel}

			if counter.count > min {
				activeChannels = append(activeChannels, decreaserChannel)
			} else {
				activeChannels = append(activeChannels, nil)
			}
			if counter.count < max {
				activeChannels = append(activeChannels, increaserChannel)
			} else {
				activeChannels = append(activeChannels, nil)
			}

			var cases []reflect.SelectCase
			for _, c := range activeChannels {
				cases = append(cases, reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(c),
				})
			}

			chosenIndex, reiksme, _ := reflect.Select(cases)
			emp := reiksme.Interface().(Employee)

			switch chosenIndex {
			case 0: 
				for z := 0; z < workerThreads; z++ {
					<-decreaserChannel
				}
				for t := 0; t < workerThreads; t++ {
					actionChannel <- emp
				}
				resultAddChannel <- emp
				break K
			case 1:
				counter.Decrease()
				fmt.Println(decreaseData())
				if dataNumber == 0 {
					messageChannel <- 0
				}
				employee := data.DeleteEmployee(counter.count)
				actionChannel <- employee

			case 2:
				data.AddEmployee(emp, counter.count)
				counter.Increase()
			}
		}
	}()

	go func() {
		lenght := len(employees.Employees)
		count := 0
		result := Employees{Employees: make([]Employee, lenght)}
	K1:
		for {
			employee := <-resultAddChannel
			switch len(employee.Name) {
			case 0:
				resultChannel <- result
				break K1
			default:
				result.AddEmployee(employee, count)
				count++
			}
		}
	}()

	//darbininkes
	for i := 0; i < workerThreads; i++ {
		go func() {
			decreaserChannel <- Employee{}
		K2:
			for {
				employee := <-actionChannel
				switch len(employee.Name) {
				case 0:
					fmt.Println("DARBININKE ISEINA")
					break K2

				default:
					stringToHash := fmt.Sprintf("%v %v %v", employee.Name, employee.Age, employee.Salary)
					var hash [32]byte
					hash = sha256.Sum256([]byte(stringToHash))
					for i := 0; i < 10; i++ {
						hash = sha256.Sum256([]byte(fmt.Sprintf("%v%x", i, hash)))
					}
					employee.ResultValue = hash
					if employee.Salary >= 1000 {
						resultAddChannel <- employee
					}
					decreaserChannel <- Employee{}
				}
			}
		}()
	}

	for _, employee := range employees.Employees {
		increaserChannel <- employee
	}

	go func() {
		message := <-messageChannel
		if message == 0 {
			finisherChannel <- Employee{}
		}
	}()

	rezultatuMas := <-resultChannel

	sort.Slice(rezultatuMas.Employees, func(i, j int) bool {
		return rezultatuMas.Employees[i].Salary < rezultatuMas.Employees[j].Salary // surikiuoja pagal salary
	})

	resultFile, err := os.OpenFile("IFF813_SlegaityteU_L1_rez.txt", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer resultFile.Close()

	resultFile.WriteString(fmt.Sprintf("Name: %-12v |Age: %-5v |Salary: %-5v |Hash: %v\n", "", "", "", ""))
	for _, employee := range rezultatuMas.Employees {
		if employee.Age > 0 {
			resultFile.WriteString(fmt.Sprintf("%-18v |%-11v |%-12v |%x\n",
				employee.Name,
				employee.Age,
				employee.Salary,
				employee.ResultValue))
		}
	}
}
