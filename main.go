package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

type Student struct {
	Id    int
	Name  string
	Age   int
	Score int
}

var students map[int]Student // 전역 변수 students : 학생의 ID를 키로 하여 Student 객체를 저장

var lastId int // 가장 최근에 추가된 마지막 학생의 ID

var lastQueriedId int // 사용자가 마지막으로 조회한 ID

func MakeWebHandler() http.Handler { // 핸들러: HTTP 요청을 받아서 그에 대응하는 작업을 수행하고, 결과를 클라이언트에게 돌려주는 역할
	mux := mux.NewRouter() // 라우터: client가 특정 URL로 요청을 보낼 때, 해당 URL에 맞는 핸들러 함수를 찾아 실행

	// 1. 학생 전체 목록 조회
	mux.HandleFunc("/students", GetStudentListHandler).Methods("GET")

	//2. 특정 학생 정보 조회
	mux.HandleFunc("/students/{id:[0-9]+}", GetStudentHandler).Methods("GET")

	// 3. 학생 데이터 추가
	mux.HandleFunc("/students", PostStudentHandler).Methods("POST")

	// 4. 학생 데이터 삭제
	mux.HandleFunc("/students/{id:[0-9]+}", DeleteStudentHandler).Methods("DELETE")

	// 5. 이전 조회한 id 다음 학생 조회 [0904]
	mux.HandleFunc("/students/next", GetNextStudentHandler).Methods("GET")

	// 초기 학생 데이터 설정
	students = make(map[int]Student)

	// 임시 학생 데이터 10개
	students[1] = Student{1, "aaa", 16, 87}
	students[2] = Student{2, "bbb", 18, 98}
	students[3] = Student{3, "ccc", 20, 85}
	students[4] = Student{4, "ccc", 11, 70}
	students[5] = Student{5, "ddd", 22, 76}
	students[6] = Student{6, "eee", 33, 82}
	students[7] = Student{7, "fff", 44, 83}
	students[8] = Student{8, "ggg", 55, 96}
	students[9] = Student{9, "hhh", 66, 62}
	students[10] = Student{10, "iii", 77, 34}

	lastId = 10

	return mux
}

type Students []Student // Students는 []Student라는 슬라이스 타입을 기반으로 만들어진 사용자 정의 타입

// sort 패키지에 정의된 sort.Interface 인터페이스 구현
func (s Students) Len() int { // 1) Len() int: 슬라이스의 길이 반환
	return len(s)
}
func (s Students) Swap(i, j int) { // 2) Swap(i, j int): 슬라이스 내 두 요소의 위치 교환
	s[i], s[j] = s[j], s[i]
}
func (s Students) Less(i, j int) bool { // 3) Less(i, j int) bool: 두 요소의 순서를 비교
	return s[i].Id < s[j].Id
}

// 1. 저장된 모든 학생 정보를 JSON 형식으로 반환하는 핸들러
func GetStudentListHandler(w http.ResponseWriter, r *http.Request) {
	list := make(Students, 0)
	for _, student := range students {
		list = append(list, student)
		sort.Sort(list) // ID 기준 -> 학생 목록 정렬
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// 2. 요청된 ID에 해당하는 학생 정보를 JSON 형식으로 반환
func GetStudentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)                 // 인수(id) 추출
	id, err := strconv.Atoi(vars["id"]) // URL에서 id를 정수로 변환
	if err != nil {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	student, ok := students[id]
	if !ok {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	lastQueriedId = id // 마지막으로 조회된 학생의 ID 업데이트

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

// 3. 새로운 학생 데이터를 받아서 시스템에 추가 & 추가한 학생 정보 볼 수 있도록 수정 (09/02)
func PostStudentHandler(w http.ResponseWriter, r *http.Request) {
	var student Student

	err := json.NewDecoder(r.Body).Decode(&student)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 디코딩 실패 시, HTTP 400 상태 코드 반환
		return
	}
	lastId++                   //새로운 학생에게 유니크한 ID 할당
	student.Id = lastId        // 새로운 ID를 학생의 Id 필드에 설정
	students[lastId] = student // 새 학생 정보 추가

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

// 4. 주어진 ID에 해당하는 학생 데이터를 제거 & 성공적으로 제거 여부 표시
func DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, ok := students[id] //// students 맵에서 해당 ID를 키로 사용하여 학생 데이터 존재 여부 확인

	if !ok {
		w.WriteHeader(http.StatusNotFound) // HTTP 404 (Not Found) 반환
		return
	}

	delete(students, id)         // 존재하는 경우, 맵에서 해당 학생 정보 삭제
	w.WriteHeader(http.StatusOK) // 성공적으로 학생 데이터를 삭제한 경우, HTTP 200 (OK)
}

// 5. next 학생 조회 핸들러
func GetNextStudentHandler(w http.ResponseWriter, r *http.Request) {
	if lastQueriedId == 0 { // 최근에 조회된 id가 없다면
		http.Error(w, "No student has been queried yet", http.StatusNotFound)
		return
	}

	nextId := lastQueriedId + 1 // 최근에 조회된 다음 id 설정
	student, ok := students[nextId]

	if !ok {
		w.WriteHeader(http.StatusNotFound) // 학생 데이터를 찾지 못한 경우 404 에러 반환
		return
	}

	lastQueriedId = nextId // 다음 학생의 ID로 업데이트

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(student)
}

// students 맵에 저장된 모든 학생의 정보를 출력 -> 테스트 코드에 출력
func PrintStudents() {
	fmt.Println("Current students in map:")
	for id, student := range students {
		fmt.Printf("ID: %d, Name: %s, Age: %d, Score: %d\n", id, student.Name, student.Age, student.Score)
	}
	fmt.Println("")
}

func main() {
	http.ListenAndServe(":3000", MakeWebHandler()) // 3000번 포트에서 HTTP 서버 시작
	// MakeWebHandler에서 반환된 핸들러를 사용하여 요청 처리
	// http://localhost:3000/students로 요청을 보내면, 서버는 두 개의 미리 정의된 학생 데이터를 JSON 형식으로 응답
	// -> 이 데이터는 Id 기준으로 정렬된 상태로 반환
}
