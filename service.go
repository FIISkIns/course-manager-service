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
	"strconv"
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
	if err != nil {
		http.Error(w, "Course does not exist", 404)
	} else {
		var courseInfo BaseCourseInfo
		resp, err := http.Get(course.URL)
		if err != nil {
			http.Error(w, "Could not access the course", http.StatusInternalServerError)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Could not read from the course", http.StatusInternalServerError)
			} else {
				json.Unmarshal(body, &courseInfo)
				course.Name = courseInfo.Title

				data, err := json.Marshal(&course)
				if err != nil {
					http.Error(w, "Could not serialize course", http.StatusInternalServerError)
				} else {

					w.Header().Set("Content-Type", "application/json")
					w.Write(data)
				}
			}
		}
	}
}

func HandleCoursePut(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	var courseInfo CourseInfo
	json.Unmarshal(body, &courseInfo)

	courseSearched := ps.ByName("course")
	course, err := getCourse(courseSearched)
	if err != nil || course.URL == "" {
		err = insertCourse(courseInfo)
		if err != nil {
			http.Error(w, "Could not insert course.", http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "Course inserted.")
		}
	} else {
		err = updateCourse(courseInfo, courseSearched)
		if err != nil {
			http.Error(w, "Could not update course.", http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "Course updated.")
		}
	}
}

func updateCourse(info CourseInfo, id string) error {
	stmt, err := database.Prepare("update courselist set id=?,url=? where id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(info.Id, info.URL, id)
	if err != nil {
		return err
	}
	return nil
}

func HandleCoursePost(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Could not insert course.", http.StatusInternalServerError)
	}
	var courseInfo CourseInfo
	json.Unmarshal(body, &courseInfo)
	err = insertCourse(courseInfo)
	if err != nil {
		http.Error(w, "Could not insert course.", http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Course inserted.")
	}

}

func insertCourse(info CourseInfo) error {
	stmt, err := database.Prepare("INSERT courselist SET id=?,url=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(info.Id, info.URL)
	if err != nil {
		return err
	}
	return nil
}

func HandleCoursesFunction(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	courses, err := getCourses("Select * from courselist")
	worked := true
	if err != nil || courses == nil {
		http.Error(w, "No courses available", 404)
	} else {
		var courseInfo BaseCourseInfo
		for index, course := range courses {
			resp, err := http.Get(course.URL)
			if err != nil {
				http.Error(w, "Could not access "+course.Id, http.StatusInternalServerError)
				worked = false
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
				}
				json.Unmarshal(body, &courseInfo)
				courses[index].Name = courseInfo.Title
			}
		}

		data, err := json.Marshal(&courses)
		if err != nil {
			http.Error(w, "Could not serialize course list", http.StatusInternalServerError)
		} else if worked == true {
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}

	}

}

func getCourse(id string) (*CourseInfo, error) {
	stmt, err := database.Prepare("Select id,url from courselist where id = ?")
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

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

func HandleCourseDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	idCourse := ps.ByName("course")

	message, err := deleteCourse(idCourse)
	if message == "failure" {
		http.Error(w, "Course doesn't exist", 404)
	} else if err != nil {
		http.Error(w, "Course couldn't be deleted", http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Course deleted.")
	}
}

func deleteCourse(idCourse string) (string, error) {
	stmt, err := database.Prepare("delete from courselist where id = ?")
	if err != nil {
		return "", err
	}
	res, err := stmt.Exec(idCourse)
	if err != nil {
		return "", err
	}

	affect, err := res.RowsAffected()
	if err != nil {
		return "", err
	}

	if affect == 0 {
		return "failure", nil
	}
	return "succes", nil
}

func getCourses(sqlString string) ([]CourseInfo, error) {
	var courses []CourseInfo
	rows, err := database.Query(sqlString)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item CourseInfo
		if err := rows.Scan(&item.Id, &item.URL); err != nil {
			return nil, err
		}
		courses = append(courses, item)
	}

	return courses, nil
}

func initDatabase() error {
	err := database.Ping()
	if err != nil {
		return err
	}
	var auxiliary int
	err = database.QueryRow("SHOW TABLES LIKE 'courselist';").Scan(&auxiliary)
	if err != nil && err == sql.ErrNoRows {
		stmt, err := database.Prepare("create table courselist (" +
			"id varchar(100) primary key," +
			"url varchar(100) not null)")
		if err != nil {
			return err
		}
		_, err = stmt.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}
func main() {

	var err error
	initConfig()
	database, err = sql.Open("mysql", config.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	if err := initDatabase(); err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	router := httprouter.New()
	router.GET("/courses", HandleCoursesFunction)
	router.GET("/courses/:course", HandleCourseGet)
	router.PUT("/courses/:course", HandleCoursePut)
	router.POST("/courses/:course", HandleCoursePost)
	router.DELETE("/courses/:course", HandleCourseDelete)

	error := http.ListenAndServe(":"+strconv.Itoa(config.Port), router)
	if error != nil {
		log.Fatal(err)
	}

}
