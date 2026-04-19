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
	vuid := util.UUID(uid)

	var resp *api.CreatePostResponse

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)

		tags := util.NormalizeStrings(req.Tags)
		visibility := db.PostVisibility(req.Visibility)
		if visibility != db.PostVisibilityPRIVATE {
			visibility = db.PostVisibilityPUBLIC
		}
		images := []string{}
		if req.Images != nil {
			images = req.Images
		}
		attachments := []string{}
		if req.Attachments != nil {
			attachments = req.Attachments
		}
		row, err := qtx.CreatePost(ctx, db.CreatePostParams{
			Uid:         uuid.New(),
			AuthorUid:   vuid,
			Text:        req.Text,
			Images:      images,
			Attachments: attachments,
			Tags:        tags,
			Visibility:  visibility,
			Pinned:      req.Pinned,
		})
		if err != nil {
			return fmt.Errorf("create post: %w", err)
		}

		if len(tags) > 0 {
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
	vuid := util.UUID(viewerUid)

	postRow, err := s.db.GetPostByUid(ctx, util.UUID(req.Uid))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}

	if postRow.Visibility == db.PostVisibilityPRIVATE && vuid != postRow.AuthorUid {
		return nil, fmt.Errorf("post not found")
	}

	authorRow, err := s.db.GetUserByUid(ctx, postRow.AuthorUid)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("post not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get post author: %w", err)
	}

	isFollowing := false
	liked := false
	collected := false
	if viewerUid != "" {
		if vuid != postRow.AuthorUid {
			isFollowing, err = s.db.IsFollowing(ctx, db.IsFollowingParams{
				FollowerUid: vuid,
				FolloweeUid: postRow.AuthorUid,
			})
			if err != nil {
				return nil, fmt.Errorf("check following: %w", err)
			}
		}
		liked, err = s.db.IsPostLiked(ctx, db.IsPostLikedParams{
			PostUid: postRow.Uid,
			UserUid: vuid,
		})
		if err != nil {
			return nil, fmt.Errorf("check post liked: %w", err)
		}
		collected, err = s.db.IsPostCollected(ctx, db.IsPostCollectedParams{
			PostUid: postRow.Uid,
			UserUid: vuid,
		})
		if err != nil {
			return nil, fmt.Errorf("check post collected: %w", err)
		}
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
			Nickname:    authorRow.Nickname,
			AvatarUrl:   authorRow.AvatarUrl,
			IsFollowing: isFollowing,
		},
		Text:            postRow.Text,
		Images:          postRow.Images,
		Attachments:     attachments,
		Tags:            postRow.Tags,
		CommentCount:    postRow.CommentCount,
		CollectionCount: postRow.CollectionCount,
		LikeCount:       postRow.LikeCount,
		Visibility:      string(postRow.Visibility),
		LatestRepliedOn: postRow.LatestRepliedOn.Time.Unix(),
		Ip:              postRow.Ip,
		Pinned:          postRow.Pinned,
		Liked:           liked,
		Collected:       collected,
		CreatedAt:       postRow.CreatedAt.Time.Unix(),
		UpdatedAt:       postRow.UpdatedAt.Time.Unix(),
	}}, nil
}

func (s *PostService) ListPosts(ctx context.Context, viewerUid string, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	vuid := util.UUID(viewerUid)

	token, err := decodePostPageToken(req.PageToken)
	if err != nil {
		return nil, err
	}
	cursorCreatedAt := pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true}
	cursorID := util.UUID(token.CursorID)

	var rows []db.Post
	switch {
	case req.AuthorUid != "":
		rows, err = s.db.ListPostsByAuthor(ctx, db.ListPostsByAuthorParams{
			AuthorUid:       util.UUID(req.AuthorUid),
			OnlyPublic:      viewerUid == "" || util.UUID(req.AuthorUid) != vuid,
			CursorCreatedAt: cursorCreatedAt,
			CursorID:        cursorID,
		})
	case req.TagName != "":
		rows, err = s.db.ListPostsByTag(ctx, db.ListPostsByTagParams{
			TagName:         req.TagName,
			CursorCreatedAt: cursorCreatedAt,
			CursorID:        cursorID,
		})
	default:
		rows, err = s.db.ListPostsPublic(ctx, db.ListPostsPublicParams{
			CursorCreatedAt: cursorCreatedAt,
			CursorID:        cursorID,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	postUIDs := make([]uuid.UUID, 0, len(rows))
	authorUIDs := make([]uuid.UUID, 0, len(rows))
	attachmentURLs := make([]string, 0)
	seenPostUIDs := make(map[uuid.UUID]struct{}, len(rows))
	seenAuthorUIDs := make(map[uuid.UUID]struct{}, len(rows))
	seenAttachmentURLs := make(map[string]struct{})
	for _, row := range rows {
		if _, ok := seenPostUIDs[row.Uid]; !ok {
			seenPostUIDs[row.Uid] = struct{}{}
			postUIDs = append(postUIDs, row.Uid)
		}
		if _, ok := seenAuthorUIDs[row.AuthorUid]; !ok {
			seenAuthorUIDs[row.AuthorUid] = struct{}{}
			authorUIDs = append(authorUIDs, row.AuthorUid)
		}
		for _, url := range row.Attachments {
			if _, ok := seenAttachmentURLs[url]; ok {
				continue
			}
			seenAttachmentURLs[url] = struct{}{}
			attachmentURLs = append(attachmentURLs, url)
		}
	}

	authorRows, err := s.db.GetUsersByUIDs(ctx, authorUIDs)
	if err != nil {
		return nil, fmt.Errorf("get post authors: %w", err)
	}
	authorMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(authorRows))
	for _, row := range authorRows {
		authorMap[row.Uid] = row
	}

	followingSet := make(map[uuid.UUID]struct{})
	likedSet := make(map[uuid.UUID]struct{})
	collectedSet := make(map[uuid.UUID]struct{})
	if viewerUid != "" && len(rows) > 0 {
		followingUIDs, err := s.db.ListFollowingUIDsByFollowerAndFolloweeUIDs(ctx, db.ListFollowingUIDsByFollowerAndFolloweeUIDsParams{
			FollowerUid:  vuid,
			FolloweeUids: authorUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list following uids: %w", err)
		}
		for _, followeeUID := range followingUIDs {
			followingSet[followeeUID] = struct{}{}
		}

		likedPostUIDs, err := s.db.ListLikedPostUIDsByUserAndPostUIDs(ctx, db.ListLikedPostUIDsByUserAndPostUIDsParams{
			UserUid:  vuid,
			PostUids: postUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list liked post uids: %w", err)
		}
		for _, postUID := range likedPostUIDs {
			likedSet[postUID] = struct{}{}
		}

		collectedPostUIDs, err := s.db.ListCollectedPostUIDsByUserAndPostUIDs(ctx, db.ListCollectedPostUIDsByUserAndPostUIDsParams{
			UserUid:  vuid,
			PostUids: postUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list collected post uids: %w", err)
		}
		for _, postUID := range collectedPostUIDs {
			collectedSet[postUID] = struct{}{}
		}
	}

	files, err := s.db.GetFilesByUrls(ctx, attachmentURLs)
	if err != nil {
		return nil, fmt.Errorf("get attachments: %w", err)
	}
	fileMap := make(map[string]db.GetFilesByUrlsRow, len(files))
	for _, file := range files {
		fileMap[file.Url] = file
	}

	posts := make([]*api.Post, 0, len(rows))
	for _, row := range rows {
		authorRow, ok := authorMap[row.AuthorUid]
		if !ok {
			continue
		}
		_, isFollowing := followingSet[row.AuthorUid]
		_, liked := likedSet[row.Uid]
		_, collected := collectedSet[row.Uid]

		attachments := make([]*api.Attachment, 0, len(row.Attachments))
		for _, url := range row.Attachments {
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

		posts = append(posts, &api.Post{
			Uid: row.Uid.String(),
			Author: &api.PostAuthor{
				Uid:         row.AuthorUid.String(),
				Nickname:    authorRow.Nickname,
				AvatarUrl:   authorRow.AvatarUrl,
				IsFollowing: isFollowing,
			},
			Text:            row.Text,
			Images:          row.Images,
			Attachments:     attachments,
			Tags:            row.Tags,
			CommentCount:    row.CommentCount,
			CollectionCount: row.CollectionCount,
			LikeCount:       row.LikeCount,
			Visibility:      string(row.Visibility),
			LatestRepliedOn: row.LatestRepliedOn.Time.Unix(),
			Ip:              row.Ip,
			Pinned:          row.Pinned,
			Liked:           liked,
			Collected:       collected,
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

func (s *PostService) ListMyCollections(ctx context.Context, uid string, req *api.ListPostsRequest) (*api.ListPostsResponse, error) {
	vuid := util.UUID(uid)

	token, err := decodePostPageToken(req.GetPageToken())
	if err != nil {
		return nil, err
	}
	cursorCreatedAt := pgtype.Timestamptz{Time: time.Unix(token.CursorCreatedAt, 0).UTC(), Valid: true}
	cursorID := util.UUID(token.CursorID)

	refs, err := s.db.ListCollectedPostRefsByUser(ctx, db.ListCollectedPostRefsByUserParams{
		UserUid:         vuid,
		CursorCreatedAt: cursorCreatedAt,
		CursorID:        cursorID,
	})
	if err != nil {
		return nil, fmt.Errorf("list collected post refs: %w", err)
	}

	postUIDs := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		postUIDs = append(postUIDs, ref.PostUid)
	}
	postRows, err := s.db.GetPostsByUIDs(ctx, postUIDs)
	if err != nil {
		return nil, fmt.Errorf("get posts by uids: %w", err)
	}
	postMap := make(map[uuid.UUID]db.Post, len(postRows))
	for _, row := range postRows {
		postMap[row.Uid] = row
	}

	rows := make([]db.Post, 0, len(refs))
	for _, ref := range refs {
		row, ok := postMap[ref.PostUid]
		if !ok {
			continue
		}
		if row.Visibility == db.PostVisibilityPRIVATE && row.AuthorUid != vuid {
			continue
		}
		rows = append(rows, row)
	}

	postUIDs = make([]uuid.UUID, 0, len(rows))
	authorUIDs := make([]uuid.UUID, 0, len(rows))
	attachmentURLs := make([]string, 0)
	seenPostUIDs := make(map[uuid.UUID]struct{}, len(rows))
	seenAuthorUIDs := make(map[uuid.UUID]struct{}, len(rows))
	seenAttachmentURLs := make(map[string]struct{})
	for _, row := range rows {
		if _, ok := seenPostUIDs[row.Uid]; !ok {
			seenPostUIDs[row.Uid] = struct{}{}
			postUIDs = append(postUIDs, row.Uid)
		}
		if _, ok := seenAuthorUIDs[row.AuthorUid]; !ok {
			seenAuthorUIDs[row.AuthorUid] = struct{}{}
			authorUIDs = append(authorUIDs, row.AuthorUid)
		}
		for _, url := range row.Attachments {
			if _, ok := seenAttachmentURLs[url]; ok {
				continue
			}
			seenAttachmentURLs[url] = struct{}{}
			attachmentURLs = append(attachmentURLs, url)
		}
	}

	authorRows, err := s.db.GetUsersByUIDs(ctx, authorUIDs)
	if err != nil {
		return nil, fmt.Errorf("get post authors: %w", err)
	}
	authorMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(authorRows))
	for _, row := range authorRows {
		authorMap[row.Uid] = row
	}

	followingSet := make(map[uuid.UUID]struct{})
	likedSet := make(map[uuid.UUID]struct{})
	collectedSet := make(map[uuid.UUID]struct{})
	if len(rows) > 0 {
		followingUIDs, err := s.db.ListFollowingUIDsByFollowerAndFolloweeUIDs(ctx, db.ListFollowingUIDsByFollowerAndFolloweeUIDsParams{
			FollowerUid:  vuid,
			FolloweeUids: authorUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list following uids: %w", err)
		}
		for _, followeeUID := range followingUIDs {
			followingSet[followeeUID] = struct{}{}
		}

		likedPostUIDs, err := s.db.ListLikedPostUIDsByUserAndPostUIDs(ctx, db.ListLikedPostUIDsByUserAndPostUIDsParams{
			UserUid:  vuid,
			PostUids: postUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list liked post uids: %w", err)
		}
		for _, postUID := range likedPostUIDs {
			likedSet[postUID] = struct{}{}
		}

		collectedPostUIDs, err := s.db.ListCollectedPostUIDsByUserAndPostUIDs(ctx, db.ListCollectedPostUIDsByUserAndPostUIDsParams{
			UserUid:  vuid,
			PostUids: postUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list collected post uids: %w", err)
		}
		for _, postUID := range collectedPostUIDs {
			collectedSet[postUID] = struct{}{}
		}
	}

	files, err := s.db.GetFilesByUrls(ctx, attachmentURLs)
	if err != nil {
		return nil, fmt.Errorf("get attachments: %w", err)
	}
	fileMap := make(map[string]db.GetFilesByUrlsRow, len(files))
	for _, file := range files {
		fileMap[file.Url] = file
	}

	posts := make([]*api.Post, 0, len(rows))
	for _, row := range rows {
		authorRow, ok := authorMap[row.AuthorUid]
		if !ok {
			continue
		}
		_, isFollowing := followingSet[row.AuthorUid]
		_, liked := likedSet[row.Uid]
		_, collected := collectedSet[row.Uid]

		attachments := make([]*api.Attachment, 0, len(row.Attachments))
		for _, url := range row.Attachments {
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

		posts = append(posts, &api.Post{
			Uid: row.Uid.String(),
			Author: &api.PostAuthor{
				Uid:         row.AuthorUid.String(),
				Nickname:    authorRow.Nickname,
				AvatarUrl:   authorRow.AvatarUrl,
				IsFollowing: isFollowing,
			},
			Text:            row.Text,
			Images:          row.Images,
			Attachments:     attachments,
			Tags:            row.Tags,
			CommentCount:    row.CommentCount,
			CollectionCount: row.CollectionCount,
			LikeCount:       row.LikeCount,
			Visibility:      string(row.Visibility),
			LatestRepliedOn: row.LatestRepliedOn.Time.Unix(),
			Ip:              row.Ip,
			Pinned:          row.Pinned,
			Liked:           liked,
			Collected:       collected,
			CreatedAt:       row.CreatedAt.Time.Unix(),
			UpdatedAt:       row.UpdatedAt.Time.Unix(),
		})
	}

	nextPageToken := ""
	if len(refs) > 0 {
		last := refs[len(refs)-1]
		token := postPageToken{
			CursorCreatedAt: last.CollectedAt.Time.Unix(),
			CursorID:        last.PostUid.String(),
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
	vuid := util.UUID(viewerUid)

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
	authorUIDs := make([]uuid.UUID, 0, len(result.Hits))
	attachmentURLs := make([]string, 0)
	seenPostUIDs := make(map[uuid.UUID]struct{}, len(result.Hits))
	seenAuthorUIDs := make(map[uuid.UUID]struct{}, len(result.Hits))
	seenAttachmentURLs := make(map[string]struct{})
	for _, hit := range result.Hits {
		postUID, err := uuid.Parse(hit.UID)
		if err == nil {
			if _, ok := seenPostUIDs[postUID]; !ok {
				seenPostUIDs[postUID] = struct{}{}
				postUIDs = append(postUIDs, postUID)
			}
		}
		authorUID, err := uuid.Parse(hit.AuthorUID)
		if err == nil {
			if _, ok := seenAuthorUIDs[authorUID]; !ok {
				seenAuthorUIDs[authorUID] = struct{}{}
				authorUIDs = append(authorUIDs, authorUID)
			}
		}
		for _, url := range hit.Attachments {
			if _, ok := seenAttachmentURLs[url]; ok {
				continue
			}
			seenAttachmentURLs[url] = struct{}{}
			attachmentURLs = append(attachmentURLs, url)
		}
	}

	authorRows, err := s.db.GetUsersByUIDs(ctx, authorUIDs)
	if err != nil {
		return nil, fmt.Errorf("get post authors: %w", err)
	}
	authorMap := make(map[uuid.UUID]db.GetUsersByUIDsRow, len(authorRows))
	for _, row := range authorRows {
		authorMap[row.Uid] = row
	}

	followingSet := make(map[uuid.UUID]struct{})
	likedSet := make(map[uuid.UUID]struct{})
	collectedSet := make(map[uuid.UUID]struct{})
	if viewerUid != "" && len(result.Hits) > 0 {
		followingUIDs, err := s.db.ListFollowingUIDsByFollowerAndFolloweeUIDs(ctx, db.ListFollowingUIDsByFollowerAndFolloweeUIDsParams{
			FollowerUid:  vuid,
			FolloweeUids: authorUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list following uids: %w", err)
		}
		for _, followeeUID := range followingUIDs {
			followingSet[followeeUID] = struct{}{}
		}

		likedPostUIDs, err := s.db.ListLikedPostUIDsByUserAndPostUIDs(ctx, db.ListLikedPostUIDsByUserAndPostUIDsParams{
			UserUid:  vuid,
			PostUids: postUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list liked post uids: %w", err)
		}
		for _, postUID := range likedPostUIDs {
			likedSet[postUID] = struct{}{}
		}

		collectedPostUIDs, err := s.db.ListCollectedPostUIDsByUserAndPostUIDs(ctx, db.ListCollectedPostUIDsByUserAndPostUIDsParams{
			UserUid:  vuid,
			PostUids: postUIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("list collected post uids: %w", err)
		}
		for _, postUID := range collectedPostUIDs {
			collectedSet[postUID] = struct{}{}
		}
	}

	files, err := s.db.GetFilesByUrls(ctx, attachmentURLs)
	if err != nil {
		return nil, fmt.Errorf("get attachments: %w", err)
	}
	fileMap := make(map[string]db.GetFilesByUrlsRow, len(files))
	for _, file := range files {
		fileMap[file.Url] = file
	}

	posts := make([]*api.Post, 0, len(result.Hits))
	for _, hit := range result.Hits {
		authorUID, authorUIDErr := uuid.Parse(hit.AuthorUID)
		postUID, postUIDErr := uuid.Parse(hit.UID)

		authorNickname := hit.AuthorNickname
		authorAvatarURL := ""
		isFollowing := false
		if authorUIDErr == nil {
			if authorRow, ok := authorMap[authorUID]; ok {
				authorNickname = authorRow.Nickname
				authorAvatarURL = authorRow.AvatarUrl
			}
			_, isFollowing = followingSet[authorUID]
		}
		liked := false
		collected := false
		if postUIDErr == nil {
			_, liked = likedSet[postUID]
			_, collected = collectedSet[postUID]
		}

		attachments := make([]*api.Attachment, 0, len(hit.Attachments))
		for _, url := range hit.Attachments {
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

		posts = append(posts, &api.Post{
			Uid: hit.UID,
			Author: &api.PostAuthor{
				Uid:         hit.AuthorUID,
				Nickname:    authorNickname,
				AvatarUrl:   authorAvatarURL,
				IsFollowing: isFollowing,
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
			Liked:           liked,
			Collected:       collected,
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
			Uid:       util.UUID(req.Uid),
			AuthorUid: util.UUID(uid),
		}
		updatedTags := []string{}
		tagUpdated := false

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
		if _, ok := paths["tags"]; ok {
			updatedTags = util.NormalizeStrings(req.Post.Tags)
			params.Tags = updatedTags
			tagUpdated = true
		}
		if _, ok := paths["visibility"]; ok {
			params.Visibility = db.NullPostVisibility{PostVisibility: db.PostVisibility(req.Post.Visibility), Valid: true}
		}
		if _, ok := paths["pinned"]; ok {
			params.Pinned = pgtype.Bool{Bool: req.Post.Pinned, Valid: true}
		}

		_, err := qtx.UpdatePostByUidAndAuthor(ctx, params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("post not found")
			}
			return fmt.Errorf("update post: %w", err)
		}
		if tagUpdated && len(updatedTags) > 0 {
			if err := s.producer.EnqueueUpdateTagSearchTx(ctx, tx, async.UpdateTagSearchArgs{
				TagNames: updatedTags,
			}); err != nil {
				return fmt.Errorf("enqueue update tag search job: %w", err)
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
	vuid := util.UUID(uid)
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)
		affected, err := qtx.ArchivePostByUidAndAuthor(ctx, db.ArchivePostByUidAndAuthorParams{
			Uid:       util.UUID(req.Uid),
			AuthorUid: vuid,
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
	vuid := util.UUID(uid)

	var count int32
	var shouldEnqueue bool

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)
		postRow, err := qtx.GetPostByUid(ctx, postUid)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("post not found")
		}
		if err != nil {
			return fmt.Errorf("get post: %w", err)
		}
		if postRow.Visibility == db.PostVisibilityPRIVATE && postRow.AuthorUid != vuid {
			return fmt.Errorf("post not found")
		}

		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			affected, err := qtx.InsertPostLikeEdge(ctx, db.InsertPostLikeEdgeParams{
				PostUid: postUid,
				UserUid: vuid,
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
				UserUid: vuid,
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
	vuid := util.UUID(uid)

	var count int32
	var shouldEnqueue bool

	if err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		qtx := s.db.WithTx(tx)
		postRow, err := qtx.GetPostByUid(ctx, postUid)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("post not found")
		}
		if err != nil {
			return fmt.Errorf("get post: %w", err)
		}
		if postRow.Visibility == db.PostVisibilityPRIVATE && postRow.AuthorUid != vuid {
			return fmt.Errorf("post not found")
		}

		switch req.Action {
		case api.ToggleAction_TOGGLE_ACTION_ADD:
			affected, err := qtx.InsertPostCollectionEdge(ctx, db.InsertPostCollectionEdgeParams{
				PostUid: postUid,
				UserUid: vuid,
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
				UserUid: vuid,
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

type postPageToken struct {
	CursorCreatedAt int64  `json:"cursor_created_at,omitempty"`
	CursorID        string `json:"cursor_id,omitempty"`
}

type postSearchPageToken struct {
	Offset int64 `json:"offset,omitempty"`
}

func decodePostPageToken(pageToken string) (postPageToken, error) {
	var token postPageToken
	if pageToken != "" {
		raw, err := base64.RawURLEncoding.DecodeString(pageToken)
		if err != nil {
			return postPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
		if err := json.Unmarshal(raw, &token); err != nil {
			return postPageToken{}, status.Error(codes.InvalidArgument, "invalid page_token")
		}
	}

	if token.CursorCreatedAt == 0 || token.CursorID == "" {
		token.CursorCreatedAt = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
		token.CursorID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
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
