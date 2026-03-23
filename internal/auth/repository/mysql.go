package repository

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/ReadingGarden/back-go/internal/auth/entity"
)

var ErrNotFound = errors.New("not found")

type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db}
}

func (r *MySQLRepository) FindUserByEmail(ctx context.Context, email string) (entity.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_no, user_nick, user_email, user_password, user_fcm, user_social_id,
		       user_social_type, user_image, user_auth_number, user_created_at
		FROM USER
		WHERE user_email = ?
		LIMIT 1
	`, email)

	return scanUser(row)
}

func (r *MySQLRepository) FindUserBySocial(ctx context.Context, socialID, socialType string) (entity.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_no, user_nick, user_email, user_password, user_fcm, user_social_id,
		       user_social_type, user_image, user_auth_number, user_created_at
		FROM USER
		WHERE user_social_id = ? AND user_social_type = ?
		LIMIT 1
	`, socialID, socialType)

	return scanUser(row)
}

func (r *MySQLRepository) FindUserByNo(ctx context.Context, userNo int64) (entity.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_no, user_nick, user_email, user_password, user_fcm, user_social_id,
		       user_social_type, user_image, user_auth_number, user_created_at
		FROM USER
		WHERE user_no = ?
		LIMIT 1
	`, userNo)

	return scanUser(row)
}

func (r *MySQLRepository) CreateUserGraph(ctx context.Context, payload entity.CreateUserPayload, nick string) (entity.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.User{}, err
	}
	defer rollback(tx)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO USER (
			user_nick, user_email, user_password, user_fcm, user_social_id, user_social_type
		) VALUES (?, ?, ?, ?, ?, ?)
	`, nick, payload.UserEmail, payload.UserPassword, payload.UserFCM, payload.UserSocialID, payload.UserSocialType)
	if err != nil {
		return entity.User{}, err
	}

	userNo, err := result.LastInsertId()
	if err != nil {
		return entity.User{}, err
	}

	gardenResult, err := tx.ExecContext(ctx, `
		INSERT INTO GARDEN (garden_title, garden_info, garden_color)
		VALUES (?, ?, ?)
	`, nick+"의 가든", "독서가든에 오신걸 환영합니다☺️", "green")
	if err != nil {
		return entity.User{}, err
	}

	gardenNo, err := gardenResult.LastInsertId()
	if err != nil {
		return entity.User{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO GARDEN_USER (garden_no, user_no, garden_leader, garden_main)
		VALUES (?, ?, ?, ?)
	`, gardenNo, userNo, true, true); err != nil {
		return entity.User{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO PUSH (user_no, push_app_ok)
		VALUES (?, ?)
	`, userNo, true); err != nil {
		return entity.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return entity.User{}, err
	}

	return r.FindUserByNo(ctx, userNo)
}

func (r *MySQLRepository) UpdateUserFCM(ctx context.Context, userNo int64, fcm string) (entity.User, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE USER SET user_fcm = ? WHERE user_no = ?`, fcm, userNo); err != nil {
		return entity.User{}, err
	}

	return r.FindUserByNo(ctx, userNo)
}

func (r *MySQLRepository) ClearUserFCM(ctx context.Context, userNo int64) (entity.User, error) {
	return r.UpdateUserFCM(ctx, userNo, "")
}

func (r *MySQLRepository) ReplaceRefreshToken(ctx context.Context, userNo int64, token string, exp time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `DELETE FROM REFRESH_TOKEN WHERE user_no = ?`, userNo); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO REFRESH_TOKEN (user_no, token, exp)
		VALUES (?, ?, ?)
	`, userNo, token, exp); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MySQLRepository) FindRefreshToken(ctx context.Context, userNo int64, token string) (entity.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_no, token, exp
		FROM REFRESH_TOKEN
		WHERE user_no = ? AND token = ?
		LIMIT 1
	`, userNo, token)

	var refreshToken entity.RefreshToken
	if err := row.Scan(&refreshToken.ID, &refreshToken.UserNo, &refreshToken.Token, &refreshToken.Exp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RefreshToken{}, ErrNotFound
		}
		return entity.RefreshToken{}, err
	}

	return refreshToken, nil
}

func (r *MySQLRepository) DeleteRefreshTokenByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM REFRESH_TOKEN WHERE id = ?`, id)
	return err
}

func (r *MySQLRepository) FindFirstRefreshTokenByUserNo(ctx context.Context, userNo int64) (entity.RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_no, token, exp
		FROM REFRESH_TOKEN
		WHERE user_no = ?
		LIMIT 1
	`, userNo)

	var refreshToken entity.RefreshToken
	if err := row.Scan(&refreshToken.ID, &refreshToken.UserNo, &refreshToken.Token, &refreshToken.Exp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RefreshToken{}, ErrNotFound
		}
		return entity.RefreshToken{}, err
	}

	return refreshToken, nil
}

func (r *MySQLRepository) FindPushByUserNo(ctx context.Context, userNo int64) (entity.Push, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT user_no, push_app_ok, push_book_ok, push_time
		FROM PUSH
		WHERE user_no = ?
		LIMIT 1
	`, userNo)

	var push entity.Push
	var pushTime sql.NullTime
	if err := row.Scan(&push.UserNo, &push.PushAppOK, &push.PushBookOK, &pushTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Push{}, ErrNotFound
		}
		return entity.Push{}, err
	}
	if pushTime.Valid {
		push.PushTime = &pushTime.Time
	}

	return push, nil
}

func (r *MySQLRepository) UpdateUserAuthNumber(ctx context.Context, userNo int64, authNumber *string) (entity.User, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE USER SET user_auth_number = ? WHERE user_no = ?`, authNumber, userNo); err != nil {
		return entity.User{}, err
	}

	return r.FindUserByNo(ctx, userNo)
}

func (r *MySQLRepository) UpdateUserPassword(ctx context.Context, userNo int64, hashedPassword string) (entity.User, error) {
	if _, err := r.db.ExecContext(ctx, `UPDATE USER SET user_password = ? WHERE user_no = ?`, hashedPassword, userNo); err != nil {
		return entity.User{}, err
	}

	return r.FindUserByNo(ctx, userNo)
}

func (r *MySQLRepository) UpdateUserProfile(ctx context.Context, userNo int64, nick, image *string) (entity.User, error) {
	if nick != nil && *nick != "" {
		if _, err := r.db.ExecContext(ctx, `UPDATE USER SET user_nick = ? WHERE user_no = ?`, *nick, userNo); err != nil {
			return entity.User{}, err
		}
	} else {
		if _, err := r.db.ExecContext(ctx, `UPDATE USER SET user_image = ? WHERE user_no = ?`, imageValue(image), userNo); err != nil {
			return entity.User{}, err
		}
	}

	return r.FindUserByNo(ctx, userNo)
}

func (r *MySQLRepository) CountGardenUsersByUser(ctx context.Context, userNo int64) (int, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM GARDEN_USER WHERE user_no = ?`, userNo)
}

func (r *MySQLRepository) CountBooksByStatus(ctx context.Context, userNo int64, status int) (int, error) {
	return r.count(ctx, `SELECT COUNT(*) FROM BOOK WHERE user_no = ? AND book_status = ?`, userNo, status)
}

func (r *MySQLRepository) DeleteUserGraph(ctx context.Context, user entity.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer rollback(tx)

	refreshToken, err := r.findFirstRefreshTokenByUserNoTx(ctx, tx, user.UserNo)
	refreshTokenExists := err == nil
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	push, err := r.findPushByUserNoTx(ctx, tx, user.UserNo)
	pushExists := err == nil
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	gardenUsers, err := r.listGardenUsersByUserTx(ctx, tx, user.UserNo)
	if err != nil {
		return err
	}

	for _, gardenUser := range gardenUsers {
		gardenUserCount, err := r.countTx(ctx, tx, `SELECT COUNT(*) FROM GARDEN_USER WHERE garden_no = ?`, gardenUser.GardenNo)
		if err != nil {
			return err
		}

		if gardenUser.GardenLeader {
			if gardenUserCount > 1 {
				secondLeader, err := r.findSecondLeaderTx(ctx, tx, gardenUser.GardenNo, user.UserNo)
				if err != nil {
					return err
				}
				if _, err := tx.ExecContext(ctx, `UPDATE GARDEN_USER SET garden_leader = ? WHERE id = ?`, true, secondLeader.ID); err != nil {
					return err
				}
			} else {
				if _, err := tx.ExecContext(ctx, `DELETE FROM GARDEN_USER WHERE id = ?`, gardenUser.ID); err != nil {
					return err
				}
			}
		} else {
			if _, err := tx.ExecContext(ctx, `DELETE FROM GARDEN_USER WHERE id = ?`, gardenUser.ID); err != nil {
				return err
			}
		}
	}

	books, err := r.listBooksByUserTx(ctx, tx, user.UserNo)
	if err != nil {
		return err
	}
	for _, book := range books {
		if _, err := tx.ExecContext(ctx, `DELETE FROM BOOK_READ WHERE book_no = ?`, book.BookNo); err != nil {
			return err
		}
		bookImage, err := r.findBookImageByBookNoTx(ctx, tx, book.BookNo)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if err == nil {
			removeImageFile(bookImage.ImageURL)
			if _, err := tx.ExecContext(ctx, `DELETE FROM BOOK_IMAGE WHERE id = ?`, bookImage.ID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM BOOK WHERE book_no = ?`, book.BookNo); err != nil {
			return err
		}
	}

	memos, err := r.listMemosByUserTx(ctx, tx, user.UserNo)
	if err != nil {
		return err
	}
	for _, memo := range memos {
		memoImage, err := r.findMemoImageByMemoNoTx(ctx, tx, memo.ID)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if err == nil {
			removeImageFile(memoImage.ImageURL)
			if _, err := tx.ExecContext(ctx, `DELETE FROM MEMO_IMAGE WHERE id = ?`, memoImage.ID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM MEMO WHERE id = ?`, memo.ID); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM USER WHERE user_no = ?`, user.UserNo); err != nil {
		return err
	}
	if refreshTokenExists && refreshToken.ID != 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM REFRESH_TOKEN WHERE id = ?`, refreshToken.ID); err != nil {
			return err
		}
	}
	if pushExists && push.UserNo != 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM PUSH WHERE user_no = ?`, push.UserNo); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLRepository) count(ctx context.Context, query string, args ...interface{}) (int, error) {
	return r.countTx(ctx, r.db, query, args...)
}

func (r *MySQLRepository) countTx(ctx context.Context, queryer queryer, query string, args ...interface{}) (int, error) {
	row := queryer.QueryRowContext(ctx, query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *MySQLRepository) listGardenUsersByUserTx(ctx context.Context, tx *sql.Tx, userNo int64) ([]entity.GardenUser, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, garden_no, user_no, garden_leader, garden_main, garden_sign_date
		FROM GARDEN_USER
		WHERE user_no = ?
	`, userNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gardenUsers []entity.GardenUser
	for rows.Next() {
		var gardenUser entity.GardenUser
		if err := rows.Scan(
			&gardenUser.ID,
			&gardenUser.GardenNo,
			&gardenUser.UserNo,
			&gardenUser.GardenLeader,
			&gardenUser.GardenMain,
			&gardenUser.GardenSignDate,
		); err != nil {
			return nil, err
		}
		gardenUsers = append(gardenUsers, gardenUser)
	}

	return gardenUsers, rows.Err()
}

func (r *MySQLRepository) findSecondLeaderTx(ctx context.Context, tx *sql.Tx, gardenNo, userNo int64) (entity.GardenUser, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, garden_no, user_no, garden_leader, garden_main, garden_sign_date
		FROM GARDEN_USER
		WHERE garden_no = ? AND user_no != ?
		ORDER BY garden_sign_date ASC
		LIMIT 1
	`, gardenNo, userNo)

	var gardenUser entity.GardenUser
	if err := row.Scan(
		&gardenUser.ID,
		&gardenUser.GardenNo,
		&gardenUser.UserNo,
		&gardenUser.GardenLeader,
		&gardenUser.GardenMain,
		&gardenUser.GardenSignDate,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.GardenUser{}, ErrNotFound
		}
		return entity.GardenUser{}, err
	}

	return gardenUser, nil
}

func (r *MySQLRepository) listBooksByUserTx(ctx context.Context, tx *sql.Tx, userNo int64) ([]entity.Book, error) {
	rows, err := tx.QueryContext(ctx, `SELECT book_no, user_no FROM BOOK WHERE user_no = ?`, userNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []entity.Book
	for rows.Next() {
		var book entity.Book
		if err := rows.Scan(&book.BookNo, &book.UserNo); err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, rows.Err()
}

func (r *MySQLRepository) findBookImageByBookNoTx(ctx context.Context, tx *sql.Tx, bookNo int64) (entity.BookImage, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, book_no, image_name, image_url
		FROM BOOK_IMAGE
		WHERE book_no = ?
		LIMIT 1
	`, bookNo)

	var image entity.BookImage
	if err := row.Scan(&image.ID, &image.BookNo, &image.ImageName, &image.ImageURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.BookImage{}, ErrNotFound
		}
		return entity.BookImage{}, err
	}

	return image, nil
}

func (r *MySQLRepository) listMemosByUserTx(ctx context.Context, tx *sql.Tx, userNo int64) ([]entity.Memo, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id, user_no FROM MEMO WHERE user_no = ?`, userNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memos []entity.Memo
	for rows.Next() {
		var memo entity.Memo
		if err := rows.Scan(&memo.ID, &memo.UserNo); err != nil {
			return nil, err
		}
		memos = append(memos, memo)
	}

	return memos, rows.Err()
}

func (r *MySQLRepository) findMemoImageByMemoNoTx(ctx context.Context, tx *sql.Tx, memoNo int64) (entity.MemoImage, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, memo_no, image_name, image_url
		FROM MEMO_IMAGE
		WHERE memo_no = ?
		LIMIT 1
	`, memoNo)

	var image entity.MemoImage
	if err := row.Scan(&image.ID, &image.MemoNo, &image.ImageName, &image.ImageURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.MemoImage{}, ErrNotFound
		}
		return entity.MemoImage{}, err
	}

	return image, nil
}

func (r *MySQLRepository) findFirstRefreshTokenByUserNoTx(ctx context.Context, tx *sql.Tx, userNo int64) (entity.RefreshToken, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT id, user_no, token, exp
		FROM REFRESH_TOKEN
		WHERE user_no = ?
		LIMIT 1
	`, userNo)

	var refreshToken entity.RefreshToken
	if err := row.Scan(&refreshToken.ID, &refreshToken.UserNo, &refreshToken.Token, &refreshToken.Exp); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.RefreshToken{}, ErrNotFound
		}
		return entity.RefreshToken{}, err
	}

	return refreshToken, nil
}

func (r *MySQLRepository) findPushByUserNoTx(ctx context.Context, tx *sql.Tx, userNo int64) (entity.Push, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT user_no, push_app_ok, push_book_ok, push_time
		FROM PUSH
		WHERE user_no = ?
		LIMIT 1
	`, userNo)

	var push entity.Push
	var pushTime sql.NullTime
	if err := row.Scan(&push.UserNo, &push.PushAppOK, &push.PushBookOK, &pushTime); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Push{}, ErrNotFound
		}
		return entity.Push{}, err
	}
	if pushTime.Valid {
		push.PushTime = &pushTime.Time
	}

	return push, nil
}

func scanUser(row *sql.Row) (entity.User, error) {
	var user entity.User
	var authNumber sql.NullString
	if err := row.Scan(
		&user.UserNo,
		&user.UserNick,
		&user.UserEmail,
		&user.UserPassword,
		&user.UserFCM,
		&user.UserSocialID,
		&user.UserSocialType,
		&user.UserImage,
		&authNumber,
		&user.UserCreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, ErrNotFound
		}
		return entity.User{}, err
	}
	if authNumber.Valid {
		user.UserAuthNumber = &authNumber.String
	}

	return user, nil
}

func removeImageFile(imageURL string) {
	if err := os.Remove(filepath.Join("images", imageURL)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return
	}
}

func imageValue(value *string) interface{} {
	if value == nil {
		return nil
	}

	return *value
}

func rollback(tx *sql.Tx) {
	_ = tx.Rollback()
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}
