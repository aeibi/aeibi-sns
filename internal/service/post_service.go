package service

import (
	"aeibi/api"
	"aeibi/internal/async"
	"aeibi/internal/repository/db"
	"aeibi/internal/repository/oss"
	searchrepo "aeibi/internal/repository/search"
	"aeibi/util"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostService struct {
	db       *db.Queries
	pool     *pgxpool.Pool
	oss      *oss.OSS
	search   *searchrepo.Search
	producer *async.Producer
}

func NewPostService(pool *pgxpool.Pool, ossClient *oss.OSS, search *searchrepo.Search, riverClient *river.Client[pgx.Tx]) *PostService {
	return &PostService{
		db:       db.New(pool),
		pool:     pool,
		oss:      ossClient,
		search:   search,
		producer: async.New(riverClient),
	}
}

func (s *PostService) CreatePost(ctx context.Context, uid string, req *api.CreatePostRequest) (*api.CreatePostResponse, error) {
	var resp *api.CreatePostResponse

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

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

		tags := util.NormalizeStrings(req.Tags)
		if len(tags) > 0 {
			if err := qtx.InsertTagsIfNotExists(ctx, tags); err != nil {
				return fmt.Errorf("insert tags if not exists: %w", err)
			}

			if err := qtx.InsertPostTagsByNames(ctx, db.InsertPostTagsByNamesParams{
				PostID: row.ID,
				Tags:   tags,
			}); err != nil {
				return fmt.Errorf("insert post tags: %w", err)
			}
			if err := s.producer.EnqueueUpdateTagSearchTx(ctx, tx, async.UpdateTagSearchArgs{
				TagNames: tags,
			}); err != nil {
				return fmt.Errorf("enqueue update tag search job: %w", err)
			}
		}
		if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
			PostUID: row.Uid,
			Action:  async.PostSearchActionUpsert,
		}); err != nil {
			return fmt.Errorf("enqueue update post search job: %w", err)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("post not found")
		}
		return nil, fmt.Errorf("get post: %w", err)
	}
	if postRow.Visibility == db.PostVisibilityPRIVATE && util.UUID(viewerUid) != postRow.Author {
		return nil, fmt.Errorf("post not found")
	}
	fileRow, err := s.db.GetFilesByUrls(ctx, postRow.Attachments)
	if err != nil {
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
		LatestRepliedOn: postRow.LatestRepliedOn.Time.Unix(),
		Ip:              postRow.Ip,
		Pinned:          postRow.Pinned,
		Liked:           postRow.Liked,
		Collected:       postRow.Collected,
		CreatedAt:       postRow.CreatedAt.Time.Unix(),
		UpdatedAt:       postRow.UpdatedAt.Time.Unix(),
	}}, nil
}

func (s *PostService) ListPosts(ctx context.Context, viewerUid string, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	token, err := decodePostPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListPosts(ctx, db.ListPostsParams{
		Viewer:          uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		AuthorUid:       uuid.NullUUID{UUID: util.UUID(req.AuthorUid), Valid: req.AuthorUid != ""},
		TagName:         pgtype.Text{String: req.TagName, Valid: req.TagName != ""},
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: token.CursorCreatedAt > 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(token.CursorID), Valid: token.CursorID != ""},
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
			LatestRepliedOn: row.LatestRepliedOn.Time.Unix(),
			Ip:              row.Ip,
			Pinned:          row.Pinned,
			Liked:           row.Liked,
			Collected:       row.Collected,
			CreatedAt:       row.CreatedAt.Time.Unix(),
			UpdatedAt:       row.UpdatedAt.Time.Unix(),
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		token := postPageToken{
			CursorCreatedAt: last.CreatedAt.Time.Unix(),
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

func (s *PostService) SearchPosts(ctx context.Context, viewerUid string, req *api.SearchPostsRequest) (*api.ListPostsResponse, error) {
	token, err := decodePostSearchPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}

	result, err := s.search.SearchPosts(searchrepo.SearchPostsParams{
		Query:     req.Query,
		ViewerUID: viewerUid,
		AuthorUID: req.AuthorUid,
		TagName:   req.TagName,
		Limit:     20,
		Offset:    token.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("search posts: %w", err)
	}

	postUIDs := make([]uuid.UUID, 0, len(result.Hits))
	for _, hit := range result.Hits {
		uid, err := uuid.Parse(hit.UID)
		if err != nil {
			continue
		}
		postUIDs = append(postUIDs, uid)
	}

	extrasRows, err := s.db.GetPostSearchExtrasByUids(ctx, db.GetPostSearchExtrasByUidsParams{
		Viewer: uuid.NullUUID{UUID: util.UUID(viewerUid), Valid: viewerUid != ""},
		Uids:   postUIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("get searched post extras: %w", err)
	}

	extrasByUID := make(map[string]db.GetPostSearchExtrasByUidsRow, len(extrasRows))
	for _, row := range extrasRows {
		extrasByUID[row.Uid.String()] = row
	}

	attachmentLists := make([][]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		if _, ok := extrasByUID[hit.UID]; !ok {
			continue
		}
		attachmentLists = append(attachmentLists, hit.Attachments)
	}

	fileMap, err := s.listAttachmentFileMap(ctx, attachmentLists...)
	if err != nil {
		return nil, err
	}

	posts := make([]*api.Post, 0, len(result.Hits))
	for _, hit := range result.Hits {
		extra, ok := extrasByUID[hit.UID]
		if !ok {
			continue
		}

		attachments := buildAttachmentsByURLOrder(hit.Attachments, fileMap)
		posts = append(posts, &api.Post{
			Uid: hit.UID,
			Author: &api.PostAuthor{
				Uid:         hit.AuthorUID,
				Nickname:    extra.AuthorNickname,
				AvatarUrl:   extra.AuthorAvatarUrl,
				IsFollowing: extra.IsFollowing,
			},
			Text:            hit.Text,
			Images:          hit.Images,
			Attachments:     attachments,
			Tags:            hit.TagNames,
			CommentCount:    int32(hit.CommentCount),
			CollectionCount: int32(hit.CollectionCount),
			LikeCount:       int32(hit.LikeCount),
			Visibility:      hit.Visibility,
			LatestRepliedOn: hit.LatestRepliedOn,
			Ip:              "",
			Pinned:          hit.Pinned,
			Liked:           extra.Liked,
			Collected:       extra.Collected,
			CreatedAt:       hit.CreatedAt,
			UpdatedAt:       hit.UpdatedAt,
		})
	}

	nextPageToken := ""
	nextOffset := token.Offset + int64(len(result.Hits))
	if len(result.Hits) > 0 && nextOffset < result.EstimatedTotalHits {
		nextPageToken, err = encodePostSearchPageToken(postSearchPageToken{Offset: nextOffset})
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
	token, err := decodePostPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}

	rows, err := s.db.ListPostsByCollector(ctx, db.ListPostsByCollectorParams{
		Collector:       util.UUID(uid),
		CursorCreatedAt: pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: token.CursorCreatedAt > 0},
		CursorID:        uuid.NullUUID{UUID: util.UUID(token.CursorID), Valid: token.CursorID != ""},
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
			LatestRepliedOn: row.LatestRepliedOn.Time.Unix(),
			Ip:              row.Ip,
			Pinned:          row.Pinned,
			Liked:           row.Liked,
			Collected:       row.Collected,
			CreatedAt:       row.CreatedAt.Time.Unix(),
			UpdatedAt:       row.UpdatedAt.Time.Unix(),
		})
	}

	var nextPageToken string
	if len(rows) > 0 {
		last := rows[len(rows)-1]
		token := postPageToken{
			CursorCreatedAt: last.CreatedAt.Time.Unix(),
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

func (s *PostService) SearchTags(_ context.Context, req *api.SearchTagsRequest) (*api.SearchTagsResponse, error) {
	result, err := s.search.SearchTags(searchrepo.SearchTagsParams{
		Query: req.Query,
		Limit: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("search tags: %w", err)
	}

	tags := make([]*api.SearchTag, 0, len(result.Hits))
	for _, hit := range result.Hits {
		tags = append(tags, &api.SearchTag{
			Name: hit.Name,
		})
	}

	return &api.SearchTagsResponse{
		Tags: tags,
	}, nil
}

func (s *PostService) SuggestTagsByPrefix(_ context.Context, req *api.SuggestTagsByPrefixRequest) (*api.SuggestTagsByPrefixResponse, error) {
	result, err := s.search.SuggestTagsByName(req.Prefix, 10)
	if err != nil {
		return nil, fmt.Errorf("suggest tags by prefix: %w", err)
	}

	tags := make([]*api.SearchTag, 0, len(result.Hits))
	for _, hit := range result.Hits {
		tags = append(tags, &api.SearchTag{
			Name: hit.Name,
		})
	}

	return &api.SuggestTagsByPrefixResponse{
		Tags: tags,
	}, nil
}

func (s *PostService) UpdatePost(ctx context.Context, uid string, req *api.UpdatePostRequest) error {
	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		params := db.UpdatePostByUidAndAuthorParams{
			Uid:    util.UUID(req.Uid),
			Author: util.UUID(uid),
		}

		paths := make(map[string]struct{}, len(req.UpdateMask.GetPaths()))
		for _, path := range req.UpdateMask.GetPaths() {
			paths[path] = struct{}{}
		}

		if _, ok := paths["text"]; ok {
			params.Text = pgtype.Text{String: req.Post.Text, Valid: true}
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
			params.Pinned = pgtype.Bool{Bool: req.Post.Pinned, Valid: true}
		}

		id, err := qtx.UpdatePostByUidAndAuthor(ctx, params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("post not found")
			}
			return fmt.Errorf("update post: %w", err)
		}

		if _, ok := paths["tags"]; ok {
			tags := util.NormalizeStrings(req.Post.Tags)

			if len(tags) > 0 {
				if err := qtx.InsertTagsIfNotExists(ctx, tags); err != nil {
					return fmt.Errorf("update post: insert tags if not exists: %w", err)
				}
			}

			if err := qtx.DeletePostTagsNotInNames(ctx, db.DeletePostTagsNotInNamesParams{
				PostID: id,
				Tags:   tags,
			}); err != nil {
				return fmt.Errorf("update post: delete obsolete post tags: %w", err)
			}

			if len(tags) > 0 {
				if err := qtx.InsertPostTagsByNames(ctx, db.InsertPostTagsByNamesParams{
					PostID: id,
					Tags:   tags,
				}); err != nil {
					return fmt.Errorf("update post: insert post tags: %w", err)
				}
				if err := s.producer.EnqueueUpdateTagSearchTx(ctx, tx, async.UpdateTagSearchArgs{
					TagNames: tags,
				}); err != nil {
					return fmt.Errorf("update post: enqueue update tag search job: %w", err)
				}
			}
		}
		if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
			PostUID: params.Uid,
			Action:  async.PostSearchActionUpsert,
		}); err != nil {
			return fmt.Errorf("enqueue update post search job: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *PostService) DeletePost(ctx context.Context, uid string, req *api.DeletePostRequest) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)
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
		if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
			PostUID: util.UUID(req.Uid),
			Action:  async.PostSearchActionDelete,
		}); err != nil {
			return fmt.Errorf("enqueue update post search job: %w", err)
		}
		return nil
	})
}

func (s *PostService) LikePost(ctx context.Context, uid string, req *api.LikePostRequest) (*api.LikePostResponse, error) {
	postUid := util.UUID(req.Uid)
	userUid := util.UUID(uid)

	var count int32
	var shouldEnqueue bool

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			affected, err := qtx.InsertPostLikeEdge(ctx, db.InsertPostLikeEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("insert post like edge: %w", err)
			}

			if affected > 0 {
				count, err = qtx.IncrementPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("increment post like count: %w", err)
				}
				shouldEnqueue = true
			} else {
				count, err = qtx.GetPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("get post like count: %w", err)
				}
			}

		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			affected, err := qtx.DeletePostLikeEdge(ctx, db.DeletePostLikeEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("delete post like edge: %w", err)
			}

			if affected > 0 {
				count, err = qtx.DecrementPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("decrement post like count: %w", err)
				}
				shouldEnqueue = true
			} else {
				count, err = qtx.GetPostLikeCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("get post like count: %w", err)
				}
			}

		default:
			return fmt.Errorf("unsupported action: %v", req.Action)
		}
		if shouldEnqueue {
			if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
				PostUID: postUid,
				Action:  async.PostSearchActionUpsert,
			}); err != nil {
				return fmt.Errorf("enqueue update post search job: %w", err)
			}
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
	var shouldEnqueue bool

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			affected, err := qtx.InsertPostCollectionEdge(ctx, db.InsertPostCollectionEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("post collection: insert post collection edge: %w", err)
			}

			if affected > 0 {
				count, err = qtx.IncrementPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: increment post collection count: %w", err)
				}
				shouldEnqueue = true
			} else {
				count, err = qtx.GetPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: get post collection count: %w", err)
				}
			}

		case api.ToggleAction_TOGGLE_ACTION_REMOVE:
			affected, err := qtx.DeletePostCollectionEdge(ctx, db.DeletePostCollectionEdgeParams{
				PostUid: postUid,
				UserUid: userUid,
			})
			if err != nil {
				return fmt.Errorf("post collection: delete post collection edge: %w", err)
			}

			if affected > 0 {
				count, err = qtx.DecrementPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: decrement post collection count: %w", err)
				}
				shouldEnqueue = true
			} else {
				count, err = qtx.GetPostCollectionCount(ctx, postUid)
				if err != nil {
					return fmt.Errorf("post collection: get post collection count: %w", err)
				}
			}

		default:
			return fmt.Errorf("post collection: unsupported action: %v", req.Action)
		}
		if shouldEnqueue {
			if err := s.producer.EnqueueUpdatePostSearchTx(ctx, tx, async.UpdatePostSearchArgs{
				PostUID: postUid,
				Action:  async.PostSearchActionUpsert,
			}); err != nil {
				return fmt.Errorf("enqueue update post search job: %w", err)
			}
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
	if err != nil {
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

type postPageToken struct {
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

type postSearchPageToken struct {
	Offset int64 `json:"offset,omitempty"`
}

func decodePostPageToken(pageToken string) (postPageToken, error) {
	if pageToken == "" {
		return postPageToken{}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return postPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	var token postPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return postPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}
	return token, nil
}

func decodePostSearchPageToken(pageToken string) (postSearchPageToken, error) {
	if pageToken == "" {
		return postSearchPageToken{}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(pageToken)
	if err != nil {
		return postSearchPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	var token postSearchPageToken
	if err := json.Unmarshal(raw, &token); err != nil {
		return postSearchPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
	}
	return token, nil
}

func encodePostPageToken(token postPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func encodePostSearchPageToken(token postSearchPageToken) (string, error) {
	raw, err := json.Marshal(token)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
