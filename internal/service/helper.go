package service

import (
	cryptoRand "crypto/rand"
	"duels-api/pkg/apperrors"
	"duels-api/pkg/mtype"
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	MediaFilesDirectory = "resources/profile-icons/"
)

func (s *UserService) getRandomIconName() (string, error) {
	files, err := os.ReadDir(MediaFilesDirectory)
	if err != nil {
		return "", apperrors.Internal("failed to chose random icon", err)
	}

	svgFiles := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".svg" {
			svgFiles = append(svgFiles, file.Name())
		}
	}

	if len(svgFiles) == 0 {
		err = fmt.Errorf("no SVG files found in directory: %s", MediaFilesDirectory)
		return "", apperrors.Internal("failed to chose random icon", err)
	}

	var seed int64
	err = binary.Read(cryptoRand.Reader, binary.LittleEndian, &seed)
	if err != nil {
		return "", apperrors.Internal("failed to chose random icon", err)
	}

	rand.New(rand.NewSource(seed))
	randomFile := svgFiles[rand.Intn(len(svgFiles))]

	return path.Join("/profile-icons", randomFile), nil
}

func generateUsername(prefix string) (mtype.Username, error) {
	n, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(1<<62))
	if err != nil {
		return "", apperrors.Internal("failed to generate username", err)
	}
	suffix := strings.ToLower(n.Text(36))
	raw := prefix + "_" + suffix
	if len(raw) > 44 {
		raw = raw[:44]
	}
	if u, ok := mtype.NewUsername(raw); ok {
		return u, nil
	}
	short := prefix + "_" + suffix[:min(len(suffix), 8)]
	if u, ok := mtype.NewUsername(short); ok {
		return u, nil
	}
	return "", apperrors.Internal("failed to build username")
}
