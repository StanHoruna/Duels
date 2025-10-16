package v1

import (
	_ "duels-api/internal/model"
	"duels-api/internal/service"
	"duels-api/pkg/apperrors"
	auth "duels-api/pkg/jwt"
	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	UserService *service.UserService
	FileService *service.FileService
}

func NewUserHandler(
	userService *service.UserService,
	fileService *service.FileService,
) *UserHandler {
	return &UserHandler{
		UserService: userService,
		FileService: fileService,
	}
}

func (h *UserHandler) RegisterRoutes(app *fiber.App, auth *AuthHandler) {
	userGroup := app.Group("/user")

	userGroup.Use(auth.AuthMiddleware)
	{
		userGroup.Get("/", h.GetUser)
		userGroup.Put("/profile-picture", h.SetProfilePicture)

		userGroup.Put("/upload-images", h.UploadImage, auth.UploadUserImageMiddleware)
	}
}

// GetUser godoc
//
//	@Summary		Get current user profile
//	@Description	Returns the authenticated user's profile data based on their JWT token.
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Success		200				{object}	object{user=model.User}		"Successfully retrieved user data"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized - missing or invalid token"
//	@Failure		404				{object}	apperrors.ErrorPublic		"User not found"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error"
//	@Router			/user/ [get]
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.UserService.GetByID(c.Context(), claims.UserID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"user": user})
}

// SetProfilePicture godoc
//
//	@Summary		Set user profile picture
//	@Description	Allows an authenticated user to upload a new profile image.
//	@Description	The uploaded image will replace the existing profile picture and return its new public URL.
//	@Tags			user
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string						true	"Authorization Bearer token"
//	@Param			image			formData	file						true	"Profile image file"
//	@Success		200				{object}	object{image_url=string}	"Profile picture successfully updated"
//	@Failure		400				{object}	apperrors.ErrorPublic		"Invalid request data or missing image file"
//	@Failure		401				{object}	apperrors.ErrorPublic		"Unauthorized - invalid or missing token"
//	@Failure		500				{object}	apperrors.ErrorPublic		"Internal server error during image upload"
//	@Router			/user/profile-picture [put]
func (h *UserHandler) SetProfilePicture(c fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	imageURL, err := h.UserService.UpdateProfilePicture(c.Context(), claims.UserID, file)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"image_url": imageURL,
	})
}

// UploadImage godoc
//
//	@Summary		Upload user images
//	@Description	Allows an authenticated user to upload one or more image files.
//	@Description	Each image is stored, and the service returns an array of accessible URLs.
//	@Description	Upload rate is limited via middleware to prevent spam (max 3 per minute).
//	@Tags			user
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Authorization Bearer token"
//	@Param			images			formData	[]file							true	"Array of images to upload"
//	@Success		200				{object}	object{image_urls=[]string}		"Images uploaded successfully, returns list of image URLs"
//	@Failure		400				{object}	apperrors.ErrorPublic			"Invalid request data or missing image files"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized - missing or invalid token"
//	@Failure		429				{object}	apperrors.ErrorPublic			"Too many uploads - rate limit exceeded"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error during file save"
//	@Router			/user/upload-images [put]
func (h *UserHandler) UploadImage(c fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	files, ok := form.File["images"]
	if !ok {
		return apperrors.BadRequest("no images provided")
	}

	fileNames, err := h.FileService.SaveUserFiles(files)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"image_urls": fileNames,
	})
}
