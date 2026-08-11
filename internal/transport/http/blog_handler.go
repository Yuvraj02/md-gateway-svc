package httpserver

import (
	"encoding/json"
	"io"
	"net/http"

	blogv1 "github.com/Yuvraj02/md-protos/proto/blog/v1"
)

func (s *Server) handleBlogPing(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("message")
	resp, err := s.clients.Blog.Ping(r.Context(), &blogv1.PingRequest{Message: msg})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": resp.GetMessage(), "service": resp.GetService()})
}

// GET /v1/blogs — list all blogs (studio home).
func (s *Server) handleListBlogs(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.ListBlogs(r.Context(), &blogv1.ListBlogsRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	out := make([]any, 0, len(resp.Blogs))
	for _, b := range resp.Blogs {
		out = append(out, blogJSON(b))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /v1/blogs/{blogID} — one blog.
func (s *Server) handleGetBlog(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.GetBlog(r.Context(), &blogv1.GetBlogRequest{BlogId: r.PathValue("blogID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, blogJSON(resp.Blog))
}

// POST /v1/blogs — create a blog (owner secret required).
func (s *Server) handleCreateBlog(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		OwnerID     string `json:"ownerId"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	resp, err := s.clients.Blog.CreateBlog(r.Context(), &blogv1.CreateBlogRequest{
		Name: body.Name, Description: body.Description, OwnerId: body.OwnerID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, blogJSON(resp.Blog))
}

// PATCH /v1/blogs/{blogID} — update blog fields (owner secret required).
func (s *Server) handleUpdateBlog(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req := &blogv1.UpdateBlogRequest{BlogId: r.PathValue("blogID")}
	if v, ok := body["name"].(string); ok {
		req.Name = v
	}
	if v, ok := body["description"].(string); ok {
		req.Description = v
	}
	if v, ok := body["coverImage"].(string); ok {
		req.CoverImage = v
	}
	if v, ok := body["slug"].(string); ok {
		req.Slug = v
	}
	resp, err := s.clients.Blog.UpdateBlog(r.Context(), req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, blogJSON(resp.Blog))
}

// POST /v1/blogs/{blogID}/upvotes — bump blog upvotes.
func (s *Server) handleUpvoteBlog(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.UpvoteBlog(r.Context(), &blogv1.UpvoteBlogRequest{BlogId: r.PathValue("blogID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, blogJSON(resp.Blog))
}

// GET /v1/blogs/{blogID}/articles — studio list for one blog.
func (s *Server) handleListArticles(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.ListArticles(r.Context(), &blogv1.ListArticlesRequest{BlogId: r.PathValue("blogID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	out := make([]any, 0, len(resp.Articles))
	for _, a := range resp.Articles {
		out = append(out, articleJSON(a))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /v1/blogs/{blogID}/articles/{articleID} — studio get (draft or published).
func (s *Server) handleGetArticle(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.GetArticle(r.Context(), &blogv1.GetArticleRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(resp.Article))
}

// POST /v1/blogs/{blogID}/articles — create empty draft (owner secret required).
func (s *Server) handleCreateArticle(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.CreateArticle(r.Context(), &blogv1.CreateArticleRequest{BlogId: r.PathValue("blogID")})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, articleJSON(resp.Article))
}

// PATCH /v1/blogs/{blogID}/articles/{articleID} — save editor changes (owner secret required).
func (s *Server) handleUpdateArticle(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req := &blogv1.UpdateArticleRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	}
	if v, ok := body["title"].(string); ok {
		req.Title = v
	}
	if v, ok := body["excerpt"].(string); ok {
		req.Excerpt = v
	}
	if tags, ok := body["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				req.Tags = append(req.Tags, s)
			}
		}
	}
	if media, ok := body["media"].(map[string]any); ok {
		req.Media = &blogv1.ArticleMedia{}
		if v, ok := media["imageUrl"].(string); ok {
			req.Media.ImageUrl = v
		}
		if v, ok := media["videoUrl"].(string); ok {
			req.Media.VideoUrl = v
		}
	}
	if content, ok := body["content"]; ok {
		b, _ := json.Marshal(content)
		req.ContentJson = string(b)
	}
	if v, ok := body["readingTimeMinutes"].(float64); ok {
		req.ReadingTimeMinutes = int32(v)
	}
	resp, err := s.clients.Blog.UpdateArticle(r.Context(), req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(resp.Article))
}

// DELETE /v1/blogs/{blogID}/articles/{articleID} — delete article (owner secret required).
func (s *Server) handleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	_, err := s.clients.Blog.DeleteArticle(r.Context(), &blogv1.DeleteArticleRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /v1/blogs/{blogID}/articles/{articleID}/upvotes
func (s *Server) handleUpvoteArticle(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.UpvoteArticle(r.Context(), &blogv1.UpvoteArticleRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(resp.Article))
}

// POST /v1/blogs/{blogID}/articles/{articleID}/publish — first publish of a DRAFT article.
func (s *Server) handlePublishArticle(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.PublishArticle(r.Context(), &blogv1.PublishArticleRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(resp.Article))
}

// POST /v1/blogs/{blogID}/articles/{articleID}/draft — create or reuse edit-draft.
func (s *Server) handleEnsureDraft(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.EnsureDraft(r.Context(), &blogv1.EnsureDraftRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, draftJSON(resp.Draft))
}

// GET /v1/blogs/{blogID}/articles/{articleID}/draft
func (s *Server) handleGetDraft(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.GetDraft(r.Context(), &blogv1.GetDraftRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, draftJSON(resp.Draft))
}

// PATCH /v1/blogs/{blogID}/articles/{articleID}/draft — save editor onto draft.
func (s *Server) handleUpdateDraft(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<20)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req := &blogv1.UpdateDraftRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	}
	if v, ok := body["title"].(string); ok {
		req.Title = v
	}
	if v, ok := body["excerpt"].(string); ok {
		req.Excerpt = v
	}
	if tags, ok := body["tags"].([]any); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok {
				req.Tags = append(req.Tags, s)
			}
		}
	}
	if media, ok := body["media"].(map[string]any); ok {
		req.Media = &blogv1.ArticleMedia{}
		if v, ok := media["imageUrl"].(string); ok {
			req.Media.ImageUrl = v
		}
		if v, ok := media["videoUrl"].(string); ok {
			req.Media.VideoUrl = v
		}
	}
	if content, ok := body["content"]; ok {
		b, _ := json.Marshal(content)
		req.ContentJson = string(b)
	}
	if v, ok := body["readingTimeMinutes"].(float64); ok {
		req.ReadingTimeMinutes = int32(v)
	}
	resp, err := s.clients.Blog.UpdateDraft(r.Context(), req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, draftJSON(resp.Draft))
}

// POST /v1/blogs/{blogID}/articles/{articleID}/draft/publish — atomic draft → live.
func (s *Server) handlePublishDraft(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.PublishDraft(r.Context(), &blogv1.PublishDraftRequest{
		BlogId: r.PathValue("blogID"), ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(resp.Article))
}

// GET /v1/articles — Discover feed (published only).
func (s *Server) handleListPublished(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.ListPublishedArticles(r.Context(), &blogv1.ListPublishedArticlesRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	out := make([]any, 0, len(resp.Articles))
	for _, a := range resp.Articles {
		out = append(out, articleJSON(a))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /v1/articles/trending — Trending by views this week.
func (s *Server) handleListTrending(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.ListTrendingArticles(r.Context(), &blogv1.ListTrendingArticlesRequest{})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	out := make([]any, 0, len(resp.Articles))
	for _, a := range resp.Articles {
		out = append(out, articleJSON(a))
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /v1/articles/{articleID} — public article page.
func (s *Server) handleGetPublishedArticle(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.GetPublishedArticle(r.Context(), &blogv1.GetPublishedArticleRequest{
		ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(resp.Article))
}

// GET /v1/articles/{articleID}/related
func (s *Server) handleListRelated(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.ListRelatedArticles(r.Context(), &blogv1.ListRelatedArticlesRequest{
		ArticleId: r.PathValue("articleID"), Limit: 3,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	out := make([]any, 0, len(resp.Articles))
	for _, a := range resp.Articles {
		out = append(out, articleJSON(a))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /v1/articles/{articleID}/views — count a page open.
func (s *Server) handleRecordView(w http.ResponseWriter, r *http.Request) {
	_, err := s.clients.Blog.RecordArticleView(r.Context(), &blogv1.RecordArticleViewRequest{
		ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /v1/articles/{articleID}/reads — count a meaningful read.
func (s *Server) handleRecordRead(w http.ResponseWriter, r *http.Request) {
	_, err := s.clients.Blog.RecordArticleRead(r.Context(), &blogv1.RecordArticleReadRequest{
		ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /v1/articles/{articleID}/comments
func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	resp, err := s.clients.Blog.ListComments(r.Context(), &blogv1.ListCommentsRequest{
		ArticleId: r.PathValue("articleID"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	out := make([]any, 0, len(resp.Comments))
	for _, c := range resp.Comments {
		out = append(out, commentJSON(c))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /v1/articles/{articleID}/comments — post a comment (public).
func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Website string `json:"website"`
		Body    string `json:"body"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	resp, err := s.clients.Blog.CreateComment(r.Context(), &blogv1.CreateCommentRequest{
		ArticleId: r.PathValue("articleID"),
		Name: body.Name, Email: body.Email, Website: body.Website, Body: body.Body,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, commentJSON(resp.Comment))
}
