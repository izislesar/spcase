package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/domain"
)

const teamColumns = `id, name, invite_code, captain_id, created_at, updated_at`

type TeamRepository interface {
	Create(context.Context, domain.Team) (domain.Team, error)
	GetByID(context.Context, uuid.UUID) (domain.Team, error)
	GetByInviteCode(context.Context, string) (domain.Team, error)
	GetByUserID(context.Context, uuid.UUID) (domain.Team, domain.TeamMembership, error)
	ListMembers(context.Context, uuid.UUID) ([]domain.TeamMember, error)
	Join(context.Context, uuid.UUID, string) (domain.Team, error)
	Leave(context.Context, uuid.UUID, time.Time) error
	Kick(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	TransferOwnership(context.Context, uuid.UUID, uuid.UUID, time.Time) error
	Disband(context.Context, uuid.UUID, time.Time) error
}

type TeamPostgres struct {
	pool *pgxpool.Pool
}

func NewTeamPostgres(pool *pgxpool.Pool) (*TeamPostgres, error) {
	if pool == nil {
		return nil, errors.New("team repository pool cannot be nil")
	}
	return &TeamPostgres{pool: pool}, nil
}

func (r *TeamPostgres) Create(ctx context.Context, team domain.Team) (domain.Team, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return domain.Team{}, err
	}
	defer tx.Rollback(ctx)

	role, err := lockUser(ctx, tx, team.CaptainID)
	if err != nil {
		return domain.Team{}, err
	}
	if role != domain.RoleUser {
		return domain.Team{}, domain.ErrForbidden
	}
	if exists, err := membershipExists(ctx, tx, team.CaptainID); err != nil {
		return domain.Team{}, err
	} else if exists {
		return domain.Team{}, domain.ErrAlreadyInTeam
	}

	const createQuery = `
		INSERT INTO teams (name, invite_code, captain_id)
		VALUES ($1, $2, $3)
		RETURNING ` + teamColumns
	created, err := scanTeam(tx.QueryRow(ctx, createQuery, team.Name, team.InviteCode, team.CaptainID))
	if err != nil {
		return domain.Team{}, mapTeamWriteError(err)
	}
	const membershipQuery = `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`
	if _, err := tx.Exec(ctx, membershipQuery, created.ID, team.CaptainID); err != nil {
		return domain.Team{}, fmt.Errorf("create captain membership: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Team{}, mapTeamWriteError(err)
	}
	return created, nil
}

func (r *TeamPostgres) GetByID(ctx context.Context, id uuid.UUID) (domain.Team, error) {
	const query = `SELECT ` + teamColumns + ` FROM teams WHERE id = $1`
	team, err := scanTeam(r.pool.QueryRow(ctx, query, id))
	return team, mapTeamReadError("get team by ID", err)
}

func (r *TeamPostgres) GetByInviteCode(ctx context.Context, inviteCode string) (domain.Team, error) {
	const query = `SELECT ` + teamColumns + ` FROM teams WHERE invite_code = $1`
	team, err := scanTeam(r.pool.QueryRow(ctx, query, inviteCode))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrInvalidInviteCode
	}
	return team, mapTeamReadError("get team by invite code", err)
}

func (r *TeamPostgres) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (domain.Team, domain.TeamMembership, error) {
	const query = `
		SELECT ` + teamColumns + `, tm.joined_at
		FROM team_members tm
		JOIN teams t ON t.id = tm.team_id
		WHERE tm.user_id = $1
	`
	var team domain.Team
	var membership domain.TeamMembership
	membership.UserID = userID
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&team.ID, &team.Name, &team.InviteCode, &team.CaptainID, &team.CreatedAt, &team.UpdatedAt,
		&membership.JoinedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.TeamMembership{}, domain.ErrNoTeam
	}
	if err != nil {
		return domain.Team{}, domain.TeamMembership{}, fmt.Errorf("get team by user ID: %w", err)
	}
	membership.TeamID = team.ID
	return team, membership, nil
}

func (r *TeamPostgres) ListMembers(ctx context.Context, teamID uuid.UUID) ([]domain.TeamMember, error) {
	const query = `
		SELECT u.id, u.full_name, COALESCE(u.telegram, ''), u.id = t.captain_id
		FROM team_members tm
		JOIN users u ON u.id = tm.user_id
		JOIN teams t ON t.id = tm.team_id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at, u.id
	`
	rows, err := r.pool.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()
	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.ID, &member.FullName, &member.Telegram, &member.IsCaptain); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}
	return members, nil
}

func (r *TeamPostgres) Join(ctx context.Context, userID uuid.UUID, inviteCode string) (domain.Team, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return domain.Team{}, err
	}
	defer tx.Rollback(ctx)

	team, err := lockTeamByInviteCode(ctx, tx, inviteCode)
	if err != nil {
		return domain.Team{}, err
	}
	role, err := lockUser(ctx, tx, userID)
	if err != nil {
		return domain.Team{}, err
	}
	if role != domain.RoleUser {
		return domain.Team{}, domain.ErrForbidden
	}
	if exists, err := membershipExists(ctx, tx, userID); err != nil {
		return domain.Team{}, err
	} else if exists {
		return domain.Team{}, domain.ErrAlreadyInTeam
	}
	count, err := countTeamMembers(ctx, tx, team.ID)
	if err != nil {
		return domain.Team{}, err
	}
	if !domain.HasCapacity(count) {
		return domain.Team{}, domain.ErrTeamFull
	}
	if _, err := tx.Exec(ctx, `INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)`, team.ID, userID); err != nil {
		return domain.Team{}, mapMembershipWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Team{}, fmt.Errorf("commit team join: %w", err)
	}
	return team, nil
}

func (r *TeamPostgres) Leave(ctx context.Context, userID uuid.UUID, lockAt time.Time) error {
	tx, team, err := r.beginMutationByMember(ctx, userID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if team.CaptainID == userID {
		return domain.ErrCaptainCannotLeave
	}
	if err := lockUsersSorted(ctx, tx, userID); err != nil {
		return err
	}
	exists, err := teamMembershipExists(ctx, tx, team.ID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrNoTeam
	}
	if err := checkMutationLock(ctx, tx, lockAt); err != nil {
		return err
	}
	if err := deleteMembership(ctx, tx, team.ID, userID); err != nil {
		return err
	}
	if err := removeSubmissionIfUndersized(ctx, tx, team.ID); err != nil {
		return err
	}
	return commitMutation(ctx, tx, "leave team")
}

func (r *TeamPostgres) Kick(
	ctx context.Context,
	captainID, memberID uuid.UUID,
	lockAt time.Time,
) error {
	tx, team, err := r.beginMutationByMember(ctx, captainID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if team.CaptainID != captainID {
		return domain.ErrNotTeamCaptain
	}
	if captainID == memberID {
		return domain.ErrCaptainCannotBeKicked
	}
	if err := lockUsersSorted(ctx, tx, captainID, memberID); err != nil {
		return err
	}
	if err := checkMutationLock(ctx, tx, lockAt); err != nil {
		return err
	}
	if err := deleteMembership(ctx, tx, team.ID, memberID); err != nil {
		return err
	}
	if err := removeSubmissionIfUndersized(ctx, tx, team.ID); err != nil {
		return err
	}
	return commitMutation(ctx, tx, "kick team member")
}

func (r *TeamPostgres) TransferOwnership(
	ctx context.Context,
	captainID, newCaptainID uuid.UUID,
	lockAt time.Time,
) error {
	tx, team, err := r.beginMutationByMember(ctx, captainID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if team.CaptainID != captainID {
		return domain.ErrNotTeamCaptain
	}
	if captainID == newCaptainID {
		return domain.ErrInvalidRequest
	}
	if err := lockUsersSorted(ctx, tx, captainID, newCaptainID); err != nil {
		return err
	}
	if err := checkMutationLock(ctx, tx, lockAt); err != nil {
		return err
	}
	exists, err := teamMembershipExists(ctx, tx, team.ID, newCaptainID)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrTeamMemberNotFound
	}
	if _, err := tx.Exec(ctx,
		`UPDATE teams SET captain_id = $2, updated_at = clock_timestamp() WHERE id = $1`,
		team.ID, newCaptainID,
	); err != nil {
		return fmt.Errorf("transfer ownership: %w", err)
	}
	return commitMutation(ctx, tx, "transfer ownership")
}

func (r *TeamPostgres) Disband(ctx context.Context, captainID uuid.UUID, lockAt time.Time) error {
	tx, team, err := r.beginMutationByMember(ctx, captainID)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if team.CaptainID != captainID {
		return domain.ErrNotTeamCaptain
	}
	if err := lockAllTeamUsers(ctx, tx, team.ID); err != nil {
		return err
	}
	if err := checkMutationLock(ctx, tx, lockAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM teams WHERE id = $1`, team.ID); err != nil {
		return fmt.Errorf("disband team: %w", err)
	}
	return commitMutation(ctx, tx, "disband team")
}

func (r *TeamPostgres) beginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin team transaction: %w", err)
	}
	return tx, nil
}

func (r *TeamPostgres) beginMutationByMember(
	ctx context.Context,
	userID uuid.UUID,
) (pgx.Tx, domain.Team, error) {
	tx, err := r.beginTx(ctx)
	if err != nil {
		return nil, domain.Team{}, err
	}
	var teamID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT team_id FROM team_members WHERE user_id = $1`, userID).Scan(&teamID); err != nil {
		tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.Team{}, domain.ErrNoTeam
		}
		return nil, domain.Team{}, fmt.Errorf("resolve member team: %w", err)
	}
	team, err := lockTeamByID(ctx, tx, teamID)
	if err != nil {
		tx.Rollback(ctx)
		return nil, domain.Team{}, err
	}
	return tx, team, nil
}

func lockTeamByInviteCode(ctx context.Context, tx pgx.Tx, inviteCode string) (domain.Team, error) {
	team, err := scanTeam(tx.QueryRow(ctx,
		`SELECT `+teamColumns+` FROM teams WHERE invite_code = $1 FOR UPDATE`, inviteCode,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrInvalidInviteCode
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("lock team by invite code: %w", err)
	}
	return team, nil
}

func lockTeamByID(ctx context.Context, tx pgx.Tx, teamID uuid.UUID) (domain.Team, error) {
	team, err := scanTeam(tx.QueryRow(ctx,
		`SELECT `+teamColumns+` FROM teams WHERE id = $1 FOR UPDATE`, teamID,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Team{}, domain.ErrTeamNotFound
	}
	if err != nil {
		return domain.Team{}, fmt.Errorf("lock team by ID: %w", err)
	}
	return team, nil
}

func lockUser(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (domain.Role, error) {
	var role domain.Role
	if err := tx.QueryRow(ctx, `SELECT role FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrUserNotFound
		}
		return "", fmt.Errorf("lock user: %w", err)
	}
	return role, nil
}

func lockUsersSorted(ctx context.Context, tx pgx.Tx, userIDs ...uuid.UUID) error {
	unique := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, id := range userIDs {
		unique[id] = struct{}{}
	}
	sorted := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		sorted = append(sorted, id)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].String() < sorted[j].String() })
	for _, id := range sorted {
		if _, err := lockUser(ctx, tx, id); err != nil {
			return err
		}
	}
	return nil
}

func lockAllTeamUsers(ctx context.Context, tx pgx.Tx, teamID uuid.UUID) error {
	const query = `
		SELECT u.id
		FROM team_members tm
		JOIN users u ON u.id = tm.user_id
		WHERE tm.team_id = $1
		ORDER BY u.id
		FOR UPDATE OF u
	`
	rows, err := tx.Query(ctx, query, teamID)
	if err != nil {
		return fmt.Errorf("lock team users: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var ignored uuid.UUID
		if err := rows.Scan(&ignored); err != nil {
			return fmt.Errorf("scan locked team user: %w", err)
		}
	}
	return rows.Err()
}

func checkMutationLock(ctx context.Context, tx pgx.Tx, lockAt time.Time) error {
	var locked bool
	if err := tx.QueryRow(ctx, `SELECT clock_timestamp() >= $1`, lockAt.UTC()).Scan(&locked); err != nil {
		return fmt.Errorf("check mutation lock: %w", err)
	}
	if locked {
		return domain.ErrMutationsLocked
	}
	return nil
}

func membershipExists(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM team_members WHERE user_id = $1)`, userID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check membership: %w", err)
	}
	return exists, nil
}

func teamMembershipExists(ctx context.Context, tx pgx.Tx, teamID, userID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM team_members WHERE team_id = $1 AND user_id = $2)`,
		teamID, userID,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("check team membership: %w", err)
	}
	return exists, nil
}

func countTeamMembers(ctx context.Context, tx pgx.Tx, teamID uuid.UUID) (int, error) {
	var count int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM team_members WHERE team_id = $1`, teamID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count team members: %w", err)
	}
	return count, nil
}

func deleteMembership(ctx context.Context, tx pgx.Tx, teamID, userID uuid.UUID) error {
	commandTag, err := tx.Exec(ctx,
		`DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`, teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete team membership: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return domain.ErrTeamMemberNotFound
	}
	return nil
}

func removeSubmissionIfUndersized(ctx context.Context, tx pgx.Tx, teamID uuid.UUID) error {
	const query = `
		DELETE FROM submissions
		WHERE team_id = $1
		  AND (SELECT COUNT(*) FROM team_members WHERE team_id = $1) < $2
	`
	if _, err := tx.Exec(ctx, query, teamID, domain.MinTeamMembers); err != nil {
		return fmt.Errorf("remove invalidated submission: %w", err)
	}
	return nil
}

func commitMutation(ctx context.Context, tx pgx.Tx, operation string) error {
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s: %w", operation, err)
	}
	return nil
}

func scanTeam(row pgx.Row) (domain.Team, error) {
	var team domain.Team
	err := row.Scan(
		&team.ID, &team.Name, &team.InviteCode, &team.CaptainID, &team.CreatedAt, &team.UpdatedAt,
	)
	return team, err
}

func mapTeamReadError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrTeamNotFound
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func mapTeamWriteError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		constraint := strings.ToLower(postgresError.ConstraintName)
		switch {
		case strings.Contains(constraint, "invite"):
			return domain.ErrInviteCodeCollision
		case strings.Contains(constraint, "name"):
			return domain.ErrTeamNameAlreadyExists
		case strings.Contains(constraint, "team_members_user"):
			return domain.ErrAlreadyInTeam
		}
	}
	return fmt.Errorf("write team: %w", err)
}

func mapMembershipWriteError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return domain.ErrAlreadyInTeam
	}
	return fmt.Errorf("write team membership: %w", err)
}
