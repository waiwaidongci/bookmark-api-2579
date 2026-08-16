package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"bookmark-api/internal/model"
	"bookmark-api/internal/repository"
	"bookmark-api/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	var input model.CreateBookmarkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, CodeBadRequest, "invalid request: "+err.Error())
		return
	}

	bookmark, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondCreated(c, bookmark)
}

func (h *Handler) List(c *gin.Context) {
	var query struct {
		Tag      string `form:"tag"`
		Keyword  string `form:"keyword"`
		Page     int    `form:"page"`
		PageSize int    `form:"page_size"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, http.StatusBadRequest, CodeBadRequest, "invalid query: "+err.Error())
		return
	}

	result, err := h.svc.List(c.Request.Context(), model.ListFilter{
		Tag:      query.Tag,
		Keyword:  query.Keyword,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, result)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	bookmark, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, bookmark)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	var input model.UpdateBookmarkInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, CodeBadRequest, "invalid request: "+err.Error())
		return
	}

	bookmark, err := h.svc.Update(c.Request.Context(), id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, bookmark)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, gin.H{"id": id})
}

func (h *Handler) IncrementClickCount(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		return
	}

	bookmark, err := h.svc.IncrementClickCount(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	respondOK(c, bookmark)
}

func (h *Handler) Health(c *gin.Context) {
	respondOK(c, gin.H{"status": "ok"})
}

func parseID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, CodeBadRequest, "invalid id")
		return 0, err
	}
	return id, nil
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrBookmarkNotFound):
		respondError(c, http.StatusNotFound, CodeNotFound, err.Error())
	case errors.Is(err, service.ErrDuplicateURL):
		respondError(c, http.StatusConflict, CodeConflict, err.Error())
	case errors.Is(err, service.ErrInvalidTitle),
		errors.Is(err, service.ErrInvalidURL),
		errors.Is(err, service.ErrInvalidTags),
		errors.Is(err, service.ErrInvalidNote),
		errors.Is(err, service.ErrNoFieldsToUpdate):
		respondError(c, http.StatusBadRequest, CodeBadRequest, err.Error())
	default:
		respondError(c, http.StatusInternalServerError, CodeInternalError, "internal server error")
	}
}
