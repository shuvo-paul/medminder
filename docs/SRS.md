# Software Requirements Specification (SRS)
## MedMinder v1.0

---

## 1. Introduction

### 1.1 Purpose
This document defines the software requirements for MedMinder, a medication reminder application that helps users track medications for themselves, their family members, and friends, set reminder schedules, and receive notifications via WhatsApp and Telegram.

### 1.2 Scope
MedMinder is a REST API backend with a Svelte single-page application frontend. The system enables users to manage multiple medication profiles for themselves, family, and friends, share profiles with caregivers with granular permissions, generate guest access for external access, and log dose history.

### 1.3 Definitions, Acronyms, and Abbreviations
| Term | Definition |
|------|------------|
| API | Application Programming Interface |
| JWT | JSON Web Token |
| Profile | A medication management context (e.g., self, family member) |
| Dose | A single instance of taking or skipping a medication |
| Reminder | A scheduled notification for a medication |
| Guest Access | Non-user profile access via refresh token (30 days) |
| PBAC | Permission-Based Access Control |

---

## 2. Overall Description

### 2.1 Product Perspective
MedMinder is a standalone medication management system consisting of:
- RESTful API backend (Go)
- Svelte single-page application frontend
- WhatsApp Business API integration for notifications
- Telegram Bot API integration for notifications
- PostgreSQL database for persistent storage
- Multi-profile management (self, family, friends)
- Profile sharing with permission-based access control
- Guest access for non-user profile access

### 2.2 User Characteristics
User access is defined by permissions on profiles (see Section 3.2.2).

### 2.3 User Stories
1. **As a caregiver**, I want to manage medication reminders for my elderly parent so that they don't miss their daily medications.
2. **As a parent**, I want to track medications for all my family members in one app so that I have a complete overview of everyone's health.
3. **As a patient**, I want to share my medication profile with my doctor so they can see my medication history.
4. **As a family member**, I want to receive WhatsApp notifications when it's time for my loved one's medication so I can remind them.
5. **As a user**, I want to log doses with notes (e.g., "took with food") so I can track side effects.
6. **As a guest**, I want to view a profile's medications without creating an account so I can help with caregiving temporarily.

### 2.4 Product Features (High-Level)
1. User registration and authentication
2. Multi-profile management per user
3. Profile sharing with permission-based access control
4. Guest access generation
5. Medication catalog and management
6. Flexible reminder scheduling
7. WhatsApp and Telegram notifications
8. Dose logging (taken, skipped, snoozed)
9. Dose history and calendar view

### 2.5 System Architecture
MedMinder is deployed as a single binary. The Go server embeds the Svelte frontend at build time using `go:embed`. The Go router handles all routes: `/api/*` and `/healthz` are handled by the backend, while all other routes are forwarded to the Svelte frontend's Client-Side Rendering (CSR).

---

## 3. Functional Requirements

### 3.1 User Authentication

#### 3.1.1 Registration
- **REQ-AUTH-001**: The system shall allow users to register with email and password.
- **REQ-AUTH-002**: The system shall validate email format and require minimum password length of 8 characters.
- **REQ-AUTH-003**: The system shall return a JWT access token and refresh token upon successful registration.

#### 3.1.2 Login
- **REQ-AUTH-004**: The system shall authenticate users with email and password.
- **REQ-AUTH-005**: The system shall return JWT access token (24h expiry) and refresh token (7 days expiry) upon successful login.
- **REQ-AUTH-006**: The system shall reject invalid credentials with appropriate error message.

#### 3.1.3 Token Refresh
- **REQ-AUTH-007**: The system shall allow users to refresh access tokens using a valid refresh token.
- **REQ-AUTH-008**: Invalid or expired refresh tokens shall be rejected.

#### 3.1.4 Logout
- **REQ-AUTH-009**: The system shall invalidate the refresh token upon logout.

#### 3.1.5 Password Reset
- **REQ-AUTH-010**: The system shall allow users to request a password reset via email.
- **REQ-AUTH-011**: Password reset tokens shall expire after 1 hour.
- **REQ-AUTH-012**: The system shall allow users to set a new password using a valid reset token.
- **REQ-AUTH-013**: The system shall invalidate all refresh tokens when password is reset.

#### 3.1.6 Email Verification
- **REQ-AUTH-014**: The system shall send an email verification link upon registration.
- **REQ-AUTH-015**: Email verification tokens shall expire after 24 hours.
- **REQ-AUTH-016**: Users shall not receive medication reminders until email is verified.
- **REQ-AUTH-017**: The system shall allow resending verification email (max 3 per day).

### 3.2 Profile Management

#### 3.2.1 Profile CRUD
- **REQ-PROF-001**: Users shall be able to create profiles with name, avatar URL, date of birth, and medical conditions.
- **REQ-PROF-002**: Users shall be able to view all profiles they own or have access to.
- **REQ-PROF-003**: Users shall be able to update profile details.
- **REQ-PROF-004**: Users with `profile:admin` permission shall be able to delete profiles.

#### 3.2.2 Profile Sharing
- **REQ-PROF-005**: Users shall be able to share profiles with other registered users.
- **REQ-PROF-005b**: Sharing with a registered user shall create a pending invitation.
- **REQ-PROF-005c**: Users shall be able to set invitation expiration when sharing: 1, 3, or 7 days.
- **REQ-PROF-005d**: Users shall be able to view pending profile invitations.
- **REQ-PROF-005e**: Users shall be able to accept or decline profile invitations.
- **REQ-PROF-005f**: Expired invitations shall be automatically declined/removed.
- **REQ-PROF-005g**: Profile access shall only be granted after invitation is accepted.
- **REQ-PROF-006**: Users shall be able to specify granular permissions for shared profiles. Available permissions:
  - `medication:read`: View medications and their details
  - `medication:write`: Add, edit, and delete medications
  - `reminder:read`: View reminders and schedules
  - `reminder:write`: Add, edit, and delete reminders
  - `dose:read`: View dose history and logs
  - `dose:write`: Log dose status (taken/skipped/snoozed)
  - `prescription:read`: View and download prescriptions
  - `prescription:write`: Upload and delete prescriptions
  - `profile:read`: View profile details
  - `profile:write`: Edit profile details
  - `profile:share`: Share profile with other users
  - `profile:admin`: Full control including transferring admin rights, revoking access, and deleting profile (implies ownership)
- **REQ-PROF-007**: Users shall be able to view all users with access to a profile.
- **REQ-PROF-008**: Users shall be able to revoke shared access.

#### 3.2.3 Guest Access
- **REQ-PROF-009**: Users shall be able to generate a guest access link for a profile.
- **REQ-PROF-010**: The link shall use a refresh token valid for 30 days.
- **REQ-PROF-010b**: Users with `profile:admin` permission shall be able to manually revoke guest access before expiration.
- **REQ-PROF-011**: Anyone with the guest access link shall be able to access the profile with `medication:read`, `reminder:read`, `dose:read`, and `prescription:read` permissions without needing a user account.

> **Note**: Guest access provides instant access and does not require invitation acceptance (unlike registered user profile sharing in Section 3.2.2).

### 3.3 Medication Management

#### 3.3.1 Medication CRUD
- **REQ-MED-001**: Users shall be able to add medications with name, dosage amount, dosage unit, and form.
- **REQ-MED-002**: Supported dosage units: mg, ml, mcg, g, tablets, capsules.
- **REQ-MED-003**: Supported medication forms: tablet, capsule, liquid, injection, topical, inhaler, drops, patch.
- **REQ-MED-004**: Users shall be able to view all medications for a profile.
- **REQ-MED-005**: Users shall be able to update medication details.
- **REQ-MED-006**: Users shall be able to delete medications (cascades to reminders).

#### 3.3.2 Medication Frequency
- **REQ-MED-007**: Users shall be able to set frequency: daily, weekly (specific days), or as-needed (PRN).
- **REQ-MED-008**: Users shall be able to set optional start and end dates for medications.

#### 3.3.3 Prescriber Information
- **REQ-MED-009**: Users shall be able to optionally link a medication to a prescriber (name, clinic, phone).

#### 3.3.4 Medication Auto-Suggestion
- **REQ-MED-010**: The system shall display medication suggestions after user types 2 or more characters.
- **REQ-MED-011**: Suggestions shall be fetched from local medicine database.
- **REQ-MED-012**: Selecting a suggestion shall auto-fill medication name.
- **REQ-MED-013**: Users shall be able to manually enter medication details if suggestion not available.
- **REQ-MED-014**: The system shall sync medicine data from DGDA to bootstrap local database.
- **REQ-MED-015**: The system shall re-sync medicine data from DGDA monthly via automated scheduled job.
- **REQ-MED-016**: The system shall alert administrators if scheduled sync fails.

#### 3.3.5 Prescription Upload
- **REQ-MED-017**: Users shall be able to upload prescription documents (PDF, JPG, PNG) for a medication.
- **REQ-MED-018**: Uploaded prescriptions shall be stored securely in external storage (Cloudflare R2).
- **REQ-MED-019**: All users with access to the profile with `prescription:read` permission shall be able to view/download prescriptions.
- **REQ-MED-020**: Users shall be able to delete uploaded prescriptions.
- **REQ-MED-021**: The system shall support a maximum file size of 10MB per prescription.
- **REQ-MED-022**: Prescriptions must be linked to a profile. Linking to a medication is optional.

### 3.4 Reminder Scheduling

#### 3.4.1 Reminder CRUD
- **REQ-REM-001**: Users shall be able to create reminders linked to a medication.
- **REQ-REM-002**: Users shall be able to specify reminder times (multiple per day supported).
- **REQ-REM-003**: Users shall be able to enable or disable reminders.
- **REQ-REM-004**: Users shall be able to view all reminders for a profile.
- **REQ-REM-005**: Users shall be able to update reminder settings.
- **REQ-REM-006**: Users shall be able to delete reminders.

#### 3.4.2 Snooze
- **REQ-REM-007**: Users shall be able to snooze reminders for 5, 10, 15, or 30 minutes.
- **REQ-REM-008**: Snoozed reminders shall trigger a new notification at the snoozed time.

### 3.5 Notifications

#### 3.5.1 WhatsApp Notifications
- **REQ-NOT-001**: The system shall send WhatsApp messages via WhatsApp Business API.
- **REQ-NOT-002**: Messages shall include medication name, dosage, and scheduled time.
- **REQ-NOT-003**: Users shall be able to opt-in/opt-out of WhatsApp notifications per profile.

#### 3.5.2 Telegram Notifications
- **REQ-NOT-004**: The system shall send messages via Telegram Bot API.
- **REQ-NOT-005**: Users shall be able to link their Telegram account to receive notifications.
- **REQ-NOT-006**: Users shall be able to opt-in/opt-out of Telegram notifications per profile.

#### 3.5.3 Notification Preferences
- **REQ-NOT-007**: Users shall be able to set quiet hours (no notifications during specified times).
- **REQ-NOT-008**: Users shall be able to set notification advance time (e.g., 5 min before scheduled time).

#### 3.5.4 Notification Delivery
- **REQ-NOT-009**: Failed notifications shall be retried up to 3 times with exponential backoff (1min, 5min, 15min).
- **REQ-NOT-010**: The system shall log all notification delivery attempts and their status.
- **REQ-NOT-011**: The system shall provide a notification delivery status endpoint for users to view failed notifications.

### 3.6 Dose Logging

#### 3.6.1 Logging
- **REQ-DOSE-001**: Users shall be able to log a dose as taken, skipped, or snoozed.
- **REQ-DOSE-002**: Users shall be able to add notes to a dose log entry (e.g., "felt dizzy", "took with food").
- **REQ-DOSE-003**: The system shall automatically log doses as "skipped" if no user action is taken within 30 minutes after the scheduled time.
  - **Acceptance Criteria**:
    - Dose is marked "skipped" at scheduled_time + 30 minutes
    - User can still log the dose as "taken" after it's auto-skipped (with original scheduled time)
    - Snoozed reminders extend the auto-skipped window

#### 3.6.2 History
- **REQ-DOSE-004**: Users shall be able to view dose history for a profile.
- **REQ-DOSE-005**: Dose history shall be filterable by date range, medication, and status (taken/skipped).
- **REQ-DOSE-006**: Users shall be able to view dose history in calendar format.

---

## 4. Non-Functional Requirements

### 4.1 Performance
- **REQ-PERF-001**: API response time shall be under 200ms at p95 for read operations.
- **REQ-PERF-002**: API response time shall be under 500ms at p95 for write operations.
- **REQ-PERF-003**: The system shall support at least 100 concurrent users per instance.
- **REQ-PERF-004**: All list endpoints shall support pagination with `limit` (default 20, max 100) and `offset` parameters.

### 4.2 Security
- **REQ-SEC-001**: All passwords shall be hashed using bcrypt with cost factor 12.
- **REQ-SEC-002**: All API endpoints shall require authentication except `/healthz`, `/api/auth/register`, `/api/auth/login`.
- **REQ-SEC-003**: All data in transit shall be encrypted using TLS 1.2 or higher.
- **REQ-SEC-004**: JWT tokens shall include expiration claims.
- **REQ-SEC-005**: Guest access shall use cryptographically random refresh tokens valid for 30 days.
- **REQ-SEC-006**: The system shall implement rate limiting: 100 requests per minute per IP for auth endpoints, 300 requests per minute per user for other endpoints.
- **REQ-SEC-007**: The system shall log all authentication events (login, logout, password reset, failed attempts) for audit purposes.
- **REQ-SEC-008**: All API requests shall include a request ID for traceability.

### 4.3 Validation
- **REQ-VAL-001**: Email addresses shall be validated for proper format.
- **REQ-VAL-002**: Passwords shall require minimum 8 characters, at least 1 uppercase, 1 lowercase, and 1 number.
- **REQ-VAL-003**: Profile names shall be limited to 100 characters.
- **REQ-VAL-004**: Medication names shall be limited to 200 characters.
- **REQ-VAL-005**: Dose notes shall be limited to 500 characters.
- **REQ-VAL-006**: All user inputs shall be sanitized to prevent XSS and SQL injection.

### 4.4 Availability
- **REQ-AVL-001**: The system shall maintain 99.9% uptime (excluding scheduled maintenance).
- **REQ-AVL-002**: The system shall have a health check endpoint at `/healthz`.

### 4.5 Data Retention
- **REQ-DATA-001**: Dose history shall be retained for 2 years by default.
- **REQ-DATA-002**: Users shall be able to request data deletion (GDPR compliance).

### 4.6 Scalability
- **REQ-SCL-001**: The system shall support horizontal scaling with stateless API instances.
- **REQ-SCL-002**: The database shall be separated from the application layer.

---

## 5. API Endpoints Overview

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/auth/register | Register new user |
| POST | /api/auth/login | User login |
| POST | /api/auth/refresh | Refresh access token |
| POST | /api/auth/logout | User logout |
| GET | /api/profiles | List user profiles |
| POST | /api/profiles | Create profile |
| GET | /api/profiles/{id} | Get profile details |
| PUT | /api/profiles/{id} | Update profile |
| DELETE | /api/profiles/{id} | Delete profile |
| POST | /api/profiles/{id}/share | Share profile with user (with optional expires_in_days param) |
| GET | /api/profiles/{id}/share | List profile sharees |
| DELETE | /api/profiles/{id}/share/{userId} | Revoke profile access |
| GET | /api/invitations | List pending invitations for current user |
| POST | /api/invitations/{id}/accept | Accept invitation |
| POST | /api/invitations/{id}/decline | Decline invitation |
| POST | /api/profiles/{id}/share/guest | Generate guest access link |
| DELETE | /api/profiles/{id}/share/guest/{tokenId} | Revoke guest access |
| GET | /api/medications | List medications for profile |
| POST | /api/medications | Create medication |
| GET | /api/medications/{id} | Get medication |
| PUT | /api/medications/{id} | Update medication |
| DELETE | /api/medications/{id} | Delete medication |
| GET | /api/profiles/{id}/prescriptions | List prescriptions |
| POST | /api/profiles/{id}/prescriptions | Upload prescription |
| GET | /api/prescriptions/{id} | Download prescription |
| DELETE | /api/prescriptions/{id} | Delete prescription |
| GET | /api/reminders | List reminders for profile |
| POST | /api/reminders | Create reminder |
| GET | /api/reminders/{id} | Get reminder |
| PUT | /api/reminders/{id} | Update reminder |
| DELETE | /api/reminders/{id} | Delete reminder |
| POST | /api/doses | Log a dose |
| GET | /api/doses | Get dose history |
| POST | /api/webhooks/whatsapp | WhatsApp webhook |
| POST | /api/webhooks/telegram | Telegram webhook |
| GET | /healthz | Health check |

---

## 5.1 API Response Format

All API responses shall follow a consistent format:

### Success Response
```json
{
  "data": { ... },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-03-14T10:30:00Z"
  }
}
```

### Paginated Response
```json
{
  "data": [...],
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-03-14T10:30:00Z",
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 150
    }
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "details": [
      {"field": "email", "message": "must be a valid email address"}
    ]
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2026-03-14T10:30:00Z"
  }
}
```

All timestamps shall be in ISO 8601 format (UTC). All datetime storage shall be in UTC; display times shall be converted to the user's configured timezone.

## 6. Data Models (Overview)

### 6.1 User
- id, email, password_hash, timezone (e.g., "America/New_York"), email_verified (boolean), created_at, updated_at

### 6.2 Profile
- id, user_id, name, avatar_url, date_of_birth, medical_conditions, created_at, updated_at

### 6.3 ProfilePermission
- id, profile_id, shared_with_user_id, permissions (JSONB array), granted_by_user_id, status (pending | accepted | declined), expires_at, created_at

### 6.4 ProfileLink
- id, profile_id, token, expires_at, used_at, created_at

### 6.5 Medication
- id, profile_id, name, dosage_amount, dosage_unit, form, frequency, frequency_days, start_date, end_date, prescriber_name, prescriber_clinic, prescriber_phone, created_at, updated_at

### 6.6 Reminder
- id, medication_id, time, enabled, snooze_minutes, created_at, updated_at

### 6.7 Dose
- id, reminder_id, scheduled_at, status (taken/skipped/snoozed), notes, logged_at

### 6.8 NotificationPreference
- id, profile_id, whatsapp_enabled, telegram_enabled, telegram_chat_id, quiet_hours_start, quiet_hours_end, advance_minutes

### 6.9 Prescription
- id, profile_id, medication_id (nullable), file_url, file_type, file_size, uploaded_at

---

## Appendix A: Feature Status Matrix

| Feature | Module | Status | Priority | Notes |
|---------|--------|--------|----------|-------|
| Email/password registration | Auth | Not Started | P0 | |
| JWT login/logout | Auth | Not Started | P0 | |
| Token refresh | Auth | Not Started | P0 | |
| Profile CRUD | Profile | Not Started | P0 | |
| Profile sharing with PBAC | Profile | Not Started | P1 | Granular permissions (medication:read/write, reminder:read/write, dose:read/write, prescription:read/write, profile:read/write/share/admin), invitation flow with user-configurable expiration (1/3/7 days) |
| Guest access | Profile | Not Started | P1 | Refresh token (30 days), admin can revoke |
| Medication CRUD | Medication | Not Started | P0 | |
| Medication frequency options | Medication | Not Started | P0 | |
| Prescriber information | Medication | Not Started | P2 | Optional feature |
| Prescription upload | Medication | Not Started | P2 | Cloudflare R2 storage |
| Medication auto-suggestion | Medication | Not Started | P2 | Local DB with DGDA data (monthly cron + admin alerts) |
| Reminder CRUD | Reminder | Not Started | P0 | |
| Reminder snooze | Reminder | Not Started | P0 | |
| WhatsApp notifications | Notification | Not Started | P0 | |
| Telegram notifications | Notification | Not Started | P0 | |
| Quiet hours | Notification | Not Started | P1 | |
| Dose logging (manual) | Dose | Not Started | P0 | |
| Auto dose logging | Dose | Not Started | P1 | |
| Dose history view | Dose | Not Started | P0 | |
| Calendar view | Dose | Not Started | P1 | |
| Health check endpoint | Infrastructure | Completed | P0 | `/healthz` exists |

---

## Appendix B: Revision History

| Version | Date | Description |
|---------|------|-------------|
| 1.0 | 2026-03-14 | Initial SRS: auth, profiles, PBAC, medications, reminders, notifications, dose logging |
