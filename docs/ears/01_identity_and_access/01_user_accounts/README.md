# User Accounts

Users create accounts to access AI Together. Accounts can be standalone (single-user) or connected to a team server.

## Purpose

Allow users to securely access the system with appropriate permissions based on their role.

## User Personas

- **Organization Creator** - First user creating a new organization
- **Manager** - Team administrator who can invite members and manage configuration
- **Member** - Team member who uses AI tools with provided configuration

## Functional Requirements (EARS Format)

### 1. Account Creation

**User Story:** As a new organization creator, I want to create my organization so my team can start using AI Together.

#### Ubiquitous Requirements (Password Rules)

- **UA-01-001:** `The system shall require a minimum password length of 12 characters.`
- **UA-01-002:** `The system shall require passwords to include uppercase letters.`
- **UA-01-003:** `The system shall require passwords to include lowercase letters.`
- **UA-01-004:** `The system shall require passwords to include numbers.`
- **UA-01-005:** `The system shall require passwords to include special characters.`
- **UA-01-006:** `The system shall prevent reuse of the last 5 passwords.`
- **UA-01-007:** `The system shall create a new organization when a user self-registers.`
- **UA-01-008:** `The system shall grant the Manager role to the first user in a new organization.`

#### Event-Driven Requirements (Registration Workflow)

- **UA-01-101:** `When a user navigates to the registration page, the system shall display the registration form.`
- **UA-01-102:** `When a user enters an email address and password, the system shall validate the password requirements.`
- **UA-01-103:** `When a user submits valid registration data, the system shall create a new organization.`
- **UA-01-104:** `When a user submits valid registration data, the system shall assign the Manager role to the user.`
- **UA-01-105:** `When a user completes registration, the system shall log the user in.`
- **UA-01-106:** `When a user completes registration, the system shall redirect the user to the dashboard.`

---

### 2. Team Invitation

**User Story:** As a manager, I want to invite team members so they can join our organization.

#### Event-Driven Requirements (Invitation Workflow)

- **UA-01-201:** `When a manager enters a team member's email address, the system shall send an invitation email.`
- **UA-01-202:** `When a manager sends an invitation, the system shall deliver the invitation email within 30 seconds.`
- **UA-01-203:** `When a manager sends an invitation, the system shall create an invitation link.`
- **UA-01-204:** `When a manager sends an invitation, the system shall set the invitation to expire after 7 days.`
- **UA-01-205:** `When an invited user clicks the invitation link, the system shall display the password setup form.`
- **UA-01-206:** `When an invited user sets their password, the system shall add the user to the organization with Member role.`
- **UA-01-207:** `When an invited user completes setup, the system shall allow the user to log in.`
- **UA-01-208:** `When a manager requests to cancel a pending invitation, the system shall invalidate the invitation link.`

#### Unwanted Behaviour Requirements (Invitation Errors)

- **UA-01-301:** `If a manager enters an email address that already exists in the organization, then the system shall display an error message.`
- **UA-01-302:** `If an invitation email is not delivered, then the system shall allow the manager to resend the invitation.`
- **UA-01-303:** `If a user clicks an expired invitation link, then the system shall display a link expiration message.`
- **UA-01-304:** `If a user clicks an expired invitation link, then the system shall allow the manager to send a new invitation.`

---

### 3. User Deactivation

**User Story:** As a manager, I want to deactivate users who leave the organization so they no longer have access.

#### Event-Driven Requirements (Deactivation Workflow)

- **UA-01-401:** `When a manager navigates to user management, the system shall display the list of users.`
- **UA-01-402:** `When a manager selects a user to deactivate, the system shall display a confirmation dialog.`
- **UA-01-403:** `When a manager confirms user deactivation, the system shall deactivate the user account.`
- **UA-01-404:** `When a user account is deactivated, the system shall terminate all active sessions for the user.`
- **UA-01-405:** `When a user account is deactivated, the system shall preserve the user's data per retention policy.`
- **UA-01-406:** `When a user account is deactivated, the system shall free up the seat for a new user.`
- **UA-01-407:** `When a manager reactivates a user within the retention period, the system shall restore the user's previous role.`

#### Unwanted Behaviour Requirements (Deactivation Restrictions)

- **UA-01-501:** `If a manager attempts to deactivate themselves when they are the last Manager, then the system shall deny the deactivation operation.`
- **UA-01-502:** `If a manager attempts to deactivate the last Manager in the organization, then the system shall deny the deactivation operation.`
- **UA-01-503:** `If a manager attempts to invite a deactivated user's email address, then the system shall offer to reactivate the existing user instead of creating a new account.`

#### Complex Requirements (Deactivated User State)

- **UA-01-1002:** `While a user account is deactivated, if the user attempts to log in, then the system shall deny the login operation.`

---

### 4. Authentication & Login

**User Story:** As a returning user, I want to log in with my email and password.

#### Event-Driven Requirements (Login Workflow)

- **UA-01-601:** `When a user enters an email and password, the system shall validate the credentials.`
- **UA-01-602:** `When a user provides valid credentials, the system shall log the user in.`
- **UA-01-603:** `When a user logs in, the system shall redirect the user to the appropriate interface based on their role.`
- **UA-01-604:** `When a user selects "Remember me" during login, the system shall extend the session duration to 7 days.`

#### Unwanted Behaviour Requirements (Failed Login Attempts)

- **UA-01-801:** `If a user has 5 failed login attempts, then the system shall lock the account for 15 minutes.`
- **UA-01-802:** `If a user account is locked, then the system shall display the message "Account locked. Try again in 15 minutes or contact support."`
- **UA-01-803:** `If a user successfully logs in after failed attempts, then the system shall reset the failed attempt counter.`

---

### 5. Session Management

**User Story:** As a user, I want to stay logged in so I don't have to re-enter my password constantly.

#### State-Driven Requirements (Active Sessions)

- **UA-01-701:** `While a user is active, the system shall maintain the login session for 24 hours.`
- **UA-01-702:** `While a user has been inactive for 24 hours, the system shall require the user to re-enter their password.`
- **UA-01-703:** `While a user session expires, the system shall redirect the user to the login page on the next action.`
- **UA-01-704:** `While a user session expires, the system shall display the message "Your session expired. Please log in again."`

#### Event-Driven Requirements (Session Security)

- **UA-01-901:** `When a user requests to log out from all devices, the system shall terminate all active sessions for the user.`

---

### 6. Password Reset

**User Story:** As a user, I want to reset my password if I forget it.

#### Event-Driven Requirements (Password Reset Workflow)

- **UA-01-902:** `When a user requests a password reset, the system shall send a password reset email.`
- **UA-01-903:** `When a system sends a password reset email, the system shall set the reset link to expire after 1 hour.`
- **UA-01-904:** `When a user clicks a valid password reset link, the system shall display the password reset form.`
- **UA-01-905:** `When a user submits a new password via the reset form, the system shall update the user's password.`

---

### 7. User Reactivation

**User Story:** As a manager, I want to reactivate a deactivated user within the retention period.

#### Complex Requirements (Reactivation Behavior)

- **UA-01-1001:** `When a deactivated user is reactivated within the retention period, the system shall restore the user's access with their previous role.`

## Business Rules

- **BR-01-001:** First user in a new organization is automatically granted the Manager role
- **BR-01-002:** An organization must have at least one Manager at all times
- **BR-01-003:** Last Manager cannot be demoted to Member
- **BR-01-004:** Last Manager cannot be deactivated
- **BR-01-005:** Email addresses must be unique within an organization
- **BR-01-006:** Users can belong to only one organization
- **BR-01-007:** Users belong to one team within their organization
- **BR-01-008:** Deactivated users' data is preserved per retention policy
- **BR-01-009:** Deactivated users no longer count against seat limit

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
