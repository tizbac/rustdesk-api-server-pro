package admin

import (
	"rustdesk-api-server-pro/app/model"
	"rustdesk-api-server-pro/config"
	"rustdesk-api-server-pro/db"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/tidwall/gjson"
	"xorm.io/xorm"
)

// ParsePeer splits the peer column written by the RustDesk client, which is a json
// array holding the remote peer of the session: ["<peer id>", "<peer name>"].
// For an audit posted by a controlled host, that peer is the controller.
func ParsePeer(peer string) (string, string) {
	if peer == "" {
		return "", ""
	}
	parsed := gjson.Parse(peer)
	if !parsed.IsArray() {
		return peer, ""
	}
	values := parsed.Array()
	peerId := ""
	peerName := ""
	if len(values) > 0 {
		peerId = values[0].String()
	}
	if len(values) > 1 {
		peerName = values[1].String()
	}
	return peerId, peerName
}

type AuditController struct {
	basicController
}

func (c *AuditController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/audit/list", "HandleList")
	b.Handle("GET", "/audit/file-transfer-list", "HandleFileTransferList")
}

func (c *AuditController) HandleList() mvc.Result {
	currentPage := c.Ctx.URLParamIntDefault("current", 1)
	pageSize := c.Ctx.URLParamIntDefault("size", 10)
	conn_id := c.Ctx.URLParamDefault("conn_id", "")
	_type := c.Ctx.URLParamDefault("type", "")
	rustdesk_id := c.Ctx.URLParamDefault("rustdesk_id", "")
	peer_id := c.Ctx.URLParamDefault("peer_id", "")
	ip := c.Ctx.URLParamDefault("ip", "")
	session_id := c.Ctx.URLParamDefault("session_id", "")
	uuid := c.Ctx.URLParamDefault("uuid", "")
	created_at_0 := c.Ctx.URLParamDefault("created_at[0]", "")
	created_at_1 := c.Ctx.URLParamDefault("created_at[1]", "")
	closed_at_0 := c.Ctx.URLParamDefault("closed_at[0]", "")
	closed_at_1 := c.Ctx.URLParamDefault("closed_at[1]", "")

	query := func() *xorm.Session {
		q := c.Db.Table(&model.Audit{})
		if conn_id != "" {
			q.Where("audit.conn_id = ?", conn_id)
		}
		if _type != "" {
			q.Where("audit.type = ?", _type)
		}
		if rustdesk_id != "" {
			q.Where("audit.rustdesk_id = ?", rustdesk_id)
		}
		if peer_id != "" {
			// the peer column is the json array ["<peer id>", "<peer name>"]
			q.Where("audit.peer LIKE ?", "[\""+peer_id+"\"%")
		}
		if ip != "" {
			q.Where("audit.ip = ?", ip)
		}
		if session_id != "" {
			q.Where("audit.session_id = ?", session_id)
		}
		if uuid != "" {
			q.Where("audit.uuid = ?", uuid)
		}
		if created_at_0 != "" && created_at_1 != "" {
			q.Where("audit.created_at BETWEEN ? AND ?", created_at_0, created_at_1)
		}
		if closed_at_0 != "" && closed_at_1 != "" {
			q.Where("audit.closed_at BETWEEN ? AND ?", closed_at_0, closed_at_1)
		}
		q.Desc("id")
		return q
	}

	pagination := db.NewPagination(currentPage, pageSize)
	auditList := make([]model.Audit, 0)
	err := pagination.Paginate(query, &model.Audit{}, &auditList)
	if err != nil {
		return c.Error(nil, err.Error())
	}

	list := make([]iris.Map, 0)
	for _, a := range auditList {
		// rustdesk_id is the host that reported the session (the controlled machine),
		// peer is the remote side of that session (the controller).
		peerId, peerName := ParsePeer(a.Peer)
		var closedAt any = nil
		if !a.ClosedAt.IsZero() {
			closedAt = a.ClosedAt.Format(config.TimeFormat)
		}
		list = append(list, iris.Map{
			"id":          a.Id,
			"conn_id":     a.ConnId,
			"rustdesk_id": a.RustdeskId,
			"peer_id":     peerId,
			"peer_name":   peerName,
			"ip":          a.IP,
			"session_id":  a.SessionId,
			"uuid":        a.Uuid,
			"type":        a.Type,
			"created_at":  a.CreatedAt.Format(config.TimeFormat),
			"closed_at":   closedAt,
		})
	}
	return c.Success(iris.Map{
		"total":   pagination.TotalCount,
		"records": list,
		"current": currentPage,
		"size":    pageSize,
	}, "ok")
}

func (c *AuditController) HandleFileTransferList() mvc.Result {
	currentPage := c.Ctx.URLParamIntDefault("current", 1)
	pageSize := c.Ctx.URLParamIntDefault("size", 10)
	_type := c.Ctx.URLParamDefault("type", "")
	rustdesk_id := c.Ctx.URLParamDefault("rustdesk_id", "")
	peer_id := c.Ctx.URLParamDefault("peer_id", "")
	uuid := c.Ctx.URLParamDefault("uuid", "")
	created_at_0 := c.Ctx.URLParamDefault("created_at[0]", "")
	created_at_1 := c.Ctx.URLParamDefault("created_at[1]", "")

	query := func() *xorm.Session {
		q := c.Db.Table(&model.FileTransfer{})
		if _type != "" {
			q.Where("type = ?", _type)
		}
		if rustdesk_id != "" {
			q.Where("rustdesk_id = ?", rustdesk_id)
		}
		if peer_id != "" {
			q.Where("peer_id = ?", peer_id)
		}
		if uuid != "" {
			q.Where("audit.uuid = ?", uuid)
		}
		if created_at_0 != "" && created_at_1 != "" {
			q.Where("audit.created_at BETWEEN ? AND ?", created_at_0, created_at_1)
		}
		q.Desc("id")
		return q
	}

	pagination := db.NewPagination(currentPage, pageSize)
	fileTransferList := make([]model.FileTransfer, 0)
	err := pagination.Paginate(query, &model.FileTransfer{}, &fileTransferList)
	if err != nil {
		return c.Error(nil, err.Error())
	}

	list := make([]iris.Map, 0)
	for _, a := range fileTransferList {
		list = append(list, iris.Map{
			"id":          a.Id,
			"rustdesk_id": a.RustdeskId,
			"peer_id":     a.PeerId,
			"path":        a.Path,
			"uuid":        a.Uuid,
			"type":        a.Type,
			"created_at":  a.CreatedAt.Format(config.TimeFormat),
		})
	}
	return c.Success(iris.Map{
		"total":   pagination.TotalCount,
		"records": list,
		"current": currentPage,
		"size":    pageSize,
	}, "ok")
}
