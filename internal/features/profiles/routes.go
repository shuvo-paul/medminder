package profiles

import (
	"database/sql"

	"github.com/danielgtaylor/huma/v2"
	"github.com/shuvo-paul/medminder/internal/database/sqlc"
	authService "github.com/shuvo-paul/medminder/internal/features/auth/service"
	"github.com/shuvo-paul/medminder/internal/features/profiles/handlers"
	"github.com/shuvo-paul/medminder/internal/features/profiles/repository"
	profileService "github.com/shuvo-paul/medminder/internal/features/profiles/service"
	"github.com/shuvo-paul/medminder/internal/middleware"
)

func RegisterRoutes(api huma.API, dbConn *sql.DB, queries *db.Queries, tokenSvc authService.TokenServiceInterface) {
	profileRepo := repository.NewProfileRepository(queries, dbConn)
	scheduleRepo := repository.NewDoseScheduleRepository(queries)
	permChecker := profileService.NewPermissionChecker(queries)
	profileSvc := profileService.NewProfileService(profileRepo, scheduleRepo, permChecker)
	scheduleSvc := profileService.NewDoseScheduleService(profileRepo, scheduleRepo)
	invitationSvc := profileService.NewInvitationService(profileRepo, scheduleRepo)
	transferRepo := repository.NewOwnershipTransferRepository(queries, dbConn)
	transferSvc := profileService.NewOwnershipTransferService(profileRepo, transferRepo, permChecker)

	readPerm := middleware.HumaRequireProfilePermission(permChecker, tokenSvc, "profile:read", "profile:owner")
	writePerm := middleware.HumaRequireProfilePermission(permChecker, tokenSvc, "profile:write", "profile:owner")
	adminPerm := middleware.HumaRequireProfilePermission(permChecker, tokenSvc, "profile:admin", "profile:owner")
	ownerPerm := middleware.HumaRequireProfilePermission(permChecker, tokenSvc, "profile:owner")

	huma.Register(api, huma.Operation{
		OperationID: "create-profile",
		Method:      "POST",
		Path:        "/api/profiles",
		Summary:     "Create a new profile",
		Tags:        []string{"profiles"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.CreateProfileHandler(profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "list-profiles",
		Method:      "GET",
		Path:        "/api/profiles",
		Summary:     "List all profiles for the authenticated user",
		Tags:        []string{"profiles"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ListProfilesHandler(profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "get-profile",
		Method:      "GET",
		Path:        "/api/profiles/{id}",
		Summary:     "Get a profile by ID",
		Tags:        []string{"profiles"},
		Middlewares: []func(huma.Context, func(huma.Context)){readPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.GetProfileHandler(profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "update-profile",
		Method:      "PUT",
		Path:        "/api/profiles/{id}",
		Summary:     "Update a profile",
		Tags:        []string{"profiles"},
		Middlewares: []func(huma.Context, func(huma.Context)){writePerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.UpdateProfileHandler(profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "delete-profile",
		Method:      "DELETE",
		Path:        "/api/profiles/{id}",
		Summary:     "Delete a profile",
		Tags:        []string{"profiles"},
		Middlewares: []func(huma.Context, func(huma.Context)){adminPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.DeleteProfileHandler(profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "list-schedules",
		Method:      "GET",
		Path:        "/api/profiles/{id}/schedules",
		Summary:     "List all dose schedules for a profile",
		Tags:        []string{"schedules"},
		Middlewares: []func(huma.Context, func(huma.Context)){readPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ListSchedulesHandler(scheduleSvc, profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "create-schedule",
		Method:      "POST",
		Path:        "/api/profiles/{id}/schedules",
		Summary:     "Create a new dose schedule",
		Tags:        []string{"schedules"},
		Middlewares: []func(huma.Context, func(huma.Context)){writePerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.CreateScheduleHandler(scheduleSvc, profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "get-schedule",
		Method:      "GET",
		Path:        "/api/profiles/{id}/schedules/{scheduleId}",
		Summary:     "Get a dose schedule by ID",
		Tags:        []string{"schedules"},
		Middlewares: []func(huma.Context, func(huma.Context)){readPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.GetScheduleHandler(scheduleSvc, profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "update-schedule",
		Method:      "PUT",
		Path:        "/api/profiles/{id}/schedules/{scheduleId}",
		Summary:     "Update a dose schedule",
		Tags:        []string{"schedules"},
		Middlewares: []func(huma.Context, func(huma.Context)){writePerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.UpdateScheduleHandler(scheduleSvc, profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "delete-schedule",
		Method:      "DELETE",
		Path:        "/api/profiles/{id}/schedules/{scheduleId}",
		Summary:     "Delete a dose schedule",
		Tags:        []string{"schedules"},
		Middlewares: []func(huma.Context, func(huma.Context)){adminPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.DeleteScheduleHandler(scheduleSvc, profileSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "share-profile",
		Method:      "POST",
		Path:        "/api/profiles/{id}/share",
		Summary:     "Share a profile with another user",
		Tags:        []string{"profiles"},
		Middlewares: []func(huma.Context, func(huma.Context)){adminPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ShareProfileHandler(invitationSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "list-invitations",
		Method:      "GET",
		Path:        "/api/invitations",
		Summary:     "List pending invitations for the authenticated user",
		Tags:        []string{"invitations"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ListInvitationsHandler(invitationSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "accept-invitation",
		Method:      "PUT",
		Path:        "/api/invitations/{invitationId}/accept",
		Summary:     "Accept a profile invitation",
		Tags:        []string{"invitations"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.AcceptInvitationHandler(invitationSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "decline-invitation",
		Method:      "PUT",
		Path:        "/api/invitations/{invitationId}/decline",
		Summary:     "Decline a profile invitation",
		Tags:        []string{"invitations"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.DeclineInvitationHandler(invitationSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "initiate-transfer",
		Method:      "POST",
		Path:        "/api/profiles/{id}/transfer-ownership",
		Summary:     "Initiate ownership transfer",
		Tags:        []string{"transfers"},
		Middlewares: []func(huma.Context, func(huma.Context)){ownerPerm},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.InitiateTransferHandler(transferSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "list-transfers",
		Method:      "GET",
		Path:        "/api/transfers",
		Summary:     "List pending ownership transfers for the authenticated user",
		Tags:        []string{"transfers"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.ListTransfersHandler(transferSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "accept-transfer",
		Method:      "POST",
		Path:        "/api/transfers/{transferId}/accept",
		Summary:     "Accept an ownership transfer",
		Tags:        []string{"transfers"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.AcceptTransferHandler(transferSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "decline-transfer",
		Method:      "POST",
		Path:        "/api/transfers/{transferId}/decline",
		Summary:     "Decline an ownership transfer",
		Tags:        []string{"transfers"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.DeclineTransferHandler(transferSvc, tokenSvc))

	huma.Register(api, huma.Operation{
		OperationID: "cancel-transfer",
		Method:      "DELETE",
		Path:        "/api/transfers/{transferId}",
		Summary:     "Cancel an ownership transfer",
		Tags:        []string{"transfers"},
		Security: []map[string][]string{
			{"bearer": {}},
		},
	}, handlers.CancelTransferHandler(transferSvc, tokenSvc))
}
