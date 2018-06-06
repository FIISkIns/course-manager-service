package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"net/http"
	"os"
)

type BaseCourseInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type CourseInfo struct {
	Id string `json:"id"`
	Name       string `json:"name"`
	URL         string `json:"url"`
}

type CourseList struct {
	Courses []CourseInfo `json:"courses"`
}

var courseList CourseList

func CourseListInit() {
	//var courseInfo BaseCourseInfo
	data, _ := ioutil.ReadFile("courses.json")
	err := json.Unmarshal(data, &courseList)
	if err != nil {
		fmt.Println(err)
	}
	/*
		Aici voi prelua titlurile si descrierile de la fiecare curs in parte. Pentru prototip le iau din fisier
		for index, course := range courseList.URLs {
			fmt.Println(course)
			resp, err := http.Get("http://localhost:8003/") // aici url-ul va fi diferit pentru fiecare curs, in functie de course
			if err != nil {
				fmt.Println(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
			}
			json.Unmarshal(body, &courseInfo)
			courseList.Titles[index] = courseInfo.Title
			courseList.Descriptions[index] = courseInfo.Description
		}*/
}

func HandleCourseGet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	courseSearched := ps.ByName("course")
	var courseInfo CourseInfo
	var indexCourse int = -1
	for index, course := range courseList.Courses{
		fmt.Println(course)
		if course.Id == courseSearched {
			indexCourse = index
		}
	}
	if indexCourse == -1 {
		http.Error(w, "Course not found", 404)
	} else {
		courseInfo.Id = courseSearched
		courseInfo.URL= courseList.Courses[indexCourse].URL
		courseInfo.Name = courseList.Courses[indexCourse].Name
		data, err := json.Marshal(&courseInfo)
		if err != nil {
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func HandleCourseDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	courseDeleted := ps.ByName("course")
	var indexDeleted int = -1
	for index, course := range courseList.Courses{
		if course.Id == courseDeleted {
			indexDeleted = index
		}
	}
	if indexDeleted == -1 {
		http.Error(w, "Course not found", 404)
	} else {
		courseList.Courses= append(courseList.Courses[:indexDeleted], courseList.Courses[indexDeleted+1:]...)
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
	data, err := json.Marshal(&courseList)
	if err != nil {
		http.Error(w, "Could not serialize course list", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)

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

func main() {
	CourseListInit()
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
