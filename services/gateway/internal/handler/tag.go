package handler

import (
	"github.com/gofiber/fiber/v2"
	metav1 "microservice-golang/gen/meta/v1"
	"microservice-golang/services/gateway/internal/dto"
	"microservice-golang/services/gateway/internal/mapper"
	"microservice-golang/services/gateway/internal/middleware"
	"microservice-golang/services/gateway/internal/request"
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

// ListTags godoc
// @Summary      Get list of tags
// @Description  Retrieve list of all registered tags, optional filter by tag type (e.g. SPORTS, PRIORITY).
// @Tags         Tags
// @Produce      json
// @Param        type      query string false "Filter by tag type"
// @Param        page      query int    false "Page number (default 1)"
// @Param        page_size query int    false "Number of items per page (default 10)"
// @Success      200  {object}  response.Response{data=[]dto.TagResponse}
// @Failure      401  {object}  response.Response
// @Router       /tags [get]
// @Security     BearerAuth
func (h *TagHandler) ListTags(c *fiber.Ctx) error {
	page, pageSize := request.ParsePagination(c)
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

// GetTagByID godoc
// @Summary      Get tag details
// @Description  Retrieve tag details by tag ID (UUID).
// @Tags         Tags
// @Produce      json
// @Param        id   path      string  true  "Tag ID (UUID)"
// @Success      200  {object}  response.Response{data=dto.TagResponse}
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Router       /tags/{id} [get]
// @Security     BearerAuth
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

// CreateTag godoc
// @Summary      Create new tag (Admin)
// @Description  Create new tag metadata for entity classification.
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateTagRequest true "New Tag Data"
// @Success      200  {object}  response.Response{data=dto.TagResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/tags [post]
// @Security     BearerAuth
func (h *TagHandler) CreateTag(c *fiber.Ctx) error {
	userID := c.Locals(middleware.ContextUserID).(string)

	var body dto.CreateTagRequest
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

// UpdateTag godoc
// @Summary      Update tag (Admin)
// @Description  Update name, type, or slug of a tag.
// @Tags         Tags
// @Accept       json
// @Produce      json
// @Param        id      path string               true "Tag ID (UUID)"
// @Param        request body dto.UpdateTagRequest true "Tag Update Data"
// @Success      200  {object}  response.Response{data=dto.TagResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/tags/{id} [put]
// @Security     BearerAuth
func (h *TagHandler) UpdateTag(c *fiber.Ctx) error {
	tagID := c.Params("id")
	userID := c.Locals(middleware.ContextUserID).(string)
	var body dto.UpdateTagRequest
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

// DeleteTag godoc
// @Summary      Delete tag (Admin)
// @Description  Delete tag via soft delete or permanently.
// @Tags         Tags
// @Produce      json
// @Param        id        path  string true  "Tag ID (UUID)"
// @Param        permanent query bool   false "Permanent delete (default false)"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Router       /admin/tags/{id} [delete]
// @Security     BearerAuth
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
