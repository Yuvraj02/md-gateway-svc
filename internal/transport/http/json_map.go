package httpserver

import (
	"encoding/json"

	authv1 "github.com/Yuvraj02/md-protos/proto/auth/v1"
	blogv1 "github.com/Yuvraj02/md-protos/proto/blog/v1"
)

// JSON shapes match the frontend TypeScript types (camelCase).

func blogJSON(b *blogv1.Blog) map[string]any {
	if b == nil {
		return nil
	}
	stats := map[string]any{"totalViews": 0, "articleCount": 0, "totalReadingTimeMinutes": 0}
	if b.Stats != nil {
		stats = map[string]any{
			"totalViews":               b.Stats.TotalViews,
			"articleCount":             b.Stats.ArticleCount,
			"totalReadingTimeMinutes":  b.Stats.TotalReadingTimeMinutes,
		}
	}
	out := map[string]any{
		"id": b.Id, "name": b.Name, "slug": b.Slug, "description": b.Description,
		"ownerId": b.OwnerId, "upvotes": b.Upvotes, "stats": stats,
		"createdAt": b.CreatedAt, "updatedAt": b.UpdatedAt,
	}
	if b.CoverImage != "" {
		out["coverImage"] = b.CoverImage
	}
	return out
}

func articleJSON(a *blogv1.Article) map[string]any {
	if a == nil {
		return nil
	}
	var content any = []any{}
	if a.ContentJson != "" {
		_ = json.Unmarshal([]byte(a.ContentJson), &content)
	}
	media := map[string]any{}
	if a.Media != nil {
		if a.Media.ImageUrl != "" {
			media["imageUrl"] = a.Media.ImageUrl
		}
		if a.Media.VideoUrl != "" {
			media["videoUrl"] = a.Media.VideoUrl
		}
	}
	comments := make([]any, 0, len(a.Comments))
	for _, c := range a.Comments {
		comments = append(comments, commentJSON(c))
	}
	out := map[string]any{
		"id": a.Id, "blogId": a.BlogId, "title": a.Title, "excerpt": a.Excerpt,
		"media": media, "tags": a.Tags, "content": content, "status": a.Status,
		"upvotes": a.Upvotes, "views": a.Views, "viewsThisWeek": a.ViewsThisWeek,
		"reads": a.Reads, "readingTimeMinutes": a.ReadingTimeMinutes,
		"comments": comments, "createdAt": a.CreatedAt, "updatedAt": a.UpdatedAt,
		"hasDraft": a.HasDraft,
	}
	if a.PublishedAt != "" {
		out["publishedAt"] = a.PublishedAt
	}
	if a.LastSaved != "" {
		out["lastSaved"] = a.LastSaved
	}
	if a.Tags == nil {
		out["tags"] = []string{}
	}
	return out
}

func commentJSON(c *blogv1.Comment) map[string]any {
	if c == nil {
		return nil
	}
	out := map[string]any{
		"id": c.Id, "articleId": c.ArticleId, "name": c.Name,
		"email": c.Email, "body": c.Body, "createdAt": c.CreatedAt,
	}
	if c.Website != "" {
		out["website"] = c.Website
	}
	return out
}

func userJSON(u *authv1.User) map[string]any {
	if u == nil {
		return nil
	}
	return map[string]any{
		"id": u.Id, "name": u.Name, "email": u.Email, "username": u.Username,
		"avatar": u.Avatar, "bio": u.Bio, "createdAt": u.CreatedAt,
	}
}

func analyticsJSON(a *authv1.UserAnalytics) map[string]any {
	if a == nil {
		return nil
	}
	tops := make([]any, 0, len(a.TopPosts))
	for _, t := range a.TopPosts {
		tops = append(tops, map[string]any{"id": t.Id, "title": t.Title, "views": t.Views})
	}
	return map[string]any{
		"userId": a.UserId, "totalViews": a.TotalViews, "totalPosts": a.TotalPosts,
		"totalLikes": a.TotalLikes, "followers": a.Followers,
		"viewsThisMonth": a.ViewsThisMonth, "postsThisMonth": a.PostsThisMonth,
		"topPosts": tops,
	}
}

func draftJSON(d *blogv1.Draft) map[string]any {
	if d == nil {
		return nil
	}
	var content any = []any{}
	if d.ContentJson != "" {
		_ = json.Unmarshal([]byte(d.ContentJson), &content)
	}
	media := map[string]any{}
	if d.Media != nil {
		if d.Media.ImageUrl != "" {
			media["imageUrl"] = d.Media.ImageUrl
		}
		if d.Media.VideoUrl != "" {
			media["videoUrl"] = d.Media.VideoUrl
		}
	}
	tags := d.Tags
	if tags == nil {
		tags = []string{}
	}
	return map[string]any{
		"id": d.Id, "articleId": d.ArticleId, "blogId": d.BlogId,
		"title": d.Title, "excerpt": d.Excerpt, "media": media, "tags": tags,
		"content": content, "readingTimeMinutes": d.ReadingTimeMinutes,
		"createdAt": d.CreatedAt, "updatedAt": d.UpdatedAt,
	}
}
