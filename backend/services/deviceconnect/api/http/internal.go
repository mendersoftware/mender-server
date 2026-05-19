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

package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mendersoftware/mender-server/pkg/identity"
	"github.com/mendersoftware/mender-server/pkg/rest.utils"
	"github.com/mendersoftware/mender-server/pkg/stream"
	"github.com/mendersoftware/mender-server/pkg/ws"
	"github.com/mendersoftware/mender-server/pkg/ws/menderclient"

	"github.com/mendersoftware/mender-server/services/deviceconnect/app"
	"github.com/mendersoftware/mender-server/services/deviceconnect/client/nats"
)

// InternalController contains status-related end-points
type InternalController struct {
	app  app.App
	nats nats.Client
}

// NewInternalController returns a new InternalController
func NewInternalController(app app.App, nc nats.Client) *InternalController {
	return &InternalController{app: app, nats: nc}
}

func (h InternalController) CheckUpdate(c *gin.Context) {
	h.sendMenderCommand(c, menderclient.MessageTypeMenderClientCheckUpdate)
}

func (h InternalController) SendInventory(c *gin.Context) {
	h.sendMenderCommand(c, menderclient.MessageTypeMenderClientSendInventory)
}

func (h InternalController) sendMenderCommand(c *gin.Context, msgType string) {
	ctx := c.Request.Context()

	tenantID := c.Param("tenantId")
	deviceID := c.Param("deviceId")

	err := sendMenderCommand(ctx, h.nats, tenantID, deviceID, msgType)
	if err != nil {
		if errors.Is(err, stream.ErrConnectionRefused) {
			rest.RenderError(c, http.StatusNotFound, app.ErrDeviceNotConnected)
			return
		}
		rest.RenderInternalError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, nil)
}

func sendMenderCommand(ctx context.Context, nc nats.Client, tenantID, deviceID, cmd string) error {
	recvAddr := fmt.Sprintf("%s:cmd%s", tenantID, uuid.NewString())
	s, err := nc.Connect(ctx, recvAddr, deviceID)
	if err != nil {
		return err
	}
	defer s.Close(ctx)

	msg := &ws.ProtoMsg{
		Header: ws.ProtoHdr{
			Proto:   ws.ProtoTypeMenderClient,
			MsgType: cmd,
		},
	}
	if idata := identity.FromContext(ctx); idata != nil {
		msg.Header.Properties = map[string]interface{}{
			PropertyUserID: idata.Subject,
		}
	}
	data, _ := msgpack.Marshal(msg)

	err = s.Send(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (h InternalController) DeleteTenant(c *gin.Context) {
	ctx := c.Request.Context()
	tenantID := c.Param("tenantId")

	err := h.app.DeleteTenant(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.Status(http.StatusNoContent)
}
