# Security

AI Together implements multiple layers of security to protect user data and system access.

## Purpose

Ensure the system is secure against unauthorized access, data breaches, and common security threats.

## Functional Requirements (EARS Format)

### 1. Authentication Security

**Purpose:** Secure user authentication and session management.

#### Ubiquitous Requirements

- **SEC-06-001:** `The system shall require a minimum password length of 12 characters.`
- **SEC-06-002:** `The system shall require passwords to include uppercase, lowercase, numbers, and special characters.`
- **SEC-06-003:** `The system shall hash passwords using bcrypt (one-way encryption) for storage.`
- **SEC-06-004:** `The system shall never store passwords in plain text.`
- **SEC-06-005:** `The system shall prevent reuse of the last 5 passwords.`

#### State-Driven Requirements (Session Management)

- **SEC-06-101:** `While a user session is active, the system shall expire the session after 24 hours of inactivity.`
- **SEC-06-102:** `While a user selects "Remember me", the system shall extend the session to a maximum of 7 days.`
- **SEC-06-103:** `While a user is logged in, the system shall allow the user to manually logout from all devices.`

#### Event-Driven Requirements (Failed Login Protection)

- **SEC-06-104:** `When a user has 5 failed login attempts, the system shall lock the account for 15 minutes.`
- **SEC-06-105:** `When an account is locked, the system shall reset the lockout after a successful login.`

---

### 2. Access Control

**Purpose:** Implement role-based permissions and multi-tenant isolation.

#### Ubiquitous Requirements

- **SEC-06-201:** `The system shall enforce role-based permissions for all protected actions.`
- **SEC-06-202:** `The system shall isolate each organization's data at the database level.`
- **SEC-06-203:** `The system shall prevent cross-tenant data access.`

#### Event-Driven Requirements (Permission Checks)

- **SEC-06-204:** `When a user attempts a protected action, the system shall verify the user's role permissions before executing.`
- **SEC-06-205:** `When a manager attempts an action, the system shall allow provider configuration and team data access.`
- **SEC-06-206:** `When a member attempts a manager-only action, the system shall deny access.`

---

### 3. Data Protection

**Purpose:** Encrypt data in transit and at rest.

#### State-Driven Requirements (Encryption in Transit)

- **SEC-06-301:** `While data is transmitted between client and server, the system shall use HTTPS/TLS encryption.`
- **SEC-06-302:** `While establishing connections, the system shall require TLS 1.3 or higher.`
- **SEC-06-303:** `While establishing connections, the system shall validate certificates properly.`

#### State-Driven Requirements (Encryption at Rest)

- **SEC-06-304:** `While API keys are stored, the system shall encrypt them at rest.`
- **SEC-06-305:** `While passwords are stored, the system shall hash them using one-way encryption.`
- **SEC-06-306:** `While database connections are established, the system shall use encryption.`

#### Event-Driven Requirements (API Key Handling)

- **SEC-06-307:** `When an API key is entered, the system shall never display it in full after initial entry.`
- **SEC-06-308:** `When an API key is retrieved, the system shall only allow replacement (not retrieval).`
- **SEC-06-309:** `When a member attempts to view provider API keys, the system shall deny access.`

---

### 4. Threat Protection

**Purpose:** Protect against common security threats.

#### Event-Driven Requirements (Brute Force Protection)

- **SEC-06-401:** `When multiple failed login attempts occur, the system shall lock the account after 5 attempts.`
- **SEC-06-402:** `When checking login credentials, the system shall not indicate whether an email is registered (prevents enumeration).`

#### Ubiquitous Requirements (SQL Injection Protection)

- **SEC-06-501:** `The system shall use parameterized queries to prevent SQL injection attacks.`
- **SEC-06-502:** `The system shall validate and sanitize all user inputs.`

#### Ubiquitous Requirements (XSS Protection)

- **SEC-06-503:** `The system shall encode user-generated content to prevent XSS attacks.`
- **SEC-06-504:** `The system shall implement Content Security Policy headers.`

---

### 5. Network Security

**Purpose:** Secure network communications.

#### Ubiquitous Requirements

- **SEC-06-601:** `The system shall enforce HTTPS for all client-to-server communications.`
- **SEC-06-602:** `The system shall implement certificate validation for all connections.`
- **SEC-06-603:** `The system shall support only TLS 1.3 or higher.`

## Security Measures

### Authentication

**Password Security:**
- Minimum 12 characters
- Requires uppercase, lowercase, number, special character
- Bcrypt hashing for storage (one-way encryption)
- Passwords never stored in plain text
- Cannot reuse last 5 passwords

**Session Management:**
- Sessions expire after 24 hours of inactivity
- "Remember me" extends to 7 days maximum
- Users can manually logout from all devices

**Failed Login Protection:**
- Account locked after 5 failed attempts
- 15-minute lockout duration
- Lockout resets after successful login

### Access Control

**Role-Based Permissions:**
- Two roles: Manager and Member
- Managers can configure providers and view team data
- Members can use providers and view personal data only
- All actions checked against permissions

**Multi-Tenant Isolation:**
- Each organization's data is completely separated
- Database-level isolation prevents cross-tenant access
- Users can never see other organizations' data

### Data Protection

**In Transit:**
- All connections use HTTPS/TLS encryption
- TLS 1.3 or higher required
- Certificates properly validated

**At Rest:**
- Database connections encrypted
- API keys stored encrypted
- Passwords hashed (one-way encryption)

**API Keys:**
- Never displayed in full after entry
- Can only be replaced, not retrieved
- Members cannot see provider API keys

## Related Documentation

- **Privacy:** [`06.01 Privacy`](../01_privacy/) - Data collection policies
- **Compliance:** [`06.04 Compliance`](../04_compliance/) - Regulatory compliance
