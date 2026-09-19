package main

import (
	"fmt"
	"os"
	"strconv"
)

type Course struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
}

type CoursesResponse []Course

func (c *Client) SetCourseId(courseId int) error {
	c.session.CourseId = &courseId
	return c.store.Save(&c.session)
}

func (c *Client) Courses() ([]Course, error) {
	courseReq, err := c.NewRequest[struct{}]("GET", "/my/courses", nil)
	if err != nil {
		return nil, err
	}

	courseResp, err := c.Do[CoursesResponse](courseReq)
	if err != nil {
		return nil, err
	}

	return courseResp, nil
}

func CourseCommand(c *Client) {
	courses, err := c.Courses()
	if err != nil {
		fmt.Println("Failed to retrieve courses:", err)
		return
	}

	if len(os.Args) == 3 {
		selectedIdx, err := strconv.Atoi(os.Args[2])
		if err != nil || selectedIdx <= 0 || selectedIdx > len(courses) {
			fmt.Println("Invalid course index")
			return
		}

		course := courses[selectedIdx-1]
		if err := c.SetCourseId(course.ID); err != nil {
			fmt.Println("Failed to save course:", err)
			return
		}

		fmt.Println("Selected course:", course.Name, "-", course.Subject)
		return
	}

	for idx, course := range courses {
		fmt.Print(idx+1, ". ")
		fmt.Println(course.Name, "-", course.Subject)
	}
}
