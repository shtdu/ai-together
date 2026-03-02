# Security Policy

## Supported Versions

Currently, only the latest version of Code Together is supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |

## Reporting a Vulnerability

We take the security of Code Together seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please send an email to: [INSERT SECURITY EMAIL HERE]

When reporting a vulnerability, please include:

* **Description**: A clear description of the vulnerability
* **Steps to Reproduce**: Detailed steps to reproduce the issue
* **Impact**: Assessment of the impact and potential risk
* **Proof of Concept**: If possible, include a proof of concept or exploit code

### What Happens Next?

1. **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
2. **Investigation**: We will investigate the vulnerability and determine its severity
3. **Resolution**: We will work to fix the vulnerability and test the fix
4. **Release**: We will release a security update as soon as possible
5. **Disclosure**: We will publicly disclose the vulnerability after the fix is released

### Security Best Practices

If you're using Code Together, we recommend following these security best practices:

* **Keep Updated**: Always use the latest version of Code Together
* **Strong Credentials**: Use strong, unique passwords for all accounts
* **JWT Secret**: Change the default `JWT_SECRET` in production environments
* **Database Credentials**: Use strong database passwords and limit database access
* **HTTPS**: Use HTTPS in production environments
* **Firewall**: Configure appropriate firewall rules to limit access
* **Regular Backups**: Maintain regular backups of your database
* **Audit Logs**: Regularly review access and usage logs

### Security Features

Code Together includes several security features:

* **JWT Authentication**: Token-based authentication for API access
* **Role-Based Access Control (RBAC)**: Fine-grained permissions using Casbin
* **License Verification**: Ed25519-based license verification
* **Multi-Tenant Isolation**: Database-level data segregation
* **Privacy-First**: No prompt/response data stored, only metadata

### Security Audits

We welcome security audits and penetration testing. If you're interested in conducting a security audit, please contact us at the email above.

### Dependency Management

We regularly update dependencies to include security patches. To update dependencies:

```bash
# Update Go dependencies
cd server && go get -u ./... && go mod tidy
cd ../member && go get -u ./... && go mod tidy
cd ../integration && go get -u ./... && go mod tidy

# Update Node.js dependencies
cd manager && pnpm update
```

## License

Code Together is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for details.

## Privacy

Code Together is designed with privacy in mind. We do not store user prompts or AI responses. Only usage metadata is collected for analytics purposes. For more information, see our [Privacy Policy](docs/privacy.md) (if available).

## Contact

For general security questions or inquiries, please contact: [INSERT SECURITY EMAIL HERE]
