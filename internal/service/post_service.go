package service

import (
	"aeibi/api"
	"aeibi/internal/repository/db"
	"aeibi/internal/repository/oss"
	"aeibi/util"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostService struct {
	db  *db.Queries
	dbx *sql.DB
	oss *oss.OSS
}

func NewPostService(dbx *sql.DB, ossClient *oss.OSS) *PostService {
	return &PostService{
		db:  db.New(dbx),
		dbx: dbx,
		oss: ossClient,
	}
}

func (s *PostService) CreatePost(ctx context.Context, uid string, req *api.CreatePostRequest) (*api.CreatePostResponse, error) {
	var resp *api.CreatePostResponse
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		row, err := qtx.CreatePost(ctx, db.CreatePostParams{
			Uid:         uuid.New(),
			Author:      util.UUID(uid),
			Text:        req.Text,
			Images:      req.Images,
			Attachments: req.Attachments,
			Visibility:  db.NullPostVisibility{PostVisibility: db.PostVisibility(req.Visibility), Valid: req.Visibility != ""},
			Pinned:      req.Pinned,
		})
		if err != nil {
			return fmt.Errorf("create post: %w", err)
		}
		err = qtx.UpsertPostTags(ctx, db.UpsertPostTagsParams{
			PostID: row.ID,
			Tags:   util.NormalizeStrings(req.Tags),
		})
		if err != nil {
			return fmt.Errorf("create post: %w", err)
		}
		resp = &api.CreatePostResponse{
			Uid: row.Uid.String(),
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *PostService) GetPost(ctx context.Context, viewerUid string, req *api.GetPostRequest) (*api.GetPostResponse, error) {
	postRow, err := s.db.GetPostByUid(ctx, db.GetPostByUidParams{
		Uid:    util.UUID(req.Uid),
		Viewer: uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	if postRow.Visibility == db.PostVisibilityPRIVATE && util.UUID(viewerUid) != postRow.Author {
		return nil, fmt.Errorf("post not found")
	}
	fileRow, err := s.db.GetFilesByUrls(ctx, postRow.Attachments)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get attachments: %w", err)
	}
	attachments := make([]*api.Attachment, 0, len(postRow.Attachments))
	for _, file := range fileRow {
		attachments = append(attachments, &api.Attachment{
			Url:         file.Url,
			Name:        file.Name,
			ContentType: file.ContentType,
			Size:        file.Size,
			Checksum:    file.Checksum,
		})
	}
	return &api.GetPostResponse{Post: &api.Post{
		Uid: postRow.Uid.String(),
		Author: &api.PostAuthor{
			Uid:         postRow.AuthorUid.String(),
			Nickname:    postRow.AuthorNickname,
			AvatarUrl:   postRow.AuthorAvatarUrl,
			IsFollowing: postRow.Following,
		},
		Text:            postRow.Text,
		Images:          postRow.Images,
		Attachments:     attachments,
		Tags:            postRow.TagNames,
		CommentCount:    postRow.CommentCount,
		CollectionCount: postRow.CollectionCount,
		LikeCount:       postRow.LikeCount,
		Visibility:      string(postRow.Visibility),
		LatestRepliedOn: postRow.LatestRepliedOn.Unix(),
		Ip:              postRow.Ip,
		Pinned:          postRow.Pinned,
		Liked:           postRow.Liked,
		Collected:       postRow.Collected,
		CreatedAt:       postRow.CreatedAt.Unix(),
		UpdatedAt:       postRow.UpdatedAt.Unix(),
	}}, nil
}

func (s *PostService) ListPosts(ctx context.Context, viewerUid string, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	filter, authorUID, err := normalizeListPostsFilter(req)
	if err != nil {
		return nil, err
	}
	cursor, err := decodeAndValidatePostPageToken(req.GetPageToken(), filter)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListPosts(ctx, db.ListPostsParams{
		Viewer:          uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		Query:           sql.NullString{String: filter.Query, Valid: filter.Query != ""},
		AuthorUid:       authorUID,
		TagName:         sql.NullString{String: filter.TagName, Valid: filter.TagName != ""},
		CursorCreatedAt: cursor.CreatedAt,
		CursorID:        cursor.ID,
		CursorScore:     cursor.Score,
	})
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	posts := make([]*api.Post, 0, len(rows))
	attachmentLists := make([][]string, 0, len(rows))
	for _, row := range rows {
		attachmentLists = append(attachmentLists, row.Attachments)
	}
	fileMap, err := s.listAttachmentFileMap(ctx, attachmentLists...)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		attachments := buildAttachmentsByURLOrder(row.Attachments, fileMap)
		posts = append(posts, &api.Post{
			Uid: row.Uid.String(),
			Author: &api.PostAuthor{
				Uid:         row.AuthorUid.String(),
				Nickname:    row.AuthorNickname,
				AvatarUrl:   row.AuthorAvatarUrl,
				IsFollowing: row.Following,
			},
			Text:            row.Text,
			Images:          row.Images,
			Attachments:     attachments,
			Tags:            row.TagNames,
			CommentCount:    row.CommentCount,
			CollectionCount: row.CollectionCount,
			LikeCount:       row.LikeCount,
			Visibility:      string(row.Visibility),
			LatestRepliedOn: row.LatestRepliedOn.Unix(),
			Ip:              row.Ip,
			Pinned:          row.Pinned,
			Liked:           row.Liked,
			Collected:       row.Collected,
			CreatedAt:       row.CreatedAt.Unix(),
			UpdatedAt:       row.UpdatedAt.Unix(),
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		token := postPageToken{
			Query:           filter.Query,
			AuthorUID:       filter.AuthorUID,
			TagName:         filter.TagName,
			CursorCreatedAt: last.CreatedAt.Unix(),
			CursorID:        last.Uid.String(),
		}
		if filter.Query != "" {
			score := last.Score
			token.CursorScore = &score
		}
		nextPageToken, err = encodePostPageToken(token)
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListPostsResponse{
		Posts:         posts,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *PostService) ListMyCollections(ctx context.Context, uid string, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	cursor, err := decodeAndValidatePostPageToken(req.GetPageToken(), listPostsFilter{})
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListPostsByCollector(ctx, db.ListPostsByCollectorParams{
		Collector:       util.UUID(uid),
		CursorCreatedAt: cursor.CreatedAt,
		CursorID:        cursor.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	posts := make([]*api.Post, 0, len(rows))
	attachmentLists := make([][]string, 0, len(rows))
	for _, row := range rows {
		attachmentLists = append(attachmentLists, row.Attachments)
	}
	fileMap, err := s.listAttachmentFileMap(ctx, attachmentLists...)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if row.Visibility == db.PostVisibilityPRIVATE && uid != row.Author.String() {
			continue
		}

		attachments := buildAttachmentsByURLOrder(row.Attachments, fileMap)

		posts = append(posts, &api.Post{
			Uid: row.Uid.String(),
			Author: &api.PostAuthor{
				Uid:         row.AuthorUid.String(),
				Nickname:    row.AuthorNickname,
				AvatarUrl:   row.AuthorAvatarUrl,
				IsFollowing: row.Following,
			},
			Text:            row.Text,
			Images:          row.Images,
			Attachments:     attachments,
			Tags:            row.TagNames,
			CommentCount:    row.CommentCount,
			CollectionCount: row.CollectionCount,
			LikeCount:       row.LikeCount,
			Visibility:      string(row.Visibility),
			LatestRepliedOn: row.LatestRepliedOn.Unix(),
			Ip:              row.Ip,
			Pinned:          row.Pinned,
			Liked:           row.Liked,
			Collected:       row.Collected,
			CreatedAt:       row.CreatedAt.Unix(),
			UpdatedAt:       row.UpdatedAt.Unix(),
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		token := postPageToken{
			CursorCreatedAt: last.CreatedAt.Unix(),
			CursorID:        last.Uid.String(),
		}
		nextPageToken, err = encodePostPageToken(token)
		if err != nil {
			return nil, fmt.Errorf("encode page token: %w", err)
		}
	}

	return &api.ListPostsResponse{
		Posts:         posts,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *PostService) SearchTags(ctx context.Context, req *api.SearchTagsRequest) (*api.SearchTagsResponse, error) {
	rows, err := s.db.SearchTags(ctx, req.Query)
	if err != nil {
		return nil, fmt.Errorf("search tags: %w", err)
	}

	tags := make([]*api.SearchTag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, &api.SearchTag{
			Name: row.Name,
		})
	}

	return &api.SearchTagsResponse{
		Tags: tags,
	}, nil
}

func (s *PostService) SuggestTagsByPrefix(ctx context.Context, req *api.SuggestTagsByPrefixRequest) (*api.SuggestTagsByPrefixResponse, error) {
	rows, err := s.db.SuggestTagsByPrefix(ctx, req.Prefix)
	if err != nil {
		return nil, fmt.Errorf("suggest tags by prefix: %w", err)
	}

	tags := make([]*api.SearchTag, 0, len(rows))
	for _, row := range rows {
		tags = append(tags, &api.SearchTag{
			Name: row.Name,
		})
	}

	return &api.SuggestTagsByPrefixResponse{
		Tags: tags,
	}, nil
}

func (s *PostService) UpdatePost(ctx context.Context, uid string, req *api.UpdatePostRequest) error {
	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		params := db.UpdatePostByUidAndAuthorParams{
			Uid:    util.UUID(req.Uid),
			Author: util.UUID(uid),
		}
		paths := make(map[string]struct{}, len(req.UpdateMask.GetPaths()))
		for _, path := range req.UpdateMask.GetPaths() {
			paths[path] = struct{}{}
		}
		if _, ok := paths["text"]; ok {
			params.Text = sql.NullString{String: req.Post.Text, Valid: true}
		}
		if _, ok := paths["images"]; ok {
			params.Images = req.Post.Images
		}
		if _, ok := paths["attachments"]; ok {
			params.Attachments = req.Post.Attachments
		}
		if _, ok := paths["visibility"]; ok {
			params.Visibility = db.NullPostVisibility{PostVisibility: db.PostVisibility(req.Post.Visibility), Valid: true}
		}
		if _, ok := paths["pinned"]; ok {
			params.Pinned = sql.NullBool{Bool: req.Post.Pinned, Valid: true}
		}

		id, err := qtx.UpdatePostByUidAndAuthor(ctx, params)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("post not found")
			}
			return fmt.Errorf("update post: %w", err)
		}
		if _, ok := paths["tags"]; ok {
			err = qtx.UpsertPostTags(ctx, db.UpsertPostTagsParams{
				PostID: id,
				Tags:   util.NormalizeStrings(req.Post.Tags),
			})
			if err != nil {
				return fmt.Errorf("update post: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *PostService) DeletePost(ctx context.Context, uid string, req *api.DeletePostRequest) error {
	return db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		affected, err := qtx.ArchivePostByUidAndAuthor(ctx, db.ArchivePostByUidAndAuthorParams{
			Uid:    util.UUID(req.Uid),
			Author: util.UUID(uid),
		})
		if err != nil {
			return fmt.Errorf("archive post: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("post not found or no permission")
		}
		return nil
	})
}

func (s *PostService) LikePost(ctx context.Context, uid string, req *api.LikePostRequest) (*api.LikePostResponse, error) {
	postUid := util.UUID(req.Uid)
	userUid := util.UUID(uid)

	var count int32

	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			applied, err := qtx.InsertPostLikeEdge(ctx, db.InsertPostLikeEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("post like: insert post like edge: %w", err)
			}

			if applied {
				count, err = qtx.IncrementPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post like: increment post like count: %w", err)
				}
			} else {
				count, err = qtx.GetPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post like: get post like count: %w", err)
				}
			}

		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			applied, err := qtx.DeletePostLikeEdge(ctx, db.DeletePostLikeEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("post like: delete post like edge: %w", err)
			}

			if applied {
				count, err = qtx.DecrementPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post like: decrement post like count: %w", err)
				}
			} else {
				count, err = qtx.GetPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post like: get post like count: %w", err)
				}
			}

		default:
			return fmt.Errorf("post like: unsupported action: %v", req.Action)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &api.LikePostResponse{
		Count: count,
	}, nil
}

func (s *PostService) CollectPost(ctx context.Context, uid string, req *api.CollectPostRequest) (*api.CollectPostResponse, error) {
	postUid := util.UUID(req.Uid)
	userUid := util.UUID(uid)

	var count int32

	if err := db.WithTx(ctx, s.dbx, s.db, func(qtx *db.Queries) error {
		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			applied, err := qtx.InsertPostCollectionEdge(ctx, db.InsertPostCollectionEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("post collection: insert post collection edge: %w", err)
			}

			if applied {
				count, err = qtx.IncrementPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: increment post collection count: %w", err)
				}
			} else {
				count, err = qtx.GetPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: get post collection count: %w", err)
				}
			}

		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			applied, err := qtx.DeletePostCollectionEdge(ctx, db.DeletePostCollectionEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("post collection: delete post collection edge: %w", err)
			}

			if applied {
				count, err = qtx.DecrementPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: decrement post collection count: %w", err)
				}
			} else {
				count, err = qtx.GetPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: get post collection count: %w", err)
				}
			}

		default:
			return fmt.Errorf("post collection: unsupported action: %v", req.Action)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &api.CollectPostResponse{
		Count: count,
	}, nil
}

func (s *PostService) listAttachmentFileMap(ctx context.Context, attachmentLists ...[]string) (map[string]db.GetFilesByUrlsRow, error) {
	attachmentUrls := make([]string, 0)
	seen := make(map[string]struct{})
	for _, list := range attachmentLists {
		for _, url := range list {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			attachmentUrls = append(attachmentUrls, url)
		}
	}

	files, err := s.db.GetFilesByUrls(ctx, attachmentUrls)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get attachments: %w", err)
	}

	fileMap := make(map[string]db.GetFilesByUrlsRow, len(files))
	for _, file := range files {
		fileMap[file.Url] = file
	}

	return fileMap, nil
}

func buildAttachmentsByURLOrder(urls []string, fileMap map[string]db.GetFilesByUrlsRow) []*api.Attachment {
	attachments := make([]*api.Attachment, 0, len(urls))
	for _, url := range urls {
		file, ok := fileMap[url]
		if !ok {
			continue
		}
		attachments = append(attachments, &api.Attachment{
			Url:         file.Url,
			Name:        file.Name,
			ContentType: file.ContentType,
			Size:        file.Size,
			Checksum:    file.Checksum,
		})
	}
	return attachments
}

type listPostsFilter struct {
	Query     string
	AuthorUID string
	TagName   string
}

type postPageToken struct {
	Query           string   `json:"query,omitempty"`
	AuthorUID       string   `json:"author_uid,omitempty"`
	TagName         string   `json:"tag_name,omitempty"`
	CursorScore     *float64 `json:"cursor_score,omitempty"`
	CursorCreatedAt int64    `json:"cursor_created_at,omitempty"`
	CursorID        string   `json:"cursor_id,omitempty"`
}

type postPageCursor struct {
	Score     sql.NullFloat64
	CreatedAt sql.NullTime
	ID        uuid.NullUUID
}

func normalizeListPostsFilter(req *api.ListPostsRequest) (listPostsFilter, uuid.NullUUID, error) {
	filter := listPostsFilter{
		Query:     strings.TrimSpace(req.GetQuery()),
		AuthorUID: strings.TrimSpace(req.GetAuthorUid()),
		TagName:   strings.TrimSpace(req.GetTagName()),
	}

	var authorUID uuid.NullUUID
	if filter.AuthorUID != "" {
		parsed, err := uuid.Parse(filter.AuthorUID)
		if err != nil {
			return listPostsFilter{}, uuid.NullUUID{}, status.Error(codes.InvalidArgument, "author_uid is invalid")
		}
		filter.AuthorUID = parsed.String()
		authorUID = uuid.NullUUID{UUID: parsed, Valid: true}
	}

	return filter, authorUID, nil
}

func decodeAndValidatePostPageToken(pageToken string, filter listPostsFilter) (postPageCursor, error) {
	if pageToken == "" {
		return postPageCursor{}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	var token postPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	if token.Query != filter.Query || token.AuthorUID != filter.AuthorUID || token.TagName != filter.TagName {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "page_token does not match current filters")
	}

	if token.AuthorUID != "" {
		if _, err := uuid.Parse(token.AuthorUID); err != nil {
			return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}

	if token.CursorCreatedAt <= 0 || token.CursorID == "" {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	cursorID, err := uuid.Parse(token.CursorID)
	if err != nil {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	hasQuery := filter.Query != ""
	if hasQuery && token.CursorScore == nil {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}
	if !hasQuery && token.CursorScore != nil {
		return postPageCursor{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	cursor := postPageCursor{
		CreatedAt: sql.NullTime{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true},
		ID:        uuid.NullUUID{UUID: cursorID, Valid: true},
	}
	if token.CursorScore != nil {
		cursor.Score = sql.NullFloat64{Float64: *token.CursorScore, Valid: true}
	}

	return cursor, nil
}

func encodePostPageToken(token postPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
