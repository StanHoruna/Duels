package service

import (
	"context"
	"duels-api/internal/model"
	"duels-api/internal/storage/cache"
	"duels-api/internal/storage/repository"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"duels-api/pkg/mtype"
	repo "duels-api/pkg/repository"
	"encoding/base64"
	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"mime/multipart"
)

type UserService struct {
	FileService        *FileService
	UserRepository     *repository.UserRepository
	DuelRepository     *repository.DuelRepository
	JWTStorage         *cache.JWTStorage
	JWTAuth            auth.JWTAuthenticator
	TransactionManager *repo.TransactionManager
}

func NewUserService(
	fileService *FileService,
	userRepository *repository.UserRepository,
	duelRepository *repository.DuelRepository,
	jwtStorage *cache.JWTStorage,
	jwtAuth auth.JWTAuthenticator,
	transactionManager *repo.TransactionManager,
) *UserService {
	return &UserService{
		UserRepository:     userRepository,
		DuelRepository:     duelRepository,
		JWTStorage:         jwtStorage,
		JWTAuth:            jwtAuth,
		TransactionManager: transactionManager,
		FileService:        fileService,
	}
}

func (s *UserService) SignInWithWallet(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (*model.User, error) {
	user, err := s.GetByPublicAddress(ctx, authWallet)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.CreateWithWallet(ctx, authWallet)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) GetByPublicAddress(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (*model.User, error) {
	ok, err := s.VerifySecret(authWallet)
	if err != nil || !ok {
		return nil, err
	}

	user, err := s.UserRepository.GetByPublicAddress(ctx, authWallet.Address)
	if err != nil {
		return nil, apperrors.NotFound("user with public address not found", err)
	}

	return user, nil
}

func (s *UserService) VerifySecret(
	authWallet model.AuthWithWallet,
) (bool, error) {
	publicKey, err := solana.PublicKeyFromBase58(authWallet.Address)
	if err != nil {
		return false, apperrors.BadRequest("invalid solana address", err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(authWallet.Secret)
	if err != nil {
		return false, apperrors.BadRequest("invalid signature format", err)
	}

	var signature solana.Signature
	copy(signature[:], signatureBytes)

	verified := signature.Verify(publicKey, []byte(authWallet.Address))
	if !verified {
		return false, apperrors.BadRequest("invalid signature")
	}

	return true, nil
}

func (s *UserService) CreateWithWallet(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (*model.User, error) {
	username, err := generateUsername("user")
	if err != nil {
		return nil, err
	}

	user := model.NewUser(username, "")
	user.PublicAddress = authWallet.Address

	err = s.TransactionManager.WithinTransaction(ctx,
		func(ctx context.Context, tx bun.Tx) error {
			if err := s.UserRepository.WithTx(tx).Create(ctx, user); err != nil {
				return apperrors.Internal("failed to create user", err)
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.User, error) {
	user, err := s.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal("failed to get user by id", err)
	}

	return user, nil
}

func (s *UserService) UpdateProfilePicture(
	ctx context.Context,
	userID uuid.UUID,
	file *multipart.FileHeader,
) (string, error) {
	if s.UserRepository == nil || s.FileService == nil {
		return "", apperrors.Internal("user service not fully initialized", nil)
	}
	if file == nil {
		return "", apperrors.BadRequest("image file is required")
	}

	user, err := s.UserRepository.GetByID(ctx, userID)
	if err != nil {
		return "", apperrors.Internal("failed to get user", err)
	}
	if user == nil {
		return "", apperrors.NotFound("user not found")
	}

	if user.ImageUrl != "" && !s.FileService.IsDefaultUserIcon(user.ImageUrl) {
		if err = s.FileService.RemoveUserFile(user.ImageUrl); err != nil {
			return "", apperrors.Internal("failed to remove previous profile image", err)
		}
	}

	newURL, err := s.FileService.SaveUserFile(file)
	if err != nil {
		return "", err
	}

	user.ImageUrl = newURL
	if err = s.UserRepository.Update(ctx, user); err != nil {
		return "", apperrors.Internal("failed to update user profile picture", err)
	}

	return newURL, nil
}

func (s *UserService) ChangeUsername(
	ctx context.Context,
	userID uuid.UUID,
	username mtype.Username,
) error {
	if !username.Valid() {
		return apperrors.BadRequest("invalid username")
	}
	user := &model.User{ID: userID, Username: username}

	err := s.UserRepository.Update(ctx, user)
	if err != nil {
		if repo.DuplicateKeyViolation(err) {
			return apperrors.AlreadyExist("this username is already taken")
		}

		return apperrors.Internal("failed to get user by id", err)
	}

	if err := s.DuelRepository.UpdateOwnerUsername(ctx, userID, username.String()); err != nil {
		return apperrors.Internal("failed to update username in duels", err)
	}

	return nil
}

func (s *UserService) GetUserStats(ctx context.Context, userID uuid.UUID) (*model.UserStats, error) {
	stats, err := s.DuelRepository.GetUserStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
