// Copyright 2022 Northern.tech AS
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
	"bytes"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strings"

	"github.com/mendersoftware/mender-server/pkg/config"
	"github.com/mendersoftware/mender-server/pkg/log"

	dconfig "github.com/mendersoftware/mender-server/services/workflows/config"
	"github.com/mendersoftware/mender-server/services/workflows/model"
)

var smtpClient SMTPClientInterface = new(SMTPClient)

func (jp *JobProcessor) processSMTPTask(
	smtpTask *model.SMTPTask,
	ps *JobStringProcessor,
	l *log.Logger,
) (*model.TaskResult, error) {
	var result *model.TaskResult = &model.TaskResult{
		SMTP: &model.TaskResultSMTP{},
	}

	l.Debugf("processSMTPTask starting")
	recipients := make([]string, 0, 10)

	to := make([]string, 0, 10)
	for _, address := range smtpTask.To {
		addresses := strings.Split(ps.ProcessJobString(address), ",")
		recipients = append(recipients, addresses...)
		to = append(to, addresses...)
	}

	cc := make([]string, 0, 10)
	for _, address := range smtpTask.Cc {
		addresses := strings.Split(ps.ProcessJobString(address), ",")
		recipients = append(recipients, addresses...)
		cc = append(cc, addresses...)
	}

	bcc := make([]string, 0, 10)
	for _, address := range smtpTask.Bcc {
		addresses := strings.Split(ps.ProcessJobString(address), ",")
		recipients = append(recipients, addresses...)
		bcc = append(bcc, addresses...)
	}

	from := ps.ProcessJobString(smtpTask.From)
	subject := ps.ProcessJobString(smtpTask.Subject)

	var err error
	var body, HTML string
	if body, err = processJobStringOrFile(smtpTask.Body, ps); err != nil {
		result.Success = false
		result.SMTP.Error = err.Error()
		l.Infof("processSMTPTask error reading file: '%s'", err.Error())
		return result, err
	}
	l.Debugf("processSMTPTask body text: '\n%s\n'", body)

	if HTML, err = processJobStringOrFile(smtpTask.HTML, ps); err != nil {
		result.Success = false
		result.SMTP.Error = err.Error()
		l.Infof("processSMTPTask error reading file: '%s'", err.Error())
		return result, err
	}
	l.Debugf("processSMTPTask body HTML: '\n%s\n'", HTML)

	altContent := &bytes.Buffer{}
	altWriter := multipart.NewWriter(altContent)
	if body != "" {
		childContent, _ := altWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/plain; charset=utf-8"},
			"Content-Transfer-Encoding": {"8bit"},
		})
		_, _ = childContent.Write([]byte(body))
	}
	if HTML != "" {
		childContent, _ := altWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {"text/html; charset=utf-8"},
			"Content-Transfer-Encoding": {"8bit"},
		})
		_, _ = childContent.Write([]byte(HTML))
	}
	altWriter.Close()

	msgContent := "From: " + from + "\r\n"
	if len(to) > 0 {
		msgContent += "To: " + strings.Join(to, ", ") + "\r\n"
	}
	if len(cc) > 0 {
		msgContent += "Cc: " + strings.Join(cc, ", ") + "\r\n"
	}
	if len(bcc) > 0 {
		msgContent += "Bcc: " + strings.Join(bcc, ", ") + "\r\n"
	}
	msgContent += "Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=" + altWriter.Boundary() + "\r\n" +
		"\r\n" +
		altContent.String()

	msgBuffer := &bytes.Buffer{}
	msgBuffer.WriteString(msgContent)

	result.SMTP.Sender = from
	result.SMTP.Recipients = recipients
	result.SMTP.Message = msgBuffer.String()

	// Set up authentication information
	smtpHostname := config.Config.GetString(dconfig.SettingSMTPHost)
	smtpUsername := config.Config.GetString(dconfig.SettingSMTPUsername)
	smtpPassword := config.Config.GetString(dconfig.SettingSMTPPassword)
	smtpAuthMechanism := config.Config.GetString(dconfig.SettingSMTPAuthMechanism)
	host, _, _ := net.SplitHostPort(smtpHostname)
	var auth smtp.Auth
	if smtpUsername != "" {
		if smtpAuthMechanism == "CRAM-MD5" {
			auth = smtp.CRAMMD5Auth(smtpUsername, smtpPassword)
		} else {
			auth = smtp.PlainAuth("", smtpUsername, smtpPassword, host)
		}
	}

	fromAddress := getEmailAddress(from)
	recipientAddresses := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		email := getEmailAddress(recipient)
		if email != "" {
			recipientAddresses = append(recipientAddresses, email)
		}
	}
	result.Success = true
	if len(recipientAddresses) > 0 {
		err = smtpClient.SendMail(
			smtpHostname,
			auth,
			fromAddress,
			recipientAddresses,
			msgBuffer.Bytes(),
		)
		l.Debugf("processSMTPTask: smtpClient.SendMail returned %v", err)
		if err != nil {
			l.Errorf("processSMTPTask: smtpClient.SendMail returned %v", err)
			result.Success = false
			result.SMTP.Error = err.Error()
		} else {
			l.Infof("processSMTPTask: email successfully sent to %v", recipients)
		}
	}

	return result, nil
}

func getEmailAddress(from string) string {
	emailAddress, err := mail.ParseAddressList(from)
	if err != nil || len(emailAddress) < 1 {
		return from
	}
	return emailAddress[0].Address
}
