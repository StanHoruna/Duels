package service

import (
	"duels-api/internal/storage/repository"
	"duels-api/pkg/apperrors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type FileService struct {
	FileRepository       *repository.FileRepository
	defaultUserIconPaths map[string]struct{}
}

func NewFileService(fileRepo *repository.FileRepository) (*FileService, error) {
	files, err := os.ReadDir(MediaFilesDirectory)
	if err != nil {
		return nil, apperrors.Internal("failed to read default user icon's dir", err)
	}

	svgFiles := make(map[string]struct{}, len(files))
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".svg" {
			svgFiles[filepath.Join(profileIconsParentDir, file.Name())] = struct{}{}
		}
	}

	return &FileService{
		defaultUserIconPaths: svgFiles,
		FileRepository:       fileRepo,
	}, nil
}

func (s *FileService) IsDefaultUserIcon(file string) bool {
	_, ok := s.defaultUserIconPaths[file]
	return ok
}

var allowedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".svg":  {},
	".png":  {},
}

const (
	mediaFilesDir      = "storage"
	userUploadMediaDir = "user_uploads"

	profileIconsParentDir = "/profile-icons"
)

func (s *FileService) SaveUserFile(
	file *multipart.FileHeader,
) (string, error) {
	return s.saveFile(userUploadMediaDir, file)
}

func (s *FileService) SaveUserFiles(
	files []*multipart.FileHeader,
) ([]string, error) {
	if len(files) > 3 {
		return nil, apperrors.Internal("no more than 3 images allowed")
	}

	filesPath := make([]string, len(files))

	for i, file := range files {
		path, err := s.saveFile(userUploadMediaDir, file)
		if err != nil {
			return nil, err
		}
		filesPath[i] = path
	}

	return filesPath, nil
}

const maxAllowedSize = 3 * 1024 * 1024 // 3 MB

func (s *FileService) saveFile(
	dir string,
	file *multipart.FileHeader,
) (string, error) {
	if file.Size > maxAllowedSize {
		return "", apperrors.BadRequest("file exceeds max allowed size")
	}

	extension := strings.ToLower(filepath.Ext(file.Filename))
	if _, ok := allowedImageExtensions[extension]; !ok {
		return "", apperrors.BadRequest("invalid file extension")
	}

	fileContent, err := file.Open()
	if err != nil {
		return "", apperrors.Internal("failed to open an image", err)
	}
	defer func() { _ = fileContent.Close() }()

	data, err := io.ReadAll(fileContent)
	if err != nil {
		return "", apperrors.Internal("failed to read an image", err)
	}

	fileName := uuid.New().String() + extension
	savePath := filepath.Join(mediaFilesDir, dir, fileName)

	if err = s.FileRepository.Save(savePath, data); err != nil {
		return "", apperrors.Internal("failed to save an image", err)
	}

	publicURL := "/" + filepath.Join(dir, fileName)

	return publicURL, nil
}

func (s *FileService) RemoveUserFile(
	fileName string,
) error {
	return s.RemoveFile(userUploadMediaDir, fileName)
}

func (s *FileService) RemoveFile(
	dir string,
	fileName string,
) error {
	path := filepath.Join(mediaFilesDir, fileName)
	cleanPath := filepath.Clean(path)

	expectedPrefix := filepath.Join(mediaFilesDir, dir)
	if !strings.HasPrefix(cleanPath, expectedPrefix) {
		return apperrors.BadRequest("invalid path: attempting directory traversal")
	}

	if _, ok := allowedImageExtensions[filepath.Ext(fileName)]; !ok {
		return apperrors.BadRequest("invalid file extension")
	}

	if err := s.FileRepository.Remove(cleanPath); err != nil {
		if os.IsNotExist(err) {
			return apperrors.NotFound("file not found")
		}

		return apperrors.Internal("failed to delete file", err)
	}

	return nil
}
