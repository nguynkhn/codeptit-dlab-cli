package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type SubmitRequest struct {
	CourseID   int    `json:"course_id"`
	QuestionID int    `json:"question_id"`
	CompilerID int    `json:"compiler_id"`
	SourceCode string `json:"source_code"`
}

type SubmitResponse struct {
	ID int `json:"id"`
}

const (
	StatusPending = "pending"
	StatusDone    = "done"
	StatusError   = "error"
)

type Submission struct {
	Status  string `json:"status"`
	Verdict string `json:"verdict"`
}

type SubmissionResponse struct {
	Submission Submission `json:"submission"`
}

func (c *Client) Submit(courseId, questionId, compilerId int, sourceCode string) (int, error) {
	submitReq, err := c.NewRequest("POST", "/submissions", &SubmitRequest{
		CourseID: courseId, QuestionID: questionId, CompilerID: compilerId,
		SourceCode: sourceCode,
	})
	if err != nil {
		return 0, err
	}

	submitResp, err := c.Do[SubmitResponse](submitReq)
	if err != nil {
		return 0, err
	}

	return submitResp.ID, err
}

func (c *Client) Submission(submissionId int) (Submission, error) {
	var submission Submission

	endpoint := fmt.Sprintf("/submissions/%d", submissionId)
	submissionReq, err := c.NewRequest[any]("GET", endpoint, nil)
	if err != nil {
		return submission, err
	}

	submissionResp, err := c.Do[SubmissionResponse](submissionReq)
	if err != nil {
		return submission, err
	}

	submission = submissionResp.Submission
	return submission, err
}

func pollSubmission(c *Client, submissionId int) {
	const (
		INTERVAL = 1500 * time.Millisecond
		TIMEOUT  = 10 * time.Second
	)

	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	ticker := time.NewTicker(INTERVAL)
	defer ticker.Stop()

	for {
		submission, err := c.Submission(submissionId)
		if err != nil {
			fmt.Println("Failed to retrieve submission:", err)
			fmt.Println("Retrying...")
		} else if submission.Status == StatusDone {
			fmt.Println("Done! Verdict:", submission.Verdict)
			return
		} else if submission.Status == StatusError {
			fmt.Println("Submission failed")
			return
		}

		select {
		case <-ctx.Done():
			fmt.Println("Polling timed out:", ctx.Err())
			return

		case <-ticker.C:
		}
	}
}

func SubmitCommand(c *Client) {
	courseID, ok := c.GetCourse()
	if !ok {
		fmt.Println("Please select a course first")
		return
	}

	if len(os.Args) < 3 {
		fmt.Println("Please enter a file")
		return
	}

	path := os.Args[2]
	lang, ok := DetectLanguage(path)
	if !ok {
		fmt.Println("No available language detected")
		return
	}

	compilers, err := c.Compilers(courseID)
	if err != nil {
		fmt.Println("Failed to retrieve compilers:", err)
		return
	}

	compilerIdx := slices.IndexFunc(compilers, func(compiler Compiler) bool {
		return compiler.ID == int(lang)
	})
	if compilerIdx == -1 {
		fmt.Println("This course does not include this language")
		return
	}
	compilerID := compilers[compilerIdx].ID

	srcBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Failed to read file:", err)
		return
	}
	sourceCode := string(srcBytes)

	problemCode := strings.TrimSuffix(path, filepath.Ext(path))
	if comment, ok := ExtractFirstComment(sourceCode, lang); ok {
		problemCode = comment
	}

	problems, err := c.Problems(courseID)
	if err != nil {
		fmt.Println("Failed to retrieve problems:", err)
		return
	}

	problemIdx := slices.IndexFunc(problems, func(problem Problem) bool {
		return problem.Code == problemCode
	})
	if problemIdx == -1 {
		fmt.Println("No problem with this code found")
		return
	}
	problemID := problems[problemIdx].ID

	submissionId, err := c.Submit(courseID, problemID, compilerID, sourceCode)
	if err != nil {
		fmt.Println("Failed to submit code:", err)
		return
	}

	fmt.Println("Code submitted! Retrieving status...")
	pollSubmission(c, submissionId)
}
