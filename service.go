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
		http.Error(w, "Could not access the database for "+courseSearched+": "+err.Error(), http.StatusInternalServerError)
		log.Println("Could not access the database for " + courseSearched + ": " + err.Error())
	} else if course == nil {
		http.Error(w, "Could not find course "+courseSearched, 404)
		log.Println("Could not find course " + courseSearched)
	} else {
		var courseInfo BaseCourseInfo
		resp, err := http.Get(course.URL)
		if err != nil {
			http.Error(w, "Could not access the course "+courseSearched, http.StatusInternalServerError)
			log.Println("Could not acces the course " + courseSearched)
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Could not read from the course "+courseSearched, http.StatusInternalServerError)
				log.Println("Could not read from the course " + courseSearched)
			} else {
				json.Unmarshal(body, &courseInfo)
				course.Name = courseInfo.Title

				data, err := json.Marshal(&course)
				if err != nil {
					http.Error(w, "Could not serialize course "+courseSearched, http.StatusInternalServerError)
					log.Println("Could not serialize course " + courseSearched)
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
	courseInfo.Id = courseSearched
	course, err := getCourse(courseSearched)
	if err != nil || course.URL == "" {
		err = insertCourse(courseInfo)
		if err != nil {
			http.Error(w, "Could not insert course "+courseSearched, http.StatusInternalServerError)
			log.Println("Could not insert course" + courseSearched)
		} else {
			fmt.Fprintf(w, "Course "+courseSearched+" inserted.")
		}
	} else {
		err = updateCourse(courseInfo, courseSearched)
		if err != nil {
			http.Error(w, "Could not update course "+courseSearched, http.StatusInternalServerError)
			log.Println("Could not update course " + courseSearched)
		} else {
			fmt.Fprintf(w, "Course "+courseSearched+" updated.")
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
		http.Error(w, "Could not  read the content to be inserted", http.StatusInternalServerError)
		log.Println("Could not read the content to be inserted")
	}
	var courseInfo CourseInfo
	json.Unmarshal(body, &courseInfo)
	idCourse := ps.ByName("course")
	courseInfo.Id = idCourse
	err = insertCourse(courseInfo)
	if err != nil {
		http.Error(w, "Could not insert course "+courseInfo.Id, http.StatusInternalServerError)
		log.Println("Could not insert course " + courseInfo.Id)
	} else {
		fmt.Fprintf(w, "Course "+courseInfo.Id+" inserted.")
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
	if err != nil {
		http.Error(w, "Could not access the database: "+err.Error(), http.StatusInternalServerError)
		log.Println("Could not access the database: " + err.Error())
	} else {
		var courseInfo BaseCourseInfo
		for index, course := range courses {
			resp, err := http.Get(course.URL)
			if err != nil {
				http.Error(w, "Could not access "+course.Id, http.StatusInternalServerError)
				log.Println("Could not acces " + course.Id)
				worked = false
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					http.Error(w, "Could not read from "+course.Id, http.StatusInternalServerError)
					log.Println("Could not read from" + course.Id)
				}
				json.Unmarshal(body, &courseInfo)
				courses[index].Name = courseInfo.Title
			}
		}

		data, err := json.Marshal(&courses)
		if err != nil {
			http.Error(w, "Could not serialize course list", http.StatusInternalServerError)
			log.Println("Could not serialize course list")
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
		http.Error(w, "Course "+idCourse+" does not exist", 404)
		log.Println("Course " + idCourse + "does not exist")
	} else if err != nil {
		http.Error(w, "Course "+idCourse+"could not be deleted", http.StatusInternalServerError)
		log.Println("Course " + idCourse + "could not be deleted")
	} else {
		fmt.Fprintf(w, "Course "+idCourse+"deleted.")
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

func checkHealth(w http.ResponseWriter, url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to communicate with: "+url+"\nCause: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to communicate with: " + url + "\nCause: " + err.Error())
		return false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response from: "+url+"\nCause: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to read response from: " + url + "\nCause: " + err.Error())
		return false
	}
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed health check on: "+url+"\nResponse: "+string(body), http.StatusInternalServerError)
		log.Println("Failed health check on: " + url + "\nResponse: " + string(body))
		return false
	}
	return true
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	if err := database.Ping(); err != nil {
		http.Error(w, "Database connection failed: "+err.Error(), http.StatusInternalServerError)
		log.Println("Database connection failed: " + err.Error())
		return
	}

	courses, _ := getCourses("Select * from courselist")
	for _, course := range courses {
		if success := checkHealth(w, course.URL); !success {
			return
		}
	}

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
	router.GET("/health", healthCheckHandler)
	router.GET("/courses/:course", HandleCourseGet)
	router.PUT("/courses/:course", HandleCoursePut)
	router.POST("/courses/:course", HandleCoursePost)
	router.DELETE("/courses/:course", HandleCourseDelete)

	error := http.ListenAndServe(":"+strconv.Itoa(config.Port), router)
	if error != nil {
		log.Fatal(err)
	}

}
