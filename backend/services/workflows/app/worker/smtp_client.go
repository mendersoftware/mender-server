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

package worker

import (
	"net/smtp"
)

// SMTPClientInterface is the interface which describes an SMTP client
type SMTPClientInterface interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// SMTPClient an SMTP client using smtp.SendMail
type SMTPClient struct {
}

// SendMail sends an email using smtp.SendMail
func (c *SMTPClient) SendMail(
	addr string,
	a smtp.Auth,
	from string,
	to []string,
	msg []byte,
) error {
	return smtp.SendMail(addr, a, from, to, msg)
}
