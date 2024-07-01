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

	"github.com/stretchr/testify/mock"
)

// SMTPClientMock is a mocked SMTPClient
type SMTPClientMock struct {
	mock.Mock
}

// SendMail is the mocked method to send emails
func (c *SMTPClientMock) SendMail(
	addr string,
	a smtp.Auth,
	from string,
	to []string,
	msg []byte,
) error {
	ret := c.Called(addr, a, from, to, msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, smtp.Auth, string, []string, []byte) error); ok {
		r0 = rf(addr, a, from, to, msg)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}
