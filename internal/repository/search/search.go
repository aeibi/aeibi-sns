package search

import (
	"fmt"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type Search struct {
	client meilisearch.ServiceManager
}

func New(client meilisearch.ServiceManager) *Search {
	return &Search{client: client}
}

func (s *Search) Close() {
	s.client.Close()
}

func (s *Search) Setup() error {
	if err := s.setupPosts(); err != nil {
		return err
	}
	if err := s.setupUsers(); err != nil {
		return err
	}
	if err := s.setupTags(); err != nil {
		return err
	}
	return nil
}

func (s *Search) ensureIndex(uid, primaryKey string) error {
	_, err := s.client.GetIndex(uid)
	if err == nil {
		return nil
	}

	task, err := s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        uid,
		PrimaryKey: primaryKey,
	})
	if err != nil {
		return err
	}
	return s.waitTaskSucceeded(task)
}

func (s *Search) waitTaskSucceeded(taskInfo *meilisearch.TaskInfo) error {
	if taskInfo == nil {
		return fmt.Errorf("nil meilisearch task")
	}

	task, err := s.client.WaitForTask(taskInfo.TaskUID, time.Second)
	if err != nil {
		return fmt.Errorf("wait meilisearch task %d: %w", taskInfo.TaskUID, err)
	}

	if task == nil {
		return fmt.Errorf("meilisearch task %d returned nil task", taskInfo.TaskUID)
	}

	taskID := task.UID
	if taskID == 0 {
		taskID = task.TaskUID
	}

	if task.Status != meilisearch.TaskStatusSucceeded {
		if task.Error.Code != "" || task.Error.Type != "" || task.Error.Message != "" || task.Error.Link != "" {
			return fmt.Errorf(
				"meilisearch task %d status=%s code=%s type=%s message=%s link=%s",
				taskID,
				task.Status,
				task.Error.Code,
				task.Error.Type,
				task.Error.Message,
				task.Error.Link,
			)
		}
		return fmt.Errorf("meilisearch task %d status=%s", taskID, task.Status)
	}

	return nil
}
