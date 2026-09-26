package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStore struct{ pool *pgxpool.Pool }

const resolveSessionQuery = `SELECT s.issuer,s.subject,u.email,u.display_name,u.local_user_id,u.role,u.status FROM auth_sessions s JOIN auth_users u ON u.issuer=s.issuer AND u.subject=s.subject WHERE s.token_hash=$1 AND s.expires_at>now()`

func NewPGStore(pool *pgxpool.Pool) *PGStore { return &PGStore{pool: pool} }

func (s *PGStore) ResolveIdentity(ctx context.Context, c Claims) (Principal, error) {
	var p Principal
	err := s.pool.QueryRow(ctx, `SELECT issuer,subject,email,display_name,local_user_id,role,status FROM auth_users WHERE issuer=$1 AND subject=$2`, c.Issuer, c.Subject).Scan(&p.Issuer, &p.Subject, &p.Email, &p.Name, &p.LocalUserID, &p.Role, &p.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Principal{}, ErrMembershipDenied
	}
	if err != nil {
		return Principal{}, err
	}
	if p.Status != "active" {
		return Principal{}, ErrMembershipDenied
	}
	return p, nil
}
func (s *PGStore) CreateSession(ctx context.Context, r SessionRecord) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO auth_sessions (token_hash,issuer,subject,local_user_id,expires_at) VALUES ($1,$2,$3,$4,$5)`, r.TokenHash, r.Principal.Issuer, r.Principal.Subject, r.Principal.LocalUserID, r.ExpiresAt)
	return err
}
func (s *PGStore) ResolveSession(ctx context.Context, hash []byte) (Principal, error) {
	var p Principal
	// Read the mapping from auth_users, rather than the historic session row:
	// owner-operated remaps take effect immediately just like status revocation.
	err := s.pool.QueryRow(ctx, resolveSessionQuery, hash).Scan(&p.Issuer, &p.Subject, &p.Email, &p.Name, &p.LocalUserID, &p.Role, &p.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return Principal{}, ErrUnauthenticated
	}
	return p, err
}
func (s *PGStore) DeleteSession(ctx context.Context, hash []byte) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash=$1`, hash)
	return err
}
func (s *PGStore) DeleteExpiredSessions(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM auth_sessions WHERE expires_at<=now()`)
	return err
}
func (s *PGStore) ProvisionIdentity(ctx context.Context, c Claims, localUserID, role string) error {
	if c.Issuer == "" || c.Subject == "" || localUserID == "" || role == "" {
		return errors.New("issuer, subject, local user ID, and role are required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var previousLocalUserID string
	err = tx.QueryRow(ctx, `SELECT local_user_id FROM auth_users WHERE issuer=$1 AND subject=$2 FOR UPDATE`, c.Issuer, c.Subject).Scan(&previousLocalUserID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		_, err = tx.Exec(ctx, `INSERT INTO auth_users (issuer,subject,email,display_name,local_user_id,role,status) VALUES ($1,$2,$3,$4,$5,$6,'active')`, c.Issuer, c.Subject, c.Email, c.Name, localUserID, role)
	} else {
		if previousLocalUserID != localUserID {
			// A remap changes account ownership. Do not let an old browser session
			// continue, even with the new local ID, after that security boundary.
			if _, err = tx.Exec(ctx, `DELETE FROM auth_sessions WHERE issuer=$1 AND subject=$2`, c.Issuer, c.Subject); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx, `UPDATE auth_users SET email=$3,display_name=$4,local_user_id=$5,role=$6,status='active',updated_at=now() WHERE issuer=$1 AND subject=$2`, c.Issuer, c.Subject, c.Email, c.Name, localUserID, role)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *PGStore) RevokeIdentity(ctx context.Context, issuer, subject string) error {
	_, err := s.pool.Exec(ctx, `UPDATE auth_users SET status='revoked' WHERE issuer=$1 AND subject=$2`, issuer, subject)
	return err
}
func (s *PGStore) DeleteExpired(ctx context.Context) error { return s.DeleteExpiredSessions(ctx) }
