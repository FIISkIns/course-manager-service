package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type BaseCourseInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CourseInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type CourseList struct {
	Courses []CourseInfo `json:"courses"`
}

var courseList CourseList
var database *sql.DB

func HandleCourseGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	courseSearched := ps.ByName("course")
	course, err := getCourse(courseSearched)
	if err != nil  {
		http.Error(w, "Course does not exist", 404)
	}else {
		var courseInfo BaseCourseInfo
		resp, err := http.Get(course.URL)
		if err != nil {
			fmt.Println(err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		json.Unmarshal(body, &courseInfo)
		course.Name = courseInfo.Title

		data, err := json.Marshal(&course)
		if err != nil {
			http.Error(w, "Could not serialize course", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func HandleCourseDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	courseDeleted := ps.ByName("course")
	var indexDeleted int = -1
	for index, course := range courseList.Courses {
		if course.Id == courseDeleted {
			indexDeleted = index
		}
	}
	if indexDeleted == -1 {
		http.Error(w, "Course not found", 404)
	} else {
		courseList.Courses = append(courseList.Courses[:indexDeleted], courseList.Courses[indexDeleted+1:]...)
		fmt.Fprintf(w, "Cursul a fost sters")
	}
	UpdateCourseList()
}

func HandleCoursePut(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var courseInfo CourseInfo
	json.Unmarshal(body, &courseInfo)

	courseSearched := ps.ByName("course")
	var indexCourse int = -1
	for index, course := range courseList.Courses {
		if course.Id == courseSearched {
			indexCourse = index
		}
	}
	if indexCourse == -1 {
		courseList.Courses = append(courseList.Courses, courseInfo)
	} else {
		courseList.Courses[indexCourse].Id = courseInfo.Id
		courseList.Courses[indexCourse].Name = courseInfo.Name
		courseList.Courses[indexCourse].URL = courseInfo.URL
	}
	fmt.Fprintf(w, "Course updated/added")
	UpdateCourseList()

}

func HandleCoursePost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var courseInfo CourseInfo
	json.Unmarshal(body, &courseInfo)
	courseList.Courses = append(courseList.Courses, courseInfo)
	fmt.Fprintf(w, "Course added.")
	UpdateCourseList()
}

func HandleCoursesFunction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	courses, err := getCourses("Select * from courselist")
	if err != nil || courses == nil{
		http.Error(w, "No courses available", 404)
	} else {
		var courseInfo BaseCourseInfo
		for index, course := range courses.Courses {
			resp, err := http.Get(course.URL)
			if err != nil {
				fmt.Println(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			}
			json.Unmarshal(body, &courseInfo)
			courses.Courses[index].Name = courseInfo.Title
		}

		data, err := json.Marshal(&courses)
		if err != nil {
			http.Error(w, "Could not serialize course list", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}

}

func UpdateCourseList() {
	f, err := os.Create("courses.json")
	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	data, err := json.Marshal(&courseList)
	if err != nil {
		fmt.Println(err)
	}

	f.Write(data)

}

func getCourse(id string) (*CourseInfo, error) {
	query := "Select id,url from courselist where id like '" + id + "'"
	rows, err := database.Query(query)

	if err != nil {
		return nil, err
	}
	var item CourseInfo
	rows.Next()
		if err := rows.Scan(&item.Id, &item.URL); err != nil {
			return nil, err
		}
	return &item, nil
}

func getCourses(sqlString string) (*CourseList, error) {
	var courses CourseList
	rows, err := database.Query(sqlString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item CourseInfo
		if err := rows.Scan(&item.Id, &item.URL); err != nil {
			return nil, err
		}
		courses.Courses = append(courses.Courses, item)
	}

	if len(courses.Courses) == 0 {
		return nil, nil
	}

	return &courses, nil
}

func main() {
	//user:password@protocol(host_ip:host_port)/database
	var err error
	database, err = sql.Open("mysql", "Gafi:bagpicioarele@tcp(127.0.0.1:3306)/courses")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	router := httprouter.New()
	router.GET("/courses", HandleCoursesFunction)
	router.GET("/courses/:course", HandleCourseGet)
	router.PUT("/courses/:course", HandleCoursePut)
	router.POST("/courses/:course", HandleCoursePost)
	router.DELETE("/courses/:course", HandleCourseDelete)

	error := http.ListenAndServe(":8001", router)
	if error != nil {
		panic(error)
	}

}
