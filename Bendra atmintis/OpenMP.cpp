#include <iostream>
#include <omp.h>
#include <nlohmann/json.hpp>
#include <vector>
#include <fstream>
#include <sstream>
#include <iomanip>
#include <thread>
#include <chrono>
#include <tuple>

using nlohmann::json;
using nlohmann::basic_json;

class Employee
{
public:
	Employee() {}
	Employee(basic_json<>& json_val) {
		name = json_val["name"];
		age = json_val["age"];
		salary = json_val["salary"];
	}

	bool compare(Employee& other) const {
		if (age == other.age)
			return salary > other.salary;
		else
			return age > other.age;
	}
	std::string print_all() const {
		std::stringstream ss;
		const std::string seperator = " |";
		ss << std::setw(12) << std::left << name << seperator << std::setw(5) << age << seperator << std::setw(10) << salary;
		return ss.str();
	}
	std::string get_name() const { return name; }
	int get_age() const { return age; }
	double get_salary() const { return salary; }

private:
	std::string name;
	int age;
	double salary;
};

class Employees
{
public:
	std::vector<Employee> employees;
};

struct Result
{
	Employee employee;
	int y;

	Result() { }

	Result(Employee& employee, int y) {
		this->employee = employee;
		this->y = y;
	}
};

class DataMonitor
{
public:
	DataMonitor(const size_t& size) {
		data = new Employee[size];
		count = 0;
		max_size = size;
		omp_init_lock(&lock);
	}
	~DataMonitor() {
		omp_destroy_lock(&lock);
		delete[] data;
	}

	bool Add(Employee& employee) {

		bool is_done = false;
        
	#pragma omp critical(monitor_lock)
		{
			if (count < max_size) {
				data[count++] = employee;
				is_done = true;
			}
		}
		return is_done;
	}
	std::tuple<Employee, bool> Take() {

		Employee tmp;
		bool is_done = false;
		#pragma omp critical(monitor_lock)
		{
			if (count > 0) {
				tmp = data[--count];
				is_done = true;
			}
		}
		return { tmp, is_done };
	}

	int get_size() const { return count; }

private:
	Employee* data;
	size_t count; 
	size_t max_size;
	omp_lock_t lock;
public:
	bool has_finished = false;
};

class ResultMonitor
{
public:
	ResultMonitor(const size_t& size) {
		data = new Result[size];
		count = 0;
		omp_init_lock(&lock);
	}
	~ResultMonitor() {
		omp_destroy_lock(&lock);
		delete[] data;
	}

	void Add(Result& result) {
		omp_set_lock(&lock);
		for (size_t i = 0; i < count; i++) {
			if (data[i].employee.compare(result.employee)) {

				Result oldRez;
				auto newRez = result;
				for (size_t j = i; j < count + 1; j++) {
					oldRez = data[j];
					data[j] = newRez;
					newRez = oldRez;
				}
				count++;
				omp_unset_lock(&lock);
				return;
			}
		}
		data[count] = result;
		count++;
		omp_unset_lock(&lock);
	}

	Result& get(size_t& index) const { return data[index]; }
	size_t get_size() const { return count; }

private:
	Result* data;
	size_t count;
	omp_lock_t lock;
};

void ReadJsonFile(const std::string& file_name, Employees& employees) {

	std::ifstream fin(file_name);
	auto j = nlohmann::json::parse(fin);
	fin.close();
	for (auto& employee : j["employees"]) {
		employees.employees.push_back(Employee(employee));
	}
	std::cout << "Done parsing Json\n";
}

void WriteTextFile(const std::string& file_name, ResultMonitor& result_monitor) {

	const std::string seperator = " |";
	std::ofstream fout(file_name);
	fout << std::setw(12) << std::left << "Name" << seperator << std::setw(17) << "Age" << seperator << std::setw(7) << "Salary" << seperator << "XOR\n";
	for (size_t i = 0; i < result_monitor.get_size(); i++)
	{
		auto rez = result_monitor.get(i);
		fout << rez.employee.print_all() << seperator << rez.y << std::endl;
	}
	fout.close();
	std::cout << "Finished writting the result file\n";
}

void XOR(DataMonitor& data_monitor, ResultMonitor& result_value) {

	while (true) {
		std::tuple<Employee, bool> stu;
		bool has_value = false;
		while (!has_value) {
			if (data_monitor.get_size() == 0 && data_monitor.has_finished) {
				return;
			}
			stu = data_monitor.Take();
			has_value = std::get<1>(stu);
		}

		auto s = std::get<0>(stu);

		std::string name = s.get_name();
		const char* c = name.c_str();
		size_t len = strlen(c);
		int suma = 0;

		for (int i = 0; i < len; i++)
		{
			suma += name[i];
		}

		int age = s.get_age();
		double salary = s.get_salary();

		salary = salary + 0.5 - (salary < 0); 
		int y = (int)salary + suma + age; 

		for (unsigned int i = 1000000; i; i--) {
			y ^= i * age;

			auto result = Result(std::get<0>(stu), y);

			if (result.employee.get_salary() > 1000) {
				result_value.Add(result);
			}
		}
	}
}
	int main() {
		Employees employees;
		ReadJsonFile("IFF813_SlegaityteU_L1_dat_1.json", employees);
		DataMonitor monitor(employees.employees.size() / 2);
		ResultMonitor result(employees.employees.size());
		#pragma omp parallel
		{
			int thread_id = omp_get_thread_num();
			if (thread_id == 0) {
				for (auto& stu : employees.employees) {
					while (!monitor.Add(stu));
				}
				monitor.has_finished = true;
			}
			else {
				XOR(monitor, result);
			}
		}
		WriteTextFile("IFF_8-13_SlegaityteU_L1_rez.txt", result);
	}
