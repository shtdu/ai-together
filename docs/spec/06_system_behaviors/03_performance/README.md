# Performance

Code Together is designed for minimal performance impact on AI tool usage.

## Purpose

Ensure the system adds negligible latency to AI tool requests while providing reliable usage tracking and configuration management.

## Performance Expectations

### AI Tool Request Routing

**User Experience:**
- No perceptible delay when using AI tools
- System works transparently in the background
- Routing happens automatically without user action

**Performance Targets:**
- Target: Additional latency < 50ms (p95)
- Maximum acceptable: < 100ms (p95)

### Configuration Sync

**User Experience:**
- Configuration updates happen automatically
- No interruption to work during sync
- Manual refresh when needed

**Performance Targets:**
- Auto-sync: Every 5 minutes (suggested)
- Manual refresh: Complete within 10 seconds
- Manager push: Complete within 10 seconds

### Dashboard & Analytics

**User Experience:**
- Dashboard loads quickly
- Analytics and charts display without long waits
- Filters and searches respond instantly

**Performance Targets:**
- Dashboard initial load: < 3 seconds
- Analytics queries: < 2 seconds
- Chart rendering: < 1 second

### Data Export

**User Experience:**
- Small exports complete quickly
- Large exports show progress
- User can continue working while export processes

**Performance Targets:**
- Small exports (< 1000 records): < 10 seconds
- Large exports: Show progress indicator
- Notification when export is ready

## What Affects Performance

### Factors Outside System Control

**Local Environment:**
- User's network speed to AI providers
- User's computer resources (CPU, memory)
- Number of concurrent AI requests

**Server Environment:**
- Number of active users in organization
- Amount of historical usage data
- Network conditions between user and server

### Factors Within System Control

- Request routing efficiency
- Database query optimization
- Batch processing for usage data
- Background task scheduling

## Performance by License Type

| Feature | Open Source | Commercial |
|---------|-------------|------------|
| Routing performance | Standard | Standard |
| Data retention | 7 days | 90 days |
| Analytics | Personal | Personal + Team |
| Dashboard load time | Standard | Standard |

## Scalability

### Current Design Supports

**Suggested Capacity Limits:**
- Up to 100 concurrent users per organization
- Up to 1,000 AI requests per minute
- Up to 100,000 usage records per day

**Beyond These Limits:**
- Additional server capacity may be needed
- Contact support for enterprise deployment options

## Known Limitations

### First-Time Setup

**Initial Configuration Sync:**
- May take 10-30 seconds
- Depends on number of providers configured
- One-time delay after initial setup

### Large Data Operations

**Exporting Large Datasets:**
- Thousands of records may take 30+ seconds
- Progress indicator shown during export
- User can continue working while export processes

**Loading Extended Date Ranges:**
- Analytics for 90+ days may load slower
- Consider using shorter date ranges for better performance

## Performance Monitoring

### What Is Tracked

- Request response times
- Configuration sync success rates
- Dashboard load times
- Error rates

### When Performance Issues Occur

**User Notifications:**
- System issues are communicated via status indicators
- Transient issues resolve automatically
- Persistent issues show helpful error messages

## Success Criteria

- Users experience no perceptible delay in AI tool usage
- Dashboard and analytics load quickly
- Configuration sync happens without disruption
- Large data operations show clear progress
- Performance degrades gracefully under load
- Users understand when delays are normal

---

**Related:** [06.01 Privacy](../01_privacy/) | [06.02 Security](../02_security/)
