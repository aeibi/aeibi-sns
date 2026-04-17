package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ListPostsParams struct {
	Viewer          uuid.NullUUID
	AuthorUid       uuid.NullUUID
	TagName         pgtype.Text
	CursorCreatedAt pgtype.Timestamptz
	CursorID        uuid.NullUUID
}

type ListPostsRow struct {
	Uid             uuid.UUID
	Author          uuid.UUID
	AuthorUid       uuid.UUID
	AuthorNickname  string
	AuthorAvatarUrl string
	Text            string
	Images          []string
	Attachments     []string
	CommentCount    int32
	CollectionCount int32
	LikeCount       int32
	Pinned          bool
	Visibility      PostVisibility
	LatestRepliedOn pgtype.Timestamptz
	Ip              string
	Status          PostStatus
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	Liked           bool
	Collected       bool
	Following       bool
	TagNames        []string
}

func (q *Queries) ListPosts(ctx context.Context, arg ListPostsParams) ([]ListPostsRow, error) {
	var (
		listPosts string
		params    []interface{}
	)
	cursorCreatedAt := arg.CursorCreatedAt
	cursorID := arg.CursorID
	if !cursorCreatedAt.Valid || !cursorID.Valid {
		cursorCreatedAt = pgtype.Timestamptz{
			Time:  time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
			Valid: true,
		}
		cursorID = uuid.NullUUID{
			UUID:  uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			Valid: true,
		}
	}

	switch {
	case arg.AuthorUid.Valid:
		listPosts = listPostsByAuthor
		params = []interface{}{
			arg.Viewer,
			arg.AuthorUid.UUID,
			cursorCreatedAt,
			cursorID.UUID,
		}
	case arg.TagName.Valid:
		listPosts = listPostsByTag
		params = []interface{}{
			arg.Viewer,
			arg.TagName.String,
			cursorCreatedAt,
			cursorID.UUID,
		}
	default:
		listPosts = listPostsPublic
		params = []interface{}{
			arg.Viewer,
			cursorCreatedAt,
			cursorID.UUID,
		}
	}

	rows, err := q.db.Query(ctx, listPosts, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListPostsRow
	for rows.Next() {
		var i ListPostsRow
		if err := rows.Scan(
			&i.Uid,
			&i.Author,
			&i.AuthorUid,
			&i.AuthorNickname,
			&i.AuthorAvatarUrl,
			&i.Text,
			&i.Images,
			&i.Attachments,
			&i.CommentCount,
			&i.CollectionCount,
			&i.LikeCount,
			&i.Pinned,
			&i.Visibility,
			&i.LatestRepliedOn,
			&i.Ip,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Liked,
			&i.Collected,
			&i.Following,
			&i.TagNames,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
