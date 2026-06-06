package handler

import (
	"github.com/gofiber/fiber/v2"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/response"
)

type TagHandler struct {
	tagClient metav1.TagServiceClient
}

func NewTagHandler(tagClient metav1.TagServiceClient) *TagHandler {
	return &TagHandler{
		tagClient: tagClient,
	}
}

func (h *TagHandler) Routes(router fiber.Router, auth fiber.Handler, admin ...fiber.Handler) {
	tag := router.Group("/tags", auth)
	tag.Get("/", h.ListTags)
	tag.Get("/:id", h.GetTagByID)

	adminMiddlewares := append([]fiber.Handler{auth}, admin...)
	adminGroup := router.Group("/admin/tags", adminMiddlewares...)
	adminGroup.Post("/", h.CreateTag)
	adminGroup.Put("/:id", h.UpdateTag)
	adminGroup.Delete("/:id", h.DeleteTag)
}

func (h *TagHandler) ListTags(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	tagType := c.Query("type")
	var tagTypePtr *string
	if tagType != "" {
		tagTypePtr = &tagType
	}

	resp, err := h.tagClient.ListTags(c.Context(), &metav1.ListTagsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Type:     tagTypePtr,
	})
	if err != nil {
		return err
	}

	tags := make([]dto.TagResponse, len(resp.Tags))
	for i, tag := range resp.Tags {
		tags[i] = mapper.ToTagResponse(tag)
	}

	return response.OKWithMeta(c, tags, response.Meta{
		Page:     int(resp.Page),
		PageSize: int(resp.PageSize),
		Total:    resp.Total,
	})
}

func (h *TagHandler) GetTagByID(c *fiber.Ctx) error {
	tagId := c.Params("id")

	resp, err := h.tagClient.GetTagByID(c.Context(), &metav1.GetTagByIDRequest{
		Id: tagId,
	})
	if err != nil {
		return err
	}
	return response.OK(c, mapper.ToTagResponse(resp.Tag))
}

func (h *TagHandler) CreateTag(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body struct {
		Name string `json:"name"`
		Type string `json:"type"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	res, err := h.tagClient.CreateTag(c.Context(), &metav1.CreateTagRequest{
		Name:        body.Name,
		Type:        body.Type,
		Slug:        body.Slug,
		CreatedById: userID,
	})
	if err != nil {
		return err
	}
	return response.OK(c, mapper.ToTagResponse(res.Tag))
}

func (h *TagHandler) UpdateTag(c *fiber.Ctx) error {
	tagID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)
	var body struct {
		Name *string `json:"name"`
		Type *string `json:"type"`
		Slug *string `json:"slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	res, err := h.tagClient.UpdateTag(c.Context(), &metav1.UpdateTagRequest{
		Id:          tagID,
		Name:        body.Name,
		Type:        body.Type,
		UpdatedById: userID,
		Slug:        body.Slug,
	})
	if err != nil {
		return err
	}
	return response.OK(c, mapper.ToTagResponse(res.Tag))
}

func (h *TagHandler) DeleteTag(c *fiber.Ctx) error {
	tagID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)
	isPermanent := c.QueryBool("permanent", false)
	_, err := h.tagClient.DeleteTag(c.Context(), &metav1.DeleteTagRequest{
		Id:          tagID,
		IsPermanent: isPermanent,
		DeletedById: userID,
	})
	if err != nil {
		return err
	}
	return response.OKWithMessage(c, "tag deleted")
}
