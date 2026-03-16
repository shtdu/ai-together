# User Accounts

Users create accounts to access AI Together. Accounts can be standalone (single-user) or connected to a team server.

## Purpose

Allow users to securely access the system with appropriate permissions based on their role.

## User Personas

- **Organization Creator** - First user creating a new organization
- **Manager** - Team administrator who can invite members and manage configuration
- **Member** - Team member who uses AI tools with provided configuration

## User Stories

- As an **organization creator**, I want to create my organization so my team can start using AI Together
- As a **manager**, I want to invite team members so they can access our team configuration
- As a **manager**, I want to deactivate users who leave our organization
- As a **user**, I want to stay logged in so I don't have to re-enter my password constantly
- As a **user**, I want to log out from all devices if my account is compromised

## Account Creation

### Organization Creation

**User Story:** As a new organization creator, I want to create my organization so my team can start using AI Together.

**Workflow:**
1. User navigates to registration page
2. Enters email address and password
3. System validates password requirements
4. Organization (tenant) is created
5. First user automatically becomes Manager
6. User is logged in and redirected to dashboard

**Password Requirements:**
- Minimum 12 characters
- Must include uppercase, lowercase, number, and special character
- Cannot reuse last 5 passwords

**Note:** Self-registration creates a new organization. Typical users join existing organizations via invitation only.

### Team Invitation

**User Story:** As a manager, I want to invite team members so they can join our organization.

**Who Can Invite:**
- Any Manager can invite team members to the organization

**Workflow:**
1. Manager enters team member's email address
2. System sends invitation email
3. Invited user clicks invitation link
4. User sets their password
5. User is added to organization with Member role (default)
6. User can log in and access team configuration

**Functional Requirements:**
- **FR-001:** System must send invitation email within 30 seconds of request
- **FR-002:** Invitation links must expire after 7 days
- **FR-003:** Managers can cancel pending invitations

### User Deactivation

**User Story:** As a manager, I want to deactivate users who leave the organization so they no longer have access.

**Who Can Deactivate:**
- Any Manager can deactivate users (except themselves if they are the last Manager)

**Workflow:**
1. Manager navigates to user management
2. Selects user to deactivate
3. System confirms deactivation action
4. User account is deactivated (not deleted)
5. User cannot log in (session terminated if active)
6. User's data is preserved for retention period
7. Seat is freed up for new user

**Deactivated User State:**
- Cannot log in (existing sessions terminated)
- Data preserved for analytics and records
- Counts against organization only if hard-deleted
- Can be reactivated by manager within retention period

**Functional Requirements:**
- **FR-004:** System must terminate all active sessions upon deactivation
- **FR-005:** System must preserve user data per retention policy
- **FR-006:** System must free up seat when user is deactivated
- **FR-007:** Last Manager cannot be deactivated

**Edge Cases:**
| Scenario | System Behavior |
|----------|----------------|
| Last manager tries to deactivate self | System prevents action with error |
| Deactivating manager with active sessions | All sessions terminated immediately |
| User reactivated within retention period | Regains access with previous role |
| Trying to invite deactivated user's email | Manager can reactivate instead of creating new account |

## Authentication

### Login

**User Story:** As a returning user, I want to log in with my email and password.

**Workflow:**
1. User enters email and password
2. System validates credentials
3. User is logged in and redirected to appropriate interface

**Session Duration:**
- User remains logged in for 24 hours of activity
- After 24 hours of inactivity, user must re-enter password
- "Remember me" option extends session to 7 days

### Failed Login Attempts

**Behavior:**
- After 5 failed login attempts: account locked for 15 minutes
- User sees message: "Account locked. Try again in 15 minutes or contact support."
- Lockout duration resets after successful login

## Session Security

### Automatic Logout

- User is logged out after 24 hours of inactivity
- On next action, user is redirected to login page
- System shows message: "Your session expired. Please log in again."

### Logout from All Devices

Users can log out from all devices simultaneously (useful if account is compromised).

### Password Reset

- User can request password reset via email
- Reset link expires after 1 hour
- User must set new password after clicking reset link

## Business Rules

- **BR-001:** First user in a new organization is automatically granted the Manager role
- **BR-002:** An organization must have at least one Manager at all times
- **BR-003:** Last Manager cannot be demoted to Member
- **BR-004:** Last Manager cannot be deactivated
- **BR-005:** Email addresses must be unique within an organization
- **BR-006:** Users can belong to only one organization
- **BR-007:** Users belong to one team within their organization
- **BR-008:** Deactivated users' data is preserved per retention policy
- **BR-009:** Deactivated users no longer count against seat limit

## Edge Cases & Error Handling

| Scenario | System Behavior |
|----------|----------------|
| Invitation email not delivered | Manager can resend invitation |
| User forgets password | User can request password reset email |
| Invitation link expired | Manager can send new invitation |
| Account already exists | System shows error to manager |
| Last manager tries to delete self | System prevents deletion with error |

## Success Criteria

- Organization creator can set up account in under 2 minutes
- Team invitations are delivered and accepted successfully
- Invitation emails delivered within 30 seconds
- Users can be deactivated and lose access immediately
- Deactivated users' data is preserved
- Failed login attempts blocked after 5 attempts
- Sessions expire after 24 hours of inactivity

---

**Related:** [01.02 Roles & Permissions](../02_roles_permissions/) | [01.03 Multi-Tenancy](../03_multi_tenancy/) | [01.04 Licensing](../04_licensing/) | [Domain 01 Overview](../README.md)
