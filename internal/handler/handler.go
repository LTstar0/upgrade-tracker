package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"upgrade-tracker/internal/model"
	"upgrade-tracker/internal/repo"
)

type Handler struct {
	clients  *repo.ClientRepo
	upgrades *repo.UpgradeRepo
	images   *repo.ImageRepo
}

func New(c *repo.ClientRepo, u *repo.UpgradeRepo, i *repo.ImageRepo) *Handler {
	return &Handler{clients: c, upgrades: u, images: i}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, map[string]any{"code": 0, "data": data})
}

func okExtra(w http.ResponseWriter, data any, extra map[string]any) {
	body := map[string]any{"code": 0, "data": data}
	for k, v := range extra {
		body[k] = v
	}
	writeJSON(w, http.StatusOK, body)
}

func fail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"code": 1, "msg": msg})
}

func pathID(r *http.Request, prefix string) (int, bool) {
	seg := strings.TrimPrefix(r.URL.Path, prefix)
	seg = strings.Split(seg, "/")[0]
	id, err := strconv.Atoi(seg)
	return id, err == nil && id > 0
}

func body[T any](r *http.Request) (T, error) {
	var v T
	err := json.NewDecoder(r.Body).Decode(&v)
	return v, err
}

// ── Routes ───────────────────────────────────────────────────────────────────

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/health",            h.health)
	mux.HandleFunc("/api/clients",           h.clientsRoot)
	mux.HandleFunc("/api/clients/",          h.clientsItem)
	mux.HandleFunc("/api/images",            h.imagesRoot)
	mux.HandleFunc("/api/images/",           h.imagesItem)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ok(w, "ok")
}

// /api/clients  →  GET list | POST create
func (h *Handler) clientsRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query().Get("q")
		clients, stats, err := h.clients.List(q)
		if err != nil {
			fail(w, 500, err.Error()); return
		}
		okExtra(w, clients, map[string]any{"stats": stats})

	case http.MethodPost:
		type req struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Contact string `json:"contact"`
			Note    string `json:"note"`
			Version string `json:"current_version"`
		}
		b, err := body[req](r)
		if err != nil || strings.TrimSpace(b.Name) == "" {
			fail(w, 400, "客户名称不能为空"); return
		}
		if b.Type == "" { b.Type = "other" }
		if b.Version == "" { b.Version = "v1.0.0" }
		c, err := h.clients.Create(b.Name, b.Type, b.Contact, b.Note, b.Version)
		if err != nil { fail(w, 500, err.Error()); return }
		ok(w, c)

	default:
		fail(w, 405, "method not allowed")
	}
}

// /api/clients/{id}           →  GET | PUT | DELETE
// /api/clients/{id}/upgrades  →  GET | POST
// /api/clients/{id}/upgrades/{uid} not needed; delete via /api/upgrades/{uid}
func (h *Handler) clientsItem(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path // e.g. /api/clients/3  or  /api/clients/3/upgrades

	// /api/clients/{id}/upgrades
	if strings.Contains(path, "/upgrades") {
		h.upgradesHandler(w, r, path)
		return
	}

	// /api/clients/{id}
	cid, ok2 := pathID(r, "/api/clients/")
	if !ok2 { fail(w, 400, "invalid id"); return }

	switch r.Method {
	case http.MethodGet:
		c, err := h.clients.Get(cid)
		if err != nil { fail(w, 404, err.Error()); return }
		ok(w, c)

	case http.MethodPut:
		type req struct {
			Name    string `json:"name"`
			Type    string `json:"type"`
			Contact string `json:"contact"`
			Note    string `json:"note"`
		}
		b, err := body[req](r)
		if err != nil || strings.TrimSpace(b.Name) == "" {
			fail(w, 400, "客户名称不能为空"); return
		}
		if err := h.clients.Update(cid, b.Name, b.Type, b.Contact, b.Note); err != nil {
			fail(w, 500, err.Error()); return
		}
		ok(w, nil)

	case http.MethodDelete:
		if err := h.clients.Delete(cid); err != nil {
			fail(w, 500, err.Error()); return
		}
		ok(w, nil)

	default:
		fail(w, 405, "method not allowed")
	}
}

func (h *Handler) upgradesHandler(w http.ResponseWriter, r *http.Request, path string) {
	// path: /api/clients/{cid}/upgrades  or  /api/clients/{cid}/upgrades/{uid}
	parts := strings.Split(strings.TrimPrefix(path, "/api/clients/"), "/")
	// parts[0]=cid, parts[1]="upgrades", parts[2]=uid (optional)
	cid, err := strconv.Atoi(parts[0])
	if err != nil || cid <= 0 { fail(w, 400, "invalid client id"); return }

	// DELETE /api/clients/{cid}/upgrades/{uid}
	if len(parts) >= 3 && parts[2] != "" {
		uid, err := strconv.Atoi(parts[2])
		if err != nil || uid <= 0 { fail(w, 400, "invalid upgrade id"); return }
		if r.Method != http.MethodDelete { fail(w, 405, "method not allowed"); return }

		clientID, err := h.upgrades.Delete(uid)
		if err != nil { fail(w, 404, err.Error()); return }
		// resync current_version
		if v := h.upgrades.LatestVersion(clientID); v != "" {
			h.clients.SetVersion(clientID, v)
		}
		ok(w, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		list, err := h.upgrades.ListByClient(cid)
		if err != nil { fail(w, 500, err.Error()); return }
		ok(w, list)

	case http.MethodPost:
		type req struct {
			Version     string   `json:"version"`
			UpgradeDate string   `json:"upgrade_date"`
			Operator    string   `json:"operator"`
			Tags        []string `json:"tags"`
			Description string   `json:"description"`
			Files       []string `json:"files"`
		}
		b, err2 := body[req](r)
		if err2 != nil { fail(w, 400, "请求格式错误"); return }
		if strings.TrimSpace(b.Version) == "" { fail(w, 400, "版本号不能为空"); return }
		if b.UpgradeDate == "" { fail(w, 400, "升级日期不能为空"); return }
		if strings.TrimSpace(b.Description) == "" { fail(w, 400, "升级说明不能为空"); return }
		if b.Operator == "" { b.Operator = "未知" }

		rec, err2 := h.upgrades.Create(cid,
			b.Version, b.UpgradeDate, b.Operator,
			model.TagsStr(b.Tags), b.Description, model.FilesStr(b.Files))
		if err2 != nil { fail(w, 500, err2.Error()); return }

		// sync current version
		h.clients.SetVersion(cid, b.Version)
		ok(w, rec)

	default:
		fail(w, 405, "method not allowed")
	}
}

// /api/images  →  GET list | POST create
func (h *Handler) imagesRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query().Get("q")
		images, err := h.images.List(q)
		if err != nil {
			fail(w, 500, err.Error()); return
		}
		ok(w, images)

	case http.MethodPost:
		type req struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Type        string `json:"type"`
			PublicURL   string `json:"public_url"`
			InternalURL string `json:"internal_url"`
			ConfigGuide string `json:"config_guide"`
			Description string `json:"description"`
		}
		b, err := body[req](r)
		if err != nil || strings.TrimSpace(b.Name) == "" || strings.TrimSpace(b.Version) == "" {
			fail(w, 400, "镜像名称和版本不能为空"); return
		}
		if b.Type == "" { b.Type = "docker" }
		i, err := h.images.Create(b.Name, b.Version, b.Type, b.PublicURL, b.InternalURL, b.ConfigGuide, b.Description)
		if err != nil { fail(w, 500, err.Error()); return }
		ok(w, i)

	default:
		fail(w, 405, "method not allowed")
	}
}

// /api/images/{id}           →  GET | PUT | DELETE
func (h *Handler) imagesItem(w http.ResponseWriter, r *http.Request) {
	id, isOK := pathID(r, "/api/images/")
	if !isOK { fail(w, 400, "invalid id"); return }

	switch r.Method {
	case http.MethodGet:
		i, err := h.images.Get(id)
		if err != nil { fail(w, 404, err.Error()); return }
		ok(w, i)

	case http.MethodPut:
		type req struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Type        string `json:"type"`
			PublicURL   string `json:"public_url"`
			InternalURL string `json:"internal_url"`
			ConfigGuide string `json:"config_guide"`
			Description string `json:"description"`
		}
		b, err := body[req](r)
		if err != nil || strings.TrimSpace(b.Name) == "" || strings.TrimSpace(b.Version) == "" {
			fail(w, 400, "镜像名称和版本不能为空"); return
		}
		if err := h.images.Update(id, b.Name, b.Version, b.Type, b.PublicURL, b.InternalURL, b.ConfigGuide, b.Description); err != nil {
			fail(w, 500, err.Error()); return
		}
		ok(w, nil)

	case http.MethodDelete:
		if err := h.images.Delete(id); err != nil {
			fail(w, 500, err.Error()); return
		}
		ok(w, nil)

	default:
		fail(w, 405, "method not allowed")
	}
}
