// Copyright 2021 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package processor

import (
	"context"
	"io/ioutil"
	"net/smtp"
	"os"
	"regexp"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	mocklib "github.com/stretchr/testify/mock"

	"github.com/mendersoftware/mender-server/services/workflows/model"
	"github.com/mendersoftware/mender-server/services/workflows/store/mock"
)

func TestProcessJobSMTP(t *testing.T) {
	var mockedSMTPClient = new(SMTPClientMock)
	var originalSMTPClient = smtpClient
	smtpClient = mockedSMTPClient

	testCases := map[string]struct {
		Body     string
		HTML     string
		Expected string
	}{
		"text and html": {
			Body: "Body",
			HTML: "<html><body>HTML</body></html>",
			Expected: "From: no-reply@hosted.mender.io\r\n" +
				"To: user@mender.io\r\n" +
				"Cc: support@mender.io\r\n" +
				"Bcc: archive@mender.io, monitor@mender.io\r\n" +
				"Subject: Subject\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/alternative; boundary=ID\r\n" +
				"\r\n" +
				"--ID\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"\r\n" +
				"Body\r\n" +
				"--ID\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n" +
				"Content-Type: text/html; charset=utf-8\r\n" +
				"\r\n" +
				"<html><body>HTML</body></html>\r\n" +
				"--ID--\r\n",
		},
		"text only": {
			Body: "Body",
			Expected: "From: no-reply@hosted.mender.io\r\n" +
				"To: user@mender.io\r\n" +
				"Cc: support@mender.io\r\n" +
				"Bcc: archive@mender.io, monitor@mender.io\r\n" +
				"Subject: Subject\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/alternative; boundary=ID\r\n" +
				"\r\n" +
				"--ID\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"\r\n" +
				"Body\r\n" +
				"--ID--\r\n",
		},
		"html only": {
			HTML: "<html><body>HTML</body></html>",
			Expected: "From: no-reply@hosted.mender.io\r\n" +
				"To: user@mender.io\r\n" +
				"Cc: support@mender.io\r\n" +
				"Bcc: archive@mender.io, monitor@mender.io\r\n" +
				"Subject: Subject\r\n" +
				"MIME-Version: 1.0\r\n" +
				"Content-Type: multipart/alternative; boundary=ID\r\n" +
				"\r\n" +
				"--ID\r\n" +
				"Content-Transfer-Encoding: 8bit\r\n" +
				"Content-Type: text/html; charset=utf-8\r\n" +
				"\r\n" +
				"<html><body>HTML</body></html>\r\n" +
				"--ID--\r\n",
		},
	}

	for i, tc := range testCases {

		t.Run(i, func(t *testing.T) {
			mockedSMTPClient.On("SendMail",
				"",
				mocklib.MatchedBy(
					func(_ smtp.Auth) bool {
						return true
					}),
				"no-reply@hosted.mender.io",
				[]string{
					"user@mender.io",
					"support@mender.io",
					"archive@mender.io",
					"monitor@mender.io",
				},
				mocklib.MatchedBy(
					func(msg []byte) bool {
						assert.NotEqual(t, "", string(msg))

						return true
					}),
			).Return(nil)

			ctx := context.Background()
			dataStore := mock.NewDataStore()
			defer dataStore.AssertExpectations(t)

			workflow := &model.Workflow{
				Name: "test",
				InputParameters: []string{
					"to",
					"cc",
					"bcc",
				},
				Tasks: []model.Task{
					{
						Name: "task_1",
						Type: model.TaskTypeSMTP,
						SMTP: &model.SMTPTask{
							From:    "no-reply@hosted.mender.io",
							To:      []string{"${workflow.input.to}"},
							Cc:      []string{"${workflow.input.cc}"},
							Bcc:     []string{"${workflow.input.bcc}"},
							Subject: "Subject",
							Body:    tc.Body,
							HTML:    tc.HTML,
						},
					},
				},
			}

			job := &model.Job{
				WorkflowName: workflow.Name,
				InputParameters: model.InputParameters{
					{
						Name:  "to",
						Value: "user@mender.io",
					},
					{
						Name:  "cc",
						Value: "support@mender.io",
					},
					{
						Name:  "bcc",
						Value: "archive@mender.io,monitor@mender.io",
					},
				},
				Status: model.StatusPending,
			}

			dataStore.On("GetWorkflowByName",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				workflow.Name,
				mocklib.AnythingOfType("string"),
			).Return(workflow, nil)

			dataStore.On("UpsertJob",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
			).Return(job, nil)

			dataStore.On("UpdateJobStatus",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
				model.StatusDone,
			).Return(nil)

			dataStore.On("UpdateJobAddResult",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
				mocklib.MatchedBy(
					func(taskResult *model.TaskResult) bool {
						assert.True(t, taskResult.Success)
						assert.Equal(t, workflow.Tasks[0].SMTP.From, taskResult.SMTP.Sender)
						assert.Equal(t, "", taskResult.SMTP.Error)
						msg := taskResult.SMTP.Message
						re := regexp.MustCompile(`[a-z0-9]{60,}`)
						msg = re.ReplaceAllString(msg, "ID")
						assert.Equal(t, tc.Expected, msg)

						return true
					}),
			).Return(nil)

			jp := mockJobProcessor(job, dataStore, nil, nil)
			err := jp.ProcessJob(ctx)

			assert.Nil(t, err)
		})
	}

	smtpClient = originalSMTPClient
}

func TestProcessJobSMTPLoadFromFile(t *testing.T) {
	var mockedSMTPClient = new(SMTPClientMock)
	defer mockedSMTPClient.AssertExpectations(t)

	var originalSMTPClient = smtpClient
	smtpClient = mockedSMTPClient

	mockedSMTPClient.On("SendMail",
		"",
		mocklib.MatchedBy(
			func(_ smtp.Auth) bool {
				return true
			}),
		"no-reply@hosted.mender.io",
		[]string{
			"user@mender.io",
			"support@mender.io",
			"archive@mender.io",
			"monitor@mender.io",
		},
		mocklib.MatchedBy(
			func(_ []byte) bool {
				return true
			}),
	).Return(nil)

	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	tmpFile, err := ioutil.TempFile("", "mail.body")
	assert.Nil(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write([]byte("Hello\n\n This is the TestProcessJobSMTPLoadFromFile" +
		"sedning greetings.\n\nTestProcessJobSMTPLoadFromFile"))
	assert.Nil(t, err)

	err = tmpFile.Close()
	assert.Nil(t, err)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeSMTP,
				SMTP: &model.SMTPTask{
					From:    "no-reply@hosted.mender.io",
					To:      []string{"user@mender.io"},
					Cc:      []string{"support@mender.io"},
					Bcc:     []string{"archive@mender.io,monitor@mender.io"},
					Subject: "Subject",
					Body:    "@" + tmpFile.Name(),
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusDone,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.True(t, taskResult.Success)
				assert.Equal(t, workflow.Tasks[0].SMTP.From, taskResult.SMTP.Sender)
				assert.Equal(t, "", taskResult.SMTP.Error)

				return true
			}),
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, nil)
	err = jp.ProcessJob(ctx)

	assert.Nil(t, err)

	smtpClient = originalSMTPClient
}

func TestProcessJobSMTPLoadFromFileFailed(t *testing.T) {
	var mockedSMTPClient = new(SMTPClientMock)
	var originalSMTPClient = smtpClient
	smtpClient = mockedSMTPClient

	var testCases = map[string]struct {
		Workflow *model.Workflow
	}{
		"body": {
			Workflow: &model.Workflow{
				Name: "test",
				Tasks: []model.Task{
					{
						Name: "task_1",
						Type: model.TaskTypeSMTP,
						SMTP: &model.SMTPTask{
							From:    "no-reply@hosted.mender.io",
							To:      []string{"user@mender.io"},
							Cc:      []string{"support@mender.io"},
							Bcc:     []string{"archive@mender.io,monitor@mender.io"},
							Subject: "Subject",
							Body:    "@/this/file/does/not/exits/for/sure",
						},
					},
				},
			},
		},
		"html": {
			Workflow: &model.Workflow{
				Name: "test",
				Tasks: []model.Task{
					{
						Name: "task_1",
						Type: model.TaskTypeSMTP,
						SMTP: &model.SMTPTask{
							From:    "no-reply@hosted.mender.io",
							To:      []string{"user@mender.io"},
							Cc:      []string{"support@mender.io"},
							Bcc:     []string{"archive@mender.io,monitor@mender.io"},
							Subject: "Subject",
							HTML:    "@/this/file/does/not/exits/for/sure",
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(i, func(t *testing.T) {
			workflow := tc.Workflow

			job := &model.Job{
				WorkflowName: workflow.Name,
				Status:       model.StatusPending,
			}

			ctx := context.Background()
			dataStore := mock.NewDataStore()
			defer dataStore.AssertExpectations(t)

			dataStore.On("GetWorkflowByName",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				workflow.Name,
				mocklib.AnythingOfType("string"),
			).Return(workflow, nil)

			dataStore.On("UpsertJob",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
			).Return(job, nil)

			dataStore.On("UpdateJobStatus",
				mocklib.MatchedBy(
					func(_ context.Context) bool {
						return true
					}),
				job,
				model.StatusFailure,
			).Return(nil)

			jp := mockJobProcessor(job, dataStore, nil, nil)
			err := jp.ProcessJob(ctx)

			assert.NotNil(t, err)
			assert.Equal(t, "open /this/file/does/not/exits/for/sure: no such file or directory", err.Error())
		})
	}

	smtpClient = originalSMTPClient
}

func TestProcessJobSMTPFailure(t *testing.T) {
	var mockedSMTPClient = new(SMTPClientMock)
	defer mockedSMTPClient.AssertExpectations(t)

	var originalSMTPClient = smtpClient
	smtpClient = mockedSMTPClient

	smtpError := errors.New("smtp error")

	mockedSMTPClient.On("SendMail",
		"",
		mocklib.MatchedBy(
			func(_ smtp.Auth) bool {
				return true
			}),
		"no-reply@hosted.mender.io",
		[]string{
			"user@mender.io",
			"support@mender.io",
			"archive@mender.io",
			"monitor@mender.io",
		},
		mocklib.MatchedBy(
			func(_ []byte) bool {
				return true
			}),
	).Return(smtpError)

	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeSMTP,
				SMTP: &model.SMTPTask{
					From:    "no-reply@hosted.mender.io",
					To:      []string{"user@mender.io"},
					Cc:      []string{"support@mender.io"},
					Bcc:     []string{"archive@mender.io,monitor@mender.io"},
					Subject: "Subject",
					Body:    "Body",
				},
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	dataStore.On("UpdateJobAddResult",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		mocklib.MatchedBy(
			func(taskResult *model.TaskResult) bool {
				assert.False(t, taskResult.Success)
				assert.Equal(t, smtpError.Error(), taskResult.SMTP.Error)

				return true
			}),
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)

	assert.Nil(t, err)

	smtpClient = originalSMTPClient
}

func TestProcessJobSMTPFailedIncompatibleDefinition(t *testing.T) {
	ctx := context.Background()
	dataStore := mock.NewDataStore()
	defer dataStore.AssertExpectations(t)

	workflow := &model.Workflow{
		Name: "test",
		Tasks: []model.Task{
			{
				Name: "task_1",
				Type: model.TaskTypeSMTP,
			},
		},
	}

	job := &model.Job{
		WorkflowName: workflow.Name,
		Status:       model.StatusPending,
	}

	dataStore.On("GetWorkflowByName",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		workflow.Name,
		mocklib.AnythingOfType("string"),
	).Return(workflow, nil)

	dataStore.On("UpsertJob",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
	).Return(job, nil)

	dataStore.On("UpdateJobStatus",
		mocklib.MatchedBy(
			func(_ context.Context) bool {
				return true
			}),
		job,
		model.StatusFailure,
	).Return(nil)

	jp := mockJobProcessor(job, dataStore, nil, nil)
	err := jp.ProcessJob(ctx)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Error: Task definition incompatible with specified type (smtp)")
}

func TestGetEmailAddress(t *testing.T) {
	testCases := map[string]struct {
		In  string
		Out string
	}{
		"ok, email": {
			In:  "test@mender.io",
			Out: "test@mender.io",
		},
		"ok, name and email": {
			In:  "Mender <test@mender.io>",
			Out: "test@mender.io",
		},
		"ko": {
			In:  "Mender <test@mender.io",
			Out: "Mender <test@mender.io",
		},
	}
	for i, tc := range testCases {
		t.Run(i, func(t *testing.T) {
			result := getEmailAddress(tc.In)
			assert.Equal(t, tc.Out, result)
		})
	}
}
