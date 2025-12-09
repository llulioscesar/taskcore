package issue

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/start-codex/taskcode/internal/store"
)

func Create(ctx context.Context, db *sqlx.DB, i *Issue) error {
	query := `
		INSERT INTO issues (project_id, issue_type_id, status_id, sprint_id, epic_id, parent_id,
		                    key, summary, description, priority, reporter_id, assignee_id,
		                    story_points, due_date, color, start_date, target_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at`

	return db.QueryRowxContext(ctx, query,
		i.ProjectID, i.IssueTypeID, i.StatusID, i.SprintID, i.EpicID, i.ParentID,
		i.Key, i.Summary, i.Description, i.Priority, i.ReporterID, i.AssigneeID,
		i.StoryPoints, i.DueDate, i.Color, i.StartDate, i.TargetDate,
	).Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
}

func GetByID(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Issue, error) {
	i := &Issue{}
	err := db.GetContext(ctx, i, `SELECT * FROM issues WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return i, err
}

func GetByKey(ctx context.Context, db *sqlx.DB, key string) (*Issue, error) {
	i := &Issue{}
	err := db.GetContext(ctx, i, `SELECT * FROM issues WHERE key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return i, err
}

func ListByProject(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	err := db.SelectContext(ctx, &issues,
		`SELECT * FROM issues WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	return issues, err
}

func ListBySprint(ctx context.Context, db *sqlx.DB, sprintID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	err := db.SelectContext(ctx, &issues,
		`SELECT * FROM issues WHERE sprint_id = $1 ORDER BY priority, created_at`, sprintID)
	return issues, err
}

func ListByAssignee(ctx context.Context, db *sqlx.DB, userID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	err := db.SelectContext(ctx, &issues,
		`SELECT * FROM issues WHERE assignee_id = $1 ORDER BY priority, created_at DESC`, userID)
	return issues, err
}

func ListBacklog(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	err := db.SelectContext(ctx, &issues,
		`SELECT * FROM issues WHERE project_id = $1 AND sprint_id IS NULL ORDER BY priority, created_at`, projectID)
	return issues, err
}

func ListChildren(ctx context.Context, db *sqlx.DB, parentID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	err := db.SelectContext(ctx, &issues,
		`SELECT * FROM issues WHERE parent_id = $1 ORDER BY created_at`, parentID)
	return issues, err
}

// Epic hierarchy functions

func ListEpics(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	query := `
		SELECT i.* FROM issues i
		INNER JOIN issue_types it ON i.issue_type_id = it.id
		WHERE i.project_id = $1 AND it.hierarchy_level = 0
		ORDER BY i.created_at DESC`
	err := db.SelectContext(ctx, &issues, query, projectID)
	return issues, err
}

func ListByEpic(ctx context.Context, db *sqlx.DB, epicID uuid.UUID) ([]*Issue, error) {
	var issues []*Issue
	err := db.SelectContext(ctx, &issues,
		`SELECT * FROM issues WHERE epic_id = $1 ORDER BY priority, created_at`, epicID)
	return issues, err
}

func UpdateEpic(ctx context.Context, db *sqlx.DB, id uuid.UUID, epicID *uuid.UUID) error {
	result, err := db.ExecContext(ctx, `UPDATE issues SET epic_id = $2 WHERE id = $1`, id, epicID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func CountByEpic(ctx context.Context, db *sqlx.DB, epicID uuid.UUID) (total int, done int, err error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE ws.category = 'done') as done
		FROM issues i
		INNER JOIN workflow_statuses ws ON i.status_id = ws.id
		WHERE i.epic_id = $1`
	err = db.QueryRowxContext(ctx, query, epicID).Scan(&total, &done)
	return
}

func GetEpicProgress(ctx context.Context, db *sqlx.DB, epicID uuid.UUID) (int, error) {
	total, done, err := CountByEpic(ctx, db, epicID)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	return (done * 100) / total, nil
}

func Update(ctx context.Context, db *sqlx.DB, i *Issue) error {
	query := `
		UPDATE issues
		SET issue_type_id = $2, status_id = $3, sprint_id = $4, epic_id = $5, parent_id = $6,
		    summary = $7, description = $8, priority = $9, assignee_id = $10,
		    story_points = $11, due_date = $12, color = $13, start_date = $14, target_date = $15
		WHERE id = $1 RETURNING updated_at`

	err := db.QueryRowxContext(ctx, query,
		i.ID, i.IssueTypeID, i.StatusID, i.SprintID, i.EpicID, i.ParentID,
		i.Summary, i.Description, i.Priority, i.AssigneeID,
		i.StoryPoints, i.DueDate, i.Color, i.StartDate, i.TargetDate,
	).Scan(&i.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func UpdateStatus(ctx context.Context, db *sqlx.DB, id, statusID uuid.UUID) error {
	result, err := db.ExecContext(ctx, `UPDATE issues SET status_id = $2 WHERE id = $1`, id, statusID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func UpdateAssignee(ctx context.Context, db *sqlx.DB, id uuid.UUID, assigneeID *uuid.UUID) error {
	result, err := db.ExecContext(ctx, `UPDATE issues SET assignee_id = $2 WHERE id = $1`, id, assigneeID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func UpdateSprint(ctx context.Context, db *sqlx.DB, id uuid.UUID, sprintID *uuid.UUID) error {
	result, err := db.ExecContext(ctx, `UPDATE issues SET sprint_id = $2 WHERE id = $1`, id, sprintID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func Delete(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM issues WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func CreateComment(ctx context.Context, db *sqlx.DB, c *Comment) error {
	query := `
		INSERT INTO issue_comments (issue_id, author_id, content)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	return db.QueryRowxContext(ctx, query, c.IssueID, c.AuthorID, c.Content).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func GetComment(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Comment, error) {
	c := &Comment{}
	err := db.GetContext(ctx, c, `SELECT * FROM issue_comments WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCommentNotFound
	}
	return c, err
}

func ListComments(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*Comment, error) {
	var comments []*Comment
	err := db.SelectContext(ctx, &comments,
		`SELECT * FROM issue_comments WHERE issue_id = $1 ORDER BY created_at`, issueID)
	return comments, err
}

func UpdateComment(ctx context.Context, db *sqlx.DB, c *Comment) error {
	err := db.QueryRowxContext(ctx,
		`UPDATE issue_comments SET content = $2 WHERE id = $1 RETURNING updated_at`,
		c.ID, c.Content).Scan(&c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrCommentNotFound
	}
	return err
}

func DeleteComment(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM issue_comments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCommentNotFound
	}
	return nil
}

func CreateRelation(ctx context.Context, db *sqlx.DB, r *Relation) error {
	query := `
		INSERT INTO issue_relations (source_issue_id, target_issue_id, relation_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (source_issue_id, target_issue_id, relation_type) DO NOTHING
		RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, r.SourceIssueID, r.TargetIssueID, r.RelationType).
		Scan(&r.ID, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}

func DeleteRelation(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	_, err := db.ExecContext(ctx, `DELETE FROM issue_relations WHERE id = $1`, id)
	return err
}

func ListRelations(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*Relation, error) {
	var relations []*Relation
	query := `SELECT * FROM issue_relations WHERE source_issue_id = $1 OR target_issue_id = $1 ORDER BY created_at`
	err := db.SelectContext(ctx, &relations, query, issueID)
	return relations, err
}

func CreateLabel(ctx context.Context, db *sqlx.DB, l *Label) error {
	query := `INSERT INTO labels (project_id, name, color) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := db.QueryRowxContext(ctx, query, l.ProjectID, l.Name, l.Color).Scan(&l.ID, &l.CreatedAt)
	if err != nil && store.IsUniqueViolation(err) {
		return ErrLabelNameExists
	}
	return err
}

func GetLabel(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Label, error) {
	l := &Label{}
	err := db.GetContext(ctx, l, `SELECT * FROM labels WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLabelNotFound
	}
	return l, err
}

func ListLabels(ctx context.Context, db *sqlx.DB, projectID uuid.UUID) ([]*Label, error) {
	var labels []*Label
	err := db.SelectContext(ctx, &labels,
		`SELECT * FROM labels WHERE project_id = $1 ORDER BY name`, projectID)
	return labels, err
}

func DeleteLabel(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM labels WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrLabelNotFound
	}
	return nil
}

func AddLabel(ctx context.Context, db *sqlx.DB, issueID, labelID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO issue_labels (issue_id, label_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		issueID, labelID)
	return err
}

func RemoveLabel(ctx context.Context, db *sqlx.DB, issueID, labelID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM issue_labels WHERE issue_id = $1 AND label_id = $2`, issueID, labelID)
	return err
}

func ListIssueLabels(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*Label, error) {
	var labels []*Label
	query := `
		SELECT l.* FROM labels l
		INNER JOIN issue_labels il ON l.id = il.label_id
		WHERE il.issue_id = $1 ORDER BY l.name`
	err := db.SelectContext(ctx, &labels, query, issueID)
	return labels, err
}

func CreateType(ctx context.Context, db *sqlx.DB, t *Type) error {
	query := `INSERT INTO issue_types (name, description, icon, color, hierarchy_level, is_subtask)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return db.QueryRowxContext(ctx, query, t.Name, t.Description, t.Icon, t.Color, t.HierarchyLevel, t.IsSubtask).Scan(&t.ID)
}

func GetType(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Type, error) {
	t := &Type{}
	err := db.GetContext(ctx, t, `SELECT * FROM issue_types WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTypeNotFound
	}
	return t, err
}

func ListTypes(ctx context.Context, db *sqlx.DB) ([]*Type, error) {
	var types []*Type
	err := db.SelectContext(ctx, &types, `SELECT * FROM issue_types ORDER BY hierarchy_level, name`)
	return types, err
}

func ListTypesByHierarchy(ctx context.Context, db *sqlx.DB, level HierarchyLevel) ([]*Type, error) {
	var types []*Type
	err := db.SelectContext(ctx, &types,
		`SELECT * FROM issue_types WHERE hierarchy_level = $1 ORDER BY name`, level)
	return types, err
}

func ListEpicTypes(ctx context.Context, db *sqlx.DB) ([]*Type, error) {
	return ListTypesByHierarchy(ctx, db, HierarchyEpic)
}

func ListStandardTypes(ctx context.Context, db *sqlx.DB) ([]*Type, error) {
	return ListTypesByHierarchy(ctx, db, HierarchyStandard)
}

func ListSubtaskTypes(ctx context.Context, db *sqlx.DB) ([]*Type, error) {
	return ListTypesByHierarchy(ctx, db, HierarchySubtask)
}

// Watchers

func AddWatcher(ctx context.Context, db *sqlx.DB, issueID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO issue_watchers (issue_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		issueID, userID)
	return err
}

func RemoveWatcher(ctx context.Context, db *sqlx.DB, issueID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM issue_watchers WHERE issue_id = $1 AND user_id = $2`, issueID, userID)
	return err
}

func ListWatchers(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT user_id FROM issue_watchers WHERE issue_id = $1`, issueID)
	return ids, err
}

func IsWatching(ctx context.Context, db *sqlx.DB, issueID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM issue_watchers WHERE issue_id = $1 AND user_id = $2)`,
		issueID, userID)
	return exists, err
}

// Voters

func AddVote(ctx context.Context, db *sqlx.DB, issueID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO issue_voters (issue_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		issueID, userID)
	return err
}

func RemoveVote(ctx context.Context, db *sqlx.DB, issueID, userID uuid.UUID) error {
	_, err := db.ExecContext(ctx,
		`DELETE FROM issue_voters WHERE issue_id = $1 AND user_id = $2`, issueID, userID)
	return err
}

func ListVoters(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := db.SelectContext(ctx, &ids,
		`SELECT user_id FROM issue_voters WHERE issue_id = $1`, issueID)
	return ids, err
}

func CountVotes(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) (int, error) {
	var count int
	err := db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM issue_voters WHERE issue_id = $1`, issueID)
	return count, err
}

// Attachments

func CreateAttachment(ctx context.Context, db *sqlx.DB, a *Attachment) error {
	query := `
		INSERT INTO issue_attachments (issue_id, user_id, filename, path, size, mime_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return db.QueryRowxContext(ctx, query,
		a.IssueID, a.UserID, a.Filename, a.Path, a.Size, a.MimeType,
	).Scan(&a.ID, &a.CreatedAt)
}

func GetAttachment(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*Attachment, error) {
	a := &Attachment{}
	err := db.GetContext(ctx, a, `SELECT * FROM issue_attachments WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrAttachmentNotFound
	}
	return a, err
}

func ListAttachments(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*Attachment, error) {
	var attachments []*Attachment
	err := db.SelectContext(ctx, &attachments,
		`SELECT * FROM issue_attachments WHERE issue_id = $1 ORDER BY created_at`, issueID)
	return attachments, err
}

func DeleteAttachment(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM issue_attachments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAttachmentNotFound
	}
	return nil
}

// Work Logs

func CreateWorkLog(ctx context.Context, db *sqlx.DB, w *WorkLog) error {
	query := `
		INSERT INTO issue_work_logs (issue_id, user_id, time_spent, description, logged_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return db.QueryRowxContext(ctx, query,
		w.IssueID, w.UserID, w.TimeSpent, w.Description, w.LoggedAt,
	).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
}

func GetWorkLog(ctx context.Context, db *sqlx.DB, id uuid.UUID) (*WorkLog, error) {
	w := &WorkLog{}
	err := db.GetContext(ctx, w, `SELECT * FROM issue_work_logs WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrWorkLogNotFound
	}
	return w, err
}

func ListWorkLogs(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) ([]*WorkLog, error) {
	var logs []*WorkLog
	err := db.SelectContext(ctx, &logs,
		`SELECT * FROM issue_work_logs WHERE issue_id = $1 ORDER BY logged_at DESC`, issueID)
	return logs, err
}

func UpdateWorkLog(ctx context.Context, db *sqlx.DB, w *WorkLog) error {
	query := `
		UPDATE issue_work_logs SET time_spent = $2, description = $3, logged_at = $4
		WHERE id = $1 RETURNING updated_at`
	err := db.QueryRowxContext(ctx, query,
		w.ID, w.TimeSpent, w.Description, w.LoggedAt,
	).Scan(&w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrWorkLogNotFound
	}
	return err
}

func DeleteWorkLog(ctx context.Context, db *sqlx.DB, id uuid.UUID) error {
	result, err := db.ExecContext(ctx, `DELETE FROM issue_work_logs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrWorkLogNotFound
	}
	return nil
}

func GetTotalTimeSpent(ctx context.Context, db *sqlx.DB, issueID uuid.UUID) (int, error) {
	var total int
	err := db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(time_spent), 0) FROM issue_work_logs WHERE issue_id = $1`, issueID)
	return total, err
}
