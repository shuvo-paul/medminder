# Software Requirements Specification (SRS)

## MedMinder v1.0

---

## 1. Introduction

### 1.1 Purpose

This document defines the software requirements for MedMinder, a medication reminder application that helps users track medications for themselves, their family members, and friends, set reminder schedules, and receive notifications via WhatsApp and Telegram.

### 1.2 Scope

MedMinder is a monolithic full-stack Go application with embedded Svelte frontend. The system enables users to manage multiple medication profiles for themselves, family, and friends, share profiles with caregivers with granular permissions, generate guest access for external access, and log dose history.

### 1.3 Definitions, Acronyms, and Abbreviations

| Term | Definition |
|------|------------|
| API | Application Programming Interface |
| JWT | JSON Web Token |
| PBAC | Permission-Based Access Control |
| Profile | A medication management context (e.g., self, family member) |
| Dose | A single instance of taking or skipping a medication |
| Dose Schedule | A named time slot (e.g., Breakfast at 08:00) that groups medications |
| Reminder | A scheduled notification for a medication |
| Guest Access | Non-user profile access via a cryptographic token (30-day expiry) |
| PRN | Pro Re Nata — Latin for "as needed"; medications taken on demand |
| Refill | The act of replenishing a medication supply (obtaining more tablets, capsules, etc.) |
| Refill Threshold | The quantity level at which a low-supply alert is triggered |
| Projected Depletion Date | The estimated date when a medication's supply will run out, calculated from current quantity and dosing frequency |
| Prescriber | A healthcare professional (doctor, specialist, etc.) who issues a prescription |
| AI Provider | An external large language model (LLM) service (e.g., Google Gemini) used to extract structured data from unstructured documents |
| BYOK | Bring Your Own Key — a model where the user supplies their own third-party API key; the system stores it encrypted and uses it on the user's behalf |
| Extraction | The automated process of identifying structured medication details from a prescription document using an AI Provider |
| Medication Candidate | A structured suggestion produced by an extraction job, representing a potential medication record pending user review and confirmation |
| Follow-up Appointment | A future visit to a prescriber requested at the time of a consultation, recorded by the user so the system can send an advance alert |
| DGDA | Directorate General of Drug Administration (Bangladesh) |
| SPA | Single-Page Application |
| PWA | Progressive Web Application — a web app installable on device with offline support |
| TLS | Transport Layer Security |
| GDPR | General Data Protection Regulation |
| VAPID | Voluntary Application Server Identification — authentication mechanism for Web Push |
| IndexedDB | Browser-native structured storage API used for offline data caching |

---

## 2. Overall Description

### 2.1 Product Perspective

MedMinder is a standalone medication management system consisting of:

- Monolithic full-stack Go application with embedded Svelte frontend delivered as a PWA
- WhatsApp Business API integration for notifications
- Telegram Bot API integration for notifications
- Web Push API (VAPID) for browser and installed-app notifications
- PostgreSQL database for persistent storage
- Multi-profile management (self, family, friends)
- Profile sharing with permission-based access control
- Guest access for non-user profile access
- Offline-capable with background sync for critical write operations

### 2.2 User Characteristics

User access is defined by permissions on profiles (see Section 3.2.2).

### 2.3 User Stories

1. **As a caregiver**, I want to manage medication reminders for my elderly parent so that they don't miss their daily medications.
2. **As a parent**, I want to track medications for all my family members in one app so that I have a complete overview of everyone's health.
3. **As a patient**, I want to share my medication profile with my doctor so they can see my medication history.
4. **As a family member**, I want to receive WhatsApp notifications when it's time for my loved one's medication so I can remind them.
5. **As a user**, I want to log doses with notes (e.g., "took with food") so I can track side effects.
6. **As a guest**, I want to view a profile's medications without creating an account so I can help with caregiving temporarily.
7. **As a user**, I want to set up meal times (breakfast, lunch, dinner) when creating a profile so that medications are automatically scheduled around my daily routine.

### 2.4 Product Features (High-Level)

1. User registration and authentication (email/password and Google OAuth)
2. Multi-profile management per user
3. Profile sharing with permission-based access control and invitation flow
4. Guest access generation (read-only, no account required)
5. Medication catalog and management with flexible frequency types
6. Dose schedule management (Breakfast, Lunch, Dinner, custom)
7. Flexible reminder scheduling (schedule, daily, weekly, monthly, interval, PRN)
8. WhatsApp, Telegram, and Web Push notifications with retry logic
9. Dose logging (taken, skipped, snoozed) with auto-skip
10. Dose history, filtering, and calendar view
11. Progressive Web App (PWA) — installable on mobile and desktop, offline-capable
12. Offline write queue with background sync for dose logging, batch logging, notes, and snooze
13. AI-assisted prescription extraction — users may connect a personal AI provider API key and trigger extraction of medication candidates from uploaded prescriptions for review and confirmation
14. Follow-up appointment alerts — track doctor-requested return visits and receive advance notifications

### 2.5 System Architecture

MedMinder is a **Monolithic Full-Stack Go Application with Embedded Frontend**:

- **Architecture Pattern**: Single binary deployment where the Go server embeds the compiled Svelte frontend at build time using Go's `go:embed` directive.
- **Build Process**:
  1. Svelte frontend is compiled to static assets (`npm run build`), including `manifest.json`, service worker (`sw.js`), and PWA icons
  2. Go server embeds the `web/dist/` directory using `go:embed`
  3. Final binary contains both backend and frontend
- **PWA Requirements**:
  - `manifest.json` must declare `name`, `short_name`, `start_url`, `display: standalone`, `theme_color`, `background_color`, and icon set (minimum 192×192 and 512×512 PNG)
  - Service worker (`sw.js`) manages app shell caching, offline fallback, background sync, and Web Push event handling
  - Service worker is registered at scope `/` and served with `Service-Worker-Allowed: /` header
- **Routing Strategy**:
  - `/api/v1/*` → REST API handlers (Go Chi router)
  - `/healthz` → Health check endpoint
  - `/sw.js` → Service worker (served with `Cache-Control: no-cache` and `Service-Worker-Allowed: /` headers)
  - All other routes → Svelte SPA app shell (client-side routing handled by service worker and Svelte router)
- **Development**: Uses `air` for live reload of Go code; Svelte Vite dev server handles frontend hot reload
- **Serving**: Go's `http.FileServer` serves embedded static files with proper MIME types and PWA-specific headers

> **Note on API versioning**: All API endpoints are prefixed with `/api/v1/` to allow non-breaking future changes. The shorthand `/api/...` used elsewhere in this document refers to `/api/v1/...`.

---

## 3. Functional Requirements

### 3.1 User Authentication

#### 3.1.1 Registration

- [x] **REQ-AUTH-001**: The system shall allow users to register with email, display name, and password.
- [x] **REQ-AUTH-002**: The system shall validate email format and require minimum password length of 8 characters with at least 1 uppercase letter, 1 lowercase letter, and 1 number.
- [x] **REQ-AUTH-003**: The system shall return a JWT access token and refresh token upon successful registration.

#### 3.1.2 Login

- [x] **REQ-AUTH-004**: The system shall authenticate users with email and password.
- [x] **REQ-AUTH-005**: The system shall return JWT access token (24h expiry) and refresh token (7 days expiry) upon successful login.
- [x] **REQ-AUTH-006**: The system shall reject invalid credentials with an appropriate error message.

#### 3.1.3 Token Refresh

- [x] **REQ-AUTH-007**: The system shall allow users to refresh access tokens using a valid refresh token.
- [x] **REQ-AUTH-008**: Invalid or expired refresh tokens shall be rejected.

#### 3.1.4 Logout

- [x] **REQ-AUTH-009**: The system shall invalidate the refresh token upon logout.

#### 3.1.5 Password Reset

- [ ] **REQ-AUTH-010**: The system shall allow users to request a password reset via email (`POST /api/auth/password/reset/request`).
- [ ] **REQ-AUTH-011**: Password reset tokens shall expire after 1 hour.
- [ ] **REQ-AUTH-012**: The system shall allow users to set a new password using a valid reset token (`POST /api/auth/password/reset/confirm`).
- [ ] **REQ-AUTH-013**: The system shall invalidate all refresh tokens when a password is reset.

#### 3.1.6 Email Verification

- [ ] **REQ-AUTH-014**: The system shall send an email verification link upon registration.
- [ ] **REQ-AUTH-015**: Email verification tokens shall expire after 24 hours.
- [ ] **REQ-AUTH-016**: Users shall not receive medication reminders until their email is verified.
- [ ] **REQ-AUTH-017**: The system shall allow resending the verification email (`POST /api/auth/email/resend-verification`), limited to 3 requests per day per user.

#### 3.1.7 OAuth Authentication (Extensible)

The system shall support OAuth 2.0 authentication with multiple providers.

##### 3.1.7.1 Supported Providers

- [ ] **REQ-OAUTH-001**: The system shall support Google as an OAuth provider.
- [ ] **REQ-OAUTH-001b**: The OAuth architecture shall be extensible to support additional providers (e.g., GitHub, Apple) without database migrations.

##### 3.1.7.2 OAuth Registration/Login

- [ ] **REQ-OAUTH-002**: The system shall allow users to register using their Google account.
- [ ] **REQ-OAUTH-003**: The system shall automatically create a user account upon successful OAuth login if no account exists.
- [ ] **REQ-OAUTH-004**: Upon successful OAuth login, the system shall issue a short-lived authorization code (5-minute expiry) and redirect the frontend to `/auth/callback?code={code}`. The frontend exchanges this code for JWT access and refresh tokens via `POST /api/auth/oauth/token`. Tokens shall never appear in redirect URLs.
- [ ] **REQ-OAUTH-005**: Users registering via OAuth shall not be required to set a password.

##### 3.1.7.3 Connecting OAuth to Existing Account

- [ ] **REQ-OAUTH-006**: Users with email/password accounts shall be able to link their Google account.
- [ ] **REQ-OAUTH-007**: Users shall be able to link only one Google account to their account (but may link other OAuth providers like GitHub, Apple).
- [ ] **REQ-OAUTH-008**: Users shall be able to unlink their Google account.
- [ ] **REQ-OAUTH-009**: The system shall record the timestamp when an OAuth provider was connected.

##### 3.1.7.4 Setting Password for OAuth Users

- [ ] **REQ-OAUTH-010**: Users who registered via OAuth shall be able to set a password.
- [ ] **REQ-OAUTH-011**: Users shall be able to change their password after setting it.
- [ ] **REQ-OAUTH-012**: Password requirements (8+ characters, uppercase, lowercase, number) shall apply to OAuth users who set a password.

##### 3.1.7.5 Account Linking by Email

- [ ] **REQ-OAUTH-013**: If a user attempts OAuth login with an email that already has a registered account, the system shall reject the OAuth login and prompt the user to link accounts via the existing account settings.

#### 3.1.8 Email Management

- [ ] **REQ-EMAIL-001**: Users shall be able to change their login email address.
- [ ] **REQ-EMAIL-002**: Changing email shall require current password confirmation (OAuth-only users without a password must set one first).
- [ ] **REQ-EMAIL-003**: The system shall send an email verification link to the new email address.
- [ ] **REQ-EMAIL-003b**: Email change verification tokens shall expire after 24 hours.
- [ ] **REQ-EMAIL-004**: The old email shall remain active until the new email is verified.
- [ ] **REQ-EMAIL-005**: If the new email is already registered, the system shall reject the change.
- [ ] **REQ-EMAIL-006**: Users with OAuth-linked accounts shall be able to change their email (OAuth account remains linked).
- [ ] **REQ-EMAIL-007**: Users shall be able to cancel pending email changes before verification via `POST /api/auth/email/cancel`.

#### 3.1.9 Account Deletion

- [ ] **REQ-AUTH-018**: Users shall be able to request permanent account deletion (`DELETE /api/auth/account`).
- [ ] **REQ-AUTH-019**: Account deletion shall require current password confirmation (or OAuth re-authentication for password-less users).
- [ ] **REQ-AUTH-020**: Upon account deletion, all profiles owned solely by the user shall be deleted, including all associated medications, reminders, doses, and prescriptions. Profiles shared with other users where the deleting user is not the sole admin shall have the user's access revoked without deleting the profile.
- [ ] **REQ-AUTH-021**: The system shall anonymize or delete all personal data within 30 days of account deletion request (GDPR compliance).

### 3.2 Profile Management

#### 3.2.1 Profile CRUD

- [ ] **REQ-PROF-001**: Users shall be able to create profiles with name, avatar URL, date of birth, medical conditions, and timezone (IANA timezone identifier, e.g., `Asia/Dhaka`).
- [ ] **REQ-PROF-001b**: When creating a profile, the system shall prompt for meal times (Breakfast, Lunch, Dinner) to set up initial dose schedules.
- [ ] **REQ-PROF-001c**: The system shall create initial dose schedules (Breakfast, Lunch, Dinner) based on user-provided meal times during profile creation.
- [ ] **REQ-PROF-001d**: Users shall be able to modify or delete initial dose schedules after profile creation.
- [ ] **REQ-PROF-001e**: Users shall be able to create additional dose schedules at any time.
- [ ] **REQ-PROF-002**: Users shall be able to view all profiles they own or have access to.
- [ ] **REQ-PROF-003**: Users shall be able to update profile details (requires `profile:write` permission).
- [ ] **REQ-PROF-004**: Users with `profile:admin` permission shall be able to delete profiles.

#### 3.2.2 Profile Ownership

- [ ] **REQ-PROF-009**: The user who creates a profile is its owner. The owner holds `profile:admin` permission and cannot have it revoked except by transferring ownership.
- [ ] **REQ-PROF-010**: A profile has exactly one owner at any time. Ownership is transferred via `POST /api/profiles/{id}/transfer-ownership`, which requires the current owner to specify the new owner (who must already have `profile:admin` permission on the profile).
- [ ] **REQ-PROF-011**: Upon ownership transfer, the previous owner retains their existing permissions but loses exclusive ownership status.
- [ ] **REQ-PROF-012**: If a profile owner deletes their account and the profile has no other `profile:admin` user, the profile is deleted (see REQ-AUTH-020).

#### 3.2.3 Profile Sharing

- [ ] **REQ-PROF-005**: Users with `profile:share` or `profile:admin` permission shall be able to share profiles with other registered users.
- [ ] **REQ-PROF-005b**: Sharing with a registered user shall create a pending invitation.
- [ ] **REQ-PROF-005c**: Users shall be able to set invitation expiration when sharing: 1, 3, or 7 days.
- [ ] **REQ-PROF-005d**: Invited users shall be able to view their pending profile invitations.
- [ ] **REQ-PROF-005e**: Invited users shall be able to accept or decline profile invitations.
- [ ] **REQ-PROF-005f**: Expired invitations shall be automatically removed. The inviting user may re-invite.
- [ ] **REQ-PROF-005g**: Profile access shall only be granted after the invitation is accepted.
- [ ] **REQ-PROF-006**: Users shall be able to specify granular permissions for shared profiles. Available permissions:
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
  - `profile:share`: Share profile with other users and manage invitations
  - `profile:admin`: Full control including revoking access, transferring ownership, and deleting the profile
- [ ] **REQ-PROF-007**: Users with `profile:share` or `profile:admin` permission shall be able to view all users with access to a profile.
- [ ] **REQ-PROF-008**: Users with `profile:admin` permission shall be able to revoke shared access.

#### 3.2.4 Guest Access

- [ ] **REQ-PROF-013**: Users with `profile:admin` permission shall be able to generate a guest access link for a profile.
- [ ] **REQ-PROF-014**: Guest access links shall use a cryptographically random token valid for 30 days.
- [ ] **REQ-PROF-014b**: Users with `profile:admin` permission shall be able to manually revoke guest access before expiration via `DELETE /api/profiles/{id}/share/guest/{tokenId}`.
- [ ] **REQ-PROF-015**: Anyone with the guest access link shall be able to access the profile with `medication:read`, `reminder:read`, `dose:read`, and `prescription:read` permissions without needing a user account.
- [ ] **REQ-PROF-015b**: Guest access endpoints shall be rate-limited to 60 requests per minute per IP.

> **Note**: Guest access provides instant read-only access and does not require invitation acceptance (unlike registered user profile sharing in Section 3.2.3).

### 3.3 Medication Management

#### 3.3.1 Medication CRUD

- [ ] **REQ-MED-001**: Users shall be able to add medications to a profile with name, dosage amount, dosage unit, and form (requires `medication:write` permission).
- [ ] **REQ-MED-002**: Supported dosage units: mg, ml, mcg, g, tablets, capsules.
- [ ] **REQ-MED-003**: Supported medication forms: tablet, capsule, liquid, injection, topical, inhaler, drops, patch.
- [ ] **REQ-MED-004**: Users shall be able to view all medications for a profile (requires `medication:read` permission).
- [ ] **REQ-MED-005**: Users shall be able to update medication details (requires `medication:write` permission).
- [ ] **REQ-MED-006**: Users shall be able to delete medications (requires `medication:write` permission). Deletion cascades to associated reminders and dose logs.

#### 3.3.2 Medication Frequency

- [ ] **REQ-MED-007**: The system shall support the following frequency types (extensible model):
  - **schedule**: Linked to a named dose schedule (e.g., Breakfast, Lunch, Dinner, or custom). `frequency_config`: `{ "schedule_id": "<uuid>" }`
  - **daily**: Every day at a fixed time. `frequency_config`: `{ "time": "08:00" }`
  - **weekly**: Specific days of the week at a fixed time. `frequency_config`: `{ "days": ["mon", "wed", "fri"], "time": "08:00" }`
  - **monthly**: A specific day of the month at a fixed time. `frequency_config`: `{ "day": 15, "time": "09:00" }`
  - **interval**: Every X units from the start date. `frequency_config`: `{ "value": 8, "unit": "hours" }`. Supported units: minutes, hours, days, weeks, months.
  - **prn**: As-needed (Pro Re Nata). No automatic notifications. `frequency_config`: `{}`
- [ ] **REQ-MED-007b**: The frequency model shall be extensible to support new types without database migrations (extensibility is achieved via the JSONB `frequency_config` field).
- [ ] **REQ-MED-008**: Users shall be able to set optional start and end dates for medications.

#### 3.3.3 Prescriber Information

- [ ] **REQ-MED-009**: Users shall be able to optionally link a medication to a Prescriber record (`prescriber_id` FK, nullable). See Section 3.10 for Prescriber management.

#### 3.3.4 Medication Auto-Suggestion

- [ ] **REQ-MED-010**: The system shall display medication name suggestions when a user types 2 or more characters (`GET /api/medications/suggest?q={query}`).
- [ ] **REQ-MED-011**: Suggestions shall be fetched from the local medicine database (bootstrapped from DGDA).
- [ ] **REQ-MED-012**: Selecting a suggestion shall auto-fill the medication name field.
- [ ] **REQ-MED-013**: Users shall be able to manually enter medication details if no suggestion is available.
- [ ] **REQ-MED-014**: The system shall sync medicine data from DGDA to bootstrap the local database on first setup.
- [ ] **REQ-MED-015**: The system shall re-sync medicine data from DGDA monthly via an automated scheduled job.
- [ ] **REQ-MED-016**: The system shall alert administrators via email if a scheduled DGDA sync fails.

#### 3.3.5 Prescription Upload

- [ ] **REQ-MED-017**: Users shall be able to upload prescription documents (PDF, JPG, PNG) for a profile, optionally linked to a specific medication (requires `prescription:write` permission).
- [ ] **REQ-MED-018**: Uploaded prescriptions shall be stored securely in external object storage (Cloudflare R2).
- [ ] **REQ-MED-019**: All users with `prescription:read` permission on the profile shall be able to view and download prescriptions.
- [ ] **REQ-MED-020**: Users shall be able to delete uploaded prescriptions (requires `prescription:write` permission).
- [ ] **REQ-MED-021**: The system shall enforce a maximum file size of 10MB per prescription upload.
- [ ] **REQ-MED-022**: Prescriptions must be linked to a profile and to a Prescriber (required). Linking to a specific medication is optional.

#### 3.3.6 Dose Schedules

- [ ] **REQ-MED-023**: Users shall be able to create dose schedules for a profile (e.g., "Breakfast", "Lunch", "Dinner", "Bedtime").
- [ ] **REQ-MED-024**: Each dose schedule shall have a name and a wall-clock time in the profile's timezone (e.g., "Breakfast" at 08:00).
- [ ] **REQ-MED-025**: Users shall be able to view all dose schedules for a profile.
- [ ] **REQ-MED-026**: Users shall be able to update a dose schedule's name and time.
- [ ] **REQ-MED-027**: Users shall be able to delete a dose schedule. Medications linked to a deleted schedule shall have their frequency type changed to `daily` with the schedule's time preserved, so no reminder data is lost.
- [ ] **REQ-MED-028**: When a medication is linked to a schedule, it shall automatically receive reminders at that schedule's time.
- [ ] **REQ-MED-029**: Users shall be able to link a medication to a schedule or set a custom reminder time (not tied to any schedule).
- [ ] **REQ-MED-030**: Users shall be able to switch a medication between schedule-based and custom time-based reminders at any time.

#### 3.3.7 PRN (As-Needed) Medications

- [ ] **REQ-MED-031**: Users shall be able to mark a medication as PRN (Pro Re Nata / as-needed).
- [ ] **REQ-MED-032**: PRN medications shall not generate automatic scheduled reminders or notifications.
- [ ] **REQ-MED-033**: Users shall be able to request an immediate notification for a PRN medication on demand via `POST /api/profiles/{id}/medications/{medId}/notify`.
- [ ] **REQ-MED-034**: The system shall log PRN medication notification requests with a timestamp.

### 3.4 Reminder Scheduling

#### 3.4.1 Reminder Generation

- [ ] **REQ-REM-001**: Reminders shall be automatically generated from medication frequency configurations.
- [ ] **REQ-REM-002**: For `schedule` frequency medications, reminders shall be created at the linked dose schedule's time (in the profile's timezone).
- [ ] **REQ-REM-003**: For `daily` frequency medications, reminders shall be created at the specified time (in the profile's timezone).
- [ ] **REQ-REM-004**: For `interval` frequency medications, reminders shall be generated at each interval from the medication's start date/time.
- [ ] **REQ-REM-005**: For `weekly` frequency medications, reminders shall be generated on the specified days at the specified time (in the profile's timezone).
- [ ] **REQ-REM-006**: For `monthly` frequency medications, reminders shall be generated on the specified day of each month at the specified time (in the profile's timezone). If the month does not contain the specified day (e.g., day 31 in February), the reminder shall fall on the last day of that month.
- [ ] **REQ-REM-007**: PRN medications shall not generate automatic reminders.

#### 3.4.2 Reminder Management

- [ ] **REQ-REM-008**: Users shall be able to enable or disable individual reminders (requires `reminder:write` permission).
- [ ] **REQ-REM-009**: Users shall be able to view all upcoming reminders for a profile (requires `reminder:read` permission).
- [ ] **REQ-REM-010**: Users shall be able to update reminder settings (time, enabled state) (requires `reminder:write` permission).
- [ ] **REQ-REM-011**: Users shall be able to delete reminders (requires `reminder:write` permission).

#### 3.4.3 Snooze

- [ ] **REQ-REM-012**: Users shall be able to snooze a reminder for 5, 10, 15, or 30 minutes via `POST /api/reminders/{id}/snooze`.
- [ ] **REQ-REM-013**: A snoozed reminder shall trigger a new notification at `original_scheduled_time + snooze_duration`.
- [ ] **REQ-REM-014**: Snoozing a reminder extends the auto-skip window (see REQ-DOSE-003) by the same snooze duration. If a reminder is snoozed at T+25min for 30 minutes, the auto-skip window becomes T+55min (original 30min window + 25min elapsed + 30min snooze).

### 3.5 Notifications

#### 3.5.1 WhatsApp Notifications

- [ ] **REQ-NOT-001**: The system shall send WhatsApp messages via WhatsApp Business API using pre-approved message templates.
- [ ] **REQ-NOT-002**: Notifications for schedule-based medications shall display the schedule name and list all medications in that schedule (e.g., "Breakfast: Aspirin 100mg tablet, Metformin 500mg tablet").
- [ ] **REQ-NOT-002b**: Notifications for non-schedule medications shall display the medication name, dosage, and scheduled time.
- [ ] **REQ-NOT-003**: Users shall be able to opt-in/opt-out of WhatsApp notifications per profile via notification preferences.
- [ ] **REQ-NOT-003b**: WhatsApp notifications require the user to provide and verify a WhatsApp-enabled phone number. The system shall only send messages within the WhatsApp 24-hour messaging window or using approved templates.

#### 3.5.2 Telegram Notifications

- [ ] **REQ-NOT-004**: The system shall send messages via Telegram Bot API.
- [ ] **REQ-NOT-005**: Users shall be able to link their Telegram account by messaging the MedMinder Telegram bot and receiving a link code to enter in the app (`POST /api/auth/telegram/link`).
- [ ] **REQ-NOT-006**: Users shall be able to opt-in/opt-out of Telegram notifications per profile via notification preferences.

#### 3.5.3 Notification Preferences

- [ ] **REQ-NOT-007**: Users shall be able to configure notification preferences per profile via `GET/PUT /api/profiles/{id}/notification-preferences`. Configurable settings:
  - WhatsApp enabled/disabled
  - WhatsApp phone number
  - Telegram enabled/disabled
  - Web Push enabled/disabled (per profile; subscription managed separately via Section 3.5.6)
  - Quiet hours start/end (in profile timezone)
  - Advance notification time (minutes before scheduled time; 0 = at scheduled time)
- [ ] **REQ-NOT-008**: The system shall not send notifications during configured quiet hours. Notifications that fall within quiet hours shall be suppressed (not deferred to after quiet hours).

#### 3.5.4 Notification Delivery

- [ ] **REQ-NOT-009**: Failed notifications shall be retried up to 3 times with exponential backoff: 1 minute, 5 minutes, 15 minutes after the initial failure.
- [ ] **REQ-NOT-010**: The system shall log all notification delivery attempts (channel, status, timestamp, error message if failed) in a `notification_logs` table.
- [ ] **REQ-NOT-011**: Users shall be able to view their recent notification delivery history, including failures, via `GET /api/profiles/{id}/notification-logs`.

#### 3.5.5 PRN Notifications

- [ ] **REQ-NOT-012**: Users shall be able to request an on-demand notification for a PRN medication via `POST /api/profiles/{id}/medications/{medId}/notify`.
- [ ] **REQ-NOT-013**: PRN notification requests shall be logged with a timestamp.

#### 3.5.6 Web Push Notifications

- [ ] **REQ-NOT-014**: The system shall support Web Push notifications using the VAPID protocol.
- [ ] **REQ-NOT-015**: Users shall be able to subscribe a browser or installed PWA to Web Push notifications via `POST /api/push/subscribe`. The request body shall include the browser-generated `endpoint`, `p256dh` key, and `auth` key.
- [ ] **REQ-NOT-016**: A single user account may hold multiple active push subscriptions (e.g., mobile browser + desktop browser + installed PWA). Each subscription is stored per device.
- [ ] **REQ-NOT-017**: Users shall be able to remove a push subscription via `DELETE /api/push/subscribe` (identified by `endpoint`). The system shall also automatically remove subscriptions that return HTTP 410 Gone from the push service.
- [ ] **REQ-NOT-018**: Web Push notifications shall follow the same quiet hours and advance time preferences configured in Section 3.5.3.
- [ ] **REQ-NOT-019**: The server VAPID public key shall be exposed at `GET /api/push/vapid-public-key` (unauthenticated) so the frontend can subscribe without a prior login round-trip.

### 3.6 Dose Logging

#### 3.6.1 Logging

- [ ] **REQ-DOSE-001**: Users shall be able to log a dose as `taken`, `skipped`, or `snoozed` (requires `dose:write` permission).
- [ ] **REQ-DOSE-002**: Users shall be able to add notes to a dose log entry (e.g., "felt dizzy", "took with food").
- [ ] **REQ-DOSE-003**: The system shall automatically log doses as `skipped` if no user action is taken within 30 minutes after the scheduled time.
  - **Acceptance Criteria**:
    - Dose is marked `skipped` at `scheduled_time + 30 minutes` (extended by any active snooze; see REQ-REM-014)
    - User can still manually log the dose as `taken` after it has been auto-skipped; the original scheduled time is preserved
    - Auto-skip is suppressed for PRN medications (they have no scheduled time)

#### 3.6.2 History

- [ ] **REQ-DOSE-004**: Users shall be able to view dose history for a profile (requires `dose:read` permission) via `GET /api/profiles/{id}/doses`.
- [ ] **REQ-DOSE-005**: Dose history shall be filterable by date range, medication ID, and status (taken/skipped/snoozed).
- [ ] **REQ-DOSE-006**: Users shall be able to view dose history in calendar format via `GET /api/profiles/{id}/doses/calendar?year={year}&month={month}`.

#### 3.6.3 Group-Level Logging

- [ ] **REQ-DOSE-007**: For schedule-based medications, users shall be able to mark all medications in a schedule as `taken` or `skipped` with a single action via `POST /api/profiles/{id}/doses/batch`.
- [ ] **REQ-DOSE-008**: The batch logging request shall accept an optional array of per-medication overrides, allowing individual medications in a schedule to have a different status than the bulk action.

### 3.7 Offline Support & Background Sync

#### 3.7.1 App Shell & Static Asset Caching

- [ ] **REQ-OFFLINE-001**: The service worker shall pre-cache the app shell (HTML, CSS, JS, fonts, icons) at install time so the application UI loads immediately without a network connection.
- [ ] **REQ-OFFLINE-002**: When a user navigates to any route while offline, the service worker shall serve the cached app shell and display an offline indicator.

#### 3.7.2 Data Caching for Offline Read

- [ ] **REQ-OFFLINE-003**: The system shall cache the following API responses in IndexedDB for offline read access:
  - Profile list and profile details
  - Medications and dose schedules for each accessed profile
  - Upcoming reminders (next 7 days) for each accessed profile
  - Dose history (last 30 days) for each accessed profile
- [ ] **REQ-OFFLINE-004**: Cached data shall be refreshed on every successful API response (cache-on-read strategy). Cache entries shall expire after 24 hours; stale entries are still served offline past expiry with a visual staleness indicator.

#### 3.7.3 Offline Write Queue

- [ ] **REQ-OFFLINE-005**: The following write operations shall be queueable when the device is offline:
  - Log a dose as `taken` or `skipped` (`POST /api/profiles/{id}/doses`)
  - Add or update a note on a dose entry
  - Batch dose logging (`POST /api/profiles/{id}/doses/batch`)
  - Snooze a reminder (`POST /api/reminders/{id}/snooze`)
- [ ] **REQ-OFFLINE-006**: Queued operations shall be stored durably in IndexedDB and shall survive app restart or browser refresh.
- [ ] **REQ-OFFLINE-007**: The UI shall display the count of pending queued operations when the device is offline.

#### 3.7.4 Background Sync

- [ ] **REQ-OFFLINE-008**: The service worker shall register a Background Sync event (`medminder-sync`) that flushes the write queue when connectivity is restored.
- [ ] **REQ-OFFLINE-009**: Each queued operation shall be retried up to 3 times. Operations that fail after 3 attempts (e.g., due to a server-side validation error) shall be moved to a "failed sync" state and surfaced to the user.
- [ ] **REQ-OFFLINE-010**: Conflict resolution shall follow last-write-wins: if a dose was auto-skipped on the server while the user had a `taken` entry queued offline, the user's queued action shall overwrite the server state on sync.

#### 3.7.5 Storage Constraints

- [ ] **REQ-OFFLINE-011**: The total IndexedDB storage used by MedMinder shall not exceed 50MB. When approaching the limit, the system shall evict the oldest cached data (dose history older than 30 days first).
- [ ] **REQ-OFFLINE-012**: The system shall request persistent storage permission (`navigator.storage.persist()`) on first install to prevent the browser from evicting data under storage pressure.

### 3.8 Medication Refill Management

#### 3.8.1 Supply Tracking

- [ ] **REQ-REFILL-001**: Users with `medication:write` permission shall be able to set a `current_quantity` and `refill_threshold` when creating or editing a medication. If `current_quantity` is non-null, `refill_threshold` must also be non-null (and vice versa).
- [ ] **REQ-REFILL-002**: Supply tracking is optional; if `current_quantity` is null, all refill features are disabled for that medication.
- [ ] **REQ-REFILL-003**: The system shall auto-decrement `current_quantity` by `dosage_amount` each time a dose is logged as `taken`.
  - **Note**: For medications where `dosage_amount` does not map 1:1 to supply units (e.g., liquid medications measured in ml), users should set `current_quantity` to the total units available and adjust `refill_threshold` accordingly. The system does not perform unit conversion; it decrements by the raw `dosage_amount` value.
- [ ] **REQ-REFILL-004**: Auto-decrement shall not reduce `current_quantity` below zero.
- [ ] **REQ-REFILL-005**: Users with `medication:read` permission shall be able to view the current supply status for each medication via `GET /api/profiles/{id}/medications/{medId}/supply`.

#### 3.8.2 Low-Supply Alerts

- [ ] **REQ-REFILL-006**: The system shall send a low-supply notification when `current_quantity` falls at or below `refill_threshold` after a dose is logged.
- [ ] **REQ-REFILL-007**: Low-supply notifications shall be delivered via the profile's configured notification channels (WhatsApp, Telegram, Web Push) and shall respect quiet hours settings.
- [ ] **REQ-REFILL-008**: The system shall not send duplicate low-supply alerts; a single alert is sent when the threshold is first crossed. Alerts reset after a refill raises `current_quantity` above `refill_threshold`.

#### 3.8.3 Refill Logging

- [ ] **REQ-REFILL-009**: Users with `medication:write` permission shall be able to log a refill (`POST /api/profiles/{id}/medications/{medId}/refills`), specifying `quantity_added` and optional `notes`.
- [ ] **REQ-REFILL-010**: Logging a refill shall add `quantity_added` to `current_quantity`.
- [ ] **REQ-REFILL-011**: Users with `medication:read` permission shall be able to view refill history (`GET /api/profiles/{id}/medications/{medId}/refills`), paginated and ordered by `logged_at` descending.

#### 3.8.4 Refill Reminders

- [ ] **REQ-REFILL-012**: The system shall calculate a `projected_depletion_date` based on `current_quantity`, `dosage_amount`, and the medication's frequency configuration.
- [ ] **REQ-REFILL-013**: Users shall be able to configure a `refill_reminder_days` value on a medication — a positive integer specifying how many days before projected depletion to send a refill reminder.
- [ ] **REQ-REFILL-014**: The system shall send a refill reminder notification at `projected_depletion_date − refill_reminder_days`, delivered via the profile's configured notification channels.
- [ ] **REQ-REFILL-015**: If `current_quantity`, `dosage_amount`, or frequency changes, the system shall recalculate `projected_depletion_date` and reschedule the refill reminder accordingly.
- [ ] **REQ-REFILL-016**: If `projected_depletion_date` cannot be calculated (e.g., PRN medications with no fixed frequency), refill reminders shall be disabled for that medication.
- [ ] **REQ-REFILL-017**: If `projected_depletion_date` is in the past (supply already depleted), the system shall send a refill reminder immediately upon the next dose logging or profile access, rather than waiting for a future date. Once a refill is logged that raises `current_quantity` above `refill_threshold`, the system shall resume normal scheduled refill reminders.

### 3.9 AI-Assisted Prescription Extraction

#### 3.9.1 AI Provider Management

- [ ] **REQ-EXTRACT-001**: An authenticated user shall be able to register an AI provider connection by supplying a `provider` name and `api_key` (`POST /api/ai-providers`).
- [ ] **REQ-EXTRACT-002**: The system shall store the API key encrypted at rest (AES-256-GCM). The plaintext key shall never be returned in any API response after registration.
- [ ] **REQ-EXTRACT-003**: A user may hold at most one active connection per provider. Registering a second key for the same provider shall replace the existing key.
- [ ] **REQ-EXTRACT-004**: The system shall initially support `gemini` as the only valid provider value. The architecture shall be extensible to add providers without database schema changes.
- [ ] **REQ-EXTRACT-005**: An authenticated user shall be able to list their connected providers — showing `provider` and `connected_at` but never the API key (`GET /api/ai-providers`).
- [ ] **REQ-EXTRACT-006**: An authenticated user shall be able to remove a provider connection (`DELETE /api/ai-providers/{providerId}`). Removal does not affect previously confirmed medications.

#### 3.9.2 Prescription Extraction

- [ ] **REQ-EXTRACT-007**: Users with `prescription:read` permission on a profile shall be able to trigger extraction on an uploaded prescription (`POST /api/prescriptions/{prescriptionId}/extract`). The requesting user must have at least one AI provider connection configured.
- [ ] **REQ-EXTRACT-008**: The system shall submit the prescription file to the user's configured AI provider using the stored (decrypted) API key and request structured medication data.
- [ ] **REQ-EXTRACT-009**: A successful extraction response shall return an ordered list of medication candidates. Each candidate shall include the following fields where identifiable:
  - `name` (string)
  - `dosage_amount` (decimal, nullable)
  - `dosage_unit` (string, nullable — mapped to a supported unit from REQ-MED-002 where possible, otherwise the raw extracted string)
  - `form` (string, nullable — mapped to a supported form from REQ-MED-003 where possible, otherwise raw string)
  - `frequency_type` (string, nullable — mapped to a supported type from REQ-MED-007, or null if not determinable)
  - `frequency_notes` (string, nullable — raw frequency text from the prescription, e.g., "twice daily after meals")
  - `prescriber_name` (string, nullable)
  - `prescriber_clinic` (string, nullable)
  - `prescriber_phone` (string, nullable)
- [ ] **REQ-EXTRACT-010**: Extraction results are suggestions only. The system shall not automatically create medication records.
- [ ] **REQ-EXTRACT-011**: A user with `medication:write` permission on the profile shall be able to confirm one or more candidates (`POST /api/prescriptions/{prescriptionId}/extractions/{extractionId}/confirm`), which creates medication records subject to the same validation as REQ-MED-001 through REQ-MED-008.
- [ ] **REQ-EXTRACT-012**: The system shall handle extraction failures gracefully, returning a descriptive error without creating records, for: unreachable provider, invalid/revoked API key, unreadable prescription file, and unparseable provider response.
- [ ] **REQ-EXTRACT-013**: Each prescription may have at most one stored extraction result at a time. Re-triggering extraction replaces the previous result and its unconfirmed candidates.
- [ ] **REQ-EXTRACT-014**: The system shall record the provider name and model identifier (if returned by the provider) in the extraction result for auditability.
- [ ] **REQ-EXTRACT-015**: Extraction is performed synchronously. If the provider does not respond within 30 seconds, the request shall time out with an error. No background retry is performed; the user may re-trigger manually.

### 3.10 Prescriber Management

- [ ] **REQ-PRESCRIBER-001**: Users with `prescription:write` permission on a profile shall be able to create a Prescriber record, specifying `name` (required), `clinic` (optional), and `phone` (optional) (`POST /api/profiles/{id}/prescribers`).
- [ ] **REQ-PRESCRIBER-002**: Users with `prescription:read` permission on a profile shall be able to list all Prescriber records for that profile (`GET /api/profiles/{id}/prescribers`).
- [ ] **REQ-PRESCRIBER-003**: Users with `prescription:write` permission shall be able to update a Prescriber record's `name`, `clinic`, or `phone` (`PUT /api/profiles/{id}/prescribers/{prescriberId}`).
- [ ] **REQ-PRESCRIBER-004**: Users with `prescription:write` permission shall be able to delete a Prescriber record (`DELETE /api/profiles/{id}/prescribers/{prescriberId}`), provided no prescriptions are currently linked to it. Attempting deletion of a linked Prescriber shall return an error.
- [ ] **REQ-PRESCRIBER-005**: When uploading a prescription, the user must supply either an existing `prescriber_id` or a new prescriber object (`name`, `clinic`, `phone`). If a new prescriber object is supplied, the system shall create the Prescriber record and link it to the prescription atomically.
- [ ] **REQ-PRESCRIBER-006**: When AI extraction returns prescriber fields (`prescriber_name`, `prescriber_clinic`, `prescriber_phone`), the confirmation response shall include those fields so the client can pre-populate a new-prescriber form or let the user select an existing Prescriber. Prescriber creation or linking is resolved at confirmation time per REQ-PRESCRIBER-005.

### 3.11 Follow-up Appointment Alerts

#### 3.11.1 Appointment Management

- [ ] **REQ-FOLLOWUP-001**: Users with `prescription:write` permission shall be able to create a follow-up appointment record (`POST /api/profiles/{id}/follow-ups`), specifying `scheduled_date` (required, ISO 8601 date), `prescriber_id` (optional FK → Prescriber), `prescription_id` (optional FK → Prescription), `notes` (optional), and `advance_notice_days` (optional integer, default 3).
- [ ] **REQ-FOLLOWUP-002**: Users with `prescription:read` permission shall be able to list all follow-up appointments for a profile (`GET /api/profiles/{id}/follow-ups`), ordered by `scheduled_date` ascending. Each record shall include a computed `is_overdue` boolean (true when `scheduled_date` is in the past and `status` is `upcoming`).
- [ ] **REQ-FOLLOWUP-003**: Users with `prescription:write` permission shall be able to update a follow-up appointment's `scheduled_date`, `prescriber_id`, `prescription_id`, `notes`, `advance_notice_days`, or `status` (`PUT /api/profiles/{id}/follow-ups/{followUpId}`). Changing `scheduled_date` or `advance_notice_days` shall reschedule the pending alert.
- [ ] **REQ-FOLLOWUP-004**: Users with `prescription:write` permission shall be able to delete a follow-up appointment (`DELETE /api/profiles/{id}/follow-ups/{followUpId}`).

#### 3.11.2 Follow-up Notifications

- [ ] **REQ-FOLLOWUP-005**: The system shall send a follow-up alert notification at 09:00 in the profile's timezone on `scheduled_date − advance_notice_days`. If `advance_notice_days` is 0, the alert fires at 09:00 on `scheduled_date` itself.
- [ ] **REQ-FOLLOWUP-006**: Follow-up alerts shall be delivered via the profile's configured notification channels (WhatsApp, Telegram, Web Push) and shall respect quiet hours settings (REQ-NOT-008).
- [ ] **REQ-FOLLOWUP-007**: The alert message shall include the appointment date, prescriber name (if linked), and notes (if present).
- [ ] **REQ-FOLLOWUP-008**: Alerts shall only be sent for follow-up appointments with `status = upcoming`. Appointments marked `completed` or `dismissed` shall not trigger alerts.
- [ ] **REQ-FOLLOWUP-009**: The system shall not resend an alert for the same follow-up appointment once it has been dispatched. If `scheduled_date` or `advance_notice_days` is updated after an alert has already fired, no second alert is sent unless the new alert date (`scheduled_date − advance_notice_days`) is in the future.
  - **Behavior on update**: If the new calculated alert date is in the future, a new alert shall be scheduled. If the new alert date is in the past (including advancing from 3 to 7 days notice, which makes the alert date earlier), no alert shall fire — the user must manually reschedule to a future date to receive an alert.
- [ ] **REQ-FOLLOWUP-010**: The system shall log all follow-up notification delivery attempts in the `notification_logs` table (with `reminder_id = null`).

---

## 4. Non-Functional Requirements

### 4.1 Performance

- [ ] **REQ-PERF-001**: API response time shall be under 200ms at p95 for read operations.
- [ ] **REQ-PERF-002**: API response time shall be under 500ms at p95 for write operations.
- [ ] **REQ-PERF-003**: The system shall support at least 100 concurrent users per instance.
- [ ] **REQ-PERF-004**: All list endpoints shall support pagination with `limit` (default 20, max 100) and `offset` parameters.

### 4.2 Security

- [ ] **REQ-SEC-001**: All passwords shall be hashed using bcrypt with cost factor 12.
- [ ] **REQ-SEC-002**: All API endpoints shall require authentication except: `/healthz`, `/api/auth/register`, `/api/auth/login`, `/api/auth/oauth/{provider}`, `/api/auth/oauth/{provider}/callback`, `/api/auth/oauth/token`, `/api/auth/refresh`, `/api/auth/email/verify`, `/api/auth/password/reset/request`, `/api/auth/password/reset/confirm`.
- [ ] **REQ-SEC-003**: All data in transit shall be encrypted using TLS 1.2 or higher.
- [ ] **REQ-SEC-004**: JWT tokens shall include expiration claims and be validated on every authenticated request.
- [ ] **REQ-SEC-005**: Guest access tokens shall be cryptographically random (minimum 32 bytes) and stored hashed in the database.
- [ ] **REQ-SEC-006**: The system shall implement rate limiting:
  - Auth endpoints: 100 requests per minute per IP
  - Password reset request: 5 requests per hour per IP (to prevent email spam)
  - Guest access endpoints: 60 requests per minute per IP
  - All other endpoints: 300 requests per minute per authenticated user
- [ ] **REQ-SEC-007**: The system shall log all authentication events (login, logout, password reset, failed login attempts, OAuth connections) for audit purposes in an `audit_logs` table.
- [ ] **REQ-SEC-008**: All API requests shall include a `X-Request-ID` header (server-generated if not provided by client) for traceability.
- [ ] **REQ-SEC-009**: The server shall generate and securely store a VAPID key pair (EC P-256) at startup if one does not exist. The private key shall never be exposed via any API endpoint. The public key is exposed via `GET /api/push/vapid-public-key`.
- [ ] **REQ-SEC-010**: AI provider API keys shall be encrypted using AES-256-GCM before storage. The encryption key shall be loaded from server configuration and shall never be stored in the database.
- [ ] **REQ-SEC-011**: The extraction endpoint (`POST /api/prescriptions/{prescriptionId}/extract`) shall be rate-limited to 10 requests per hour per authenticated user.

### 4.3 Validation

- [ ] **REQ-VAL-001**: Email addresses shall be validated for proper format.
- [ ] **REQ-VAL-002**: Passwords shall require minimum 8 characters, at least 1 uppercase letter, 1 lowercase letter, and 1 number.
- [ ] **REQ-VAL-003**: Profile names shall be limited to 100 characters.
- [ ] **REQ-VAL-004**: Medication names shall be limited to 200 characters.
- [ ] **REQ-VAL-005**: Dose notes shall be limited to 500 characters.
- [ ] **REQ-VAL-006**: All user inputs shall be sanitized to prevent XSS and SQL injection.
- [ ] **REQ-VAL-007**: Timezone values shall be validated against the IANA timezone database.
- [ ] **REQ-VAL-008**: Display names shall be limited to 100 characters and may not be empty.

### 4.4 Availability

- [ ] **REQ-AVL-001**: The system shall maintain 99.9% uptime (excluding scheduled maintenance).
- [ ] **REQ-AVL-002**: The system shall have a health check endpoint at `/healthz` that returns the database connectivity status.

### 4.5 Data Retention

- [ ] **REQ-DATA-001**: Dose history shall be retained for 2 years by default.
- [ ] **REQ-DATA-002**: Users shall be able to request full data deletion (GDPR Right to Erasure) via `DELETE /api/auth/account`. Personal data shall be deleted or anonymized within 30 days.

### 4.6 Scalability

- [ ] **REQ-SCL-001**: The system shall support horizontal scaling with stateless API instances (JWT-based auth requires no server-side session state).
- [ ] **REQ-SCL-002**: The database shall be separated from the application layer to allow independent scaling.

### 4.7 Observability

- [ ] **REQ-OBS-001**: The system shall emit structured JSON logs for all requests, including method, path, status code, latency, and request ID.
- [ ] **REQ-OBS-002**: The system shall log all background job executions (reminder generation, DGDA sync, auto-skip) with outcome and duration.
- [ ] **REQ-OBS-003**: The `/healthz` endpoint shall report application version, uptime, and database connectivity status.

---

## 5. Data Models (Overview)

All `id` fields use UUID v4. All timestamps are stored in UTC.

### 5.1 User

- `id`, `email`, `display_name`, `password_hash` (nullable), `email_verified` (boolean), `created_at`, `updated_at`

### 5.2 OAuthAccount

Extensible — supports multiple providers per user.

- `id`, `user_id` (FK → User), `provider` (e.g., "google", "github"), `provider_user_id`, `connected_at`, `created_at`
- UNIQUE(`provider`, `provider_user_id`)
- UNIQUE(`user_id`, `provider`)

### 5.3 OAuthAuthorizationCode

Short-lived codes for the OAuth token exchange flow (REQ-OAUTH-004).

- `id`, `user_id` (FK → User), `code` (hashed), `expires_at`, `created_at`

### 5.4 EmailChangeRequest

- `id`, `user_id` (FK → User), `new_email`, `verification_token` (hashed), `expires_at`, `created_at`

### 5.5 AuditLog

- `id`, `user_id` (FK → User, nullable for unauthenticated events), `event_type` (e.g., `login`, `logout`, `password_reset`, `login_failed`, `oauth_connected`), `ip_address`, `user_agent`, `created_at`

### 5.6 Profile

- `id`, `owner_user_id` (FK → User), `name`, `avatar_url`, `date_of_birth`, `medical_conditions`, `timezone` (IANA identifier, e.g., `Asia/Dhaka`), `created_at`, `updated_at`

### 5.7 ProfilePermission

- `id`, `profile_id` (FK → Profile), `shared_with_user_id` (FK → User), `permissions` (JSONB array of permission strings), `granted_by_user_id` (FK → User), `status` (`pending` | `accepted` | `declined`), `expires_at` (nullable — invitation expiry only; null after acceptance), `created_at`, `updated_at`

### 5.8 ProfileLink (Guest Access)

- `id`, `profile_id` (FK → Profile), `token_hash` (hashed token; the raw token is only returned once at creation), `expires_at`, `created_at`

### 5.9 DoseSchedule

- `id`, `profile_id` (FK → Profile), `name` (e.g., "Breakfast"), `time` (wall-clock time as `HH:MM:SS`, interpreted in profile timezone), `description` (optional), `created_at`, `updated_at`

### 5.10 Medication

- `id`, `profile_id` (FK → Profile), `name`, `dosage_amount` (decimal), `dosage_unit` (enum), `form` (enum), `frequency_type` (enum: `schedule`, `daily`, `weekly`, `monthly`, `interval`, `prn`), `frequency_config` (JSONB — schema defined per frequency type in Section 3.3.2), `start_date` (nullable), `end_date` (nullable), `prescriber_id` (FK → Prescriber, nullable)
- `current_quantity` (decimal, nullable — **nullable; null disables all refill tracking features**), `refill_threshold` (decimal, nullable — requires `current_quantity` to be non-null), `refill_reminder_days` (integer, nullable — requires `current_quantity` to be non-null), `created_at`, `updated_at`

### 5.11 Reminder

- `id`, `medication_id` (FK → Medication), `scheduled_at` (UTC timestamp), `enabled` (boolean), `created_at`, `updated_at`

### 5.12 Dose

- `id`, `reminder_id` (FK → Reminder, nullable for manual/PRN doses), `medication_id` (FK → Medication), `scheduled_at` (UTC timestamp), `status` (`taken` | `skipped` | `snoozed`), `notes` (nullable), `logged_at` (UTC timestamp), `is_auto_logged` (boolean)

### 5.13 NotificationPreference

- `id`, `profile_id` (FK → Profile), `whatsapp_enabled` (boolean), `whatsapp_phone` (nullable), `telegram_enabled` (boolean), `telegram_chat_id` (nullable), `web_push_enabled` (boolean, default false), `quiet_hours_start` (nullable, `HH:MM`), `quiet_hours_end` (nullable, `HH:MM`), `advance_minutes` (integer, default 0)

### 5.17 PushSubscription

Per-device Web Push subscription. A user may have multiple active subscriptions.

- `id`, `user_id` (FK → User), `endpoint` (text, unique), `p256dh` (text), `auth` (text), `created_at`, `last_used_at`

### 5.14 NotificationLog

- `id`, `profile_id` (FK → Profile), `reminder_id` (FK → Reminder, nullable), `channel` (`whatsapp` | `telegram`), `status` (`sent` | `failed` | `suppressed`), `attempt_number` (integer), `error_message` (nullable), `created_at`

### 5.15 Prescription

- `id`, `profile_id` (FK → Profile), `prescriber_id` (FK → Prescriber, **required**), `medication_id` (FK → Medication, nullable), `file_url`, `file_type` (`pdf` | `jpg` | `png`), `file_size` (bytes), `uploaded_by_user_id` (FK → User), `uploaded_at`

### 5.18 RefillLog

- `id`, `medication_id` (FK → Medication), `quantity_added` (decimal), `notes` (nullable), `logged_by_user_id` (FK → User), `logged_at` (UTC timestamp)

### 5.19 AIProvider

Stores a user's connection to an external AI provider. One record per user per provider.

- `id` (UUID), `user_id` (FK → User), `provider` (string — e.g., `gemini`; validated at application layer for extensibility without schema changes), `encrypted_api_key` (text — AES-256-GCM ciphertext including nonce), `connected_at` (UTC timestamp), `updated_at` (UTC timestamp)
- UNIQUE(`user_id`, `provider`)

### 5.20 ExtractionResult

Stores the result of a single AI extraction job for a prescription. At most one record per prescription at a time.

- `id` (UUID), `prescription_id` (FK → Prescription), `performed_by_user_id` (FK → User), `ai_provider` (string — provider name used, e.g., `gemini`), `ai_model` (string, nullable — model identifier returned by provider), `candidates` (JSONB — ordered array of candidate objects per REQ-EXTRACT-009), `status` (enum: `success` | `failed`), `error_message` (text, nullable), `created_at` (UTC timestamp)

### 5.21 Prescriber

Profile-scoped record of a healthcare professional who issues prescriptions.

- `id`, `profile_id` (FK → Profile), `name`, `clinic` (nullable), `phone` (nullable), `created_at`, `updated_at`

### 5.22 FollowUpAppointment

Doctor-requested return visit recorded against a profile.

- `id`, `profile_id` (FK → Profile), `prescriber_id` (FK → Prescriber, nullable), `prescription_id` (FK → Prescription, nullable), `scheduled_date` (date), `notes` (nullable text), `advance_notice_days` (integer, default 3), `status` (enum: `upcoming` | `completed` | `dismissed`), `alert_sent_at` (UTC timestamp, nullable — set when the alert is dispatched), `created_at`, `updated_at`

### 5.16 MedicineDatabase

Local copy of DGDA medicine data used for auto-suggestion.

- `id`, `name`, `generic_name` (nullable), `manufacturer` (nullable), `last_synced_at`

---

## 6. API Endpoints Overview

All endpoints are prefixed with `/api/v1`. Authenticated endpoints require a valid JWT access token in the `Authorization: Bearer {token}` header unless otherwise noted.

### 6.1 Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/auth/register | No | Register new user |
| POST | /api/auth/login | No | Login with email/password |
| POST | /api/auth/refresh | No | Refresh access token |
| POST | /api/auth/logout | Yes | Invalidate refresh token |
| GET | /api/auth/oauth/{provider} | No | Redirect to OAuth provider |
| GET | /api/auth/oauth/{provider}/callback | No | OAuth callback; issues auth code, redirects to frontend |
| POST | /api/auth/oauth/token | No | Exchange OAuth auth code for JWT tokens |
| POST | /api/auth/connect/{provider} | Yes | Link OAuth provider to existing account |
| POST | /api/auth/disconnect/{provider} | Yes | Unlink OAuth provider |
| POST | /api/auth/set-password | Yes | Set password for OAuth-only user |
| PUT | /api/auth/password | Yes | Change existing password |
| POST | /api/auth/password/reset/request | No | Request password reset email |
| POST | /api/auth/password/reset/confirm | No | Set new password using reset token |
| POST | /api/auth/email/resend-verification | No | Resend email verification link |
| PUT | /api/auth/email | Yes | Request email address change |
| POST | /api/auth/email/verify | No | Verify new email with token |
| POST | /api/auth/email/cancel | Yes | Cancel pending email change |
| POST | /api/auth/telegram/link | Yes | Initiate Telegram account linking |
| DELETE | /api/auth/account | Yes | Delete account and all owned data |

### 6.2 Profiles

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles | Yes | List profiles user owns or has access to |
| POST | /api/profiles | Yes | Create profile |
| GET | /api/profiles/{id} | Yes | Get profile details |
| PUT | /api/profiles/{id} | Yes | Update profile (`profile:write`) |
| DELETE | /api/profiles/{id} | Yes | Delete profile (`profile:admin`) |
| POST | /api/profiles/{id}/transfer-ownership | Yes | Transfer profile ownership (`profile:admin`) |
| POST | /api/profiles/{id}/share | Yes | Share profile with user (`profile:share` or `profile:admin`) |
| GET | /api/profiles/{id}/share | Yes | List users with profile access (`profile:share` or `profile:admin`) |
| PUT | /api/profiles/{id}/share/{userId} | Yes | Update a user's permissions (`profile:admin`) |
| DELETE | /api/profiles/{id}/share/{userId} | Yes | Revoke a user's access (`profile:admin`) |
| POST | /api/profiles/{id}/share/guest | Yes | Generate guest access link (`profile:admin`) |
| DELETE | /api/profiles/{id}/share/guest/{tokenId} | Yes | Revoke guest access token (`profile:admin`) |
| GET | /api/invitations | Yes | List pending invitations for current user |
| POST | /api/invitations/{id}/accept | Yes | Accept profile invitation |
| POST | /api/invitations/{id}/decline | Yes | Decline profile invitation |

### 6.3 Dose Schedules

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/schedules | Yes | List dose schedules for profile |
| POST | /api/profiles/{id}/schedules | Yes | Create dose schedule |
| GET | /api/profiles/{id}/schedules/{scheduleId} | Yes | Get dose schedule |
| PUT | /api/profiles/{id}/schedules/{scheduleId} | Yes | Update dose schedule |
| DELETE | /api/profiles/{id}/schedules/{scheduleId} | Yes | Delete dose schedule |

### 6.4 Medications

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/medications | Yes | List medications for profile (`medication:read`) |
| POST | /api/profiles/{id}/medications | Yes | Create medication (`medication:write`) |
| GET | /api/profiles/{id}/medications/{medId} | Yes | Get medication details (`medication:read`) |
| PUT | /api/profiles/{id}/medications/{medId} | Yes | Update medication (`medication:write`) |
| DELETE | /api/profiles/{id}/medications/{medId} | Yes | Delete medication (`medication:write`) |
| POST | /api/profiles/{id}/medications/{medId}/notify | Yes | Request PRN on-demand notification (`dose:write`) |
| GET | /api/medications/suggest | Yes | Suggest medication names by query (`?q=`) |
| GET | /api/profiles/{id}/medications/{medId}/supply | Yes | Get supply status and projected depletion date (`medication:read`) |
| GET | /api/profiles/{id}/medications/{medId}/refills | Yes | List refill history (`medication:read`) |
| POST | /api/profiles/{id}/medications/{medId}/refills | Yes | Log a refill (`medication:write`) |

### 6.5 Prescriptions

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/prescriptions | Yes | List prescriptions (`prescription:read`) |
| POST | /api/profiles/{id}/prescriptions | Yes | Upload prescription (`prescription:write`) |
| GET | /api/prescriptions/{prescriptionId} | Yes | Download prescription (`prescription:read`) |
| DELETE | /api/prescriptions/{prescriptionId} | Yes | Delete prescription (`prescription:write`) |
| POST | /api/prescriptions/{prescriptionId}/extract | Yes | Trigger AI extraction (`prescription:read`; requires AI provider configured) |
| GET | /api/prescriptions/{prescriptionId}/extractions/latest | Yes | Get most recent extraction result (`prescription:read`) |
| POST | /api/prescriptions/{prescriptionId}/extractions/{extractionId}/confirm | Yes | Confirm candidates to create medications (`medication:write`) |

### 6.6 Reminders

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/reminders | Yes | List upcoming reminders for profile (`reminder:read`) |
| GET | /api/reminders/{reminderId} | Yes | Get reminder details (`reminder:read`) |
| PUT | /api/reminders/{reminderId} | Yes | Update reminder (`reminder:write`) |
| DELETE | /api/reminders/{reminderId} | Yes | Delete reminder (`reminder:write`) |
| POST | /api/reminders/{reminderId}/snooze | Yes | Snooze reminder (`reminder:write`) |

### 6.7 Doses

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/profiles/{id}/doses | Yes | Log a dose (`dose:write`) |
| GET | /api/profiles/{id}/doses | Yes | Get dose history (`dose:read`) |
| GET | /api/profiles/{id}/doses/calendar | Yes | Get dose calendar view (`dose:read`) |
| POST | /api/profiles/{id}/doses/batch | Yes | Batch log doses for a schedule (`dose:write`) |

### 6.8 Notification Preferences & Logs

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/notification-preferences | Yes | Get notification preferences |
| PUT | /api/profiles/{id}/notification-preferences | Yes | Update notification preferences |
| GET | /api/profiles/{id}/notification-logs | Yes | View notification delivery history |

### 6.9 Web Push

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/push/vapid-public-key | No | Get server VAPID public key for browser subscription |
| POST | /api/push/subscribe | Yes | Register a push subscription (`endpoint`, `p256dh`, `auth`) |
| DELETE | /api/push/subscribe | Yes | Remove a push subscription (by `endpoint`) |

### 6.10 Webhooks & Infrastructure

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | /api/webhooks/whatsapp | Signature | WhatsApp incoming webhook |
| POST | /api/webhooks/telegram | Signature | Telegram incoming webhook |
| GET | /healthz | No | Health check (DB status, uptime, version) |

### 6.11 AI Providers

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/ai-providers | Yes | List user's connected AI providers (keys never returned) |
| POST | /api/ai-providers | Yes | Register or replace an AI provider API key |
| DELETE | /api/ai-providers/{providerId} | Yes | Remove an AI provider connection |

### 6.12 Prescribers

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/prescribers | Yes | List prescribers for profile (`prescription:read`) |
| POST | /api/profiles/{id}/prescribers | Yes | Create prescriber (`prescription:write`) |
| PUT | /api/profiles/{id}/prescribers/{prescriberId} | Yes | Update prescriber (`prescription:write`) |
| DELETE | /api/profiles/{id}/prescribers/{prescriberId} | Yes | Delete prescriber (`prescription:write`) |

### 6.13 Follow-up Appointments

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | /api/profiles/{id}/follow-ups | Yes | List follow-up appointments (`prescription:read`) |
| POST | /api/profiles/{id}/follow-ups | Yes | Create follow-up appointment (`prescription:write`) |
| PUT | /api/profiles/{id}/follow-ups/{followUpId} | Yes | Update follow-up appointment (`prescription:write`) |
| DELETE | /api/profiles/{id}/follow-ups/{followUpId} | Yes | Delete follow-up appointment (`prescription:write`) |

---

## 7. API Response Format

All API responses use [Huma v2](https://huma.rocks/) framework which provides a consistent, simple JSON format. The response body is defined by each endpoint's output struct.

### Simple Response

```json
{
  "status": "ok",
  "timestamp": "2026-03-14T10:30:00Z"
}
```

### List Response (Paginated)

```json
{
  "items": [],
  "limit": 20,
  "offset": 0,
  "total": 150
}
```

### Error Response

Huma uses [RFC 9457](https://tools.ietf.org/html/rfc9457) Problem Details for HTTP APIs:

```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "Invalid email format",
  "errors": [
    {"field": "email", "message": "must be a valid email address"}
  ]
}
```

> **Note**: Unlike the traditional `{ "error": { "code": ..., "message": ... } }` format, Huma returns errors as Problem Details with `application/problem+json` content-type. Clients should handle both formats for compatibility.

All timestamps in API responses are in ISO 8601 format (UTC). All datetime storage in the database is in UTC. Display times in client applications shall be converted to the **profile's configured timezone** (not the viewing user's timezone), since profiles represent specific individuals who live in a specific timezone.

---

