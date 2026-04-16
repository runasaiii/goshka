package service

import (
	"errors"
	"testing"

	"goshka/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserServiceGetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)
	user := &repository.User{ID: 2, Name: "Test User"}

	mockRepo.EXPECT().
		GetUserByID(2).
		Return(user, nil)

	result, err := service.GetUserByID(2)

	require.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestUserServiceCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repository.NewMockUserRepository(ctrl)
	service := NewUserService(mockRepo)
	user := &repository.User{ID: 2, Name: "Test User"}

	mockRepo.EXPECT().
		CreateUser(user).
		Return(nil)

	err := service.CreateUser(user)

	require.NoError(t, err)
}

func TestUserServiceRegisterUser(t *testing.T) {
	t.Run("user already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		user := &repository.User{ID: 2, Name: "New User"}
		email := "user@example.com"

		mockRepo.EXPECT().
			GetByEmail(email).
			Return(&repository.User{ID: 1, Name: "Existing User"}, nil)

		err := service.RegisterUser(user, email)

		require.Error(t, err)
		assert.EqualError(t, err, "user with this email already exists")
	})

	t.Run("new user success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		user := &repository.User{ID: 2, Name: "New User"}
		email := "user@example.com"

		gomock.InOrder(
			mockRepo.EXPECT().
				GetByEmail(email).
				Return(nil, nil),
			mockRepo.EXPECT().
				CreateUser(user).
				Return(nil),
		)

		err := service.RegisterUser(user, email)

		require.NoError(t, err)
	})

	t.Run("repository error on CreateUser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		user := &repository.User{ID: 2, Name: "New User"}
		email := "user@example.com"
		createErr := errors.New("create user failed")

		gomock.InOrder(
			mockRepo.EXPECT().
				GetByEmail(email).
				Return(nil, nil),
			mockRepo.EXPECT().
				CreateUser(user).
				Return(createErr),
		)

		err := service.RegisterUser(user, email)

		require.Error(t, err)
		assert.ErrorIs(t, err, createErr)
	})

	t.Run("repository error on GetByEmail", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		user := &repository.User{ID: 2, Name: "New User"}
		email := "user@example.com"

		mockRepo.EXPECT().
			GetByEmail(email).
			Return(nil, errors.New("repository failed"))

		err := service.RegisterUser(user, email)

		require.Error(t, err)
		assert.EqualError(t, err, "error getting user with this email")
	})
}

func TestUserServiceUpdateUserName(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)

		err := service.UpdateUserName(2, "")

		require.Error(t, err)
		assert.EqualError(t, err, "name cannot be empty")
	})

	t.Run("user not found repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		repoErr := errors.New("user not found")

		mockRepo.EXPECT().
			GetUserByID(2).
			Return(nil, repoErr)

		err := service.UpdateUserName(2, "New Name")

		require.Error(t, err)
		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("successful update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		user := &repository.User{ID: 2, Name: "Old Name"}

		gomock.InOrder(
			mockRepo.EXPECT().
				GetUserByID(2).
				Return(user, nil),
			mockRepo.EXPECT().
				UpdateUser(gomock.Any()).
				DoAndReturn(func(updated *repository.User) error {
					assert.Same(t, user, updated)
					assert.Equal(t, "New Name", updated.Name)
					return nil
				}),
		)

		err := service.UpdateUserName(2, "New Name")

		require.NoError(t, err)
		assert.Equal(t, "New Name", user.Name)
	})

	t.Run("UpdateUser fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		user := &repository.User{ID: 2, Name: "Old Name"}
		updateErr := errors.New("update failed")

		gomock.InOrder(
			mockRepo.EXPECT().
				GetUserByID(2).
				Return(user, nil),
			mockRepo.EXPECT().
				UpdateUser(gomock.Any()).
				DoAndReturn(func(updated *repository.User) error {
					assert.Same(t, user, updated)
					assert.Equal(t, "New Name", updated.Name)
					return updateErr
				}),
		)

		err := service.UpdateUserName(2, "New Name")

		require.Error(t, err)
		assert.ErrorIs(t, err, updateErr)
		assert.Equal(t, "New Name", user.Name)
	})
}

func TestUserServiceDeleteUser(t *testing.T) {
	t.Run("attempt to delete admin", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)

		err := service.DeleteUser(1)

		require.Error(t, err)
		assert.EqualError(t, err, "it is not allowed to delete admin user")
	})

	t.Run("successful delete", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		called := false

		mockRepo.EXPECT().
			DeleteUser(2).
			DoAndReturn(func(id int) error {
				called = true
				assert.Equal(t, 2, id)
				return nil
			}).
			Times(1)

		err := service.DeleteUser(2)

		require.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		service := NewUserService(mockRepo)
		deleteErr := errors.New("delete failed")

		mockRepo.EXPECT().
			DeleteUser(2).
			Return(deleteErr).
			Times(1)

		err := service.DeleteUser(2)

		require.Error(t, err)
		assert.ErrorIs(t, err, deleteErr)
	})
}
